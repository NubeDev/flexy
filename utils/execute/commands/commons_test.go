package commands

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	var err error

	resp, err := c.SystemdStatus("mosquitto")
	pprint.PrintJSON(resp)
	fmt.Println(err)

	err = c.SystemdCommand("mosquitto", "status")
	fmt.Println(err)

	//s, err := c.SystemdShow("mosquitto", "NRestarts")
	//fmt.Println(s)
	//fmt.Println(err)

	r := c.Run(&CommandBody{
		Command: "pwd",
		Args:    nil,
		Timeout: 0,
	})

	pprint.PrintJSON(r)
}
