package rqlclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/NubeDev/flexy/utils/execute/commands"
	"github.com/nats-io/nats.go"
)

// RQLClient struct to hold the NATS connection
type RQLClient struct {
	natsClient *nats.Conn
}

// New initializes a new RQLClient
func New(natsURL string) (*RQLClient, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}
	return &RQLClient{natsClient: nc}, nil
}

// SystemdStatus sends a request to the NATS subject to get the systemd status of a unit
func (inst *RQLClient) SystemdStatus(clientUUID, unit string, timeout time.Duration) (*commands.StatusResp, error) {
	subject := "host." + clientUUID + ".rql"
	requestPayload := map[string]interface{}{
		"script": fmt.Sprintf("ctl.SystemdStatus(\"%s\")", unit),
	}
	reqData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}
	msg, err := inst.natsClient.Request(subject, reqData, timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	var statusResp *commands.StatusResp
	err = json.Unmarshal(msg.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return statusResp, nil
}
