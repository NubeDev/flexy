package natsforwarder

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

// Forwarder manages the connection to the target NATS server
type Forwarder struct {
	natsClient *nats.Conn
	timeout    time.Duration
}

// NewForwarder creates a new NATS forwarder with the given target server URL and timeout
func NewForwarder(url string, timeout time.Duration) (*Forwarder, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %v", err)
	}
	log.Info().Msgf("start nats proxy on url: %s", url)
	return &Forwarder{
		natsClient: nc,
		timeout:    timeout,
	}, nil
}

// Close closes the connection to the NATS server
func (f *Forwarder) Close() {
	if f.natsClient != nil {
		f.natsClient.Close()
	}
}

// ForwardRequest forwards the incoming NATS message to the target server and responds back
func (f *Forwarder) ForwardRequest(m *nats.Msg, targetSubject string) error {
	// Forward the message to the target NATS server and wait for the response
	msg, err := f.natsClient.Request(targetSubject, m.Data, f.timeout)
	if err != nil {
		log.Printf("Error forwarding request to target NATS server: %v", err)
		m.Respond([]byte(fmt.Sprintf("Error forwarding request: %v", err)))
		return err
	}
	// Respond with the message received from the forwarded request
	err = m.Respond(msg.Data)
	return err
}
