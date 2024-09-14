package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/modules/bios/appmanager"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/subjects"
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
	natsConn           *nats.Conn
	natsStore          *natsrouter.NatsRouter
	systemD            *systemctl.CTL
	appManager         *appmanager.AppManager
	biosSubjectBuilder *subjects.SubjectBuilder
}

// NewService initializes the NATS connection and returns the Service
func NewService(natsURL, dataPath, systemPath string) (*Service, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	appManager, err := appmanager.NewAppManager(dataPath, systemPath)
	if err != nil {
		return nil, err
	}
	return &Service{
		natsConn:           nc,
		systemD:            systemctl.New(),
		appManager:         appManager,
		biosSubjectBuilder: subjects.NewSubjectBuilder(globalUUID, "bios", subjects.IsBios),
	}, nil
}

// Common error handling method
func (inst *Service) handleError(reply string, responseCode int, details string) {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"details": details,
	}
	respJSON, _ := json.Marshal(response)
	inst.natsConn.Publish(reply, respJSON)
}

// Common NATS publish method
func (inst *Service) publish(reply string, content string, responseCode int) {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"content": content,
	}
	respJSON, _ := json.Marshal(response)
	inst.natsConn.Publish(reply, respJSON)
}

// GetCommandValue extracts a value from the command body by key
func (inst *Service) GetCommandValue(cmd *Command, key string) (string, error) {
	value, ok := cmd.Body[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("'%s' is required and must be a non-empty string", key)
	}
	return value, nil
}

// DecodeCommand decodes the incoming NATS message into a Command struct
func (inst *Service) DecodeCommand(m *nats.Msg) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

// HandleCommand processes incoming JSON commands
func (inst *Service) HandleCommand(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	// Call the appropriate method based on the command
	switch cmd.Command {
	case "ping":
		inst.handlePing(m)
	case "list_library_apps":
		inst.handleListLibraryApps(m)
	case "list_installed_apps":
		inst.handleListInstalledApps(m)
	case "install_app":
		inst.handleInstallApp(m)
	case "uninstall_app":
		inst.handleUninstallApp(m)
	case "read_file":
		inst.handleReadFile(m)
	case "make_dir":
		inst.handleMakeDir(m)
	case "delete_dir":
		inst.handleDeleteDir(m)
	case "zip_folder":
		inst.handleZipFolder(m)
	case "unzip_folder":
		inst.handleUnzipFolder(m)
	default:
		inst.handleError(m.Reply, code.UnknownCommand, "Unknown command")
	}
}

// Individual command handlers

func (inst *Service) handlePing(m *nats.Msg) {
	inst.publish(m.Reply, "PONG", code.SUCCESS)
}

func (inst *Service) handleListLibraryApps(m *nats.Msg) {
	apps, err := inst.appManager.ListLibraryApps()
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error listing library apps: %v", err))
		return
	}
	respJSON, _ := json.Marshal(apps)
	inst.publish(m.Reply, string(respJSON), code.SUCCESS)
}

func (inst *Service) handleListInstalledApps(m *nats.Msg) {
	apps, err := inst.appManager.ListInstalledApps()
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error listing installed apps: %v", err))
		return
	}
	respJSON, _ := json.Marshal(apps)
	inst.publish(m.Reply, string(respJSON), code.SUCCESS)
}

func (inst *Service) handleInstallApp(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	appName, err := inst.GetCommandValue(cmd, "name")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	version, err := inst.GetCommandValue(cmd, "version")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	app := &appmanager.App{Name: appName, Version: version}
	err = inst.appManager.Install(app)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error installing app: %v", err))
	} else {
		inst.publish(m.Reply, fmt.Sprintf("App %s version %s installed", appName, version), code.SUCCESS)
	}
}

func (inst *Service) handleUninstallApp(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	appName, err := inst.GetCommandValue(cmd, "name")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	version, err := inst.GetCommandValue(cmd, "version")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	app := &appmanager.App{Name: appName, Version: version}
	err = inst.appManager.Uninstall(app)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error uninstalling app: %v", err))
	} else {
		inst.publish(m.Reply, fmt.Sprintf("App %s version %s uninstalled", appName, version), code.SUCCESS)
	}
}

func (inst *Service) handleReadFile(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	path, err := inst.GetCommandValue(cmd, "path")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	content, err := inst.ReadFile(path)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error reading file: %v", err))
	} else {
		inst.publish(m.Reply, content, code.SUCCESS)
	}
}

func (inst *Service) handleMakeDir(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	path, err := inst.GetCommandValue(cmd, "path")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = inst.MakeDir(path)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error creating directory: %v", err))
	} else {
		inst.publish(m.Reply, "Directory created", code.SUCCESS)
	}
}

func (inst *Service) handleDeleteDir(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	path, err := inst.GetCommandValue(cmd, "path")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = inst.DeleteDir(path)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error deleting directory: %v", err))
	} else {
		inst.publish(m.Reply, "Directory deleted", code.SUCCESS)
	}
}

func (inst *Service) handleZipFolder(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	srcDir, err := inst.GetCommandValue(cmd, "srcDir")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	dstZip, err := inst.GetCommandValue(cmd, "dstZip")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = inst.ZipFolder(srcDir, dstZip)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error zipping folder: %v", err))
	} else {
		inst.publish(m.Reply, "Folder zipped", code.SUCCESS)
	}
}

func (inst *Service) handleUnzipFolder(m *nats.Msg) {
	cmd, err := inst.DecodeCommand(m)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	srcZip, err := inst.GetCommandValue(cmd, "srcZip")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	destDir, err := inst.GetCommandValue(cmd, "destDir")
	if err != nil {
		inst.handleError(m.Reply, code.InvalidParams, err.Error())
		return
	}

	err = inst.UnzipFolder(srcZip, destDir)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error unzipping folder: %v", err))
	} else {
		inst.publish(m.Reply, "Folder unzipped", code.SUCCESS)
	}
}
