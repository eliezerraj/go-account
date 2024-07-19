package service

import (
	"time"
	"os"
	"testing"
	"context"

	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog"
	"github.com/joho/godotenv"

	"github.com/go-account/internal/core"
	"github.com/go-account/internal/repository/pg"
)

var(
	logLevel = zerolog.DebugLevel
	appServer	core.AppServer
	databaseRDS core.DatabaseRDS
	server  core.Server
)

func getEnv() {
	log.Debug().Msg("1. getEnv")

	if os.Getenv("DB_HOST") !=  "" {
		databaseRDS.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") !=  "" {
		databaseRDS.Port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_NAME") !=  "" {	
		databaseRDS.DatabaseName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_SCHEMA") !=  "" {	
		databaseRDS.Schema = os.Getenv("DB_SCHEMA")
	}
	if os.Getenv("DB_DRIVER") !=  "" {	
		databaseRDS.Postgres_Driver = os.Getenv("DB_DRIVER")
	}
	if os.Getenv("DB_USER") !=  "" {	
		databaseRDS.User = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PASS") !=  "" {	
		databaseRDS.Password = os.Getenv("DB_PASS")
	}
	server.ReadTimeout=60

	databaseRDS.Host = "rds-proxy-db-arch.proxy-couoacqalfwt.us-east-2.rds.amazonaws.com"

	appServer.Server = &server
	appServer.Database = &databaseRDS
}

func TestGetAccount(t *testing.T) {
	zerolog.SetGlobalLevel(logLevel)	
	t.Setenv("AWS_REGION", "us-east-2")

	err := godotenv.Load("../../cmd/.env")
	if err != nil {
		t.Errorf("Load Env")
		log.Info().Msg("no .env file, proceed using os env")
	}

	getEnv()

	log.Debug().Msg("2. Start TestGetAccount")

	log.Debug().Interface("2. AppServer.Database :", appServer.Database).Msg("")

	ctx, cancel := context.WithTimeout(	context.Background(), 
										time.Duration( appServer.Server.ReadTimeout ) * time.Second)
	defer cancel()

	databasePG, err := pg.NewDatabasePGServer(ctx,appServer.Database)
	repoDB := pg.NewWorkerRepository(databasePG)
	
	workerService := NewWorkerService(&repoDB)

	account := core.Account{AccountID: os.Getenv("TST_ACCOUNT_ID") }

	res, err := workerService.Get(ctx, account)
	if err != nil {
		t.Errorf("Error -TestGetAccount GET erro: %v ", err)
	} else {
		if (account.AccountID == res.AccountID) {
			t.Logf("Success result : %v :", res)
		} else {
			t.Errorf("Error account not found : %v ", account.AccountID)
		}
	}
}