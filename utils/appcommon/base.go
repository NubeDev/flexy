package appcommon

import (
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type App struct {
	AppID    string
	Config   *viper.Viper
	NatsConn *nats.Conn
	RootCmd  *cobra.Command
}

func NewApp() *App {
	app := &App{
		Config:  viper.New(),
		RootCmd: &cobra.Command{},
	}
	return app
}
