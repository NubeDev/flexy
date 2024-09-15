package main

import (
	"fmt"
	"github.com/NubeDev/flexy/app/startup"
	"github.com/nats-io/nats.go"
	"log"
)

func main() {
	startup.BootLogger()
	const proxyNatsPort = 4223
	opts := &Opts{
		GlobalUUID:      "abc",
		NatsURL:         nats.DefaultURL,
		DataPath:        "/data",
		SystemPath:      "",
		GitToken:        "",
		GitDownloadPath: "/data/library",
		ProxyNatsPort:   proxyNatsPort,
	}
	service, err := NewService(opts)
	if err != nil {
		fmt.Printf("Failed to connect to NATS: %v\n", err)
		return
	}
	defer service.natsConn.Close()
	go service.natsForwarder(service.natsConn, fmt.Sprintf("nats://127.0.0.1:%d", opts.ProxyNatsPort))
	err = service.natsStoreInit(service.natsConn)
	if err != nil {
		log.Fatal("failed to init nats-store")
		return
	}
	err = service.StartService()
	if err != nil {
		log.Fatal("failed to init nats-service", err)
		return
	}

	// Block the main thread forever
	select {}
}
