package rqlclient

import (
	"encoding/json"
	"fmt"
	model "github.com/NubeDev/flexy/app/models"
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	githubdownloader "github.com/NubeDev/flexy/utils/gitdownloader"
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/NubeDev/flexy/utils/subjects"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

// Client struct to hold the NATS connection
type Client struct {
	globalUUID         string
	natsConn           *nats.Conn
	biosSubjectBuilder *subjects.SubjectBuilder
	gitDownloader      *githubdownloader.GitHubDownloader
	natsClient         natlib.NatLib
}

// New initializes a new Client
func New(natsURL, globalUUID string) (*Client, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}
	return &Client{
		globalUUID:         globalUUID,
		natsConn:           nc,
		biosSubjectBuilder: subjects.NewSubjectBuilder(globalUUID, "bios", subjects.IsBios),
		natsClient: natlib.New(natlib.NewOpts{
			EnableJetStream: true,
		}),
	}, nil
}

// sendNATSRequest is a reusable helper function to send a request to a NATS subject
// and unmarshal the response.
func (inst *Client) sendNATSRequest(clientUUID, script string, timeout time.Duration) (*nats.Msg, error) {
	subject := "host." + clientUUID + ".flex.rql"
	requestPayload := map[string]interface{}{
		"script": script,
	}
	reqData, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %v", err)
	}
	msg, err := inst.natsConn.Request(subject, reqData, timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return msg, nil
}

// RequestToApp allows sending a NATS request with a dynamic subject and JSON body
func (inst *Client) RequestToApp(appID, subject string, body []byte, timeout time.Duration) (*nats.Msg, error) {
	msg, err := inst.natsConn.Request(fmt.Sprintf("%s.%s", appID, subject), body, timeout)
	if err != nil {
		return nil, fmt.Errorf("NATS request to subject %s failed: %v", subject, err)
	}
	return msg, nil
}

// RequestWithSubject allows sending a NATS request with a dynamic subject and JSON body
func (inst *Client) RequestWithSubject(subject string, body []byte, timeout time.Duration) (*nats.Msg, error) {
	msg, err := inst.natsConn.Request(subject, body, timeout)
	if err != nil {
		return nil, fmt.Errorf("NATS request to subject %s failed: %v", subject, err)
	}
	return msg, nil
}

// Helper to build a request and handle the response
func (inst *Client) biosCommandRequest(body map[string]string, action, entity, op string, timeout time.Duration) (interface{}, error) {
	// Marshal the request body
	requestData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s command: %v", body, err)
	}

	// Build the subject
	subject := inst.biosSubjectBuilder.BuildSubject(action, entity, op)

	log.Info().Msgf("bios-command nats subject: %s", subject)
	// Send the request
	request, err := inst.natsConn.Request(subject, requestData, timeout)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	// Unmarshal and return the response
	var statusResp interface{}
	err = json.Unmarshal(request.Data, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("error: %v", string(request.Data))
	}

	return statusResp, nil
}

func (inst *Client) PingHostAllCore(timeout time.Duration) ([]natlib.Response, error) {
	data := []byte("ping")
	all, err := inst.natsClient.RequestAll("global.get.system.ping", data, timeout)
	if err != nil {
		return nil, err
	}
	var out []natlib.Response
	for _, msg := range all {
		var m natlib.Response
		err := json.Unmarshal(msg.Data, &m)
		if err == nil {
			out = append(out, m)
		}
	}
	return out, nil
}

func (inst *Client) ModuleHelp(clientUUID, moduleUUID string, args []string, timeout time.Duration) (interface{}, error) {
	request, err := inst.natsConn.Request("subject", []byte(""), timeout)
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

func (inst *Client) GetHosts(clientUUID string, timeout time.Duration) ([]*model.Host, error) {
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

func (inst *Client) GetHost(clientUUID, hostUUID string, timeout time.Duration) (*model.Host, error) {
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

func (inst *Client) DeleteHost(clientUUID, hostUUID string, timeout time.Duration) (interface{}, error) {
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
func (inst *Client) CreateHost(clientUUID string, host *hostService.Fields, timeout time.Duration) (*model.Host, error) {
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
