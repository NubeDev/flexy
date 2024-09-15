package main

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/appcommon"
	"github.com/nats-io/nats.go"
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
	if err := inst.SetupSubscriptions(); err != nil {
		fmt.Println("Error setting up subscriptions:", err)
		return
	}

	if err := inst.SetupResponders(); err != nil {
		fmt.Println("Error setting up responders:", err)
		return
	}

	// Graceful shutdown handling
	inst.WaitForShutdown()
}

func (inst *App) SetupSubscriptions() error {
	// Subscribe to the "put.ufw" topic
	_, err := inst.app.Subscribe("put", "ufw", inst.HandleSubscriptionMessage)
	if err != nil {
		return fmt.Errorf("error subscribing to put.ufw: %w", err)
	}
	return nil
}

func (inst *App) HandleSubscriptionMessage(msg *nats.Msg) {
	fmt.Printf("Received message on %s: %s\n", msg.Subject, string(msg.Data))
}

func (inst *App) SetupResponders() error {
	// Subscribe to the "get.ufw" topic and respond to requests
	_, err := inst.app.Respond("get", "ufw", inst.HandleRequestMessage)
	if err != nil {
		return fmt.Errorf("error setting up responder for get.ufw: %w", err)
	}
	return nil
}

func (inst *App) handlePing(msg *nats.Msg) []byte {
	fmt.Printf("Received request on %s: %s\n", msg.Subject, string(msg.Data))
	// Process the request and return a response
	response := []byte(fmt.Sprintf("Hello back with: %s", string(msg.Data)))
	return response
}

func (inst *App) HandleRequestMessage(msg *nats.Msg) []byte {
	fmt.Printf("Received request on %s: %s\n", msg.Subject, string(msg.Data))
	// Process the request and return a response
	response := []byte(fmt.Sprintf("Hello back with: %s", string(msg.Data)))
	return response
}

func (inst *App) WaitForShutdown() {
	// Block until a signal is received (for graceful shutdown)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down App A")
	inst.app.NatsConn.Close()
}
