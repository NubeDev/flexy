package natsapis

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"log"
)

type RequestBody struct {
	Script string `json:"script"`
}

type Response struct {
	Error bool        `json:"error"`
	Data  interface{} `json:"data"`
}

func RQLHandler() func(m *nats.Msg) {
	return func(m *nats.Msg) {
		var reqBody RequestBody
		err := json.Unmarshal(m.Data, &reqBody)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			response := Response{Error: true, Data: "Error unmarshalling message"}
			respBytes, _ := json.Marshal(response)
			m.Respond(respBytes)
			return
		}

		responseData, err := rqlHandler(reqBody.Script)
		if err != nil {
			log.Printf("Error processing request: %v", err)
			response := Response{Error: true, Data: err.Error()}
			respBytes, _ := json.Marshal(response)
			m.Respond(respBytes)
			return
		}
		m.Respond(responseData)
	}
}
