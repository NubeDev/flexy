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
	AppID    string `json:"appID"`
	Version  string `json:"version"`
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

func (s *Service) handleSystemctlGet(m *nats.Msg) {
	decoded, action, err := s.decodeAndSetAction(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	switch action {
	case "status":
		status, err := s.systemctlService.SystemdStatus(decoded.Name)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error getting status of service %s: %v", decoded.Name, err))
		} else {
			s.publishResponse(m, status, code.SUCCESS)
		}

	case "is-enabled":
		enabled, err := s.systemctlService.SystemdIsEnabled(decoded.Name)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error checking if service %s is enabled: %v", decoded.Name, err))
		} else {
			out := Message{
				fmt.Sprintf("Service %s is-enabled: %v", decoded.Name, enabled),
			}
			s.publishResponse(m, out, code.SUCCESS)
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
			out := Message{
				fmt.Sprintf("Service %s property %s: %s", decoded.Name, decoded.Property, result),
			}
			s.publishResponse(m, out, code.SUCCESS)
		}
	default:
		message := fmt.Sprintf("Unknown GET action in systemctl manager: %s", action)
		log.Error().Msg(message)
		s.handleError(m.Reply, code.UnknownCommand, message)
	}
}

func (s *Service) handleSystemctlPost(m *nats.Msg) {
	decoded, action, err := s.decodeAndSetAction(m)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	switch action {
	case "start", "stop", "restart", "enable", "disable":
		err := s.systemctlService.SystemdCommand(decoded.Name, action)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error performing %s on service %s: %v", decoded.Action, decoded.Name, err))
		} else {
			out := Message{
				fmt.Sprintf("Service %s %sed successfully", decoded.Name, decoded.Action),
			}
			s.publishResponse(m, out, code.SUCCESS)
		}

	default:
		s.handleError(m.Reply, code.InvalidParams, fmt.Sprintf("Unknown action: %s", decoded.Action))
	}
}

// New method to handle setting the decoded.Name based on AppID
func (s *Service) setAppName(decoded *Systemd) (*Systemd, error) {
	if decoded.Name == "" {
		if decoded.AppID != "" {
			if decoded.Version == "" {
				return nil, fmt.Errorf("app version is required")
			}
			app, err := s.appManager.GetAppByID(decoded.AppID, decoded.Version)
			if err != nil {
				return nil, err
			}
			if app == nil {
				return nil, fmt.Errorf("failed to get app by id: %s", decoded.AppID)
			}
			decoded.Name = app.Name
		} else {
			return nil, fmt.Errorf("app name is required")
		}
	}
	return decoded, nil
}

// Helper function to decode the message and determine the action
func (s *Service) decodeAndSetAction(m *nats.Msg) (*Systemd, string, error) {
	decoded, err := s.DecodeSystemd(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			return nil, "", fmt.Errorf("failed to parse JSON")
		}
		return nil, "", err
	}

	decoded, err = s.setAppName(decoded)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return nil, "", err
	}

	// Determine the action from the decoded message or the NATS subject
	var action string
	if decoded.Action != "" {
		action = decoded.Action
	} else {
		subjectParts := strings.Split(m.Subject, ".")
		action = subjectParts[len(subjectParts)-1]
		decoded.Action = action
	}

	return decoded, action, nil
}
