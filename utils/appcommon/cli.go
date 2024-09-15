package appcommon

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

func (app *App) SetupCLI(customSetup func(app *App)) {
	app.RootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is ./config.yaml)")
	app.RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		if err := app.InitConfig(configFile); err != nil {
			return err
		}
		app.AppID = app.Config.GetString("id")
		if app.AppID == "" {
			return fmt.Errorf("app_id is not set in the configuration")
		}

		natsURL := app.Config.GetString("url")
		if natsURL == "" {
			natsURL = nats.DefaultURL
		}
		if err := app.InitNATS(natsURL); err != nil {
			return err
		}
		return nil
	}

	// Call the custom setup function for app-specific CLI setup
	if customSetup != nil {
		customSetup(app)
	}
}
