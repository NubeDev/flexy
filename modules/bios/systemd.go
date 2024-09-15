package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
)

type Systemd struct {
	Name     string `json:"name"`
	Action   string `json:"action"`
	Property string `json:"property,omitempty"`
}

func (s *Service) DecodeSystemd(m *nats.Msg) (*Systemd, error) {
	var cmd Systemd
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

func (s *Service) BiosSystemdCommand(m *nats.Msg) {
	// Decode the incoming message into the Systemd struct
	decoded, err := s.DecodeSystemd(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			s.handleError(m.Reply, code.ERROR, "failed to parse JSON")
			return
		}
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	if decoded.Name == "" {
		s.handleError(m.Reply, code.InvalidParams, "service name is required")
		return
	}
	if decoded.Action == "" {
		s.handleError(m.Reply, code.InvalidParams, "action is required, e.g., start, stop, restart, enable, disable, status, is-enabled, show")
		return
	}

	// Switch based on the action provided
	switch decoded.Action {
	case "start", "stop", "restart", "enable", "disable":
		err := s.systemctlService.SystemdCommand(decoded.Name, decoded.Action)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error performing %s on service %s: %v", decoded.Action, decoded.Name, err))
		} else {
			s.publish(m.Reply, fmt.Sprintf("Service %s %sed successfully", decoded.Name, decoded.Action), code.SUCCESS)
		}

	case "status":
		status, err := s.systemctlService.SystemdStatus(decoded.Name)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error getting status of service %s: %v", decoded.Name, err))
		} else {
			respJSON, _ := json.Marshal(status)
			s.publish(m.Reply, string(respJSON), code.SUCCESS)
		}

	case "is-enabled":
		enabled, err := s.systemctlService.SystemdIsEnabled(decoded.Name)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error checking if service %s is enabled: %v", decoded.Name, err))
		} else {
			s.publish(m.Reply, fmt.Sprintf("Service %s is-enabled: %v", decoded.Name, enabled), code.SUCCESS)
		}

	case "show":
		if decoded.Property == "" {
			s.handleError(m.Reply, code.InvalidParams, "'property' is required for the show action")
			return
		}
		result, err := s.systemctlService.SystemdShow(decoded.Name, decoded.Property)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error showing property %s of service %s: %v", decoded.Property, decoded.Name, err))
		} else {
			s.publish(m.Reply, fmt.Sprintf("Service %s property %s: %s", decoded.Name, decoded.Property, result), code.SUCCESS)
		}

	default:
		s.handleError(m.Reply, code.InvalidParams, fmt.Sprintf("Unknown action: %s", decoded.Action))
	}
}
