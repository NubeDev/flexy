package main

import (
	"fmt"
	"github.com/NubeDev/flexy/modules/bios/appmanager"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
)

func main() {
	var err error

	libraryPath := "library" // Path where the zip files are stored
	installPath := "install" // Path where the apps will be installed
	backupPath := "backup"   // Path where backups will be stored

	// Create a new AppManager instance
	manager := appmanager.NewAppManager("/home/user/code/go/nube/flex/flexy/modules/bios/testbios/data/", libraryPath, installPath, backupPath, "tmp", "/etc/systemd/system")

	// Install the app
	appName := "my-app"
	version := "v1.1"
	libraryApps, err := manager.ListLibraryApps()
	if err != nil {
		return
	}
	pprint.PrintJSON(libraryApps)

	apps, err := manager.ListInstalledApps()
	if err != nil {
		return
	}

	pprint.PrintJSON(apps)

	fmt.Printf("Installing %s version %s...\n", appName, version)
	//err = manager.Install(appName, version)
	//if err != nil {
	//	log.Fatalf("Failed to install the app: %v", err)
	//}
	//fmt.Printf("Successfully installed %s version %s!\n", appName, version)

	// Uninstall the app
	//fmt.Printf("Uninstalling %s version %s...\n", appName, version)
	//err = manager.Uninstall(appName, version)
	//if err != nil {
	//	log.Fatalf("Failed to uninstall the app: %v", err)
	//}
	//fmt.Printf("Successfully uninstalled %s version %s!\n", appName, version)
}
