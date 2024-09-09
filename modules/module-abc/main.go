package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/modules/module-abc/ufwcommand"
	"github.com/nats-io/nats.go"
	"log"
)

// ModuleID is a unique identifier for this module
const ModuleID = "my-module"

// ufw is the instance of the UFW command system
var ufw = ufwcommand.New()

// Body is used to parse incoming NATS messages
type Body struct {
	Command string      `json:"command"`
	Payload interface{} `json:"body"`
}

// UFWCommandPayload is a struct for parsing the body when sending UFW commands
type UFWCommandPayload struct {
	Port int `json:"port"`
}

func main() {
	// Connect to NATS server
	nc, err := nats.Connect("nats://127.0.0.1:4223")
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()
	// Start the module by subscribing to NATS topics
	StartModule(nc, ModuleID)

	// Keep the module running
	select {}
}

// StartModule starts the NATS subscriptions for the module
func StartModule(nc *nats.Conn, moduleID string) {
	log.Printf("Starting module with ID: %s", moduleID)

	// Subscribe to a method called "ping" for the module
	nc.QueueSubscribe("module."+moduleID+".ping", "module_queue", func(m *nats.Msg) {
		// Handle the method and return a response
		response := fmt.Sprintf("PONG from: %s", ModuleID)
		log.Printf("Received 'ping' method call with payload: %s", string(m.Data))
		err := m.Respond([]byte(response))

		if err != nil {
			fmt.Printf("Error responding: %v", err)
		}
	})

	// Subscribe to a method called "command" for the module
	nc.QueueSubscribe("module."+moduleID+".command", "module_queue", func(m *nats.Msg) {
		var body Body
		err := json.Unmarshal(m.Data, &body)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			m.Respond([]byte(fmt.Sprintf("Error parsing command: %v", err)))
			return
		}

		// Handle the command using the new HandleUFWCommand function
		err = HandleUFWCommand(m, body)
		if err != nil {
			m.Respond([]byte(fmt.Sprintf("Error handling command: %v", err)))
		}
	})

	log.Printf("Module %s is now listening for NATS requests...", moduleID)
}

// HandleUFWCommand handles all UFW related commands
// example ./nats req module.my-module.command "{\"command\": \"ufw\", \"body\": {\"subCommand\": \"open\", \"port\": 8080}}"
func HandleUFWCommand(m *nats.Msg, body Body) error {
	if body.Command != "ufw" {
		return fmt.Errorf("unknown command: %s", body.Command)
	}

	// Convert Payload to the expected format
	payloadMap, subCommand, err := extractPayloadAndSubCommand(body.Payload)
	if err != nil {
		return err
	}
	switch subCommand {
	case "open":
		err := handlePort(m, payloadMap, true)
		if err != nil {
			return err
		}

	case "close":
		err := handlePort(m, payloadMap, false)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid UFW subcommand")
	}

	return nil
}

// handlePort handles extracting the port, opening it, and sending the response
func handlePort(m *nats.Msg, payloadMap map[string]interface{}, open bool) error {
	ufwPayload, err := extractPortFromPayload(payloadMap)
	if err != nil {
		return err
	}
	if open {
		// Open the UFW port
		resp, err := ufw.UWFOpenPort(ufwPayload)
		if err != nil {
			return fmt.Errorf("error opening port: %v", err)
		}
		err = marshalAndRespond(m, resp)
		if err != nil {
			return err
		}
	} else {
		// Close the UFW port
		resp, err := ufw.UWFClosePort(ufwPayload)
		if err != nil {
			return fmt.Errorf("error opening port: %v", err)
		}
		err = marshalAndRespond(m, resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractPayloadAndSubCommand extracts the payload map and subCommand from the body
func extractPayloadAndSubCommand(body interface{}) (map[string]interface{}, string, error) {
	// Convert Payload to the expected format
	payloadMap, ok := body.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("invalid payload format")
	}

	// Extract the subCommand
	subCommand, ok := payloadMap["subCommand"].(string)
	if !ok {
		return nil, "", fmt.Errorf("invalid subcommand format")
	}

	// Return the extracted payload map and subCommand
	return payloadMap, subCommand, nil
}

func marshalAndRespond(m *nats.Msg, resp interface{}) error {
	// Marshal the response to JSON
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("error marshalling response to JSON: %v", err)
	}
	// Send the JSON response
	err = m.Respond(jsonResp)
	if err != nil {
		return fmt.Errorf("error sending response: %v", err)
	}
	return nil
}

func extractPortFromPayload(payloadMap map[string]interface{}) (ufwcommand.UFWBody, error) {
	port, ok := payloadMap["port"].(float64) // JSON unmarshal numbers as float64
	if !ok {
		return ufwcommand.UFWBody{}, fmt.Errorf("invalid port format")
	}
	// Return UFWBody with the extracted port
	return ufwcommand.UFWBody{
		Port: int(port),
	}, nil
}
