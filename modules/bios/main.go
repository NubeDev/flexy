package main

import (
	"fmt"
	"github.com/NubeDev/flexy/app/startup"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

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

func (s *Service) Run() {
	if err := s.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Block the main thread
	select {}
}
