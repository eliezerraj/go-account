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
	"github.com/rs/zerolog"

	"github.com/eliezerraj/go-core/middleware"

	// trace
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/propagation"
	 sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	// Metrics
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	childLogger  = zerolog.New(os.Stdout).
						With().
						Str("component","go-account").
						Str("package","internal.core.service").
						Timestamp().
						Logger()

	core_middleware middleware.ToolsMiddleware
	tracerProvider 	go_core_observ.TracerProvider
	infoTrace 		go_core_observ.InfoTrace
	tracer			trace.Tracer
)

type HttpServer struct {
	httpServer	*model.Server
}

// About create new http server
func NewHttpAppServer(httpServer *model.Server) HttpServer {
	childLogger.Info().
				Str("func","NewHttpAppServer").Send()
	
	return HttpServer{httpServer: httpServer }
}

// About initialize MeterProvider with Prometheus exporter
func initMeterProvider(ctx context.Context, serviceName string) (*sdkmetric.MeterProvider, error) {
	childLogger.Info().
				Str("func","initMeterProvider").Send()

	// 1. Configurar o Recurso OTel
	res, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("env", "DEV"),
		),
	)
	if err != nil {
		return nil, err
	}

	// 2. Criar o Prometheus Exporter
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	// 3. Criar o MeterProvider, usando o Prometheus Exporter como Reader.
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(exporter),
	)

	return provider, nil
}

//About Create Custom Metrics
var httpRequestsCounter metric.Int64Counter
var httpLatencyHistogram metric.Float64Histogram
var err_metric error

func setupCustomMetrics(meter metric.Meter) error {
	childLogger.Info().
				Str("func","setupCustomMetrics").Send()

	httpRequestsCounter, err_metric = meter.Int64Counter("eliezer-http_requests_total",
				metric.WithDescription("Total number of HTTP requests by path"),
				metric.WithUnit("1"),
	)
	if err_metric != nil {
		childLogger.Error().
					Err(err_metric).
					Msg("Erro Create Custom Metrics")
		return err_metric
	}

	httpLatencyHistogram, err_metric = meter.Float64Histogram("eliezer-http_server_latency_seconds",
		metric.WithDescription("Latency of HTTP server requests by path"),
		metric.WithUnit("s"),
	)	
	if err_metric != nil {
		childLogger.Error().
					Err(err_metric).
					Msg("Erro Create Custom Metrics")
		return err_metric
	}

	return nil
}

