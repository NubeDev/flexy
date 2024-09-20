package main

import (
	"fmt"
	"github.com/NubeDev/flexy/app/startup"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

const responseTypeKey = "Debug"

type Message struct {
	Message string `json:"message"`
}

func main() {
	startup.BootLogger()

	service := &Service{
		Config: viper.New(),
		RootCmd: &cobra.Command{
			Use:   "myapp",
			Short: "My Application",
		},
	}

	service.SetupCLI()
	service.Run()
}

func (s *Service) InitNATS(natsURL string) error {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return err
	}
	s.natsConn = nc
	return nil
}

func (s *Service) StartGinServer() {
	r := s.SetupGin() // Setup Gin server with routes

	port := s.Config.GetString("gin_port")
	if port == "" {
		port = "8080" // Default to port 8080 if not specified
	}
	// Start Gin server
	go func() {
		if err := r.Run(":" + port); err != nil {
			fmt.Printf("Failed to start Gin server: %v\n", err)
			os.Exit(1)
		}
	}()
}

func (s *Service) Run() {
	if err := s.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start the Gin server after initializing everything else
	s.StartGinServer()

	// Block the main thread
	select {}
}
