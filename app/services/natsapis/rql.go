package natsapis

import (
	"encoding/json"
	"github.com/NubeDev/flexy/app/services/rqlservice"
	"github.com/NubeDev/flexy/utils/helpers"
)

func rqlHandler(body string) func() ([]byte, error) {
	return func() ([]byte, error) {
		destroy, err := rqlservice.RQL().RunAndDestroy(helpers.UUID(), body, nil)
		if err != nil {
			return nil, err
		}
		return json.Marshal(destroy)
	}

}
