package rqlclient

import (
	"fmt"
	"time"
)

func (inst *Client) BiosSystemdCommandGet(serviceName, action, property, appID string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "property": property, "appID": appID, "action": action}
	if appID != "" {
		return inst.biosCommandRequest(body, "get", "apps", fmt.Sprintf("manager.%s", "systemctl"), timeout)
	}
	return inst.biosCommandRequest(body, "get", "system", fmt.Sprintf("systemctl.%s", action), timeout)
}

func (inst *Client) BiosSystemdCommandPost(serviceName, action, property, appID string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "property": property, "appID": appID}
	return inst.biosCommandRequest(body, "post", "system", fmt.Sprintf("systemctl.%s", action), timeout)
}
