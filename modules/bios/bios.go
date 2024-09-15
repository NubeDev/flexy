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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	systemctlService   systemctl.Commands
	appManager         *appmanager.AppManager
	biosSubjectBuilder *subjects.SubjectBuilder
	githubDownloader   *githubdownloader.GitHubDownloader
	services           []string
	Config             *viper.Viper
	RootCmd            *cobra.Command
}

type Opts struct {
	GlobalUUID string

	NatsURL         string
	RootPath        string
	AppsPath        string
	SystemPath      string
	GitToken        string
	GitDownloadPath string
	ProxyNatsPort   int
}

func (s *Service) NewService(opts *Opts) error {
	var globalUUID string
	var natsURL string
	var dataPath string
	var systemPath string
	var gitToken string
	var gitDownloadPath string

	if opts != nil {
		globalUUID = opts.GlobalUUID
		natsURL = opts.NatsURL
		dataPath = opts.AppsPath
		systemPath = opts.SystemPath
		gitToken = opts.GitToken
		gitDownloadPath = opts.GitDownloadPath
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return err
	}
	appManager, err := appmanager.NewAppManager(dataPath, systemPath)
	if err != nil {
		return err
	}
	log.Info().Msgf("start bios NATS server: %v", natsURL)

	// Assign initialized components to the Service struct
	s.globalUUID = globalUUID
	s.gitDownloadPath = gitDownloadPath
	s.natsConn = nc
	s.systemctlService = systemctl.New()
	s.appManager = appManager
	s.biosSubjectBuilder = subjects.NewSubjectBuilder(globalUUID, "bios", subjects.IsBios)
	s.githubDownloader = githubdownloader.New(gitToken, gitDownloadPath)

	return nil
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

// DecodeApps decodes the incoming NATS message into a Command struct
func (s *Service) DecodeApps(m *nats.Msg) (*App, error) {
	var cmd App
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

// DecodeCommand decodes the incoming NATS message into a Command struct
func (s *Service) DecodeCommand(m *nats.Msg) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}
	return &cmd, nil
}

func (s *Service) GetCommandValue(cmd *Command, key string) (string, error) {
	value, ok := cmd.Body[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("'%s' is required and must be a non-empty string", key)
	}
	return value, nil
}

func (s *Service) handlePing(m *nats.Msg) {
	s.publish(m.Reply, "PONG", code.SUCCESS)
}

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
