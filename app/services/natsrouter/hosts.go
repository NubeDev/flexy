package natsrouter

import (
	"encoding/json"
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
)

func hostHandler(endpoint, method, body string) func() ([]byte, error) {
	switch endpoint {
	case "hosts":
		if method == "GET" {
			//return func() ([]byte, error) {
			//	allHosts := hostService.Create()
			//	return json.Marshal(allHosts)
			//}
		}
	case "addHost":
		if method == "POST" {
			return func() ([]byte, error) {
				var b *hostService.Fields
				err := json.Unmarshal([]byte(body), &b)
				if err != nil {
					return nil, err
				}
				create, err := hostService.Create(b)
				if err != nil {
					return nil, err
				}
				return json.Marshal(create)
			}
		}
	}
	return nil
}
