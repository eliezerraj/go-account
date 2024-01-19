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
	"github.com/go-account/internal/repository/postgre"
)

var(
	logLevel = zerolog.DebugLevel
	envDB	 				core.DatabaseRDS
	dataBaseHelper 			postgre.DatabaseHelper
	infoPod					core.InfoPod
	server					core.Server
	repoDB					postgre.WorkerRepository
)

func getEnv() {
	if os.Getenv("DB_HOST") !=  "" {
		envDB.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") !=  "" {
		envDB.Port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_NAME") !=  "" {	
		envDB.DatabaseName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_SCHEMA") !=  "" {	
		envDB.Schema = os.Getenv("DB_SCHEMA")
	}
	if os.Getenv("DB_DRIVER") !=  "" {	
		envDB.Postgres_Driver = os.Getenv("DB_DRIVER")
	}
	if os.Getenv("DB_USER") !=  "" {	
		envDB.User = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PASS") !=  "" {	
		envDB.Password = os.Getenv("DB_PASS")
	}
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

	server.ReadTimeout = 60
	server.WriteTimeout = 60
	server.IdleTimeout = 60
	server.CtxTimeout = 60

	infoPod.Database = &envDB

	log.Debug().Interface("infoPod.Database :", infoPod.Database).Msg("init")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration( server.ReadTimeout ) * time.Second)
	defer cancel()

	dataBaseHelper, err = postgre.NewDatabaseHelper(ctx, envDB)
	repoDB = postgre.NewWorkerRepository(dataBaseHelper)
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