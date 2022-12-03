package console

import (
	"log"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/luckyAkbar/central-worker-service/internal/config"
	"github.com/luckyAkbar/central-worker-service/internal/db"
	"github.com/luckyAkbar/central-worker-service/internal/delivery/rest"
	"github.com/luckyAkbar/central-worker-service/internal/helper"
	"github.com/luckyAkbar/central-worker-service/internal/middleware"
	"github.com/luckyAkbar/central-worker-service/internal/repository"
	"github.com/luckyAkbar/central-worker-service/internal/usecase"
	"github.com/luckyAkbar/central-worker-service/internal/worker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the server",
	Run:   server,
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

func server(c *cobra.Command, args []string) {
	db.InitializePostgresConn()
	sqlDB, err := db.PostgresDB.DB()
	if err != nil {
		logrus.Fatal("unable to start server. reason: ", err.Error())
	}

	defer helper.WrapCloser(sqlDB.Close)

	HTTPServer := echo.New()

	HTTPServer.Pre(echoMiddleware.AddTrailingSlash())
	HTTPServer.Use(middleware.RequestID())
	HTTPServer.Use(echoMiddleware.Logger())
	HTTPServer.Use(echoMiddleware.Recover())
	HTTPServer.Use(echoMiddleware.CORS())

	mailRepo := repository.NewMailRepository(db.PostgresDB)

	workerClient, err := worker.NewClient(config.WorkerBrokerRedisHost())
	if err != nil {
		logrus.Fatal(err)
	}

	mailUsecase := usecase.NewMailUsecase(mailRepo, workerClient)

	apiGroup := HTTPServer.Group("api")

	rest.Init(apiGroup, mailUsecase)

	logrus.Info("starting the server...")
	if err := HTTPServer.Start(config.ServerPort()); err != nil {
		log.Fatal("unable to start server: ", err)
	}
}
