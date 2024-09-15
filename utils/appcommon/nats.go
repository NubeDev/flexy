package appcommon

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"time"
)

func (app *App) InitNATS(natsURL string) error {
	var err error
	app.NatsConn, err = nats.Connect(natsURL)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) Publish(operation string, subject string, msg []byte) error {
	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
	return app.NatsConn.Publish(topic, msg)
}

func (app *App) Subscribe(operation string, subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
	return app.NatsConn.Subscribe(topic, handler)
}

// Respond sets up a subscription that responds to requests
func (app *App) Respond(operation string, subject string, handler func(msg *nats.Msg) []byte) (*nats.Subscription, error) {
	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
	return app.NatsConn.Subscribe(topic, func(msg *nats.Msg) {
		// Call the handler to generate a response
		response := handler(msg)
		// Reply with the response data
		msg.Respond(response)
	})
}

// Request sends a request message and waits for a response
func (app *App) Request(operation string, subject string, msg []byte, timeout time.Duration) (*nats.Msg, error) {
	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
	return app.NatsConn.Request(topic, msg, timeout)
}
