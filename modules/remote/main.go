package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// The subject to listen for SSH commands
	edgeSubject := "ssh.edge.device123"

	// Subscribe to SSH commands
	nc.Subscribe(edgeSubject, func(msg *nats.Msg) {
		command := string(msg.Data)
		fmt.Printf("Received SSH command: %s\n", command)

		// Execute the command locally
		out, err := exec.Command("bash", "-c", command).CombinedOutput()
		var response string
		if err != nil {
			response = fmt.Sprintf("Error: %v\n", err)
		} else {
			response = string(out)
		}

		// Send the response back to the cloud
		responseSubject := "ssh.response.device123"
		err = nc.Publish(responseSubject, []byte(response))
		if err != nil {
			log.Printf("Error sending response: %v\n", err)
		}
	})

	// Keep the edge device running
	select {}
}
