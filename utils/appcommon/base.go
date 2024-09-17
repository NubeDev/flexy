package appcommon

import (
	"github.com/NubeDev/flexy/utils/natlib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type App struct {
	AppID       string
	Description string
	NatsConn    natlib.NatLib
	Config      *viper.Viper
	RootCmd     *cobra.Command
}

type Opts struct {
	NatsConn natlib.NatLib
}

func NewApp() *App {
	app := &App{
		Config:  viper.New(),
		RootCmd: &cobra.Command{},
	}
	return app
}
