package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/modules/bios/appmanager"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/systemctl"
	"github.com/nats-io/nats.go"
)

// Command structure to decode the incoming JSON
type Command struct {
	Command string                 `json:"command"`
	Body    map[string]interface{} `json:"body"`
}

// Service struct to handle NATS and file operations
type Service struct {
	natsConn   *nats.Conn
	natsStore  *natsrouter.NatsRouter
	systemD    *systemctl.CTL
	appManager *appmanager.AppManager
}

// NewService initializes the NATS connection and returns the Service
func NewService(natsURL, dataPath, systemPath string) (*Service, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	return &Service{
		natsConn:   nc,
		systemD:    systemctl.New(),
		appManager: appmanager.NewAppManager(dataPath, systemPath),
	}, nil
}

// Common error handling method
func (s *Service) handleError(reply string, responseCode int, details string) {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"details": details,
	}
	respJSON, _ := json.Marshal(response)
	s.natsConn.Publish(reply, respJSON)
}

// Common NATS publish method
func (s *Service) publish(reply string, content string, responseCode int) {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"content": content,
	}
	respJSON, _ := json.Marshal(response)
	s.natsConn.Publish(reply, respJSON)
}

// HandleCommand processes incoming JSON commands
func (s *Service) HandleCommand(m *nats.Msg) {
	var cmd Command

	// Unmarshal the received JSON message
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}

	switch cmd.Command {
	case "ping":
		s.publish(m.Reply, "PONG", code.SUCCESS)
	case "list_library_apps":
		apps, err := s.appManager.ListLibraryApps()
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error listing library apps: %v", err))
		} else {
			respJSON, _ := json.Marshal(apps)
			s.publish(m.Reply, string(respJSON), code.SUCCESS)
		}

	case "list_installed_apps":
		apps, err := s.appManager.ListInstalledApps()
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error listing installed apps: %v", err))
		} else {
			respJSON, _ := json.Marshal(apps)
			s.publish(m.Reply, string(respJSON), code.SUCCESS)
		}

	case "install_app":
		appName, ok := cmd.Body["name"].(string)
		version, versionOk := cmd.Body["version"].(string)
		if !ok || !versionOk || appName == "" || version == "" {
			s.handleError(m.Reply, code.InvalidParams, "'name' and 'version' are required for install_app")
			return
		}
		app := &appmanager.App{Name: appName, Version: version}
		err := s.appManager.Install(app)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error installing app: %v", err))
		} else {
			s.publish(m.Reply, fmt.Sprintf("App %s version %s installed", appName, version), code.SUCCESS)
		}

	case "uninstall_app":
		appName, ok := cmd.Body["name"].(string)
		version, versionOk := cmd.Body["version"].(string)
		if !ok || !versionOk || appName == "" || version == "" {
			s.handleError(m.Reply, code.InvalidParams, "'name' and 'version' are required for uninstall_app")
			return
		}
		app := &appmanager.App{Name: appName, Version: version}
		err := s.appManager.Uninstall(app)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error uninstalling app: %v", err))
		} else {
			s.publish(m.Reply, fmt.Sprintf("App %s version %s uninstalled", appName, version), code.SUCCESS)
		}
	case "read_file":
		path, ok := cmd.Body["path"].(string)
		if !ok || path == "" {
			s.handleError(m.Reply, code.InvalidParams, "'path' is required for read_file")
			return
		}
		content, err := s.ReadFile(path)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error reading file: %v", err))
		} else {
			s.publish(m.Reply, content, code.SUCCESS)
		}
	case "make_dir":
		path, ok := cmd.Body["path"].(string)
		if !ok || path == "" {
			s.handleError(m.Reply, code.InvalidParams, "'path' is required for make_dir")
			return
		}
		err := s.MakeDir(path)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error creating directory: %v", err))
		} else {
			s.publish(m.Reply, "Directory created", code.SUCCESS)
		}
	case "delete_dir":
		path, ok := cmd.Body["path"].(string)
		if !ok || path == "" {
			s.handleError(m.Reply, code.InvalidParams, "'path' is required for delete_dir")
			return
		}
		err := s.DeleteDir(path)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error deleting directory: %v", err))
		} else {
			s.publish(m.Reply, "Directory deleted", code.SUCCESS)
		}
	case "zip_folder":
		srcDir, srcOk := cmd.Body["srcDir"].(string)
		dstZip, dstOk := cmd.Body["dstZip"].(string)
		if !srcOk || !dstOk || srcDir == "" || dstZip == "" {
			s.handleError(m.Reply, code.InvalidParams, "'srcDir' and 'dstZip' are required for zip_folder")
			return
		}
		err := s.ZipFolder(srcDir, dstZip)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error zipping folder: %v", err))
		} else {
			s.publish(m.Reply, "Folder zipped", code.SUCCESS)
		}
	case "unzip_folder":
		srcZip, srcOk := cmd.Body["srcZip"].(string)
		destDir, dstOk := cmd.Body["destDir"].(string)
		if !srcOk || !dstOk || srcZip == "" || destDir == "" {
			s.handleError(m.Reply, code.InvalidParams, "'srcZip' and 'destDir' are required for unzip_folder")
			return
		}
		err := s.UnzipFolder(srcZip, destDir)
		if err != nil {
			s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error unzipping folder: %v", err))
		} else {
			s.publish(m.Reply, "Folder unzipped", code.SUCCESS)
		}
	default:
		s.handleError(m.Reply, code.UnknownCommand, "Unknown command")
	}
}
