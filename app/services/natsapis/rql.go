package natsapis

import (
	"encoding/json"
	"github.com/NubeDev/flexy/app/services/rqlservice"
	"github.com/NubeDev/flexy/utils/helpers"
)

func rqlHandler(endpoint, method, body string) func() ([]byte, error) {
	if endpoint == "rql" {
		return func() ([]byte, error) {
			destroy, err := rqlservice.RQL().RunAndDestroy(helpers.UUID(), body, nil)
			if err != nil {
				return nil, err
			}
			return json.Marshal(destroy)
		}
	}
	return nil
}
