package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/modules/bios/appmanager"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
)

func (s *Service) handleListLibraryApps(m *nats.Msg) {
	apps, err := s.appManager.ListLibraryApps()
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error listing library apps: %v", err))
		return
	}
	respJSON, _ := json.Marshal(apps)
	s.publish(m.Reply, string(respJSON), code.SUCCESS)
}

func (s *Service) handleListInstalledApps(m *nats.Msg) {
	apps, err := s.appManager.ListInstalledApps()
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error listing installed apps: %v", err))
		return
	}
	respJSON, _ := json.Marshal(apps)
	s.publish(m.Reply, string(respJSON), code.SUCCESS)
}

func (s *Service) handleInstallApp(m *nats.Msg) {
	decoded, err := s.DecodeApps(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			s.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	decoded, err = s.getAppName(decoded)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Name == "" {
		s.handleError(m.Reply, code.InvalidParams, "app name is required")
		return
	}
	if decoded.Version == "" {
		s.handleError(m.Reply, code.InvalidParams, "app version is required")
		return
	}
	app := &appmanager.App{Name: decoded.Name, Version: decoded.Version}
	err = s.appManager.Install(app)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error installing app: %v", err))
	} else {
		s.publish(m.Reply, fmt.Sprintf("App %s version %s installed", decoded.Name, decoded.Version), code.SUCCESS)
	}
}

// New method to handle setting the decoded.Name based on AppID
func (s *Service) getAppName(decoded *App) (*App, error) {
	if decoded.Name == "" {
		if decoded.AppID != "" {
			app, err := s.appManager.GetAppByID(decoded.AppID, decoded.Version)
			if err != nil {
				return nil, err
			}
			if app == nil {
				return nil, fmt.Errorf("%s not found", decoded.AppID)
			}
			decoded.Name = app.Name

		} else {
			return nil, fmt.Errorf("app name is required")
		}
	}
	return decoded, nil
}

func (s *Service) handleUninstallApp(m *nats.Msg) {
	decoded, err := s.DecodeApps(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			s.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	decoded, err = s.getAppName(decoded)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Name == "" {
		s.handleError(m.Reply, code.InvalidParams, "app name is required")
		return
	}
	if decoded.Version == "" {
		s.handleError(m.Reply, code.InvalidParams, "app version is required")
		return
	}
	app := &appmanager.App{Name: decoded.Name, Version: decoded.Version}
	err = s.appManager.Uninstall(app)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error uninstalling app: %v", err))
	} else {
		s.publish(m.Reply, fmt.Sprintf("App %s version %s uninstalled", decoded.Name, decoded.Version), code.SUCCESS)
	}
}
