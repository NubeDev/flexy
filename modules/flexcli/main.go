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
	Use:   "global-system-ping",
	Short: "Ping all the apps",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			resp, err := client.PingHostAllCore(timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appInstall = &cobra.Command{
	Use:   "app-path-install",
	Short: "Install an app by its zip folder name",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: appName and appVersion are required")
			}
			appName := args[0]
			appVersion := args[1]
			resp, err := client.BiosInstallApp(appName, appVersion, "", timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appInstallByID = &cobra.Command{
	Use:   "app-install",
	Short: "Install an app by its appID",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: appName and appVersion are required")
			}
			appID := args[0]
			appVersion := args[1]
			resp, err := client.BiosInstallApp("", appVersion, appID, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appUninstall = &cobra.Command{
	Use:   "app-path-uninstall",
	Short: "Uninstall an app by its zip folder name",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: appName and appVersion are required")
			}
			appName := args[0]
			appVersion := args[1]
			resp, err := client.BiosUninstallApp(appName, appVersion, "", timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var appUninstallByID = &cobra.Command{
	Use:   "app-uninstall",
	Short: "Uninstall an app by its appID",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments: appName and appVersion are required")
			}
			appID := args[0]
			appVersion := args[1]
			resp, err := client.BiosUninstallApp("", appVersion, appID, timeout)
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
	Use:   "apps-library",
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

var appSystemctl = &cobra.Command{
	Use:   "app-systemctl",
	Short: "Run systemd/systemctl commands eg; start, stop, restart, enable, disable",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("not enough arguments: service and action are required. eg my-service start")
			}
			service := args[0]
			version := args[1]
			action := args[2]
			var property string
			if len(args) > 2 {
				property = args[2]
			}
			if action == "status" || action == "is-enabled" {
				resp, err := client.BiosSystemdCommandGet("", action, property, service, version, timeout)
				if err != nil {
					return err
				}
				pprint.PrintJSON(resp)
			} else {
				resp, err := client.BiosSystemdCommandPost("", action, property, service, version, timeout)
				if err != nil {
					return err
				}
				pprint.PrintJSON(resp)
			}

			return nil
		})
	},
}

// go run main.go --url=nats://localhost:4222 --client-uuid=abc systemctl my-app start
var systemctlAction = &cobra.Command{
	Use:   "systemctl",
	Short: "Run systemd/systemctl commands eg; start, stop, restart, enable, disable",
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
			if action == "status" || action == "is-enabled" {
				resp, err := client.BiosSystemdCommandGet(service, action, property, "", "", timeout)
				if err != nil {
					return err
				}
				pprint.PrintJSON(resp)
			} else {
				resp, err := client.BiosSystemdCommandPost(service, action, property, "", "", timeout)
				if err != nil {
					return err
				}
				pprint.PrintJSON(resp)
			}

			return nil
		})
	},
}

var downloadReleaseCmd = &cobra.Command{
	Use:   "github-download",
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

var listGitReleaseCmd = &cobra.Command{
	Use:   "github-releases",
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

			resp, err := client.GitListAsset(owner, repo, tag, arch, token, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var natsRequestCmd = &cobra.Command{
	Use:   "nats-request",
	Short: "Send a request to a NATS subject with a provided body",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("not enough arguments, required: appID, subject, body. try; app-abc post.math.add.run 100")
			}
			appID := args[0]
			subject := args[1]
			var body string
			if len(args) > 2 {
				body = args[2]
			}

			// Create the request payload (convert the string body to JSON)
			payload := []byte(body)

			// Send the request to the NATS subject
			msg, err := client.RequestToApp(appID, subject, payload, timeout)
			if err != nil {
				return fmt.Errorf("NATS request failed: %v", err)
			}

			// Print the response from the NATS request
			pprint.PrintJSON(string(msg.Data))
			return nil
		})
	},
}

var getStoresCmd = &cobra.Command{
	Use:   "store-get-stores",
	Short: "Retrieve all object stores",
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			resp, err := client.GetStores(timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var getObjectsCmd = &cobra.Command{
	Use:   "store-get-objects",
	Short: "Retrieve all objects in a store  [storeName]",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			storeName := args[0]
			resp, err := client.GetStoreObjects(storeName, timeout)
			if err != nil {
				return err
			}
			pprint.PrintJSON(resp)
			return nil
		})
	},
}

var addObjectCmd = &cobra.Command{
	Use:   "store-add-object",
	Short: "Add an object to the store from where the cli is being excepted (eg; not on the bios) [storeName] [objectName] [filePath]",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("not enough arguments")
			}
			storeName := args[0]
			objectName := args[1]
			filePath := args[2]
			overwriteIfExisting := args[3]
			var overwrite bool
			if overwriteIfExisting == "true" {
				overwrite = true
			}
			_, err := client.AddObject(storeName, objectName, filePath, overwrite)
			if err != nil {
				return err
			}
			fmt.Printf("Object %s added successfully to store %s\n", objectName, storeName)
			return nil
		})
	},
}

var downloadObjectCmd = &cobra.Command{
	Use:   "store-download-object",
	Short: "Download an object from the store and save it locally  [storeName] [objectName] [destinationPath]",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			storeName := args[0]
			objectName := args[1]
			destinationPath := args[2]

			resp, err := client.DownloadObject(storeName, objectName, destinationPath, timeout)
			if err != nil {
				return err
			}
			fmt.Printf("downloaded successfully to %s \n", resp)
			return nil
		})
	},
}

var deleteObjectCmd = &cobra.Command{
	Use:   "store-delete-object",
	Short: "Delete an object from the store  [storeName] [objectName]",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runCommand(cmd, args, func(client *rqlclient.Client, args []string) error {
			storeName := args[0]
			objectName := args[1]
			resp, err := client.DeleteObject(storeName, objectName, timeout)
			if err != nil {
				fmt.Println(err)
				pprint.PrintJSON(resp)
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
	rootCmd.PersistentFlags().StringVarP(&globalUUID, "global-uuid", "c", "", "global UUID")
	createHostCmd.MarkFlagRequired("global-uuid")
	createHostCmd.Flags().StringVarP(&jsonInput, "json", "j", "", "JSON input")
	createHostCmd.MarkFlagRequired("json")

	// Add the new command to rootCmd
	rootCmd.AddCommand(downloadReleaseCmd)
	rootCmd.AddCommand(listGitReleaseCmd)

	// Add createHostCmd as a subcommand
	rootCmd.AddCommand(createHostCmd)

	// Register commands
	rootCmd.AddCommand(getHostsCmd)
	rootCmd.AddCommand(getHostCmd)
	rootCmd.AddCommand(createHostCmd)
	rootCmd.AddCommand(deleteHostCmd)
	rootCmd.AddCommand(modulesPing)

	rootCmd.AddCommand(appInstallByID)
	rootCmd.AddCommand(appInstall)
	rootCmd.AddCommand(appUninstall)
	rootCmd.AddCommand(appUninstallByID)
	rootCmd.AddCommand(appList)
	rootCmd.AddCommand(appInstalled)
	rootCmd.AddCommand(appSystemctl)
	rootCmd.AddCommand(systemctlAction)
	rootCmd.AddCommand(natsRequestCmd)
	rootCmd.AddCommand(getStoresCmd)
	rootCmd.AddCommand(getObjectsCmd)
	rootCmd.AddCommand(addObjectCmd)
	rootCmd.AddCommand(downloadObjectCmd)
	rootCmd.AddCommand(deleteObjectCmd)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
