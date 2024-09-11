package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// The subject to send the command to the edge device
	edgeSubject := "ssh.edge.device123"

	// Subscribe to the response from the edge device
	responseSubject := "ssh.response.device123"
	sub, err := nc.SubscribeSync(responseSubject)
	if err != nil {
		log.Fatal(err)
	}

	// Keep the session running
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("NATS SSH-like session started. Type your commands or 'exit' to quit.")

	for {
		// Read command from user input
		fmt.Print("Enter command: ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		// Exit if the user types 'exit'
		if command == "exit" {
			fmt.Println("Exiting session.")
			break
		}

		// Send the SSH command to the edge device
		fmt.Printf("Sending SSH command: %s\n", command)
		err = nc.Publish(edgeSubject, []byte(command))
		if err != nil {
			log.Fatal(err)
		}

		// Wait for the response with a timeout
		msg, err := sub.NextMsg(10 * time.Second)
		if err != nil {
			fmt.Println("Timed out waiting for response from edge device.")
		} else {
			// Print the response
			fmt.Printf("Response from edge device: %s\n", string(msg.Data))
		}
	}
}
