package setting

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"time"
)

type App struct {
	JwtSecret        string `yaml:"jwtsecret"`
	PasswordSalt     string `yaml:"passwordsalt"`
	PrefixUrl        string `yaml:"prefixurl"`
	TimeFormat       string `yaml:"timeformat"`
	EnabledCORS      bool   `yaml:"enabledcors"`
	ExpireTimeFormat string `yaml:"expiretimeformat"`
}

type Server struct {
	RunMode      string        `yaml:"runmode"`
	HttpPort     int           `yaml:"httpport"`
	ReadTimeout  time.Duration `yaml:"readtimeout"`
	WriteTimeout time.Duration `yaml:"writetimeout"`
}

type Database struct {
	EchoSql     bool   `yaml:"echosql"`
	Type        string `yaml:"type"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Host        string `yaml:"host"`
	Name        string `yaml:"name"`
	TablePrefix string `yaml:"tableprefix"`
}

type Config struct {
	App      *App      `yaml:"app"`
	Server   *Server   `yaml:"server"`
	Database *Database `yaml:"database"`
}

var AppSetting = &App{}
var ServerSetting = &Server{}
var DatabaseSetting = &Database{}

// Setup initialize the configuration instance
func Setup() {
	var err error
	data, err := ioutil.ReadFile("config/app.yaml")
	if err != nil {
		log.Fatalf("setting.Setup, fail to read 'conf/app.yaml': %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("setting.Setup, fail to unmarshal yaml: %v", err)
	}

	// Assign the unmarshalled configurations to your global variables
	AppSetting = cfg.App
	ServerSetting = cfg.Server
	DatabaseSetting = cfg.Database

	// Convert time durations from seconds to time.Duration
	ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second
}
