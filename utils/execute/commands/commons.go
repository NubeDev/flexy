package commands

import (
	"github.com/NubeDev/flexy/utils/execute"
	"regexp"
	"strings"
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
