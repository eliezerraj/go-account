package main

import(
	"time"
	"os"
	"os/signal"
	"syscall"
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-account/internal/infra/configuration"
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/service"
	"github.com/go-account/internal/infra/server"
	"github.com/go-account/internal/adapter/api"
	"github.com/go-account/internal/adapter/database"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"  
)

var(
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer
	childLogger = log.With().Str("component","go-account").Str("package", "main").Logger()
)

// About initialize the enviroment var
func init(){
	childLogger.Info().Str("func","init").Send()

	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv()

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.ConfigOTEL = &configOTEL
	appServer.DatabaseConfig = &databaseConfig
}

// About main
func main (){
	childLogger.Info().Str("func","main").Interface("appServer",appServer).Send()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				childLogger.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				childLogger.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	// wire
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database)
	httpRouters := api.NewHttpRouters(workerService, time.Duration(appServer.Server.CtxTimeout))
	httpServer := server.NewHttpAppServer(appServer.Server)

	// start server
	httpServer.StartHttpAppServer(ctx, &httpRouters, &appServer)
}