package server

import (
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"

	"github.com/go-account/internal/adapter/api"
	"github.com/go-account/internal/core/model"
	go_core_observ "github.com/eliezerraj/go-core/observability"  

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/eliezerraj/go-core/middleware"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

var childLogger = log.With().Str("handler", "api").Logger()
var core_middleware middleware.ToolsMiddleware
var tracerProvider go_core_observ.TracerProvider
var infoTrace go_core_observ.InfoTrace

type HttpServer struct {
	httpServer	*model.Server
}

func NewHttpAppServer(httpServer *model.Server) HttpServer {
	return HttpServer{httpServer: httpServer }
}

func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpRouters *api.HttpRouters,
										appServer *model.AppServer) {
	childLogger.Info().Msg("StartHttpAppServer")
			
	// ---------------------- OTEL ---------------
	childLogger.Info().Str("OTEL_EXPORTER_OTLP_ENDPOINT :", appServer.ConfigOTEL.OtelExportEndpoint).Msg("")
	
	infoTrace.PodName = appServer.InfoPod.PodName
	infoTrace.PodVersion = appServer.InfoPod.ApiVersion
	infoTrace.ServiceType = "k8-workload"
	infoTrace.Env = appServer.InfoPod.Env
	infoTrace.AccountID = appServer.InfoPod.AccountID

	tp := tracerProvider.NewTracerProvider(	ctx, 
											appServer.ConfigOTEL, 
											&infoTrace)
	defer func() { 
		err := tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Msg("error closing OTEL tracer !!!")
		}
	}()
	
	otel.SetTextMapPropagator(xray.Propagator{})
	otel.SetTracerProvider(tp)

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(core_middleware.MiddleWareHandlerHeader)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(appServer)
	})

	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpRouters.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpRouters.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpRouters.Header)

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/info")
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(appServer)
	})
	
	addAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addAccount.HandleFunc("/add", core_middleware.MiddleWareErrorHandler(httpRouters.AddAccount))		
	addAccount.Use(otelmux.Middleware("go-account"))

	getAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getAccount.HandleFunc("/get/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetAccount))		
	getAccount.Use(otelmux.Middleware("go-account"))

	deleteAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	deleteAccount.HandleFunc("/delete", core_middleware.MiddleWareErrorHandler(httpRouters.DeleteAccount))		
	deleteAccount.Use(otelmux.Middleware("go-account"))

	updateAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	updateAccount.HandleFunc("/update/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.UpdateAccount))		
	updateAccount.Use(otelmux.Middleware("go-account"))

	listAccountPerPerson := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	listAccountPerPerson.HandleFunc("/list/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.ListAccountPerPerson))		
	listAccountPerPerson.Use(otelmux.Middleware("go-account"))

	//----------------------------------

	addAccountBalance := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addAccountBalance.HandleFunc("/add/accountBalance", core_middleware.MiddleWareErrorHandler(httpRouters.AddAccountBalance))
	addAccountBalance.Use(otelmux.Middleware("go-account"))

	getAccountBalance := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getAccountBalance.HandleFunc("/accountBalance/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetAccountBalance))		
	getAccountBalance.Use(otelmux.Middleware("go-account"))

	getMovimentAccountBalance := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getMovimentAccountBalance.HandleFunc("/get/movimentAccountBalance/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetMovimentAccountBalance))		
	getMovimentAccountBalance.Use(otelmux.Middleware("go-account"))

	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpServer.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpServer.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpServer.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpServer.IdleTimeout) * time.Second, 
	}

	childLogger.Info().Str("Service Port : ", strconv.Itoa(h.httpServer.Port)).Msg("Service Port")

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Error().Err(err).Msg("canceling http mux server !!!")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("warning dirty shutdown !!!")
		return
	}
	childLogger.Info().Msg("stop done !!!")
}