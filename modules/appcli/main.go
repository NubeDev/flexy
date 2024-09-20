package main

import (
	"archive/zip"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/*
go run main.go generate --id=app-abc --version=v1.0.3 --desc="A demo app"


go build main.go && sudo ./main build --id=app-abc --version=v1.0.3 --arch=amd64 --go-path=/home/user/sdk/go1.23.1/bin/go

*/

var appID, appVersion, appArch, appDesc, goPath string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app-cli",
		Short: "App CLI Tool",
	}

	// Generate command
	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate an app template",
		Run: func(cmd *cobra.Command, args []string) {
			generateApp(appID, appVersion, appDesc)
		},
	}

	// Build command
	var buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build and zip the app",
		Run: func(cmd *cobra.Command, args []string) {
			buildApp(appID, appVersion, appArch, goPath)
		},
	}

	// Flags for generate command
	generateCmd.Flags().StringVar(&appID, "id", "", "App ID (required)")
	generateCmd.Flags().StringVar(&appVersion, "version", "v1.0.0", "App version")
	generateCmd.Flags().StringVar(&appDesc, "desc", "A demo application", "App description")
	generateCmd.MarkFlagRequired("id")

	// Flags for build command
	buildCmd.Flags().StringVar(&appID, "id", "", "App ID (required)")
	buildCmd.Flags().StringVar(&appVersion, "version", "v1.0.0", "App version")
	buildCmd.Flags().StringVar(&appArch, "arch", "amd64", "App architecture (e.g., amd64, arm64)")
	buildCmd.Flags().StringVar(&goPath, "go-path", "", "Path to go executable (optional)")

	buildCmd.MarkFlagRequired("id")

	// Add commands to root
	rootCmd.AddCommand(generateCmd, buildCmd)
	rootCmd.Execute()
}

// generateApp creates the folder and files for the app template
func generateApp(id, version, description string) {
	appDir := filepath.Join(".", id)

	// Create directory
	err := os.Mkdir(appDir, 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// Create config.yaml
	configContent := fmt.Sprintf(`id: "%s"
version: "%s"
description: "%s"
url: "nats://localhost:4222"`, id, version, description)
	configPath := filepath.Join(appDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		fmt.Println("Error writing config.yaml:", err)
		return
	}

	// Create main.go
	mainGoContent :=
		`
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
		app.RootCmd.Use = "rubix app"      // replace as id
		app.RootCmd.Short = "App cli tool" // replace as description
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


`

	mainGoPath := filepath.Join(appDir, "main.go")
	err = os.WriteFile(mainGoPath, []byte(mainGoContent), 0644)
	if err != nil {
		fmt.Println("Error writing main.go:", err)
		return
	}

	myFigure := figure.NewFigure(fmt.Sprintf(":) %s", strings.ToUpper(appID)), "", true)
	myFigure.Print()

}

// buildApp builds and zips the app, then moves it to /ros/apps/library
// buildApp builds and zips the app, then moves it to /ros/apps/library
func buildApp(id, version, arch, goPath string) {
	appDir := filepath.Join(".", id)
	// Build file name format: [id]-[arch]
	buildOutputFile := fmt.Sprintf("%s-%s", id, arch)

	// If no specific goPath is provided, use "go" from the default PATH
	if goPath == "" {
		goPath = "go"
	}

	// Change the working directory to appDir to build all Go files in that directory
	buildCmd := exec.Command(goPath, "build", "-o", buildOutputFile)
	buildCmd.Dir = appDir // Set the working directory to the appDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err := buildCmd.Run()
	if err != nil {
		fmt.Println("Error building the app:", err)
		return
	}

	// Zip file name format: [id]-[version]-[arch]
	zipOutputFile := fmt.Sprintf("%s-%s-%s.zip", id, version, arch)
	zipPath := filepath.Join("/ros/apps/library", zipOutputFile)
	//zipPath := filepath.Join("/home/user", zipOutputFile) // Optional alternative path for testing

	// Zip the built file and config.yaml
	err = zipFiles(zipPath, []string{filepath.Join(appDir, buildOutputFile), filepath.Join(appDir, "config.yaml")})
	if err != nil {
		fmt.Println("Error zipping the app:", err)
		return
	}

	fmt.Println("App built and moved to:", zipPath)
}

// zipFiles creates a zip archive from a list of files
func zipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		err = addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}
	return nil
}

// addFileToZip adds a file to the zip archive
func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
