package main

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/natsforwarder"
	"github.com/NubeDev/flexy/utils/subjects"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"time"
)

// natsForwarder
// will forward a message to ROS or a module
func (inst *Service) natsForwarder(globalUUID string, sourceNATS *nats.Conn, targetURL string) {
	// Connect to the NATS server for the source
	// Create a forwarder to forward requests to the target NATS server
	timeout := 5 * time.Second
	log.Info().Msgf("connected to NATS modules server: %s", targetURL)
	forwarder, err := natsforwarder.NewForwarder(targetURL, timeout)
	if err != nil {
		log.Fatal().Msgf("Error creating NATS forwarder: %v", err)
	}
	defer forwarder.Close()

	subjectModules := fmt.Sprintf("%s.proxy.>", globalUUID)
	// Set up a handler to forward requests to the target subject
	_, err = sourceNATS.QueueSubscribe(subjectModules, "rql_queue", func(m *nats.Msg) {
		forwardSubject := subjects.GetSubjectParts(m.Subject)
		if err != nil {
			log.Info().Msgf("failed to get app-id from subject: %s", m.Subject)
		}
		log.Info().Msgf("module foward message subject: %s", forwardSubject)
		log.Info().Msgf("module foward message data: %s", string(m.Data))
		err := forwarder.ForwardRequest(m, forwardSubject)
		if err != nil {
			log.Error().Msgf("NATS forwarder failed to foward message: %v", err)
			return
		}
	})
	if err != nil {
		log.Fatal().Msgf("Error creating NATS forwarder: %v", err)
		return
	}

	// Keep the program running
	select {}
}
