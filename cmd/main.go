package main

import(
	"time"
	"os"
	"os/signal"
	"syscall"
	"context"
	"crypto/tls"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/afiskon/promtail-client/promtail"

	"github.com/go-account/internal/infra/configuration"
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/service"
	"github.com/go-account/internal/infra/server"
	"github.com/go-account/internal/adapter/api"
	"github.com/go-account/internal/adapter/database"

	redis "github.com/redis/go-redis/v9"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"  
	go_core_cache "github.com/eliezerraj/go-core/cache/redis_cluster"
)

var(
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	childLogger = log.With().Str("component","go-account").Str("package", "main").Logger()
	
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer

	promtailClient   promtail.Client
)

// About initialize the enviroment var
func init(){
	childLogger.Info().Str("func","init").Send()
	
	zerolog.SetGlobalLevel(logLevel)

	cfg := promtail.ClientConfig{
		PushURL:            "http://localhost:3100/api/prom/push",
		BatchWait:          3 * time.Second,
		BatchEntriesNumber: 100,
		SendLevel:          promtail.INFO, // ⚠️ Use promtail.LogLevel, not zerolog
		PrintLevel:         promtail.DEBUG,
	}
	var err error
	promtailClient, err = promtail.NewClientProto(cfg)
	if err != nil {
		panic(err)
	}

	// You’ll send logs manually using promtailClient.Send()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv()
	cacheConfig 	:= configuration.GetCacheEnv()

	appServer.CacheConfig = &cacheConfig
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

	// Open Valkey
	var redisClientCache 	go_core_cache.RedisClient
	var optRedisClient		redis.Options

	optRedisClient.Username = appServer.CacheConfig.Username
	optRedisClient.Password = appServer.CacheConfig.Password
	optRedisClient.Addr = appServer.CacheConfig.Host
	optRedisClient.PoolSize =     10  // Maximum number of connections in the pool
	optRedisClient.MinIdleConns = 5   // Minimum number of idle connections
	optRedisClient.PoolTimeout =  5 * time.Second // Timeout for getting a connection from the pool

	if true {
		optRedisClient.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	workerCache := redisClientCache.NewRedisClientCache(&optRedisClient)
	
	// wire
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(database, workerCache)
	httpRouters := api.NewHttpRouters(workerService, time.Duration(appServer.Server.CtxTimeout))
	httpServer := server.NewHttpAppServer(appServer.Server)

		// Services Health Check
	err = workerService.HealthCheck(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("fatal error health check aborting")
	} else {
		childLogger.Info().Msg("SERVICES HEALTH CHECK OK")
	}

	// start server
	httpServer.StartHttpAppServer(ctx, &httpRouters, &appServer)
}