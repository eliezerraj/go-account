package main

import(
	"time"
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-account/internal/util"
	"github.com/go-account/internal/handler"
	"github.com/go-account/internal/core"
	"github.com/go-account/internal/service"
	"github.com/go-account/internal/repository/postgre"
	"github.com/go-account/internal/repository/pg"
)

var(
	logLevel 	= 	zerolog.DebugLevel
	appServer	core.AppServer
)

func init(){
	log.Debug().Msg("init")
	zerolog.SetGlobalLevel(logLevel)

	infoPod , server, _ := util.GetInfoPod()
	database := util.GetDatabaseEnv()
	configOTEL := util.GetOtelEnv()
	cert := util.GetCertEnv()

	//appServer.Cert = &cert
	appServer.InfoPod = &infoPod
	appServer.Database = &database
	appServer.Server = &server
	appServer.Server.Cert = &cert
	appServer.ConfigOTEL = &configOTEL
}

func main() {
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Msg("main")
	log.Debug().Msg("----------------------------------------------------")
	log.Debug().Interface("appServer :",appServer).Msg("")
	log.Debug().Msg("----------------------------------------------------")

	ctx, cancel := context.WithTimeout(	context.Background(), 
										time.Duration( appServer.Server.ReadTimeout ) * time.Second)
	defer cancel()

	// Open Database
	count := 1
	var databaseHelper	postgre.DatabaseHelper
	var databasePG	pg.DatabasePG
	var err error
	for {
		databaseHelper, err = postgre.NewDatabaseHelper(ctx, appServer.Database)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro open Database... trying again !!")
			} else {
				log.Error().Err(err).Msg("Fatal erro open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		//break
		databasePG, err = pg.NewDatabasePGServer(ctx, appServer.Database)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("Erro open Database... trying again !!")
			} else {
				log.Error().Err(err).Msg("Fatal erro open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second)
			count = count + 1
			continue
		}
		break
	}

	repoDB 			:= postgre.NewWorkerRepository(databaseHelper)
	repoDatabase := pg.NewWorkerRepository(databasePG)

	workerService 	:= service.NewWorkerService(&repoDatabase, &repoDB )
	
	httpWorkerAdapter 	:= handler.NewHttpWorkerAdapter(workerService)
	httpServer 			:= handler.NewHttpAppServer(appServer.Server)

	httpServer.StartHttpAppServer(ctx, &httpWorkerAdapter, &appServer)
}