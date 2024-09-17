package main

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/appcommon"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	appCommon := appcommon.NewApp()
	app := &App{
		appCommon,
	}

	customSetup := func(app *appcommon.App) {
		app.RootCmd.Use = "app-abc"
		app.RootCmd.Short = "App A Command Line Interface"
		app.RootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
		app.Config.BindPFlag("debug", app.RootCmd.PersistentFlags().Lookup("debug"))
	}

	appCommon.SetupCLI(customSetup)

	// Define the main command action
	appCommon.RootCmd.Run = func(cmd *cobra.Command, args []string) {
		app.RunApp()
	}

	if err := appCommon.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type App struct {
	app *appcommon.App
}

func (inst *App) RunApp() {
	debug := inst.app.Config.GetBool("debug")
	if debug {
		fmt.Println("Debug mode is enabled for App A")
	}
	fmt.Printf("App '%s' is running\n", inst.app.AppID)

	// Setup NATS subscriptions and responders
	if err := inst.SetupResponders(); err != nil {
		fmt.Println("Error setting up responders:", err)
		return
	}

	// Graceful shutdown handling
	inst.WaitForShutdown()
}

func (inst *App) HandleSubscriptionMessage(msg *nats.Msg) {
	fmt.Printf("Received message on %s: %s\n", msg.Subject, string(msg.Data))
}

func (inst *App) SetupResponders() error {
	// Use global subject for ping
	err := inst.app.NatsConn.SubscribeWithRespond("global.get.system.ping", inst.globalPing, nil)
	if err != nil {
		return fmt.Errorf("error subscribing to global.get.system.ping: %w", err)
	}
	return nil
}

func (inst *App) globalPing(msg *nats.Msg) ([]byte, error) {
	// Create a Ping response in JSON format
	response := natlib.NewResponse(code.SUCCESS, inst.app.AppID, natlib.Args{Description: inst.app.Description})
	// Marshal to JSON and handle error
	jsonResponse, err := response.ToJSONError()
	if err != nil {
		log.Error().Msgf("Error converting ping to JSON: %v", err)
		return nil, err
	}

	return jsonResponse, nil
}

func (inst *App) WaitForShutdown() {
	// Block until a signal is received (for graceful shutdown)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down App A")
	inst.app.NatsConn.Close()
}
