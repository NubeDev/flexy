package main

import (
	"context"
	"fmt"
	"github.com/NubeDev/flexy/app/middleware"
	models "github.com/NubeDev/flexy/app/models"
	"github.com/NubeDev/flexy/app/services/natsapis"
	"github.com/NubeDev/flexy/app/services/natsrouter"
	"github.com/NubeDev/flexy/app/services/rqlservice"
	"github.com/NubeDev/flexy/app/startup"
	"github.com/NubeDev/flexy/common"
	"github.com/NubeDev/flexy/routers"
	"github.com/NubeDev/flexy/utils/casbin"
	"github.com/NubeDev/flexy/utils/setting"
	"github.com/NubeDev/flexy/utils/subjects"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var globalUUID string
var appID = "ros"
var port int
var natsModulePort int
var useAuth bool

func main() {

	cli()
	startup.BootLogger()

	setting.Setup(useAuth)
	models.Setup()
	common.InitValidate()
	startup.InitServices()
	// Initialize route permissions. The purpose of this initialization is to avoid querying the database for route permissions on every access.
	// If you change route permissions, you need to call this method again.
	casbin.SetupCasbin()
	rqlservice.BootRQL()
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
	natsURL := fmt.Sprintf("nats://127.0.0.1:%d", natsModulePort)
	// Initialize NATS cloud connection
	natsCloud, err := setupNATS(natsURL)
	if err != nil {
		log.Fatal().Msgf("error setting up NATS cloud: %v", err)
	}
	defer natsCloud.Close()
	natsRouterCloud := natsrouter.New(natsCloud)

	go bootNatsCloud(globalUUID, natsRouterCloud)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal().Msgf("server: %v", err)
		}
	}()

	log.Info().Msgf("start http server listening %s", endPoint)
	log.Info().Msgf("actual pid is %d", os.Getpid())

	// Wait for an interrupt signal to gracefully shut down the server (with a 5-second timeout)
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("server shutdown: %v", err)
	}

	log.Info().Msg("server exiting")
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
	rootCmd.Flags().IntVar(&natsModulePort, "natsModulePort", 4223, "nats module server port")
	rootCmd.Flags().BoolVar(&useAuth, "auth", true, "use auth")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
	}

}

func bootNatsCloud(uuid string, natsRouter *natsrouter.NatsRouter) {
	log.Info().Msgf("starting edge device with UUID: %s", uuid)
	subject := subjects.NewSubjectBuilder(globalUUID, appID, subjects.IsApp)
	natsRouter.Handle(fmt.Sprintf("%s.", setting.NatsSettings.TopicPrefix)+uuid+".flex.rql", natsapis.RQLHandler())
	natsRouter.Handle(subject.BuildSubject("get", "system", "ping"), natsrouter.PingHandler(uuid))
	select {}
}

func setupNATS(url string) (*nats.Conn, error) {
	if url == "" {
		url = nats.DefaultURL
	}
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatal().Msgf("error connecting to NATS: %v", err)
		return nil, err
	}
	log.Info().Msgf("connected to NATS: %s", url)
	return nc, nil
}
