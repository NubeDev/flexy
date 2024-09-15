package systemctl

import (
	"errors"
	"fmt"
	"github.com/sergeymakinen/go-systemdconf/v2"
	"github.com/sergeymakinen/go-systemdconf/v2/unit"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ServiceFile struct {
	Name                        string   `json:"name"`
	Version                     string   `json:"version"`
	ServiceDescription          string   `json:"ServiceDescription"`
	RunAsUser                   string   `json:"runAsUser"`
	ServiceWorkingDirectory     string   `json:"serviceWorkingDirectory"`     // /ros/apps/installed/rubix-os/v0.6.1/
	ExecStart                   string   `json:"execStart"`                   // app -p 1660 -g <data_dir> -d data -prod
	AttachWorkingDirOnExecStart bool     `json:"attachWorkingDirOnExecStart"` // true, false
	EnvironmentVars             []string `json:"environmentVars"`             // Environment="g=/data/bacnet-server-c"
	FileNameWithVersion         bool     `json:"FileNameWithVersion"`         // if true service file name will include the version number eg; my-app-v1.1.1.service
}

func GenerateServiceFile(app *ServiceFile, writePath string) (string, error) {
	if app.Name == "" {
		return "", errors.New("name cannot be empty")
	}
	if app.Version == "" {
		return "", errors.New("version cannot be empty")
	}

	workingDirectory := app.ServiceWorkingDirectory
	if workingDirectory == "" {
		workingDirectory = fmt.Sprintf("/ros/apps/installed/%s/%s", app.Name, app.Version)
	}

	user := app.RunAsUser
	if user == "" {
		user = "root"
	}

	execCmd := app.ExecStart
	if app.AttachWorkingDirOnExecStart {
		execCmd = path.Join(workingDirectory, execCmd)
	}

	rootDir := "/ros/apps/installed"
	dataDir := fmt.Sprintf("%s/%s/data", rootDir, app.Name)
	dataDirName := app.Name

	execCmd = strings.ReplaceAll(execCmd, "<root_dir>", rootDir)
	execCmd = strings.ReplaceAll(execCmd, "<data_dir>", dataDir)
	execCmd = strings.ReplaceAll(execCmd, "<data_dir_name>", dataDirName)

	description := app.ServiceDescription
	if description == "" {
		description = fmt.Sprintf("Rubix-OS application %s", app.Name)
	}

	var env systemdconf.Value
	for _, s := range app.EnvironmentVars {
		env = append(env, s)
	}

	service := unit.ServiceFile{
		Unit: unit.UnitSection{
			Description: systemdconf.Value{description},
			After:       systemdconf.Value{"network.target"},
		},
		Service: unit.ServiceSection{
			Type: systemdconf.Value{"simple"},
			ExecOptions: unit.ExecOptions{
				User:             systemdconf.Value{user},
				WorkingDirectory: systemdconf.Value{workingDirectory},
				Environment:      env,
				StandardOutput:   systemdconf.Value{"syslog"},
				StandardError:    systemdconf.Value{"syslog"},
				SyslogIdentifier: systemdconf.Value{app.Name},
			},
			ExecStart:  systemdconf.Value{execCmd},
			Restart:    systemdconf.Value{"always"},
			RestartSec: systemdconf.Value{"10"},
		},
		Install: unit.InstallSection{
			WantedBy: systemdconf.Value{"multi-user.target"},
		},
	}

	b, err := systemdconf.Marshal(service)
	if err != nil {
		return "", err
	}

	var serviceFilePath string
	if writePath != "" {
		if app.FileNameWithVersion {
			serviceFilePath = filepath.Join(writePath, fmt.Sprintf("%s-%s.service", app.Name, app.Version))
		} else {
			serviceFilePath = filepath.Join(writePath, fmt.Sprintf("%s.service", app.Name))
		}
		err = os.WriteFile(serviceFilePath, b, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write service file: %w", err)
		}
	}
	return string(b), nil
}
