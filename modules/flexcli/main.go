package main

import (
	"encoding/json"
	"fmt"
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"log"
	"os"
	"time"

	"github.com/NubeDev/flexy/utils/rqlclient"
	"github.com/spf13/cobra"
)

var (
	natsURL    string
	clientUUID string
	timeout    time.Duration
	jsonInput  string // The JSON input as a string

)

// rootCmd is the main command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rql-client",
	Short: "RQLClient CLI is a tool to interact with NATS services",
	Long:  `This CLI tool allows you to interact with the NATS services using RQLClient library.`,
}

func runCommand(cmd *cobra.Command, args []string, execFunc func(client *rqlclient.RQLClient, args []string) error) {
	client, err := rqlclient.New(natsURL)
	if err != nil {
		log.Fatalf("failed to create RQLClient: %v", err)
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
		runCommand(cmd, args, func(client *rqlclient.RQLClient, args []string) error {
			unit := args[0] // The unit is passed as a positional argument
			status, err := client.SystemdStatus(clientUUID, unit, timeout)
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
		runCommand(cmd, args, func(client *rqlclient.RQLClient, args []string) error {
			resp, err := client.GetHosts(clientUUID, timeout)
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
		runCommand(cmd, args, func(client *rqlclient.RQLClient, args []string) error {
			hostUUID := args[0] // The host UUID is passed as a positional argument
			resp, err := client.GetHost(clientUUID, hostUUID, timeout)
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
		runCommand(cmd, args, func(client *rqlclient.RQLClient, args []string) error {
			hostUUID := args[0] // The host UUID is passed as a positional argument
			resp, err := client.DeleteHost(clientUUID, hostUUID, timeout)
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
		runCommand(cmd, args, func(client *rqlclient.RQLClient, args []string) error {
			// Parse the JSON input into the Fields struct
			var fields *hostService.Fields
			err := json.Unmarshal([]byte(jsonInput), &fields)
			if err != nil {
				return fmt.Errorf("failed to parse JSON input: %v", err)
			}
			if fields == nil {
				return fmt.Errorf("fields are empty: %v", err)
			}
			resp, err := client.CreateHost(clientUUID, fields, timeout)
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
	rootCmd.PersistentFlags().StringVarP(&natsURL, "nats-url", "n", "nats://localhost:4222", "NATS server URL")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 5*time.Second, "Request timeout")
	rootCmd.PersistentFlags().StringVarP(&clientUUID, "client-uuid", "c", "", "Client UUID")
	rootCmd.MarkPersistentFlagRequired("client-uuid")

	createHostCmd.Flags().StringVarP(&jsonInput, "json", "j", "", "JSON input")
	createHostCmd.MarkFlagRequired("json")
	// Add createHostCmd as a subcommand
	rootCmd.AddCommand(createHostCmd)

	// Register commands
	rootCmd.AddCommand(systemdStatusCmd)
	rootCmd.AddCommand(getHostsCmd)
	rootCmd.AddCommand(getHostCmd)
	rootCmd.AddCommand(createHostCmd)
	rootCmd.AddCommand(deleteHostCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
