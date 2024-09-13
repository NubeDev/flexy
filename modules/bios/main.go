package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
)

const globalUUID = "abc"

func main() {
	service, err := NewService(nats.DefaultURL, "/data", "")
	if err != nil {
		fmt.Printf("Failed to connect to NATS: %v\n", err)
		return
	}
	defer service.natsConn.Close()
	go natsForwarder(globalUUID, service.natsConn, fmt.Sprintf("nats://127.0.0.1:%d", 4222))
	err = service.natsStoreInit(service.natsConn)
	if err != nil {
		log.Fatal("failed to init nats-store")
		return
	}
	fmt.Println("Service started...")
	service.StartService(globalUUID)

	// Block the main thread forever
	select {}
}
