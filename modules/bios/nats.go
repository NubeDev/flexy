package main

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"strings"
)

// StartService starts the NATS subscription and listens for commands
// these are the bios commands. eg; install and app
func (s *Service) StartService() error {
	var err error

	go s.natsForwarder(s.natsConn, fmt.Sprintf("nats://127.0.0.1:%d", s.Config.GetInt("proxy_port")))

	// Initialize NATS store
	err = s.natsStoreInit(s.natsConn)
	if err != nil {
		return fmt.Errorf("failed to init nats-store: %v", err)
	}

	err = s.natsClient.SubscribeWithRespond(s.biosSubjectBuilder.GlobalSubject("get", "system", "ping"), s.handlePing, &natlib.Opts{})
	if err != nil {
		return err
	}

	err = s.natsClient.SubscribeWithRespond(s.biosSubjectBuilder.BuildSubject("get", "system", "ping"), s.handlePing, &natlib.Opts{})
	if err != nil {
		return err
	}

	err = s.addNatsSubscribe(s.biosSubjectBuilder.BuildSubject("get", "system", "systemctl.*"), s.handleSystemctlGet)
	if err != nil {
		return err
	}

	err = s.addNatsSubscribe(s.biosSubjectBuilder.BuildSubject("post", "system", "systemctl.*"), s.handleSystemctlPost)
	if err != nil {
		return err
	}

	// Apps-related subscriptions (centralized handlers)
	err = s.addNatsSubscribe(s.biosSubjectBuilder.BuildSubject("get", "apps", "manager.*"), s.handleAppsGet)
	if err != nil {
		return err
	}
	err = s.addNatsSubscribe(s.biosSubjectBuilder.BuildSubject("post", "apps", "manager.*"), s.handleAppsPost)
	if err != nil {
		return err
	}

	// Git-related subscriptions (centralized handlers)
	err = s.addNatsSubscribe(s.biosSubjectBuilder.BuildSubject("get", "git", "manager.*"), s.handleGitGet)
	if err != nil {
		return err
	}
	err = s.addNatsSubscribe(s.biosSubjectBuilder.BuildSubject("post", "git", "manager.*"), s.handleGitPost)
	if err != nil {
		return err
	}

	// Store handler
	err = s.addNatsSubscribe(s.biosSubjectBuilder.AddGlobalUUID("bios.%s.store"), s.HandleStore)
	if err != nil {
		return err
	}

	return nil
}

// Central handler for "GET" requests for apps
func (s *Service) handleAppsGet(m *nats.Msg) {

	subjectParts := strings.Split(m.Subject, ".")
	action := subjectParts[len(subjectParts)-1]

	switch action {
	case "installed":
		s.handleListInstalledApps(m)
	case "library":
		s.handleListLibraryApps(m)
	default:
		message := fmt.Sprintf("Unknown GET action in apps manager: %s", action)
		log.Error().Msg(message)
		s.handleError(m.Reply, code.UnknownCommand, message)
	}
}

// Central handler for "POST" requests for apps
func (s *Service) handleAppsPost(m *nats.Msg) {
	subjectParts := strings.Split(m.Subject, ".")
	action := subjectParts[len(subjectParts)-1]

	switch action {
	case "install":
		s.handleInstallApp(m)
	case "uninstall":
		s.handleUninstallApp(m)
	default:
		message := fmt.Sprintf("Unknown POST action in apps manager: %s", action)
		log.Error().Msg(message)
		s.handleError(m.Reply, code.UnknownCommand, message)

	}
}

// Central handler for "GET" requests for git
func (s *Service) handleGitGet(m *nats.Msg) {
	subjectParts := strings.Split(m.Subject, ".")
	action := subjectParts[len(subjectParts)-1]

	switch action {
	case "asset", "assets":
		s.gitListAllAssets(m)
	default:
		message := fmt.Sprintf("Unknown GET action in git: %s", action)
		log.Error().Msg(message)
		s.handleError(m.Reply, code.UnknownCommand, message)
	}
}

// Central handler for "POST" requests for git
func (s *Service) handleGitPost(m *nats.Msg) {
	subjectParts := strings.Split(m.Subject, ".")
	action := subjectParts[len(subjectParts)-1]

	switch action {
	case "asset":
		s.gitDownloadAsset(m)
	default:
		message := fmt.Sprintf("Unknown POST action in git: %s", action)
		log.Error().Msg(message)
		s.handleError(m.Reply, code.UnknownCommand, message)
	}
}

func (s *Service) addNatsSubscribe(subj string, cb nats.MsgHandler) error {
	_, err := s.natsConn.Subscribe(subj, cb)
	return err
}
