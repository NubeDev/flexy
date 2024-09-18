package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/appcommon"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/guides"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/NubeDev/flexy/utils/subjects"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	appCommon := appcommon.NewApp()
	app := &App{
		app: appCommon,
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
	app      *appcommon.App
	subjects *subjects.SubjectBuilder
}

func (inst *App) RunApp() {
	debug := inst.app.Config.GetBool("debug")
	if debug {
		fmt.Println("Debug mode is enabled for App A")
	}
	inst.subjects = subjects.NewSubjectBuilder(inst.app.AppID, inst.app.AppID, subjects.IsApp)

	fmt.Printf("App '%s' is running\n", inst.app.AppID)

	// Setup NATS subscriptions and responders
	if err := inst.SetupResponders(); err != nil {
		fmt.Println("Error setting up responders:", err)
		return
	}

	// Graceful shutdown handling
	inst.WaitForShutdown()
}

func (inst *App) SetupResponders() error {
	// Use global subject for ping
	err := inst.app.NatsConn.SubscribeWithRespond("global.get.system.ping", inst.globalPing, nil)
	if err != nil {
		return fmt.Errorf("error subscribing to global.get.system.ping: %w", err)
	}

	err = inst.app.NatsConn.SubscribeWithRespond(inst.subjects.BuildSubject("get", "system", "help"), inst.getHelp, nil)
	if err != nil {
		return fmt.Errorf("error subscribing to global.get.system.ping: %w", err)
	}

	err = inst.app.NatsConn.SubscribeWithRespond(inst.subjects.BuildSubject("post", "math", "add.*"), inst.mathAdd, nil)
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

// mathAdd receives a message and either returns help or executes the math operation based on the subject
func (inst *App) mathAdd(msg *nats.Msg) ([]byte, error) {
	// Split the subject to check for help or run commands
	subjectParts := strings.Split(msg.Subject, ".")
	if len(subjectParts) < 3 {
		return nil, fmt.Errorf("invalid subject format")
	}

	command := subjectParts[4] // e.g., "help" or "run"
	switch command {
	case "help":
		// Provide help details for mathAdd
		helpDetails, err := inst.getMathAddHelp("mathAdd")
		if err != nil {
			return nil, fmt.Errorf("unknown help guild: %s", command)
		}
		// Return the help details as JSON
		return json.Marshal(helpDetails)

	case "run":
		// Run the math operation (2x number)
		return inst.runMathAdd(msg)

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

// runMathAdd performs the actual math operation of adding the number to itself (2x)
func (inst *App) runMathAdd(msg *nats.Msg) ([]byte, error) {
	// Extract the string number from the message
	numberStr := string(msg.Data)

	// Convert the string to an integer
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		log.Error().Msgf("Error converting string to integer: %v", err)
		return nil, fmt.Errorf("invalid number format: %s", numberStr)
	}

	// Perform the math: add the number to itself (multiply by 2)
	result := number * 2

	// Create a response with the result
	response := natlib.NewResponse(200, fmt.Sprint(result))

	// Marshal the response to JSON
	return json.Marshal(response)
}

func (inst *App) getHelp(msg *nats.Msg) ([]byte, error) {
	// Split the subject to check for help or run commands
	return json.Marshal(inst.allHelp())
}

func (inst *App) allHelp() *guides.HelpGuide {
	arg1 := guides.NewArgFloat("num1")
	method := guides.NewMethod("mathAdd", "Will multiply the number pass in by two", "APP_ID.post.math.add.run", false, "", []guides.Args{arg1})
	module := guides.NewModule("MathOperations", []guides.Method{method})
	return guides.NewHelpGuide([]guides.Module{module})

}

// getMathAddHelp returns help information for the mathAdd function
func (inst *App) getMathAddHelp(methodName string) (map[string]interface{}, error) {
	return inst.allHelp().GetMethodDetails(methodName)
}

func (inst *App) WaitForShutdown() {
	// Block until a signal is received (for graceful shutdown)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Shutting down App A")
	inst.app.NatsConn.Close()
}
