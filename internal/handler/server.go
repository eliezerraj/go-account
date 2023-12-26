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

	"github.com/go-account/internal/service"
	"github.com/go-account/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) *HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")
	return &HttpWorkerAdapter{
		workerService: workerService,
	}
}

type HttpServer struct {
	start 			time.Time
	httpAppServer 	core.HttpAppServer
}

func NewHttpAppServer(httpAppServer core.HttpAppServer) HttpServer {
	childLogger.Debug().Msg("NewHttpAppServer")

	return HttpServer{	start: time.Now(), 
						httpAppServer: httpAppServer,
					}
}

func (h HttpServer) StartHttpAppServer(ctx context.Context, httpWorkerAdapter *HttpWorkerAdapter) {
	childLogger.Info().Msg("StartHttpAppServer")
		
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(MiddleWareHandlerHeader)

	myRouter.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/")
		json.NewEncoder(rw).Encode(h.httpAppServer)
	})

	myRouter.HandleFunc("/info", func(rw http.ResponseWriter, req *http.Request) {
		childLogger.Debug().Msg("/info")
		json.NewEncoder(rw).Encode(h.httpAppServer)
	})
	
	health := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    health.HandleFunc("/health", httpWorkerAdapter.Health)

	live := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    live.HandleFunc("/live", httpWorkerAdapter.Live)

	header := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    header.HandleFunc("/header", httpWorkerAdapter.Header)
	header.Use(MiddleWareHandlerHeader)

	addAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addAccount.Handle("/add", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "account:", h.httpAppServer.InfoPod.AvailabilityZone, ".add")), 
						http.HandlerFunc(httpWorkerAdapter.Add),
						),
	)
	addAccount.Use(httpWorkerAdapter.DecoratorDB)

	getAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
    //getAccount.HandleFunc("/get/{id}", httpWorkerAdapter.Get)
	getAccount.Handle("/get/{id}",
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "account:", h.httpAppServer.InfoPod.AvailabilityZone, ".getId")),
						http.HandlerFunc(httpWorkerAdapter.Get),
						),
	)
	getAccount.Use(MiddleWareHandlerHeader)

	updateAccount := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	updateAccount.Handle("/update/{id}", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "account:", h.httpAppServer.InfoPod.AvailabilityZone, ".updateId")),
						http.HandlerFunc(httpWorkerAdapter.Update),
						),
	)
	updateAccount.Use(httpWorkerAdapter.DecoratorDB)

	listAccount := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	listAccount.Handle("/list/{id}", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "account:", h.httpAppServer.InfoPod.AvailabilityZone, ".list")),
						http.HandlerFunc(httpWorkerAdapter.List),
						),
	)
	listAccount.Use(MiddleWareHandlerHeader)
	
	deleteAccount := myRouter.Methods(http.MethodDelete, http.MethodOptions).Subrouter()
    deleteAccount.HandleFunc("/delete/{id}", httpWorkerAdapter.Delete)
	deleteAccount.Use(MiddleWareHandlerHeader)

	//---------------
	addFundAcc := myRouter.Methods(http.MethodPost, http.MethodOptions).Subrouter()
	addFundAcc.Handle("/add/fund", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "account:", h.httpAppServer.InfoPod.AvailabilityZone, ".add.fund")), 
						http.HandlerFunc(httpWorkerAdapter.AddFundBalanceAccount),
						),
	)
	addFundAcc.Use(httpWorkerAdapter.DecoratorDB)

	getMovAcc := myRouter.Methods(http.MethodGet, http.MethodOptions).Subrouter()
	getMovAcc.Handle("/get/movimentBalanceAccount/{id}", 
						xray.Handler(xray.NewFixedSegmentNamer(fmt.Sprintf("%s%s%s", "account:", h.httpAppServer.InfoPod.AvailabilityZone, ".get.movimentBalanceAccount")), 
						http.HandlerFunc(httpWorkerAdapter.GetMovimentBalanceAccount),
						),
	)
	getMovAcc.Use(httpWorkerAdapter.DecoratorDB)

	// -------------------

	var serverTLSConf *tls.Config
	if string(h.httpAppServer.Cert.CertPEM) != "" {
		certPEM_Raw, err := base64.StdEncoding.DecodeString(string(h.httpAppServer.Cert.CertPEM))
		if err != nil {
			panic(err)
		}
		certPrivKeyPEM_Raw, err := base64.StdEncoding.DecodeString(string(h.httpAppServer.Cert.CertPrivKeyPEM))
		if err != nil {
			panic(err)
		}

		fmt.Println("------------------------------------------------")
		fmt.Println(string(certPEM_Raw))
		fmt.Println("------------------------------------------------")
		fmt.Println(string(certPrivKeyPEM_Raw))
		fmt.Println("------------------------------------------------")

		serverCert, err := tls.X509KeyPair( certPEM_Raw, certPrivKeyPEM_Raw)
		if err != nil {
			childLogger.Error().Err(err).Msg("Erro Load X509 KeyPair")
		}
		serverTLSConf = &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			//InsecureSkipVerify: true,
		}
	}

	// ---------------
	srv := http.Server{
		Addr:         ":" +  strconv.Itoa(h.httpAppServer.Server.Port),      	
		Handler:      myRouter,                	          
		ReadTimeout:  time.Duration(h.httpAppServer.Server.ReadTimeout) * time.Second,   
		WriteTimeout: time.Duration(h.httpAppServer.Server.WriteTimeout) * time.Second,  
		IdleTimeout:  time.Duration(h.httpAppServer.Server.IdleTimeout) * time.Second,
		TLSConfig: serverTLSConf,
	}

	childLogger.Info().Str("Service Port : ", strconv.Itoa(h.httpAppServer.Server.Port)).Msg("Service Port")

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