package main

import (
	"encoding/json"
	"fmt"
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"github.com/NubeDev/flexy/utils/rqlclient"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

var (
	natsURL    string
	globalUUID string
	timeout    time.Duration
	jsonInput  string // The JSON input as a string
)

// rootCmd is the main command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rql-client",
	Short: "Client CLI is a tool to interact with NATS services",
	Long:  `This CLI tool allows you to interact with the NATS services using Client library.`,
}

func runCommand(cmd *cobra.Command, args []string, execFunc func(client *rqlclient.Client, args []string) error) {
	client, err := rqlclient.New(natsURL, globalUUID)
	if err != nil {
		log.Fatalf("failed to create Client: %v", err)
	}

	err = execFunc(client, args)
	if err != nil {
		log.Fatalf("command execution failed: %v", err)
	}
}

// systemdStatusCmd for fetching systemd status using positional arguments
var systemdStatusCmd = &cobra.Command{
	Use:   "systemd-status [unit]",
	Short: "Get systemd status for a unit",
	Args:  cobra.ExactArgs(1), // Expect exactly 1 argument (unit name)
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			unit := args[0] // The unit is passed as a positional argument
			status, err := client.SystemdStatus(globalUUID, unit, timeout)
			if err != nil {
				return err
			}
			fmt.Printf("Systemd Status: %+v\n", status)
			return nil
		})
	},
}

// getHostsCmd for fetching all hosts
var getHostsCmd = &cobra.Command{
	Use:   "get-hosts",
	Short: "Retrieve all hosts",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			resp, err := client.GetHosts(globalUUID, timeout)
			if err != nil {
				return err
			}
			fmt.Printf("Hosts count: %+v\n", len(resp))
			// Print each host individually with dereferencing pointers
			for _, host := range resp {
				fmt.Printf("Host: %+v\n", *host) // Dereference the pointer to print actual data
			}
			return nil
		})
	},
}

// getHostCmd for fetching a specific host using its UUID
var getHostCmd = &cobra.Command{
	Use:   "get-host [host-uuid]",
	Short: "Retrieve details of a specific host",
	Args:  cobra.ExactArgs(1), // Expect exactly 1 argument (host UUID)
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			hostUUID := args[0] // The host UUID is passed as a positional argument
			resp, err := client.GetHost(globalUUID, hostUUID, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var deleteHostCmd = &cobra.Command{
	Use:   "delete-host [host-uuid]",
	Short: "Delete a host via its uuid",
	Args:  cobra.ExactArgs(1), // Expect exactly 1 argument (host UUID)
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			hostUUID := args[0] // The host UUID is passed as a positional argument
			resp, err := client.DeleteHost(globalUUID, hostUUID, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var createHostCmd = &cobra.Command{
	Use:   "create-host",
	Short: "Create a host with the provided JSON data",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			// Parse the JSON input into the Fields struct
			var fields *hostService.Fields
			err := json.Unmarshal([]byte(jsonInput), &fields)
			if err != nil {
				return fmt.Errorf("failed to parse JSON input: %v", err)
			}
			if fields == nil {
				return fmt.Errorf("fields are empty: %v", err)
			}
			resp, err := client.CreateHost(globalUUID, fields, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var modulesPing = &cobra.Command{
	Use:   "ping-all-modules",
	Short: "Ping all the modules",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			resp, err := client.PingHostAllCore()
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appInstall = &cobra.Command{
	Use:   "apps-install",
	Short: "Install an app",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: appName and appVersion are required")
			}
			appName := args[0]
			appVersion := args[1]
			resp, err := client.BiosInstallApp(appName, appVersion, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appUninstall = &cobra.Command{
	Use:   "apps-uninstall",
	Short: "Uninstall an app",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: appName and appVersion are required")
			}
			appName := args[0]
			appVersion := args[1]
			resp, err := client.BiosUninstallApp(appName, appVersion, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appInstalled = &cobra.Command{
	Use:   "apps-installed",
	Short: "List all apps that are installed",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			resp, err := client.BiosInstalledApps(timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appList = &cobra.Command{
	Use:   "apps-list",
	Short: "List all available apps from the library that can be installed",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			resp, err := client.BiosLibraryApps(timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var systemCTL = &cobra.Command{
	Use:   "systemctl",
	Short: "Run systemd/systemctl commands",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: service and action are required. eg my-service start")
			}
			service := args[0]
			action := args[1]
			var property string
			if len(args) > 2 {
				property = args[2]
			}

			resp, err := client.BiosSystemdCommand(service, action, property, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var downloadReleaseCmd = &cobra.Command{
	Use:   "download-release",
	Short: "Download a GitHub release asset",
	Long:  `Download a GitHub release asset by specifying owner, repo, tag, architecture, and other options.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("not enough arguments, required: owner, repo, tag, arch, token")
			}
			owner := args[0]
			repo := args[1]
			tag := args[2]
			arch := args[3]
			token := args[4]

			resp, err := client.GitDownloadAsset(owner, repo, tag, arch, token, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

func init() {
	// Define persistent flags common to all commands
	rootCmd.PersistentFlags().StringVarP(&natsURL, "url", "u", "nats://localhost:4222", "NATS server URL")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 5*time.Second, "Request timeout")
	rootCmd.PersistentFlags().StringVarP(&globalUUID, "client-uuid", "c", "", "Client UUID")
	createHostCmd.MarkFlagRequired("client-uuid")
	createHostCmd.Flags().StringVarP(&jsonInput, "json", "j", "", "JSON input")
	createHostCmd.MarkFlagRequired("json")

	// Add the new command to rootCmd
	rootCmd.AddCommand(downloadReleaseCmd)

	// Add createHostCmd as a subcommand
	rootCmd.AddCommand(createHostCmd)

	// Register commands
	rootCmd.AddCommand(systemdStatusCmd)
	rootCmd.AddCommand(getHostsCmd)
	rootCmd.AddCommand(getHostCmd)
	rootCmd.AddCommand(createHostCmd)
	rootCmd.AddCommand(deleteHostCmd)
	rootCmd.AddCommand(modulesPing)

	rootCmd.AddCommand(appInstall)
	rootCmd.AddCommand(appUninstall)
	rootCmd.AddCommand(appList)
	rootCmd.AddCommand(appInstalled)
	rootCmd.AddCommand(systemCTL)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
