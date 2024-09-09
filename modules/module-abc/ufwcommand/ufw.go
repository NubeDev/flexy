package ufwcommand

import (
	"errors"
	"github.com/NubeIO/lib-ufw/ufw"
)

func New() *System {
	inst := &System{}
	inst.ufw = ufw.New(&ufw.System{})
	return inst
}

type UFWBody struct {
	Port int `json:"port"`
}

type System struct {
	ufw *ufw.System
}

// UWFActive check status and also if ufwcommand is installed
func (inst *System) UWFActive() (bool, error) {
	return inst.ufw.UWFActive()
}

// UWFStatus check status and also if ufwcommand is installed
func (inst *System) UWFStatus() (*ufw.Message, error) {
	return inst.ufw.UWFStatus()
}

// UWFStatusList check status and also if ufwcommand is installed
func (inst *System) UWFStatusList() ([]ufw.UFWStatus, error) {
	return inst.ufw.UWFStatusList()
}

func (inst *System) UWFOpenPort(body UFWBody) (*ufw.Message, error) {
	return inst.ufw.UWFOpenPort(body.Port)
}

func (inst *System) UWFClosePort(body UFWBody) (*ufw.Message, error) {
	if body.Port == 22 {
		return nil, errors.New("ufw: port 22 is not allowed to be closed")
	}
	return inst.ufw.UWFClosePort(body.Port)
}

func (inst *System) UWFEnable() (*ufw.Message, error) {
	return inst.ufw.UWFEnable()
}

func (inst *System) UWFDisable() (*ufw.Message, error) {
	return inst.ufw.UWFDisable()
}
