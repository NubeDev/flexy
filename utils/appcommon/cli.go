package appcommon

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

func (app *App) SetupCLI(customSetup func(app *App)) {
	// Add persistent flags for config and id
	app.RootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is ./config.yaml)")
	app.RootCmd.PersistentFlags().String("id", "", "Application ID")

	app.RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Get config file path
		configFile, _ := cmd.Flags().GetString("config")
		if err := app.InitConfig(configFile); err != nil {
			return err
		}

		// Get AppID from CLI flag or config file
		cliID, _ := cmd.Flags().GetString("id")
		if cliID != "" {
			app.AppID = cliID
		} else {
			app.AppID = app.Config.GetString("id")
		}

		// Check if AppID is set
		if app.AppID == "" {
			return fmt.Errorf("id is not set in the configuration or CLI")
		}

		// Get description from config
		app.Description = app.Config.GetString("description")

		// Initialize NATS with URL from config or default
		natsURL := app.Config.GetString("url")
		if natsURL == "" {
			natsURL = nats.DefaultURL
		}
		app.NatsConn = natlib.New(natlib.NewOpts{URL: natsURL})
		return nil
	}

	// Call the custom setup function for app-specific CLI setup
	if customSetup != nil {
		customSetup(app)
	}
}
