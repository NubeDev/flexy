package appmanager

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/NubeDev/flexy/utils/execute/commands"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type AppManager struct {
	LibraryPath      string // Path to the library directory (e.g., data/library)
	InstallPath      string // Path to the install directory (e.g., data/install)
	BackupPath       string // Path to the backup directory (e.g., data/backup)
	TmpPath          string // Path to the backup directory (e.g., data/tmp)
	SystemPath       string // Path to the backup directory (e.g., /lib/systemd/system/)
	cmd              commands.Commands
	systemctlService *commands.SystemctlService
}

// App struct to hold application details
type App struct {
	Path    string `json:"path,omitempty"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewAppManager creates a new AppManager instance
func NewAppManager(rootPath, systemPath string) (*AppManager, error) {
	var libraryPath = "library"
	var installPath = "install"
	var backupPath = "backup"
	var tmpPath = "tmp"
	if rootPath == "" {
		rootPath = "/data"
	}
	if systemPath == "" {
		systemPath = "/etc/systemd/system"
	}
	am := &AppManager{
		LibraryPath: fmt.Sprintf("%s/%s", rootPath, libraryPath),
		InstallPath: fmt.Sprintf("%s/%s", rootPath, installPath),
		BackupPath:  fmt.Sprintf("%s/%s", rootPath, backupPath),
		TmpPath:     fmt.Sprintf("%s/%s", rootPath, tmpPath),
		SystemPath:  systemPath,
		cmd:         commands.New(),
	}
	err := am.ensureDirectories()
	return am, err
}

func (inst *AppManager) ensureDirectories() error {
	dirs := []string{
		inst.LibraryPath,
		inst.InstallPath,
		inst.BackupPath,
		inst.TmpPath,
	}

	// Loop through each directory and create it if it doesn't exist
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			fmt.Printf("Created directory: %s\n", dir)
		}
	}
	return nil
}

func (inst *AppManager) ListLibraryApps() ([]*App, error) {
	return getAppsFromDir(inst.LibraryPath)
}

func getAppsFromDir(dir string) ([]*App, error) {
	// Read the directory contents
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	// Regex to capture version strings with "v" prefix and additional qualifiers (e.g., rc, beta, etc.)
	versionRegex := regexp.MustCompile(`v?\d+(\.\d+)*([.-][a-zA-Z0-9]+)*`)
	var apps []*App
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".zip" {
			fullPath := filepath.Join(dir, file.Name())
			versionMatches := versionRegex.FindStringSubmatch(file.Name())
			if len(versionMatches) > 0 {
				verStr := versionMatches[0]
				appName := strings.Replace(file.Name(), verStr, "", 1)
				appName = strings.TrimSuffix(appName, filepath.Ext(appName)) // Remove the file extension
				appName = strings.Trim(appName, "-_")                        // Clean up extra characters
				verStr = strings.Trim(verStr, ".zip")                        // Clean up extra characters

				// Add the app to the list
				apps = append(apps, &App{
					Path:    fullPath,
					Name:    appName,
					Version: verStr,
				})
			}
		}
	}

	return apps, nil
}

func (inst *AppManager) ListInstalledApps() ([]*App, error) {
	var apps []*App

	// Read the contents of the installation directory
	files, err := os.ReadDir(inst.InstallPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read install directory: %w", err)
	}

	// Loop through the directories (app names)
	for _, appDir := range files {
		if appDir.IsDir() {
			appName := appDir.Name()
			// Read the versions within each app directory
			versions, err := os.ReadDir(filepath.Join(inst.InstallPath, appName))
			if err != nil {
				return nil, fmt.Errorf("failed to read versions for app %s: %w", appName, err)
			}
			// Loop through the versions
			for _, versionDir := range versions {
				if versionDir.IsDir() {
					apps = append(apps, &App{
						Name:    appName,
						Version: versionDir.Name(),
					})
				}
			}
		}
	}

	return apps, nil
}

// Install installs the specified app version
func (inst *AppManager) Install(app *App) error {
	if app == nil {
		return errors.New("app can not bre empty")
	}
	var appName = app.Name
	var version = app.Version

	zipFilePath := filepath.Join(inst.LibraryPath, fmt.Sprintf("%s-%s.zip", appName, version))
	installPath := filepath.Join(inst.InstallPath, appName, version)

	// Step 1: Check if the app exists in the library
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		return fmt.Errorf("app %s version %s not found in the library", appName, version)
	}

	// Step 2: Stop and delete old app version (if exists)
	if err := inst.stopAndRemoveOldApp(appName); err != nil {
		return err
	}

	// Step 3: Unzip the new app to the install directory with the correct structure
	if err := inst.unzipApp(zipFilePath, installPath, appName, version); err != nil {
		return fmt.Errorf("failed to unzip app: %w", err)
	}

	// Step 4: Generate systemd service file
	serviceFilePath, err := inst.createSystemdService(appName, installPath)
	if err != nil {
		return fmt.Errorf("failed to generate systemctl service file: %w", err)
	}

	// Step 5: Move service file and enable it
	if err := inst.setupAndStartService(appName, serviceFilePath); err != nil {
		return fmt.Errorf("failed to setup and start service: %w", err)
	}

	return nil
}

// Uninstall uninstalls the specified app version
func (inst *AppManager) Uninstall(app *App) error {
	if app == nil {
		return errors.New("app can not bre empty")
	}
	var appName = app.Name
	var version = app.Version

	installPath := filepath.Join(inst.InstallPath, appName, version)
	backupPath := filepath.Join(inst.BackupPath, appName, version)

	// Step 1: Check if the app is installed
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return fmt.Errorf("app %s version %s is not installed", appName, version)
	}

	// Step 2: Stop and disable the service
	if err := inst.stopAndDisableService(appName); err != nil {
		return err
	}

	// Step 3: Delete the systemd service file
	if err := inst.deleteSystemdService(appName); err != nil {
		return fmt.Errorf("failed to delete systemd service file: %w", err)
	}

	// Step 4: Backup the app
	if err := inst.backupApp(installPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup app: %w", err)
	}

	// Step 5: Delete the app from the installation directory
	if err := os.RemoveAll(installPath); err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}

	return nil
}

func (inst *AppManager) DeleteSystemFile(appName string) error {
	// Construct the full path to the systemd service file
	serviceFilePath := filepath.Join(inst.SystemPath, fmt.Sprintf("%s.service", appName))

	// Check if the service file exists
	if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
		return fmt.Errorf("system file for app %s not found", appName)
	}

	// Remove the service file
	if err := os.Remove(serviceFilePath); err != nil {
		return fmt.Errorf("failed to remove system file for app %s: %w", appName, err)
	}

	fmt.Printf("System file for app %s successfully deleted.\n", appName)
	return nil
}

func (inst *AppManager) DeleteApp(appName string) error {
	// Construct the full path to the install directory for the app
	appInstallDir := filepath.Join(inst.InstallPath, appName)

	// Check if the install directory exists
	if _, err := os.Stat(appInstallDir); os.IsNotExist(err) {
		return fmt.Errorf("install directory for app %s not found", appName)
	}

	// Remove the install directory
	if err := os.RemoveAll(appInstallDir); err != nil {
		return fmt.Errorf("failed to remove install directory for app %s: %w", appName, err)
	}

	fmt.Printf("Install directory for app %s successfully deleted.\n", appName)
	return nil
}

func (inst *AppManager) DeleteAppBackup(appName string) error {
	// Construct the full path to the backup directory for the app
	appBackupDir := filepath.Join(inst.BackupPath, appName)

	// Check if the backup directory exists
	if _, err := os.Stat(appBackupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory for app %s not found", appName)
	}

	// Remove the backup directory
	if err := os.RemoveAll(appBackupDir); err != nil {
		return fmt.Errorf("failed to remove backup directory for app %s: %w", appName, err)
	}

	fmt.Printf("Backup directory for app %s successfully deleted.\n", appName)
	return nil
}

func (inst *AppManager) DeleteLibraryApp(appName string) error {
	// Construct the full path to the library files for the app
	files, err := os.ReadDir(inst.LibraryPath)
	if err != nil {
		return fmt.Errorf("failed to read library directory: %w", err)
	}

	// Loop through the files and find the ones matching the app name
	for _, file := range files {
		// Match files that start with the app name and have a .zip extension
		if !file.IsDir() && strings.HasPrefix(file.Name(), appName) && filepath.Ext(file.Name()) == ".zip" {
			filePath := filepath.Join(inst.LibraryPath, file.Name())

			// Remove the file
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to remove library file for app %s: %w", appName, err)
			}

			fmt.Printf("Library file %s for app %s successfully deleted.\n", file.Name(), appName)
		}
	}

	return nil
}

// stopAndRemoveOldApp stops and removes an old app version (if exists)
func (inst *AppManager) stopAndRemoveOldApp(appName string) error {
	// Stop and remove the old version of the app
	return inst.stopAndDisableService(appName)
}

// deleteSystemdService deletes the systemd service file for the app
func (inst *AppManager) deleteSystemdService(appName string) error {
	serviceFilePath := filepath.Join(inst.SystemPath, fmt.Sprintf("%s.service", appName))
	if _, err := os.Stat(serviceFilePath); err == nil {
		// The service file exists, attempt to delete it
		if err := os.Remove(serviceFilePath); err != nil {
			return fmt.Errorf("failed to delete service file: %w", err)
		}
		fmt.Printf("Deleted systemd service file: %s\n", serviceFilePath)
	} else if os.IsNotExist(err) {
		// Service file does not exist, nothing to delete
		fmt.Printf("Systemd service file %s does not exist, skipping deletion\n", serviceFilePath)
	} else {
		return err
	}
	return nil
}

// stopAndDisableService stops and disables the systemd service for the app
func (inst *AppManager) stopAndDisableService(appName string) error {
	serviceName := fmt.Sprintf("%s.service", appName)
	if err := inst.cmd.SystemdCommand(serviceName, "stop"); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	if err := inst.cmd.SystemdCommand(serviceName, "disable"); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}
	return nil
}

// unzipApp unzips the app from a zip file into the correct install directory structure
func (inst *AppManager) unzipApp(zipFilePath, destPath, appName, version string) error {
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Ensure the destination directory exists
	fmt.Println("INSTALL DIR", destPath)
	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return err
	}

	// Extract the files directly into destPath (without creating additional directories)
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue // Skip directories, no need to create them
		}

		// Extract the file directly to destPath, ignoring the zip's internal folder structure
		fpath := filepath.Join(destPath, filepath.Base(file.Name))

		// Extract the file
		if err := extractFile(file, fpath); err != nil {
			return err
		}
		if file.Name != appName {
			newBinaryPath := filepath.Join(destPath, appName)
			if err := os.Rename(fpath, newBinaryPath); err != nil {
				return err
			}
		}
		newBinaryPath := filepath.Join(destPath, appName)
		if err := inst.setExecutable(newBinaryPath); err != nil {
			return err
		}

	}

	return nil
}

// extractFile extracts a single file from a zip archive
func extractFile(file *zip.File, dest string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create the destination file
	outFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy the file contents
	_, err = io.Copy(outFile, rc)
	return err
}

// createSystemdService creates a systemd service file for the app
func (inst *AppManager) createSystemdService(appName, execPath string) (string, error) {
	// Use the existing systemctl library for generating a service file
	service := commands.NewSystemctlService(appName, fmt.Sprintf("Service for %s", appName), fmt.Sprintf("%s/%s", execPath, appName), "always")
	inst.systemctlService = service
	tmpServiceFilePath, err := service.GenerateServiceFile(inst.TmpPath) // temp path for generating service file
	if err != nil {
		return "", err
	}
	fmt.Printf("create systemd file on tmp path: %s \n", tmpServiceFilePath)
	return tmpServiceFilePath, nil
}

// setupAndStartService moves the service file to the appropriate location, enables, and starts it
func (inst *AppManager) setupAndStartService(appName, serviceFilePath string) error {
	// Move the service file to /etc/systemd/system/
	if err := inst.systemctlService.MoveServiceFile(serviceFilePath, inst.SystemPath); err != nil {
		return err
	}
	fmt.Printf("move systemd file to from: %s to: %s \n", serviceFilePath, inst.SystemPath)
	serviceName := fmt.Sprintf("%s.service", appName)

	// Enable and start the service
	if err := inst.cmd.SystemdCommand(serviceName, "enable"); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	if err := inst.cmd.SystemdCommand(serviceName, "start"); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

// backupApp makes a backup of the app to the backup directory
func (inst *AppManager) backupApp(installPath, backupPath string) error {
	if err := os.MkdirAll(backupPath, os.ModePerm); err != nil {
		return err
	}

	// Copy files to the backup directory
	err := filepath.Walk(installPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(installPath, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(backupPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})

	return err
}

func (inst *AppManager) setExecutable(path string) error {
	return os.Chmod(path, 0755) // Set as executable
}

// copyFile copies a file from source to destination
func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