// About start http server
func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpRouters *api.HttpRouters,
										appServer *model.AppServer) {
	childLogger.Info().
				Str("func","StartHttpAppServer").Send()
			
	// --------- OTEL traces ---------------
	var initTracerProvider *sdktrace.TracerProvider
	
	if appServer.InfoPod.OtelTraces {
		infoTrace.PodName = appServer.InfoPod.PodName
		infoTrace.PodVersion = appServer.InfoPod.ApiVersion
		infoTrace.ServiceType = "k8-workload"
		infoTrace.Env = appServer.InfoPod.Env
		infoTrace.AccountID = appServer.InfoPod.AccountID

		initTracerProvider = tracerProvider.NewTracerProvider(	ctx, 
																appServer.ConfigOTEL, 
																&infoTrace)

		otel.SetTextMapPropagator(propagation.TraceContext{})
		otel.SetTracerProvider(initTracerProvider)
		tracer = initTracerProvider.Tracer(appServer.InfoPod.PodName)
	}

	// --------- OTEL metrics ---------------
	var meterProvider *sdkmetric.MeterProvider

	if appServer.InfoPod.OtelMetrics {
		meterProvider, err := initMeterProvider(ctx, infoTrace.PodName)
		if err != nil {
			childLogger.Error().
						Err(err).
						Msg("Error start Otel Metrics Provider")
		} else {
			meter := meterProvider.Meter(infoTrace.PodName)

			setupCustomMetrics(meter)
			if err != nil {
				childLogger.Info().
							Msg("Erro Create Custom Metrics")
			}

			childLogger.Info().
						Msg("Otel Metrics Provider started SUCCESSFULL")
		}
	}

	// handle defer
	defer func() { 

		if meterProvider != nil {
			if err := meterProvider.Shutdown(ctx); err != nil {
				childLogger.Error().
							Err(err).
							Msg("Erro to stop metrics provider")
			}
		}

		if initTracerProvider != nil {
			err := initTracerProvider.Shutdown(ctx)
			if err != nil{
				childLogger.Error().
							Err(err).
							Msg("Erro to shutdown tracer provider")
			}
		}
		childLogger.Info().
					Msg("stop done !!!")
	}()

	// Router
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(core_middleware.MiddleWareHandlerHeader)
	myRouter.Handle("/metrics", promhttp.Handler())

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Info().Str("HandleFunc","/").Send()
		
		json.NewEncoder(rw).Encode(appServer)
	})
	
	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpRouters.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpRouters.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	header.HandleFunc("/header", httpRouters.Header)	

	wk_ctx := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    wk_ctx.HandleFunc("/context", httpRouters.Context)
	
	stat := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    stat.HandleFunc("/stat", httpRouters.Stat)
	
	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Info().
					Str("HandleFunc","/info").Send()
		start := time.Now()
		targetPath := "/info"

		req_ctx, cancel := context.WithTimeout(req.Context(), 5 * time.Second)
    	defer cancel()

		req_ctx, span := tracerProvider.SpanCtx(req_ctx, "adapter.api.info")
		defer span.End()

		defer func() {
			if httpLatencyHistogram != nil {
				duration := time.Since(start).Seconds()
				httpLatencyHistogram.Record(req.Context(), duration, metric.WithAttributes(attribute.String("http.target", targetPath)))
			}
		}()

		if httpRequestsCounter != nil {
			httpRequestsCounter.Add(req.Context(), 1, metric.WithAttributes(attribute.String("http.target", "/info")))
		}

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(appServer)
	})
	
	addAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addAccount.HandleFunc("/add", core_middleware.MiddleWareErrorHandler(httpRouters.AddAccount))		
	addAccount.Use(otelmux.Middleware("go-account"))

	getAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getAccount.HandleFunc("/get/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetAccount))		
	getAccount.Use(otelmux.Middleware("go-account"))

	getAccountId := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getAccountId.HandleFunc("/getId/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.GetAccountId))		
	getAccountId.Use(otelmux.Middleware("go-account"))

	deleteAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	deleteAccount.HandleFunc("/delete", core_middleware.MiddleWareErrorHandler(httpRouters.DeleteAccount))		
	deleteAccount.Use(otelmux.Middleware("go-account"))

	updateAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	updateAccount.HandleFunc("/update/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.UpdateAccount))		
	updateAccount.Use(otelmux.Middleware("go-account"))

	listAccountPerPerson := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	listAccountPerPerson.HandleFunc("/list/{id}", core_middleware.MiddleWareErrorHandler(httpRouters.ListAccountPerPerson))		
	listAccountPerPerson.Use(otelmux.Middleware("go-account"))
		
	// start http server
	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpServer.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpServer.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpServer.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpServer.IdleTimeout) * time.Second, 
	}

	childLogger.Info().
				Str("Service Port", strconv.Itoa(h.httpServer.Port)).Send()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Warn().
						Err(err).Msg("Canceling http mux server !!!")
		}
	}()
	
	// Get SIGNALS
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-ch

		switch sig {
		case syscall.SIGHUP:
			childLogger.Info().
						Msg("Received SIGHUP: reloading configuration...")
		case syscall.SIGINT, syscall.SIGTERM:
			childLogger.Info().
						Msg("Received SIGINT/SIGTERM termination signal. Exiting")
			return
		default:
			childLogger.Info().
						Interface("Received signal:", sig).Send()
		}
	}

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Warn().
					Err(err).
					Msg("Dirty shutdown WARNING !!!")
		return
	}
}