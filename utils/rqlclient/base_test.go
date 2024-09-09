package rqlclient

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"github.com/nats-io/nats.go"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client, err := New(nats.DefaultURL)
	if err != nil {
		return
	}
	status, err := client.SystemdStatus("abc", "mosquitto", time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	pprint.PrintJSON(status)
}
