package appmanager

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/NubeDev/flexy/utils/systemctl"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ManagerInterface interface {
	ListLibraryApps() ([]*App, error)
	ListInstalledApps() ([]*App, error)
	Install(app *App) error
	Uninstall(app *App) error
	DeleteSystemFile(appName string) error
	DeleteApp(appName string) error
	DeleteAppBackup(appName string) error
	DeleteLibraryApp(appName string) error
	RestoreBackup(name, version string) error
	ListBackups() ([]*App, error)
}

type AppManager struct {
	LibraryPath      string // Path to the library directory (e.g., data/library)
	InstallPath      string // Path to the install directory (e.g., data/install)
	BackupPath       string // Path to the backup directory (e.g., data/backup)
	TmpPath          string // Path to the backup directory (e.g., data/tmp)
	SystemPath       string // Path to the backup directory (e.g., /lib/systemd/system/)
	systemctlService systemctl.Commands
}

// App struct to hold application details
type App struct {
	Path    string `json:"path,omitempty"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Config struct {
	ID          string          `yaml:"id"`
	URL         string          `yaml:"url"`
	ServiceFile ServiceFileYAML `yaml:"service_file"`
}

type ServiceFileYAML struct {
	Env string `yaml:"env"`
}

// NewAppManager creates a new AppManager instance
func NewAppManager(rootPath, systemPath string) (ManagerInterface, error) {
	var libraryPath = "library"
	var installPath = "installed"
	var backupPath = "backups"
	var tmpPath = "/ros/tmp"
	if rootPath == "" {
		rootPath = "/ros/apps"
	}
	if systemPath == "" {
		systemPath = "/etc/systemd/system"
	}
	am := &AppManager{
		LibraryPath:      fmt.Sprintf("%s/%s", rootPath, libraryPath),
		InstallPath:      fmt.Sprintf("%s/%s", rootPath, installPath),
		BackupPath:       fmt.Sprintf("%s/%s", rootPath, backupPath),
		TmpPath:          tmpPath,
		SystemPath:       systemPath,
		systemctlService: systemctl.New(),
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
			log.Info().Msgf("bios created directory: %s", dir)
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

	// Regex pattern to capture app name and version
	versionedFilenameRegex := regexp.MustCompile(`^(.+?)[-_]?v(\d+(\.\d+)*)(.*)$`)

	var apps []*App
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".zip" {
			fullPath := filepath.Join(dir, file.Name())
			filename := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			matches := versionedFilenameRegex.FindStringSubmatch(filename)
			if matches != nil {
				appName := matches[1] + matches[4]    // Combine name and any suffix
				version := "v" + matches[2]           // Add back 'v' to version
				appName = strings.Trim(appName, "-_") // Clean up any leading/trailing hyphens or underscores
				// Add the app to the list
				apps = append(apps, &App{
					Path:    fullPath,
					Name:    appName,
					Version: version,
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
		return errors.New("app cannot be empty")
	}
	var appName = app.Name
	var version = app.Version
	apps, err := inst.ListLibraryApps()
	if err != nil {
		return err
	}
	for _, appList := range apps {
		fmt.Println(appList.Name, appList.Version)
		if appList.Name == appName {
			if appList.Version == version {
				app.Path = appList.Path
			}
		}

	}
	zipFilePath := app.Path
	installPath := filepath.Join(inst.InstallPath, appName, version)

	// Step 1: Check if the app exists in the library
	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		return fmt.Errorf("app %s version %s not found in the library", appName, version)
	}

	// Step 2: Stop and delete old app version (if exists)
	if err := inst.stopAndRemoveOldApp(appName); err != nil {
		return err
	}

	// Step 3: Extract the binary
	if err := inst.unzipApp(zipFilePath, installPath, appName); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	// Step 4: Extract config.yaml if exists
	configFilePath, err := inst.extractConfigFile(zipFilePath, installPath)
	if err != nil {
		return fmt.Errorf("failed to extract config.yaml: %w", err)
	}
	// Step 5: Parse config.yaml
	var config *Config
	if configFilePath != "" {
		log.Info().Msgf("transfer config file: %s", configFilePath)
		config, err = inst.parseConfigFile(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to parse config.yaml: %w", err)
		}
	}
	// Step 6: Generate systemd service file
	err = inst.createSystemdService(appName, installPath, version, config)
	if err != nil {
		return fmt.Errorf("failed to generate systemctl service file: %w", err)
	}

	// Step 7: Move service file and enable it
	if err := inst.setupAndStartService(appName); err != nil {
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

	log.Info().Msgf("System file for app %s successfully deleted.", appName)
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

	log.Info().Msgf("Install directory for app %s successfully deleted.", appName)
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

	log.Info().Msgf("Backup directory for app %s successfully deleted. ", appName)
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

			log.Info().Msgf("Library file %s for app %s successfully deleted. ", file.Name(), appName)
		}
	}

	return nil
}

// RestoreBackup restores a specific app version from the backup directory
func (inst *AppManager) RestoreBackup(name, version string) error {
	backupPath := filepath.Join(inst.BackupPath, name, version)
	installPath := filepath.Join(inst.InstallPath, name, version)

	// Check if the backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup for app %s version %s not found", name, version)
	}

	// Stop and remove old app version (if exists)
	if err := inst.stopAndRemoveOldApp(name); err != nil {
		return err
	}

	// Restore the app from the backup directory
	err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(backupPath, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(installPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})

	if err != nil {
		return fmt.Errorf("failed to restore app %s version %s: %w", name, version, err)
	}

	// Setup and start the service after restoring
	if err := inst.setupAndStartService(name); err != nil {
		return fmt.Errorf("failed to setup and start service: %w", err)
	}

	return nil
}

// ListBackups lists all the available app backups
func (inst *AppManager) ListBackups() ([]*App, error) {
	return getAppsFromDir(inst.BackupPath)
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
		log.Info().Msgf("Deleted systemd service file: %s ", serviceFilePath)
	} else if os.IsNotExist(err) {
		// Service file does not exist, nothing to delete
		log.Info().Msgf("Systemd service file %s does not exist, skipping deletion ", serviceFilePath)
	} else {
		return err
	}
	return nil
}

// stopAndDisableService stops and disables the systemd service for the app
func (inst *AppManager) stopAndDisableService(appName string) error {
	serviceName := fmt.Sprintf("%s.service", appName)
	if err := inst.systemctlService.SystemdCommand(serviceName, "stop"); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	if err := inst.systemctlService.SystemdCommand(serviceName, "disable"); err != nil {
		return fmt.Errorf("failed to disable service: %w", err)
	}
	return nil
}

// unzipApp unzips the app from a zip file into the correct install directory structure
// unzipApp unzips the app from a zip file into the correct install directory structure
func (inst *AppManager) unzipApp(zipFilePath, destPath, appName string) error {
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Ensure the destination directory exists
	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return err
	}

	// Extract all files from the zip archive
	for _, file := range reader.File {
		filePath := filepath.Join(destPath, file.Name)

		// Check if the current file is a directory
		if file.FileInfo().IsDir() {
			// Create directory
			if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// Ensure the directory for the file exists
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		// Extract the file
		if err := extractFile(file, filePath); err != nil {
			return err
		}

		// If this is the app binary, set it as executable
		if file.Name == appName || filepath.Ext(file.Name) == "" {
			if err := inst.setExecutable(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (inst *AppManager) extractConfigFile(zipFilePath, destPath string) (string, error) {
	// Open the zip file
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// Get the folder name from the zip file (without the .zip extension)
	zipFolderName := strings.TrimSuffix(filepath.Base(zipFilePath), filepath.Ext(zipFilePath))

	var configFilePath string

	// Iterate through files in the zip archive
	for _, file := range reader.File {
		// Check if the file path matches the expected format: <zipFolderName>/config.yaml
		expectedConfigPath := filepath.Join(zipFolderName, "config.yaml")
		if file.Name == expectedConfigPath {
			// Extract config.yaml to destPath
			fpath := filepath.Join(destPath, filepath.Base(file.Name))
			if err := extractFile(file, fpath); err != nil {
				return "", err
			}
			configFilePath = fpath
			break
		}
	}

	if configFilePath == "" {
		//return "", fmt.Errorf("config.yaml not found in the zip file")
	}

	return configFilePath, nil
}

func (inst *AppManager) parseConfigFile(configFilePath string) (*Config, error) {
	var config *Config

	// Read and parse config.yaml
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return config, err
	}

	return config, nil
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
func (inst *AppManager) createSystemdService(appName, execPath, version string, config *Config) error {
	serviceFile := &systemctl.ServiceFile{
		Name:                        appName,
		Version:                     version,
		ServiceDescription:          "",
		RunAsUser:                   "",
		ServiceWorkingDirectory:     "",
		ExecStart:                   fmt.Sprintf("%s/%s", execPath, appName),
		AttachWorkingDirOnExecStart: false,
		EnvironmentVars:             nil,
		FileNameWithVersion:         false,
	}

	if config != nil {
		// Use data from config to populate serviceFile
		if config.ServiceFile.Env != "" {
			serviceFile.EnvironmentVars = []string{config.ServiceFile.Env}
		}
		// serviceFile.ServiceDescription = config.ServiceDescription
		// serviceFile.RunAsUser = config.RunAsUser
	}

	// Generate and move the service file
	_, err := systemctl.GenerateServiceFile(serviceFile, inst.SystemPath)
	return err
}

// setupAndStartService moves the service file to the appropriate location, enables, and starts it
func (inst *AppManager) setupAndStartService(appName string) error {
	serviceName := fmt.Sprintf("%s.service", appName)
	// Enable and start the service
	if err := inst.systemctlService.SystemdCommand(serviceName, "enable"); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}
	if err := inst.systemctlService.SystemdCommand(serviceName, "start"); err != nil {
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
