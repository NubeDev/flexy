package rqlclient

import (
	"encoding/json"
	"fmt"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

type StoreRequest struct {
	Action          string `json:"action"`
	StoreName       string `json:"storeName,omitempty"`
	ObjectName      string `json:"objectName,omitempty"`
	DestinationPath string `json:"destinationPath,omitempty"`
	Data            string `json:"data,omitempty"`
}

func (inst *Client) storeCommandRequest(body map[string]interface{}, action string, timeout time.Duration) (*natlib.Response, error) {
	// Marshal the request body
	requestData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Build the subject
	subject := fmt.Sprintf("%s.post.system.store.%s", inst.globalUUID, action)
	log.Info().Msgf("store-command NATS subject: %s", subject)

	// Send the request
	request, err := inst.natsConn.Request(subject, requestData, timeout)
	if err != nil {
		return nil, fmt.Errorf("NATS request failed: %v", err)
	}

	// Unmarshal the response into natlib.Response
	var response natlib.Response
	err = json.Unmarshal(request.Data, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Check if the response code indicates an error
	if response.Code != code.SUCCESS {
		return nil, fmt.Errorf("error from server: %s", response.Payload)
	}

	return &response, nil
}

func (inst *Client) GetStores(timeout time.Duration) ([]string, error) {
	body := map[string]interface{}{
		"action": "get.stores",
	}

	response, err := inst.storeCommandRequest(body, "get.stores", timeout)
	if err != nil {
		return nil, err
	}

	// The payload should contain the list of stores as JSON string
	var stores []string
	err = json.Unmarshal([]byte(response.Payload), &stores)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response payload: %v", err)
	}

	return stores, nil
}

func (inst *Client) GetStoreObjects(storeName string, timeout time.Duration) ([]*nats.ObjectInfo, error) {
	body := map[string]interface{}{
		"action":    "get.object",
		"storeName": storeName,
	}

	response, err := inst.storeCommandRequest(body, "get.object", timeout)
	if err != nil {
		return nil, err
	}

	// The payload should contain the list of object info
	var objects []*nats.ObjectInfo
	err = json.Unmarshal([]byte(response.Payload), &objects)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response payload: %v", err)
	}

	return objects, nil
}

func (inst *Client) AddObject(storeName, objectName, path string, overwriteIfExisting bool) (string, error) {
	err := inst.natsClient.NewObject(storeName, objectName, path, overwriteIfExisting)
	if err != nil {
		return "", err
	}
	return "uploaded to server ok", nil
}

func (inst *Client) DeleteObject(storeName, objectName string, timeout time.Duration) (*natlib.Response, error) {
	body := map[string]interface{}{
		"action":     "delete.object",
		"storeName":  storeName,
		"objectName": objectName,
	}

	response, err := inst.storeCommandRequest(body, "delete.object", timeout)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (inst *Client) DownloadObject(storeName, objectName, destinationPath string, timeout time.Duration) (*natlib.Response, error) {
	body := map[string]interface{}{
		"action":          "download.object",
		"storeName":       storeName,
		"objectName":      objectName,
		"destinationPath": destinationPath,
	}

	response, err := inst.storeCommandRequest(body, "download.object", timeout)
	if err != nil {
		return response, err
	}

	return response, nil
}
