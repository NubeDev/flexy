package main

import (
	"context"
	"fmt"
	"github.com/NubeDev/flexy/app/middleware"
	models "github.com/NubeDev/flexy/app/models"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/routers"
	"github.com/NubeDev/flexy/utils/casbin"
	"github.com/NubeDev/flexy/utils/setting"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var globalUUID string
var port int
var useAuth bool

func main() {
	cli()

	setting.Setup(useAuth)
	models.Setup()
	common.InitValidate()
	// Initialize route permissions. The purpose of this initialization is to avoid querying the database for route permissions on every access.
	// If you change route permissions, you need to call this method again.
	casbin.SetupCasbin()

	gin.SetMode(setting.ServerSetting.RunMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	if setting.AppSetting.EnabledCORS {
		r.Use(middleware.CORS())
	}

	routersInit := routers.InitRouter(r)

	readTimeout := setting.ServerSetting.ReadTimeout
	writeTimeout := setting.ServerSetting.WriteTimeout
	if port == 0 {
		port = setting.ServerSetting.HttpPort
	}
	endPoint := fmt.Sprintf(":%d", port)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	// Initialize NATS connection
	nc, err := SetupNATS()
	if err != nil {
		log.Fatal().Msgf("Error setting up NATS: %v", err)
	}
	defer nc.Close()
	natsRouter := natsrouter.New(nc)

	go bootNats(globalUUID, natsRouter)

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		PartsOrder: []string{
			"time",
			"level",
			"caller",
			"message",
		},
		FormatLevel: func(i interface{}) string {
			var levelColor int
			switch i {
			case zerolog.LevelTraceValue:
				levelColor = 36 // Cyan
			case zerolog.LevelDebugValue:
				levelColor = 90 // Dark grey
			case zerolog.LevelInfoValue:
				levelColor = 32 // Green
			case zerolog.LevelWarnValue:
				levelColor = 33 // Yellow
			case zerolog.LevelErrorValue:
				levelColor = 31 // Red
			case zerolog.LevelFatalValue:
				levelColor = 35 // Magenta
			case zerolog.LevelPanicValue:
				levelColor = 95 // Bright magenta
			default:
				levelColor = 37 // White
			}
			return fmt.Sprintf("\x1b[%dm%s\x1b[0m", levelColor, strings.ToUpper(fmt.Sprintf("%-6s", i)))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("\x1b[34m%s:\x1b[0m", i) // Field names in blue
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			// Extract the function and line number only
			parts := strings.Split(fmt.Sprintf("%s", i), "/")
			return parts[len(parts)-1]
		},
	}
	// Set global logger with above configuration
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	debugMode := true
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal().Msgf("server: %v", err)
		}
	}()

	log.Info().Msgf("start http server listening %s", endPoint)
	log.Info().Msgf("Actual pid is %d", os.Getpid())

	// Wait for an interrupt signal to gracefully shut down the server (with a 5-second timeout)
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server Shutdown: %v", err)
	}

	log.Info().Msg("Server exiting")
}

func bootNats(uuid string, natsRouter *natsrouter.NatsRouter) {
	log.Printf("Starting edge device with UUID: %s", uuid)
	// Register NATS routes
	natsRouter.Handle("host."+uuid+".server", natsrouter.ServerHandler())
	natsRouter.Handle("host."+uuid+".ping", natsrouter.PingHandler(uuid))
	// Keep the edge device running indefinitely
	select {}
}

func cli() {
	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "A brief description of your application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("UUID:", globalUUID)
			fmt.Println("Port:", port)
			fmt.Println("Disable Auth:", useAuth)
		},
	}

	rootCmd.Flags().StringVar(&globalUUID, "uuid", "", "UUID for the edge device")
	rootCmd.Flags().IntVar(&port, "port", 0, "HTTP server port")
	rootCmd.Flags().BoolVar(&useAuth, "auth", true, "use auth")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
	}

}

func SetupNATS() (*nats.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal().Msgf("Error connecting to NATS: %v", err)
		return nil, err
	}
	log.Info().Msgf("Connected to NATS")
	return nc, nil
}
