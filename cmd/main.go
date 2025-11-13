package main

import(
	"time"
	"os"
	"os/signal"
	"syscall"
	"context"
	"crypto/tls"
	
	"github.com/rs/zerolog"

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

	childLogger  = zerolog.New(os.Stdout).
						With().
						Str("component","go-account").
						Str("package", "main").
						Timestamp().
						Logger()
						
	appServer	model.AppServer
	databaseConfig go_core_pg.DatabaseConfig
	databasePGServer go_core_pg.DatabasePGServer
)

// About initialize the enviroment var
func init(){
	zerolog.SetGlobalLevel(logLevel)

	childLogger.Info().
				Str("func","init").Send()
	
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
	childLogger.Info().
				Str("func","main").
				Interface("appServer",appServer).Send()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				childLogger.Warn().
							Err(nil).Msg("Unable to open database... trying again WARNING !!")
			} else {
				childLogger.Error().
							Err(err).
							Msg("Fatal Error open Database ABORTING !!")
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
		childLogger.Error().
					Err(err).Msg("Error health check support services ERROR")
	} else {
		childLogger.Info().
					Msg("SERVICES HEALTH CHECK OK")
	}

	// start server
	httpServer.StartHttpAppServer(ctx, &httpRouters, &appServer)
}