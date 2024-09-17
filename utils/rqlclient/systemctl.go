package rqlclient

import (
	"fmt"
	"time"
)

func (inst *Client) BiosSystemdCommandGet(serviceName, action, property string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "property": property}
	return inst.biosCommandRequest(body, "get", "system", fmt.Sprintf("systemctl.%s", action), timeout)
}

func (inst *Client) BiosSystemdCommandPost(serviceName, action, property string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "property": property}
	return inst.biosCommandRequest(body, "post", "system", fmt.Sprintf("systemctl.%s", action), timeout)
}
