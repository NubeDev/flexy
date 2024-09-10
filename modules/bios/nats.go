package main

import (
	"fmt"
	"github.com/NubeDev/flexy/utils/natsforwarder"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

/*
 */

// StartService starts the NATS subscription and listens for commands
// these are the bios commands. eg; install and app
func (s *Service) StartService(globalUUID string) {
	s.natsConn.Subscribe(fmt.Sprintf("bios.%s.command", globalUUID), s.HandleCommand)
	s.natsConn.Subscribe(fmt.Sprintf("bios.%s.store", globalUUID), s.HandleStore)
}

// natsForwarder
// will forward a message to ROS or a module
func natsForwarder(uuid string, sourceNATS *nats.Conn, targetURL string) {
	// Connect to the NATS server for the source
	// Create a forwarder to forward requests to the target NATS server
	timeout := 5 * time.Second
	log.Info().Msgf("connected to NATS modules server: %s", targetURL)
	forwarder, err := natsforwarder.NewForwarder(targetURL, timeout)
	if err != nil {
		log.Fatal().Msgf("Error creating NATS forwarder: %v", err)
	}
	defer forwarder.Close()

	/*
		----------- PROXY TO FLEX/ROS SERVER ------------------
	*/

	subject := fmt.Sprintf("host.%s.flex.*", uuid)
	// Set up a handler to forward requests to the target subject
	sourceNATS.QueueSubscribe(subject, "rql_queue", func(m *nats.Msg) {
		// Split the subject based on "modules."
		parts := strings.Split(m.Subject, ".")
		fmt.Println(parts, len(parts))
		if len(parts) != 4 {
			fmt.Println("No part found after 'modules.'")
			return
		}
		nextPart := parts[3]
		targetSubject := fmt.Sprintf("host.%s.flex.%s", uuid, nextPart)
		log.Info().Msgf("module foward message subject: %s", targetSubject)
		log.Info().Msgf("module foward message data: %s", string(m.Data))
		forwarder.ForwardRequest(m, targetSubject)
	})

	/*
		----------- PROXY TO MODULES ------------------
	*/
	// hosts.abc.modules
	subjectModules := fmt.Sprintf("host.%s.modules.*", uuid)

	// Set up a handler to forward requests to the target subject
	sourceNATS.QueueSubscribe(subjectModules, "rql_queue", func(m *nats.Msg) {
		// Split the subject based on "modules."
		parts := strings.SplitN(m.Subject, "modules.", 2)
		fmt.Println(parts)
		if len(parts) < 2 {
			fmt.Println("No part found after 'modules.'")
			return
		}
		moduleUUID := parts[1]
		// host.<GLOBAL-UUID>.modules.<MODULE-UUID>
		targetSubject := fmt.Sprintf("module.%s.command", moduleUUID)
		log.Info().Msgf("module foward message subject: %s", targetSubject)
		log.Info().Msgf("module foward message data: %s", string(m.Data))
		forwarder.ForwardRequest(m, targetSubject)
	})

	// Keep the program running
	select {}
}
