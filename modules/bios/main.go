package main

import (
	"fmt"
	"github.com/NubeDev/flexy/app/startup"
	"github.com/nats-io/nats.go"
	"log"
)

const globalUUID = "abc"
const proxyNatsPort = 4223

func main() {
	startup.BootLogger()
	service, err := NewService(nats.DefaultURL, "/data", "")
	if err != nil {
		fmt.Printf("Failed to connect to NATS: %v\n", err)
		return
	}
	defer service.natsConn.Close()
	go service.natsForwarder(globalUUID, service.natsConn, fmt.Sprintf("nats://127.0.0.1:%d", proxyNatsPort))
	err = service.natsStoreInit(service.natsConn)
	if err != nil {
		log.Fatal("failed to init nats-store")
		return
	}
	fmt.Println("Service started...")
	err = service.StartService(globalUUID)
	if err != nil {
		log.Fatal("failed to init nats-service", err)
		return
	}

	// Block the main thread forever
	select {}
}
