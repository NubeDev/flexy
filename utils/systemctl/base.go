package systemctl

import (
	"errors"
	"fmt"
	"github.com/NubeDev/flexy/utils/execute"
	"github.com/NubeDev/flexy/utils/times"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var defaultTimeout = 2

type CommandBody struct {
	Command string
	Arg     string
	Args    []string
	Timeout int
}

type Commands interface {
	Run(body *CommandBody) *execute.Response
	Uptime(timeout ...int) (*UptimeInfo, error)
	SystemdStatus(unit string) (*StatusResp, error)
	// SystemdCommand start, stop, restart, enable, disable
	SystemdCommand(unit, commandType string) error
	SystemdShow(unit, property string) (string, error)
	SystemdIsEnabled(unit string) (bool, error)
}

type commands struct {
	ex execute.Execute
}

func New() Commands {
	return &commands{
		ex: execute.New(),
	}
}

type UptimeInfo struct {
	UpTime       string
	Users        string
	LoadAverages [3]string
}

func (cmd *commands) Run(body *CommandBody) *execute.Response {
	if body == nil {
		return &execute.Response{
			Error: "the command body can not be empty",
		}
	}
	return cmd.ex.AddTimeout(body.Timeout).Run(body.Command, body.Args...)
}

func (cmd *commands) Uptime(timeout ...int) (*UptimeInfo, error) {
	if len(timeout) > 0 {
		defaultTimeout = timeout[0]
	}
	if defaultTimeout < 0 {
		defaultTimeout = 2
	}
	c := cmd.ex.AddTimeout(defaultTimeout).Run("uptime")
	if c.AsError() != nil {
		return nil, c.AsError()
	}
	resp := parseUptimeOutput(c.AsString())
	return resp, nil

}

func parseUptimeOutput(output string) *UptimeInfo {
	uptimeInfo := &UptimeInfo{}

	// Use regular expressions to extract relevant information
	re := regexp.MustCompile(`up (.+?), ([0-9]+) user`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 3 {
		uptimeInfo.UpTime = matches[1]
		uptimeInfo.Users = matches[2] + " user"
	}

	// Extract load averages
	fields := strings.Fields(output)
	if len(fields) >= 11 {
		copy(uptimeInfo.LoadAverages[:], fields[len(fields)-3:])
	}

	return uptimeInfo
}

// SystemdCommand start, stop, restart, enable, disable
func (cmd *commands) SystemdCommand(unit, commandType string) error {
	err := isValidAction(commandType)
	if err != nil {
		return err
	}
	c := cmd.ex.Run("systemctl", commandType, unit)
	if c.AsError() != nil {
		return c.AsError()
	}
	return nil
}

func isValidAction(action string) error {
	validActions := []string{"start", "stop", "restart", "enable", "disable"}
	for _, valid := range validActions {
		if action == valid {
			return nil // Match found, no error
		}
	}
	return fmt.Errorf("invalid action: %s, try: %s or %s", action, validActions[0], validActions[1])
}

func (cmd *commands) SystemdShow(unit, property string) (string, error) {
	c := cmd.ex.Run("systemctl", "show", "-p", property, unit)
	if c.AsError() != nil {
		return "", c.AsError()
	}
	return strings.TrimSpace(c.AsString()), nil
}

func (cmd *commands) SystemdIsEnabled(unit string) (bool, error) {
	c := cmd.ex.Run("systemctl", "is-enabled", unit)
	if c.AsError() != nil {
		return false, c.AsError()
	}

	// Trim and convert the output to lowercase
	output := strings.ToLower(strings.TrimSpace(c.AsString()))

	// Check if the output is "enabled"
	if output == "enabled" {
		return true, nil
	}

	return false, nil
}

func (cmd *commands) SystemdStatus(unit string) (*StatusResp, error) {
	c := cmd.ex.Run("systemctl", "status", unit)
	if c.AsError() != nil {
		return nil, c.AsError()
	}
	out := parseSystemdStatusOutput(c.AsString())
	s, err := cmd.SystemdShow(unit, "NRestarts")
	count, err := parseRestartCount(s)
	if err == nil {
		out.RestartCount = count
	}
	enabled, err := cmd.SystemdIsEnabled(unit)
	if err == nil {
		out.IsEnabled = enabled
	}

	return out, nil
}

func parseRestartCount(output string) (int, error) {
	parts := strings.Split(output, "=")

	if len(parts) != 2 {
		return 0, errors.New("invalid output format")
	}

	return strconv.Atoi(parts[1])
}

type StatusResp struct {
	Status       string    `json:"status,omitempty"`
	RunningSince time.Time `json:"runningSince,omitempty"`
	Uptime       string    `json:"uptime,omitempty"`
	PID          int       `json:"pid,omitempty"`
	Memory       string    `json:"memory,omitempty"`
	CPU          string    `json:"cpu,omitempty"`
	IsEnabled    bool      `json:"isEnabled"`
	IsActive     bool      `json:"isActive"`
	IsFailed     bool      `json:"isFailed"`
	RestartCount int       `json:"restartCount"`
}

func parseSystemdStatusOutput(output string) *StatusResp {
	statusInfo := &StatusResp{}

	// Use regular expressions to extract relevant information
	reStatus := regexp.MustCompile(`Active: (.+) \((.+)\) since (.+);`)
	rePID := regexp.MustCompile(`Main PID: (\d+)`)
	reMemory := regexp.MustCompile(`Memory: (.+)`)
	reCPU := regexp.MustCompile(`CPU: (.+)`)

	// Find matches for each pattern
	matchesStatus := reStatus.FindStringSubmatch(output)
	matchesPID := rePID.FindStringSubmatch(output)
	matchesMemory := reMemory.FindStringSubmatch(output)
	matchesCPU := reCPU.FindStringSubmatch(output)

	// Populate StatusResp fields
	if len(matchesStatus) >= 4 {
		statusInfo.Status = matchesStatus[1]
		if statusInfo.Status == "active" {
			statusInfo.IsActive = true
		}
		if statusInfo.Status == "failed" {
			statusInfo.IsFailed = true
		}
		statusInfo.RunningSince, _ = time.Parse("Mon 2006-01-02 15:04:05 MST", matchesStatus[3])
		statusInfo.Uptime = times.New(statusInfo.RunningSince).TimeSince()
	}
	if len(matchesPID) >= 2 {
		statusInfo.PID, _ = strconv.Atoi(matchesPID[1])
	}
	if len(matchesMemory) >= 2 {
		statusInfo.Memory = matchesMemory[1]
	}
	if len(matchesCPU) >= 2 {
		statusInfo.CPU = matchesCPU[1]
	}

	return statusInfo
}
