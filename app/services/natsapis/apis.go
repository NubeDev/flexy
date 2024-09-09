package natsapis

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"log"
)

type RequestBody struct {
	Script string `json:"script"`
}

func RQLHandler() func(m *nats.Msg) {
	return func(m *nats.Msg) {
		var reqBody RequestBody
		err := json.Unmarshal(m.Data, &reqBody)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			m.Respond([]byte("Error unmarshalling message"))
			return
		}

		handlerFunc := rqlHandler(reqBody.Script)

		if handlerFunc == nil {
			m.Respond([]byte("Unknown endpoint or method"))
			return
		}

		response, err := handlerFunc()
		if err != nil {
			log.Printf("Error processing request: %v", err)
			m.Respond([]byte("Error processing request"))
			return
		}

		m.Respond(response)
	}
}
