package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

func main() {
	service, err := NewService(nats.DefaultURL)
	if err != nil {
		fmt.Printf("Failed to connect to NATS: %v\n", err)
		return
	}
	defer service.natsConn.Close()
	go natsForwarder("abc", service.natsConn, fmt.Sprintf("nats://127.0.0.1:%d", 4223))
	fmt.Println("Service started...")
	service.StartService()

	// Block the main thread forever
	select {}
}
