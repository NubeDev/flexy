package main

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
)

func (s *Service) natsStoreInit(nc *nats.Conn) error {
	// Create a NatsRouter
	router := natsrouter.New(nc)
	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	router.JetStreamContext = js
	s.natsStore = router
	storeName := "mystore"
	err = s.natsStore.CreateObjectStore(storeName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) HandleStore(m *nats.Msg) {
	var cmd Command
	// Unmarshal the received JSON message
	if err := json.Unmarshal(m.Data, &cmd); err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}
	switch cmd.Command {
	case "get_stores":
		content := s.natsStore.GetStores()
		marshal, err := json.Marshal(content)
		if err != nil {
			return
		}
		s.publish(m.Reply, string(marshal), code.SUCCESS)
	case "get_store_objects":
		path, ok := cmd.Body["store"].(string)
		if !ok || path == "" {
			s.handleError(m.Reply, code.InvalidParams, "'store' is required for read_file")
			return
		}
		objects, err := s.natsStore.GetStoreObjects(path)
		if err != nil {
			s.handleError(m.Reply, code.InvalidParams, err.Error())
			return
		}
		marshal, err := json.Marshal(objects)
		if err != nil {
			s.handleError(m.Reply, code.InvalidParams, err.Error())
			return
		}
		s.publish(m.Reply, string(marshal), code.SUCCESS)

	default:
		s.handleError(m.Reply, code.UnknownCommand, "Unknown command")
	}

}
