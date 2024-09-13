package rqlclient

import (
	"encoding/json"
	"fmt"
	model "github.com/NubeDev/flexy/app/models"
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	"log"
	"sync"
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

// sendNATSRequest is a reusable helper function to send a request to a NATS subject
// and unmarshal the response.
func (inst *RQLClient) sendNATSRequest(clientUUID, script string, timeout time.Duration) (*nats.Msg, error) {
	subject := "host." + clientUUID + ".flex.rql"
	requestPayload := map[string]interface{}{
		"script": script,
	}
	reqData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}
	msg, err := inst.natsClient.Request(subject, reqData, timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return msg, nil
}

// PingHostAllCore pings all hosts and collects responses from multiple clients.
func (inst *RQLClient) PingHostAllCore() ([]string, error) {
	responseChan := make(chan string, 10) // buffer of 10, can be adjusted
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to prevent race conditions

	// Subscribe to the shared response subject (module.global.response)
	sub, err := inst.natsClient.SubscribeSync("module.global.response")
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	// Publish the request to the shared subject (module.global)
	err = inst.natsClient.Publish("module.global.ping", []byte("hello"))
	if err != nil {
		return nil, err
	}

	// Set a timeout for collecting responses
	timeout := time.NewTimer(5 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-timeout.C:
				return
			default:
				// Wait for the next message with a short timeout
				msg, err := sub.NextMsg(1 * time.Second)
				if err == nats.ErrTimeout {
					continue
				}
				if err != nil {
					log.Println("Error receiving message:", err)
					return
				}
				if msg != nil && len(msg.Data) > 0 {
					// Lock access to the response slice to prevent race condition
					mu.Lock()
					responseChan <- string(msg.Data)
					mu.Unlock()
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Collect the responses
	var responses []string
	for res := range responseChan {
		if res != "" {
			mu.Lock() // Lock the slice while appending to prevent race condition
			responses = append(responses, res)
			mu.Unlock()
		}
	}

	return responses, nil
}

func (inst *RQLClient) ModuleHelp(clientUUID, moduleUUID string, args []string, timeout time.Duration) (interface{}, error) {
	request, err := inst.natsClient.Request("subject", []byte(""), timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	var statusResp interface{}
	err = json.Unmarshal(request.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("error: %v", string(request.Data))
	}
	return statusResp, err
}

// SystemdStatus sends a request to get the systemd status of a unit
func (inst *RQLClient) SystemdStatus(clientUUID, unit string, timeout time.Duration) (*commands.StatusResp, error) {
	script := fmt.Sprintf("ctl.SystemdStatus(\"%s\")", unit)
	request, err := inst.sendNATSRequest(clientUUID, script, timeout)
	if err != nil {
		return nil, err
	}
	var statusResp *commands.StatusResp
	err = json.Unmarshal(request.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return statusResp, err
}

func (inst *RQLClient) GetHosts(clientUUID string, timeout time.Duration) ([]*model.Host, error) {
	script := fmt.Sprintf("hosts.GetHosts()")
	request, err := inst.sendNATSRequest(clientUUID, script, timeout)
	if err != nil {
		return nil, err
	}
	var statusResp []*model.Host
	err = json.Unmarshal(request.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("error: %v", string(request.Data))
	}
	return statusResp, err
}

func (inst *RQLClient) GetHost(clientUUID, hostUUID string, timeout time.Duration) (*model.Host, error) {
	script := fmt.Sprintf("hosts.GetHost(\"%s\")", hostUUID)
	request, err := inst.sendNATSRequest(clientUUID, script, timeout)
	if err != nil {
		return nil, err
	}
	var statusResp *model.Host
	err = json.Unmarshal(request.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	return statusResp, err
}

func (inst *RQLClient) DeleteHost(clientUUID, hostUUID string, timeout time.Duration) (interface{}, error) {
	script := fmt.Sprintf("hosts.Delete(\"%s\")", hostUUID)
	request, err := inst.sendNATSRequest(clientUUID, script, timeout)
	if err != nil {
		return nil, err
	}
	var statusResp interface{}
	err = json.Unmarshal(request.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("error: %v", string(request.Data))
	}
	return statusResp, err
}

// CreateHost creates a host by passing hostService.Fields into the JavaScript script
func (inst *RQLClient) CreateHost(clientUUID string, host *hostService.Fields, timeout time.Duration) (*model.Host, error) {
	// Prepare the script with host details injected into it
	script := fmt.Sprintf(`
		let host = {
			Name: "%s",
			Ip: "%s"
		};
		hosts.Create(host);
	`, host.Name, host.Ip)

	// Send the request with the script
	request, err := inst.sendNATSRequest(clientUUID, script, timeout)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response into a Host struct
	var createdHost *model.Host
	err = json.Unmarshal(request.Data, &createdHost)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return createdHost, nil
}
