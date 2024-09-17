package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func (s *Service) InitConfig(configFile string) error {
	if configFile != "" {
		s.Config.SetConfigFile(configFile)
	} else {
		s.Config.AddConfigPath(".")
		s.Config.SetConfigName("config")
	}
	s.Config.SetConfigType("yaml")
	if err := s.Config.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func (s *Service) SetupCLI() {
	s.RootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is ./config.yaml)")
	s.RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		if err := s.InitConfig(configFile); err != nil {
			return err
		}

		s.globalUUID = s.Config.GetString("id")
		if s.globalUUID == "" {
			return fmt.Errorf("id is not set in the configuration")
		}

		natsURL := s.Config.GetString("nats_url")
		if natsURL == "" {
			natsURL = nats.DefaultURL
		}

		// Initialize options from the configuration
		opts := &Opts{
			GlobalUUID:      s.globalUUID,
			NatsURL:         natsURL,
			RootPath:        s.Config.GetString("root_path"),
			AppsPath:        fmt.Sprintf("%s/%s", s.Config.GetString("root_path"), s.Config.GetString("apps_path")),
			SystemPath:      s.Config.GetString("system_path"),
			GitToken:        s.Config.GetString("git_token"),
			GitDownloadPath: s.Config.GetString("git_download_path"),
			ProxyNatsPort:   s.Config.GetInt("proxy_port"),
		}

		// Initialize services using NewService
		if err := s.NewService(opts); err != nil {
			return fmt.Errorf("failed to initialize services: %v", err)
		}

		// Retrieve services from the configuration
		s.services = s.Config.GetStringSlice("services")
		s.description = s.Config.GetString("description")

		return nil
	}

	s.RootCmd.Run = func(cmd *cobra.Command, args []string) {
		// Start the service
		if err := s.StartService(); err != nil {
			log.Fatal().Msgf("Failed to start service: %s", err)
		}
	}
}
