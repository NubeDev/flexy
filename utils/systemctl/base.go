package systemctl

import (
	"github.com/NubeDev/flexy/utils/execute/commands"
)

type CTL struct {
	cmd commands.Commands
}

type SystemdCommand struct {
	Unit        string `json:"unit"`
	CommandType string `json:"commandType"`
}

func New() *CTL {
	return &CTL{
		cmd: commands.New(),
	}
}

// SystemdStatus  "mosquitto"
func (inst *CTL) SystemdStatus(unit string) (*commands.StatusResp, error) {
	return inst.cmd.SystemdStatus(unit)
}

// SystemdCommand start, stop, restart, enable, disable "mosquitto"
func (inst *CTL) SystemdCommand(unit, commandType string) error {
	return inst.cmd.SystemdCommand(unit, commandType)
}

// SystemdShow ("mosquitto", "NRestarts")
func (inst *CTL) SystemdShow(unit, property string) (string, error) {
	return inst.cmd.SystemdShow(unit, property)
}

// SystemdIsEnabled "mosquitto"
func (inst *CTL) SystemdIsEnabled(unit string) (bool, error) {
	return inst.cmd.SystemdIsEnabled(unit)
}
