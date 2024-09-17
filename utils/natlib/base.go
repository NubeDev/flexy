package natlib

import (
	"encoding/json"
	"errors"
	"github.com/NubeDev/flexy/utils/code"
	"github.com/nats-io/nats.go"
	"time"
)

// Opts represents options for NATS operations.
type Opts struct {
	Timeout int // seconds
}

// Subjects represents a NATS subject with its type.
type Subjects struct {
	Type    string // "Subscribe" or "SubscribeWithRespond"
	Subject string
}

// NatLib is the common interface with NATS methods.
type NatLib interface {
	Publish(subj string, data []byte, cb nats.MsgHandler, opts *Opts) error
	Subscribe(subj string, cb nats.MsgHandler, opts *Opts) error
	SubscribeWithRespond(subj string, handler func(msg *nats.Msg) ([]byte, error), opts *Opts) error
	RequestAll(subj string, data []byte, timeout time.Duration) ([]*nats.Msg, error)
	Close()
}

var uuidName = "Global-UUID"

// natsLib implements the NatLib interface.
type natsLib struct {
	nc         *nats.Conn
	subjects   []*Subjects
	globalUUID string
}

type NewOpts struct {
	URL        string
	GlobalUUID string
	NatsConn   *nats.Conn
}

// New creates a new instance of NatLib.
func New(opts NewOpts) NatLib {
	var url = opts.URL
	if url == "" {
		url = nats.DefaultURL
	}
	var err error
	var nc *nats.Conn
	if opts.NatsConn == nil {
		nc, err = nats.Connect(url)
		if err != nil {
			panic(err)
		}
	} else {
		nc = opts.NatsConn
	}
	return &natsLib{
		nc:       nc,
		subjects: []*Subjects{},
	}
}

// Close will close the connection to the server.
func (nl *natsLib) Close() {
	nl.nc.Close()
}

// Publish sends a request and waits for a response.
func (nl *natsLib) Publish(subj string, data []byte, cb nats.MsgHandler, opts *Opts) error {
	timeout := 5 * time.Second
	if opts != nil && opts.Timeout > 0 {
		timeout = time.Duration(opts.Timeout) * time.Second
	}
	msg, err := nl.nc.Request(subj, data, timeout)
	if err != nil {
		return err
	}
	if nl.globalUUID != "" {
		msg.Header.Set(uuidName, nl.globalUUID)
	}
	cb(msg)
	return nil
}

// Subscribe listens for messages on a subject.
func (nl *natsLib) Subscribe(subj string, cb nats.MsgHandler, opts *Opts) error {
	nl.subjects = append(nl.subjects, &Subjects{
		Type:    "subscribe",
		Subject: subj,
	})
	_, err := nl.nc.Subscribe(subj, cb)
	return err
}

// SubscribeWithRespond subscribes and responds to incoming messages.
func (nl *natsLib) SubscribeWithRespond(subj string, handler func(msg *nats.Msg) ([]byte, error), opts *Opts) error {
	nl.subjects = append(nl.subjects, &Subjects{
		Type:    "subscribeWithRespond",
		Subject: subj,
	})

	_, err := nl.nc.Subscribe(subj, func(msg *nats.Msg) {
		responseData, err := handler(msg)
		if err != nil {
			// Handle error appropriately
			return
		}
		if nl.globalUUID != "" {
			msg.Header.Set(uuidName, nl.globalUUID)
		}
		if err := msg.Respond(responseData); err != nil {
			// Handle error appropriately
		}
	})
	return err
}

// RequestAll sends a request to the subject and collects all responses
// received within the specified timeout duration.
func (nl *natsLib) RequestAll(subj string, data []byte, timeout time.Duration) ([]*nats.Msg, error) {
	// Create a unique inbox
	inbox := nats.NewInbox()
	sub, err := nl.nc.SubscribeSync(inbox)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	// Publish the message with the reply set to the inbox
	m := &nats.Msg{
		Subject: subj,
		Reply:   inbox,
		Data:    data,
	}
	if nl.globalUUID != "" {
		m.Header.Set(uuidName, nl.globalUUID)
	}
	err = nl.nc.PublishMsg(m)
	if err != nil {
		return nil, err
	}

	// Buffer to store the responses
	var responses []*nats.Msg

	// Collect responses until timeout
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		msg, err := sub.NextMsg(remaining)
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				// Timeout reached, exit loop
				break
			}
			return nil, err
		}
		responses = append(responses, msg)
	}

	return responses, nil
}

type Response struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Payload     string `json:"payload"`
	Description string `json:"description,omitempty"`
}

type Args struct {
	Description string `json:"description"`
}

func NewResponse(messageCode int, payload string, args ...Args) *Response {
	resp := &Response{
		Code:    messageCode,
		Message: code.GetMsg(messageCode),
		Payload: payload,
	}
	if len(args) > 0 {
		resp.Description = args[0].Description
	}
	return resp
}

func (p *Response) ToJSONError() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Response) ToJSON() []byte {
	jsonData, _ := json.Marshal(p)
	return jsonData
}

func (p *Response) ToJSONString() (string, error) {
	jsonData, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
