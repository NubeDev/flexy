package natlib

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"testing"
)

func TestSub2(t *testing.T) {
	nl := New(NewOpts{})

	nl.SubscribeWithRespond("test", func(msg *nats.Msg) ([]byte, error) {
		fmt.Println("Received ping request")
		responseData := []byte("pong from 2")
		return responseData, nil
	}, nil)
	select {}
}
