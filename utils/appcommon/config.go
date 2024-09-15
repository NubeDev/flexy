package appcommon

func (app *App) InitConfig(configFile string) error {
	if configFile != "" {
		app.Config.SetConfigFile(configFile)
	} else {
		app.Config.AddConfigPath(".")
		app.Config.SetConfigName("config")
	}
	app.Config.SetConfigType("yaml")
	if err := app.Config.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
