package handler

import (
	"time"
	"encoding/json"
	"net/http"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"context"
	"fmt"
	"encoding/base64"
	"crypto/tls"

	"github.com/gorilla/mux"

	"github.com/go-account/internal/lib"
	"github.com/go-account/internal/service"
	"github.com/go-account/internal/core"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)
// ----------------------------------------------------
type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")

	return HttpWorkerAdapter{
		workerService: workerService,
	}
}
// ----------------------------------------------------
type HttpServer struct {
	httpServer	*core.Server
}

func NewHttpAppServer(httpServer *core.Server) HttpServer {
	childLogger.Debug().Msg("NewHttpAppServer")

	return HttpServer{httpServer: httpServer }
}
// ----------------------------------------------------
func (h HttpServer) StartHttpAppServer(	ctx context.Context, 
										httpWorkerAdapter *HttpWorkerAdapter,
										appServer *core.AppServer) {
	childLogger.Info().Msg("StartHttpAppServer")
	// ---------------------- OTEL ---------------
	childLogger.Info().Str("OTEL_EXPORTER_OTLP_ENDPOINT :", appServer.ConfigOTEL.OtelExportEndpoint).Msg("")
	
	tp := lib.NewTracerProvider(ctx, appServer.ConfigOTEL, appServer.InfoPod)
	defer func() { 
		err := tp.Shutdown(ctx)
		if err != nil{
			childLogger.Error().Err(err).Msg("Erro closing OTEL tracer !!!")
		}
	}()
	otel.SetTextMapPropagator(xray.Propagator{})
	otel.SetTracerProvider(tp)

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(MiddleWareHandlerHeader)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(appServer)
	})

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/info")
		childLogger.Debug().Interface("===> : ", req.TLS).Msg("")
		json.NewEncoder(rw).Encode(&appServer)
	})
	
	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpWorkerAdapter.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpWorkerAdapter.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpWorkerAdapter.Header)

	addAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addAccount.Handle("/add", 
						http.HandlerFunc(httpWorkerAdapter.Add),)
	addAccount.Use(httpWorkerAdapter.DecoratorDB)
	addAccount.Use(otelmux.Middleware("go-account"))

	getAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getAccount.Handle("/get/{id}",
						http.HandlerFunc(httpWorkerAdapter.Get),)
	getAccount.Use(otelmux.Middleware("go-account"))

	getIdAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getIdAccount.Handle("/getId/{id}",
						http.HandlerFunc(httpWorkerAdapter.GetId),)
	getIdAccount.Use(otelmux.Middleware("go-account"))

	updateAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	updateAccount.Handle("/update/{id}", 
						http.HandlerFunc(httpWorkerAdapter.Update),)
	updateAccount.Use(otelmux.Middleware("go-account"))

	listAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	listAccount.Handle("/list/{id}", 
						http.HandlerFunc(httpWorkerAdapter.List),)
	listAccount.Use(otelmux.Middleware("go-account"))

	deleteAccount := myRouter.Methods(http.MethodDelete, http.MethodOptions).Subrouter()
    deleteAccount.HandleFunc("/delete/{id}", httpWorkerAdapter.Delete)
	deleteAccount.Use(otelmux.Middleware("go-account"))

	//---------------
	addFundAcc := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addFundAcc.Handle("/add/fund", 
						http.HandlerFunc(httpWorkerAdapter.AddFundBalanceAccount),)
	//addFundAcc.Use(httpWorkerAdapter.DecoratorDB)
	addFundAcc.Use(otelmux.Middleware("go-account"))

	getMovAcc := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getMovAcc.Handle("/get/movimentBalanceAccount/{id}", 
						http.HandlerFunc(httpWorkerAdapter.GetMovimentBalanceAccount),)
	//getMovAcc.Use(httpWorkerAdapter.DecoratorDB)
	getMovAcc.Use(otelmux.Middleware("go-account"))

	getFundAcc := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getFundAcc.Handle("/fundBalanceAccount/{id}", 
						http.HandlerFunc(httpWorkerAdapter.GetFundBalanceAccount),)
	//getFundAcc.Use(httpWorkerAdapter.DecoratorDB)
	getFundAcc.Use(otelmux.Middleware("go-account"))

	transferFundAcc := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	transferFundAcc.Handle("/transferFund", 
						http.HandlerFunc(httpWorkerAdapter.TransferFundAccount),)
	//transferFundAcc.Use(httpWorkerAdapter.DecoratorDB)
	transferFundAcc.Use(otelmux.Middleware("go-account"))

	// -------------------

	var serverTLSConf *tls.Config
	if h.httpServer.Cert.IsTLS == true {

		certPEM_Raw, err := base64.StdEncoding.DecodeString(string(h.httpServer.Cert.CertPEM))
		if err != nil {
			panic(err)
		}
		certPrivKeyPEM_Raw, err := base64.StdEncoding.DecodeString(string(h.httpServer.Cert.CertPrivKeyPEM))
		if err != nil {
			panic(err)
		}

		fmt.Println("---------------- Server CRT ------------------------")
		fmt.Println(string(certPEM_Raw))
		fmt.Println("-----------------Server Key ----------------------")
		fmt.Println(string(certPrivKeyPEM_Raw))
		fmt.Println("------------------------------------------------")

		serverCert, err := tls.X509KeyPair( certPEM_Raw, certPrivKeyPEM_Raw)
		if err != nil {
			childLogger.Error().Err(err).Msg("Erro Load X509 KeyPair")
		}
		serverTLSConf = &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			MinVersion:       tls.VersionTLS13,
			InsecureSkipVerify: false,
		}
	}

	// ---------------
	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpServer.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpServer.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpServer.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpServer.IdleTimeout) * time.Second,
		TLSConfig: serverTLSConf,
	}

	childLogger.Info().Str("Service Port : ", strconv.Itoa(h.httpServer.Port)).Msg("Service Port")

	go func() {
		var err error
		if serverTLSConf == nil {
			err = srv.ListenAndServe()
		}else {
			err = srv.ListenAndServeTLS("","")
		}
		if err != nil {
			childLogger.Error().Err(err).Msg("Cancel http mux server !!!")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		childLogger.Error().Err(err).Msg("WARNING Dirty Shutdown !!!")
		return
	}
	childLogger.Info().Msg("Stop Done !!!!")
}