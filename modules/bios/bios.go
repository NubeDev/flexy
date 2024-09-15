package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/modules/bios/appmanager"
	"github.com/NubeDev/flexy/utils/code"
	githubdownloader "github.com/NubeDev/flexy/utils/gitdownloader"
	"github.com/NubeDev/flexy/utils/subjects"
	"github.com/NubeDev/flexy/utils/systemctl"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Command structure to decode the incoming JSON
type Command struct {
	Command string                 `json:"command"`
	Body    map[string]interface{} `json:"body"`
}

type App struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Service struct to handle NATS and file operations
type Service struct {
	globalUUID         string
	gitDownloadPath    string
	natsConn           *nats.Conn
	natsStore          *natsrouter.NatsRouter
	systemD            *systemctl.CTL
	appManager         *appmanager.AppManager
	biosSubjectBuilder *subjects.SubjectBuilder
	githubDownloader   *githubdownloader.GitHubDownloader
}

type Opts struct {
	GlobalUUID      string
	NatsURL         string
	DataPath        string
	SystemPath      string
	GitToken        string
	GitDownloadPath string
	ProxyNatsPort   int
}

// NewService initializes the NATS connection and returns the Service
func NewService(opts *Opts) (*Service, error) {
	var globalUUID string
	var natsURL string
	var dataPath string
	var systemPath string
	var gitToken string
	var gitDownloadPath string

	if opts != nil {
		globalUUID = opts.GlobalUUID
		natsURL = opts.NatsURL
		dataPath = opts.DataPath
		systemPath = opts.SystemPath
		gitToken = opts.GitToken
		gitDownloadPath = opts.GitDownloadPath
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}
	appManager, err := appmanager.NewAppManager(dataPath, systemPath)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("start bios NATS server: %v", natsURL)
	ser := &Service{
		globalUUID:         globalUUID,
		gitDownloadPath:    gitDownloadPath,
		natsConn:           nc,
		systemD:            systemctl.New(),
		appManager:         appManager,
		biosSubjectBuilder: subjects.NewSubjectBuilder(globalUUID, "bios", subjects.IsBios),
		githubDownloader:   githubdownloader.New(gitToken, gitDownloadPath),
	}

	return ser, nil
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

// DecodeApps decodes the incoming NATS message into a Command struct
func (inst *Service) DecodeApps(m *nats.Msg) (*App, error) {
	var cmd App
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

// DecodeCommand decodes the incoming NATS message into a Command struct
func (inst *Service) DecodeCommand(m *nats.Msg) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

func (inst *Service) GetCommandValue(cmd *Command, key string) (string, error) {
	value, ok := cmd.Body[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("'%s' is required and must be a non-empty string", key)
	}
	return value, nil
}

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
	decoded, err := inst.DecodeApps(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			inst.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Name == "" {
		inst.handleError(m.Reply, code.InvalidParams, "app name is required")
		return
	}
	if decoded.Version == "" {
		inst.handleError(m.Reply, code.InvalidParams, "app version is required")
		return
	}
	app := &appmanager.App{Name: decoded.Name, Version: decoded.Version}
	err = inst.appManager.Install(app)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error installing app: %v", err))
	} else {
		inst.publish(m.Reply, fmt.Sprintf("App %s version %s installed", decoded.Name, decoded.Version), code.SUCCESS)
	}
}

func (inst *Service) handleUninstallApp(m *nats.Msg) {
	decoded, err := inst.DecodeApps(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			inst.handleError(m.Reply, code.ERROR, "failed to parse json")
			return
		}
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	if decoded.Name == "" {
		inst.handleError(m.Reply, code.InvalidParams, "app name is required")
		return
	}
	if decoded.Version == "" {
		inst.handleError(m.Reply, code.InvalidParams, "app version is required")
		return
	}
	app := &appmanager.App{Name: decoded.Name, Version: decoded.Version}
	err = inst.appManager.Uninstall(app)
	if err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error uninstalling app: %v", err))
	} else {
		inst.publish(m.Reply, fmt.Sprintf("App %s version %s uninstalled", decoded.Name, decoded.Version), code.SUCCESS)
	}
}

type Systemd struct {
	Name     string `json:"name"`
	Action   string `json:"action"`
	Property string `json:"property,omitempty"`
}

func (inst *Service) DecodeSystemd(m *nats.Msg) (*Systemd, error) {
	var cmd Systemd
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

func (inst *Service) BiosSystemdCommand(m *nats.Msg) {
	// Decode the incoming message into the Systemd struct
	decoded, err := inst.DecodeSystemd(m)
	if decoded == nil || err != nil {
		if decoded == nil {
			inst.handleError(m.Reply, code.ERROR, "failed to parse JSON")
			return
		}
		inst.handleError(m.Reply, code.ERROR, err.Error())
		return
	}

	if decoded.Name == "" {
		inst.handleError(m.Reply, code.InvalidParams, "service name is required")
		return
	}
	if decoded.Action == "" {
		inst.handleError(m.Reply, code.InvalidParams, "action is required, e.g., start, stop, restart, enable, disable, status, is-enabled, show")
		return
	}

	// Switch based on the action provided
	switch decoded.Action {
	case "start", "stop", "restart", "enable", "disable":
		err := inst.systemD.SystemdCommand(decoded.Name, decoded.Action)
		if err != nil {
			inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error performing %s on service %s: %v", decoded.Action, decoded.Name, err))
		} else {
			inst.publish(m.Reply, fmt.Sprintf("Service %s %sed successfully", decoded.Name, decoded.Action), code.SUCCESS)
		}

	case "status":
		status, err := inst.systemD.SystemdStatus(decoded.Name)
		if err != nil {
			inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error getting status of service %s: %v", decoded.Name, err))
		} else {
			respJSON, _ := json.Marshal(status)
			inst.publish(m.Reply, string(respJSON), code.SUCCESS)
		}

	case "is-enabled":
		enabled, err := inst.systemD.SystemdIsEnabled(decoded.Name)
		if err != nil {
			inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error checking if service %s is enabled: %v", decoded.Name, err))
		} else {
			inst.publish(m.Reply, fmt.Sprintf("Service %s is-enabled: %v", decoded.Name, enabled), code.SUCCESS)
		}

	case "show":
		if decoded.Property == "" {
			inst.handleError(m.Reply, code.InvalidParams, "'property' is required for the show action")
			return
		}
		result, err := inst.systemD.SystemdShow(decoded.Name, decoded.Property)
		if err != nil {
			inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Error showing property %s of service %s: %v", decoded.Property, decoded.Name, err))
		} else {
			inst.publish(m.Reply, fmt.Sprintf("Service %s property %s: %s", decoded.Name, decoded.Property, result), code.SUCCESS)
		}

	default:
		inst.handleError(m.Reply, code.InvalidParams, fmt.Sprintf("Unknown action: %s", decoded.Action))
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
