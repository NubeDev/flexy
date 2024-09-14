package main

import (
	"fmt"
)

// StartService starts the NATS subscription and listens for commands
// these are the bios commands. eg; install and app
func (inst *Service) StartService(globalUUID string) error {
	var err error
	_, err = inst.natsConn.Subscribe(inst.biosSubjectBuilder.BuildSubject("get", "system", "ping"), inst.handlePing)
	if err != nil {
		return err
	}
	_, err = inst.natsConn.Subscribe(inst.biosSubjectBuilder.BuildSubject("get", "apps", "installed"), inst.handleListInstalledApps)
	if err != nil {
		return err
	}
	_, err = inst.natsConn.Subscribe(inst.biosSubjectBuilder.BuildSubject("get", "apps", "library"), inst.handleListLibraryApps)
	if err != nil {
		return err
	}
	_, err = inst.natsConn.Subscribe(inst.biosSubjectBuilder.BuildSubject("post", "apps", "install"), inst.handleInstallApp)
	if err != nil {
		return err
	}
	_, err = inst.natsConn.Subscribe(inst.biosSubjectBuilder.BuildSubject("post", "apps", "uninstall"), inst.handleUninstallApp)
	if err != nil {
		return err
	}
	_, err = inst.natsConn.Subscribe(fmt.Sprintf("bios.%s.store", globalUUID), inst.HandleStore)
	if err != nil {
		return err
	}
	return nil
}
