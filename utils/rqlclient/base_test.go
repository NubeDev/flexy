package rqlclient

import (
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"testing"
)

func TestNew(t *testing.T) {
	client, err := New("nats://127.0.0.1:4223", "")
	if err != nil {
		return
	}
	//status, err := client.SystemdStatus("abc", "mosquitto", time.Second)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//pprint.PrintJSON(status)
	core, err := client.PingHostAllCore()
	if err != nil {
		return
	}
	pprint.PrintJSON(core)
}
