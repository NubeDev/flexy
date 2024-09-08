package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
)

// ModuleID is a unique identifier for this module
const ModuleID = "my-module"

func main() {
	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	// Start the module by subscribing to NATS topics
	StartModule(nc, ModuleID)

	// Keep the module running
	select {}
}

type Host struct {
	UUID string `gorm:"primary_key" json:"uuid"`
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// StartModule starts the NATS subscriptions for the module
func StartModule(nc *nats.Conn, moduleID string) {
	log.Printf("Starting module with ID: %s", moduleID)

	// Subscribe to a method called "hello" for the module
	nc.QueueSubscribe("module."+moduleID+".hello", "module_queue", func(m *nats.Msg) {
		// Handle the method and return a response
		response := fmt.Sprintf("Hello from %s, you said: %s", moduleID, string(m.Data))
		log.Printf("Received 'hello' method call with payload: %s", string(m.Data))
		err := m.Respond([]byte(response))

		if err != nil {
			fmt.Printf("Error responding: %v", err)
		}
	})

	// Subscribe to another method, "status", for the module
	nc.QueueSubscribe("module."+moduleID+".status", "module_queue", func(m *nats.Msg) {
		// Respond with module status
		response := fmt.Sprintf("Module %s is running perfectly!", moduleID)
		log.Printf("Received 'status' method call")
		m.Respond([]byte(response))
	})

	log.Printf("Module %s is now listening for NATS requests...", moduleID)
}
