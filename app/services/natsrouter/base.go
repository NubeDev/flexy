package natsrouter

import (
	"github.com/nats-io/nats.go"
	"log"
)

type NatsRouter struct {
	nc               *nats.Conn
	JetStreamContext nats.JetStreamContext
}

// NatsHandlerFunc is the handler type for NATS messages.
type NatsHandlerFunc func(*nats.Msg)

// New creates a new NatsRouter
func New(nc *nats.Conn) *NatsRouter {
	// Initialize JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Error initializing JetStream: %v", err)
	}

	return &NatsRouter{nc: nc, JetStreamContext: js}
}

// Handle registers a handler for a NATS subject
func (r *NatsRouter) Handle(subject string, handler NatsHandlerFunc) {
	r.nc.Subscribe(subject, nats.MsgHandler(handler))
}

// QueueHandle registers a handler for a NATS subject with a queue group
func (r *NatsRouter) QueueHandle(subject string, queue string, handler NatsHandlerFunc) {
	r.nc.QueueSubscribe(subject, queue, nats.MsgHandler(handler))
}

// Publish publishes a message on a NATS subject
func (r *NatsRouter) Publish(subject string, reply string, data []byte) {
	err := r.nc.Publish(subject, data)
	if err != nil {
		log.Println("Error publishing message:", err)
	}
}

func PingHandler(uuid string) func(m *nats.Msg) {
	return func(m *nats.Msg) {
		log.Printf("Ping received for UUID %s, replying with pong", uuid)
		m.Respond([]byte("pong"))
	}
}
