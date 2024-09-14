package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
)

func (inst *Service) natsStoreInit(nc *nats.Conn) error {
	// Create a NatsRouter
	router := natsrouter.New(nc)
	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	router.JetStreamContext = js
	inst.natsStore = router
	storeName := "mystore"
	err = inst.natsStore.CreateObjectStore(storeName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (inst *Service) HandleStore(m *nats.Msg) {
	var cmd Command
	// Unmarshal the received JSON message
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		inst.handleError(m.Reply, code.ERROR, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}
	switch cmd.Command {
	case "get_stores":
		content := inst.natsStore.GetStores()
		marshal, err := json.Marshal(content)
		if err != nil {
			return
		}
		inst.publish(m.Reply, string(marshal), code.SUCCESS)
	case "get_store_objects":
		path, ok := cmd.Body["store"].(string)
		if !ok || path == "" {
			inst.handleError(m.Reply, code.InvalidParams, "'store' is required for read_file")
			return
		}
		objects, err := inst.natsStore.GetStoreObjects(path)
		if err != nil {
			inst.handleError(m.Reply, code.InvalidParams, err.Error())
			return
		}
		marshal, err := json.Marshal(objects)
		if err != nil {
			inst.handleError(m.Reply, code.InvalidParams, err.Error())
			return
		}
		inst.publish(m.Reply, string(marshal), code.SUCCESS)

	default:
		inst.handleError(m.Reply, code.UnknownCommand, "Unknown command")
	}

}
