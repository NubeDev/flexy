package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
	"strings"
)

/*
Usage

./nats req abc.post.system.store.get.object '{"storeName": "bios"}'

./nats req abc.post.system.store.download.object \
'{
  "storeName": "bios",
  "objectName": "flexy-app-flexy-app-v1.0.3-amd64.zip",
  "destinationPath": "/home/user/app.zip"
}'

*/

func (s *Service) natsStoreInit(storeName string) error {
	// Create a NatsRouter
	err := s.natsClient.CreateObjectStore(storeName, nil)
	if err != nil {
		return err
	}
	return nil
}

type StoreRequest struct {
	Action          string `json:"action"` // e.g., "add.object", "delete.object", "download.object"
	StoreName       string `json:"storeName"`
	ObjectName      string `json:"objectName"`
	DestinationPath string `json:"destinationPath"` // Used for download
	Data            string `json:"data"`            // Base64-encoded data for add.object
}

func (s *Service) handleStore(m *nats.Msg) {
	if s.natsStore == nil {
		s.handleError(m.Reply, code.InvalidParams, "Store is not enabled in the config file")
		return
	}

	var decoded StoreRequest
	// Unmarshal the received JSON message
	if err := json.Unmarshal(m.Data, &decoded); err != nil {
		s.handleError(m.Reply, code.ERROR, fmt.Sprintf("Invalid JSON format: %v", err))
		return
	}

	decoded.Action = s.getActionFromMessage(decoded.Action, m.Subject)
	if decoded.Action == "" {
		s.handleError(m.Reply, code.ERROR, "failed to find a valid action, try get.stores")
		return
	}
	actionHandlers := map[string]func(*nats.Msg, StoreRequest){
		"get.stores":      s.handleGetStores,
		"get.object":      s.handleGetObject,
		"add.object":      s.handleAddObject,
		"delete.object":   s.handleDeleteObject,
		"download.object": s.handleDownloadObject,
	}

	if handler, found := actionHandlers[decoded.Action]; found {
		handler(m, decoded)
	} else {
		s.handleError(m.Reply, code.UnknownCommand, "Unknown command")
	}
}

func (s *Service) getActionFromMessage(action, subject string) string {
	if action != "" {
		return action
	}
	const storePrefix = "post.system.store."
	if idx := strings.Index(subject, storePrefix); idx != -1 {
		return subject[idx+len(storePrefix):]
	}
	return ""
}

func (s *Service) handleGetStores(m *nats.Msg, _ StoreRequest) {
	content, err := s.natsClient.GetStores()
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	s.publishResponse(m, content, code.SUCCESS)
}

func (s *Service) handleGetObject(m *nats.Msg, decoded StoreRequest) {
	storeName := s.validateField(m.Reply, decoded.StoreName, "Store name is required")
	if storeName == "" {
		return
	}
	objects, err := s.natsClient.GetStoreObjects(storeName)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	s.publishResponse(m, objects, code.SUCCESS)
}

func (s *Service) handleAddObject(m *nats.Msg, decoded StoreRequest) {
	storeName := s.validateField(m.Reply, decoded.StoreName, "Store name is required")
	objectName := s.validateField(m.Reply, decoded.ObjectName, "Object name is required")
	if storeName == "" || objectName == "" || s.validateField(m.Reply, decoded.Data, "Data is required for adding an object") == "" {
		return
	}
	dataBytes, err := base64.StdEncoding.DecodeString(decoded.Data)
	if err != nil {
		s.handleError(m.Reply, code.InvalidParams, "Invalid base64 data: "+err.Error())
		return
	}
	err = s.natsClient.PutBytes(storeName, objectName, dataBytes, true)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	out := Message{
		"Object added successfully",
	}
	s.publishResponse(m, out, code.SUCCESS)
}

func (s *Service) handleDeleteObject(m *nats.Msg, decoded StoreRequest) {
	storeName := s.validateField(m.Reply, decoded.StoreName, "Store name is required")
	objectName := s.validateField(m.Reply, decoded.ObjectName, "Object name is required")
	if storeName == "" || objectName == "" {
		return
	}
	err := s.natsClient.DeleteObject(storeName, objectName)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	out := Message{
		"Object deleted successfully",
	}
	s.publishResponse(m, out, code.SUCCESS)
}

func (s *Service) handleDownloadObject(m *nats.Msg, decoded StoreRequest) {
	storeName := s.validateField(m.Reply, decoded.StoreName, "Store name is required")
	objectName := s.validateField(m.Reply, decoded.ObjectName, "Object name is required")
	destinationPath := s.validateField(m.Reply, decoded.DestinationPath, "Destination path is required")
	if storeName == "" || objectName == "" || destinationPath == "" {
		return
	}
	err := s.natsClient.DownloadObject(storeName, objectName, destinationPath)
	if err != nil {
		s.handleError(m.Reply, code.ERROR, err.Error())
		return
	}
	out := Message{
		"Object downloaded successfully",
	}
	s.publishResponse(m, out, code.SUCCESS)
}

func (s *Service) validateField(reply string, field, errorMsg string) string {
	if field == "" {
		s.handleError(reply, code.InvalidParams, errorMsg)
	}
	return field
}

func (s *Service) processResult(reply string, result interface{}, err error) {
	if err != nil {
		s.handleError(reply, code.ERROR, err.Error())
		return
	}
	marshal, err := json.Marshal(result)
	if err != nil {
		s.handleError(reply, code.InvalidParams, err.Error())
		return
	}
	s.publish(reply, string(marshal), code.SUCCESS)

}
