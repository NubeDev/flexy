package rqlclient

import (
	"fmt"
	"time"
)

func (inst *Client) BiosSystemdCommandGet(serviceName, action, property, appID, version string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "property": property, "appID": appID, "version": version, "action": action}
	if appID != "" {
		return inst.biosCommandRequest(body, "get", "apps", fmt.Sprintf("manager.%s", "systemctl"), timeout)
	}
	return inst.biosCommandRequest(body, "get", "system", fmt.Sprintf("systemctl.%s", action), timeout)
}

func (inst *Client) BiosSystemdCommandPost(serviceName, action, property, appID, version string, timeout time.Duration) (interface{}, error) {
	body := map[string]string{"name": serviceName, "property": property, "appID": appID, "version": version, "action": action}
	return inst.biosCommandRequest(body, "post", "system", fmt.Sprintf("systemctl.%s", action), timeout)
}
