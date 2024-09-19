package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/modules/bios/appmanager"
	"github.com/NubeDev/flexy/utils/code"
	githubdownloader "github.com/NubeDev/flexy/utils/gitdownloader"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/NubeDev/flexy/utils/subjects"
	"github.com/NubeDev/flexy/utils/systemctl"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Command structure to decode the incoming JSON
type Command struct {
	Action  string                 `json:"action"`
	Command string                 `json:"command"`
	Body    map[string]interface{} `json:"body"`
}

type App struct {
	Name    string `json:"name"`
	AppID   string `json:"appID"`
	Version string `json:"version"`
}

// Service struct to handle NATS and file operations
type Service struct {
	globalUUID      string
	description     string
	gitDownloadPath string
	natsClient      natlib.NatLib
	natsConn        *nats.Conn
	//natsStore          *natsrouter.NatsRouter
	systemctlService   systemctl.Commands
	appManager         appmanager.ManagerInterface
	biosSubjectBuilder *subjects.SubjectBuilder
	githubDownloader   *githubdownloader.GitHubDownloader
	services           []string
	Config             *viper.Viper
	RootCmd            *cobra.Command
	natsSubjects       []string
	natsStore          *natsStore
}

type Opts struct {
	GlobalUUID      string
	NatsURL         string
	RootPath        string
	AppsPath        string
	SystemPath      string
	GitToken        string
	GitDownloadPath string
	ProxyNatsPort   int
	EnableNatsStore bool
}

type natsStore struct {
	name            string
	enableNatsStore bool
}

func (s *Service) NewService(opts *Opts) error {
	if opts == nil {
		return fmt.Errorf("no options provided")
	}
	var globalUUID string
	var natsURL string
	var dataPath string
	var systemPath string
	var gitToken string
	var gitDownloadPath string

	globalUUID = opts.GlobalUUID
	natsURL = opts.NatsURL
	dataPath = opts.AppsPath
	systemPath = opts.SystemPath
	gitToken = opts.GitToken
	gitDownloadPath = opts.GitDownloadPath

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
	s.natsClient = natlib.New(natlib.NewOpts{
		EnableJetStream: opts.EnableNatsStore,
	})
	return nil
}

// Common error handling method
func (s *Service) handleError(reply string, responseCode int, details string) {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"payload": details,
	}
	respJSON, _ := json.Marshal(response)
	s.natsConn.Publish(reply, respJSON)
}

// Common NATS publish method
func (s *Service) publish(reply string, content string, responseCode int) {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"payload": content,
	}
	respJSON, _ := json.Marshal(response)
	s.natsConn.Publish(reply, respJSON)
}

func (s *Service) responded(content string, responseCode int) []byte {
	response := map[string]interface{}{
		"code":    responseCode,
		"message": code.GetMsg(responseCode),
		"payload": content,
	}
	respJSON, _ := json.Marshal(response)
	return respJSON
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

func (s *Service) handlePing(m *nats.Msg) ([]byte, error) {
	//return s.responded(fmt.Sprintf("pong from service: %s", s.globalUUID), code.SUCCESS), nil
	response := natlib.NewResponse(code.SUCCESS, s.globalUUID, natlib.Args{Description: s.description})
	return response.ToJSON(), nil
}
