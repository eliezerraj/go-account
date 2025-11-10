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

	// trace
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	// metrics
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/attribute"

	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	clientprom "github.com/prometheus/client_golang/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	 semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"github.com/prometheus/client_golang/prometheus/promhttp"

)

var (
	childLogger = log.With().Str("component","go-account").Str("package","internal.infra.server").Logger()
	core_middleware middleware.ToolsMiddleware
	tracerProvider go_core_observ.TracerProvider
	infoTrace go_core_observ.InfoTrace
	tracer	trace.Tracer
)

type HttpServer struct {
	httpServer	*model.Server
}

// About create new http server
func NewHttpAppServer(httpServer *model.Server) HttpServer {
	childLogger.Info().Str("func","NewHttpAppServer").Send()
	
	return HttpServer{httpServer: httpServer }
}

// About start http server
func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpRouters *api.HttpRouters,
										appServer *model.AppServer) {
	childLogger.Info().Str("func","StartHttpAppServer").Send()
			
	// Otel tracer
	infoTrace.AccountID = appServer.InfoPod.AccountID
	infoTrace.PodName = appServer.InfoPod.PodName
	infoTrace.PodVersion = appServer.InfoPod.ApiVersion
	infoTrace.ServiceType = "k8-workload"
	infoTrace.Env = appServer.InfoPod.Env

	tp := tracerProvider.NewTracerProvider(	ctx, 
											appServer.ConfigOTEL, 
											&infoTrace)

	if tp != nil {
		otel.SetTextMapPropagator(propagation.TraceContext{}) //  propagation.TraceContext{}
		otel.SetTracerProvider(tp)
		tracer = tp.Tracer(appServer.InfoPod.PodName)
	}

	// Metrics
	exporterMetrics, err := otelprom.New()
	if err != nil {
		childLogger.Error().Err(err).Msg("failed to initialize prometheus exporter")
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("go-account"),
	)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporterMetrics),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	meter := otel.Meter("go-account")

	// --- Start collecting Go runtime metrics
	if err := runtime.Start(runtime.WithMeterProvider(meterProvider)); err != nil {
		childLogger.Error().Err(err).Msg("failed to start runtime instrumentation")
	}

	// metrics
	requestCounter, err := meter.Int64Counter("http_requests_total")
	if err != nil {
		childLogger.Error().Err(err).Msg("failed to create request counter")
	}

	latencyHistogram, err := meter.Float64Histogram("http_request_duration_seconds")
	if err != nil {
		childLogger.Error().Err(err).Msg("failed to create histogram")
	}

	cpuUsage := clientprom.NewGauge(clientprom.GaugeOpts{
		Name: "go_process_cpu_usage_percent",
		Help: "Approximate CPU usage percentage of the Go process",
	})

	// ---- Prometheus HTTP Handler (/metrics) ----
	var metricsHandler http.Handler
	// Create a separate client_golang registry for process/go collectors and custom gauges.
	registry := clientprom.NewRegistry()
	registry.MustRegister(clientprom.NewProcessCollector(clientprom.ProcessCollectorOpts{}))
	registry.MustRegister(clientprom.NewGoCollector())
	registry.MustRegister(cpuUsage)

	// Use the client_golang registry as the /metrics handler. In some versions of the
	// OpenTelemetry Prometheus exporter the exporter type does not implement the
	// client_golang Gatherer interface, so combining them fails. To ensure Prometheus
	// receives process/go metrics and custom gauges, expose the client_golang registry.
	metricsHandler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	// handle defer
	defer func() { 
		if tp != nil {
			err := tp.Shutdown(ctx)
			if err != nil{
				childLogger.Error().Err(err).Send()
			}
		}
		childLogger.Info().Msg("stop done !!!")
	}()

	// Router
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(core_middleware.MiddleWareHandlerHeader)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Info().Str("HandleFunc","/").Send()
		
		json.NewEncoder(rw).Encode(appServer)
	})
	
	metrics := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	metrics.HandleFunc("/metrics", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Info().Str("HandleFunc","/metrics").Send()
		metricsHandler.ServeHTTP(rw, req)
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
		childLogger.Info().Str("HandleFunc","/info").Send()

		start := time.Now()
		defer func() {
			duration := time.Since(start).Seconds()
			latencyHistogram.Record(ctx, duration, metric.WithAttributes(attribute.String("path", "/info")))
		}()

		requestCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("path", "/info"),
			attribute.String("method", req.Method),
		))

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

	childLogger.Info().Str("Service Port", strconv.Itoa(h.httpServer.Port)).Send()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			childLogger.Info().Err(err).Msg("canceling http mux server !!!")
		}
	}()
	
	// Get SIGNALS
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	for {
		sig := <-ch

		switch sig {
		case syscall.SIGHUP:
			childLogger.Info().Msg("Received SIGHUP: reloading configuration...")
		case syscall.SIGINT, syscall.SIGTERM:
			childLogger.Info().Msg("Received SIGINT/SIGTERM termination signal. Exiting")
			return
		default:
			childLogger.Info().Interface("Received signal:", sig).Send()
		}
	}

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("warning dirty shutdown !!!")
		return
	}
}