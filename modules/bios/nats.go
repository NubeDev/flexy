package main

import (
	"fmt"
)

// StartService starts the NATS subscription and listens for commands
// these are the bios commands. eg; install and app
func (s *Service) StartService() error {
	var err error

	go s.natsForwarder(s.natsConn, fmt.Sprintf("nats://127.0.0.1:%d", s.Config.GetInt("bios.proxy_port")))

	// Initialize NATS store
	err = s.natsStoreInit(s.natsConn)
	if err != nil {
		return fmt.Errorf("failed to init nats-store: %v", err)
	}

	var globalUUID = s.globalUUID
	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("get", "system", "ping"), s.handlePing)
	if err != nil {
		return err
	}
	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("get", "apps", "installed"), s.handleListInstalledApps)
	if err != nil {
		return err
	}
	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("get", "apps", "library"), s.handleListLibraryApps)
	if err != nil {
		return err
	}
	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("post", "apps", "install"), s.handleInstallApp)
	if err != nil {
		return err
	}
	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("post", "apps", "uninstall"), s.handleUninstallApp)
	if err != nil {
		return err
	}

	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("get", "system", "systemctl"), s.BiosSystemdCommand)
	if err != nil {
		return err
	}

	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("post", "git", "asset"), s.gitDownloadAsset)
	if err != nil {
		return err
	}
	_, err = s.natsConn.Subscribe(s.biosSubjectBuilder.BuildSubject("get", "git", "assets"), s.gitListAllAssets)
	if err != nil {
		return err
	}

	_, err = s.natsConn.Subscribe(fmt.Sprintf("bios.%s.store", globalUUID), s.HandleStore)
	if err != nil {
		return err
	}
	return nil
}
