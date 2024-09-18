package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"strings"
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

// Central handler for "POST" requests for systemctl
func (s *Service) handleSystemctlGet(m *nats.Msg) {
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

	subjectParts := strings.Split(m.Subject, ".")
	action := subjectParts[len(subjectParts)-1]
	switch action {
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
		message := fmt.Sprintf("Unknown GET action in systemctl manager: %s", action)
		log.Error().Msg(message)
		s.handleError(m.Reply, code.UnknownCommand, message)
	}
}

func (s *Service) handleSystemctlPost(m *nats.Msg) {
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
	subjectParts := strings.Split(m.Subject, ".")
	action := subjectParts[len(subjectParts)-1]
	if decoded.Name == "" {
		s.handleError(m.Reply, code.InvalidParams, "service name is required")
		return
	}
	// Switch based on the action provided
	switch action {
	case "start", "stop", "restart", "enable", "disable":
		err := s.systemctlService.SystemdCommand(decoded.Name, action)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error performing %s on service %s: %v", decoded.Action, decoded.Name, err))
		} else {
			s.publish(m.Reply, fmt.Sprintf("Service %s %sed successfully", decoded.Name, decoded.Action), code.SUCCESS)
		}

	default:
		s.handleError(m.Reply, code.InvalidParams, fmt.Sprintf("Unknown action: %s", decoded.Action))
	}
}
