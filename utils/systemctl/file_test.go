package systemctl

import (
	"fmt"
	"testing"
)

func TestGenerateServiceFileContent(t *testing.T) {
	app := &ServiceFile{
		Name:      "rubix-os",
		Version:   "v0.6.1",
		ExecStart: "app -p 1660 -g <data_dir> -d data -prod",
		EnvironmentVars: []string{
			`ENV_VAR1=value1`,
			`ENV_VAR2=value2`,
		},
	}

	serviceContent, err := GenerateServiceFile(app, "./")
	if err != nil {
		// handle error
	}

	// Use serviceContent as needed
	fmt.Println(serviceContent)
}
