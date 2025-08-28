package api

import (
	"fmt"
	"time"
	"context"
	"net/http"
	"strconv"
	"reflect"	
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/go-account/internal/core/service"
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"

	"github.com/eliezerraj/go-core/coreJson"
	"github.com/gorilla/mux"
)

var (
	childLogger = log.With().Str("component", "go-account").Str("package", "internal.adapter.api").Logger()
	core_json coreJson.CoreJson
	core_apiError coreJson.APIError
	core_tools go_core_tools.ToolsCore
	tracerProvider go_core_observ.TracerProvider
)

type HttpRouters struct {
	workerService 	*service.WorkerService
	ctxTimeout		time.Duration
}

// Initialize router
func NewHttpRouters(workerService *service.WorkerService,
					ctxTimeout	time.Duration) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
		ctxTimeout: ctxTimeout,
	}
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Health").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About show all context values
func (h *HttpRouters) Context(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Context").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	contextValues := reflect.ValueOf(req.Context()).Elem()
	json.NewEncoder(rw).Encode(fmt.Sprintf("%v",contextValues))
}

// About show pgx stats
func (h *HttpRouters) Stat(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	res := h.workerService.Stat(req.Context())

	json.NewEncoder(rw).Encode(res)
}

// About handle error
func (h *HttpRouters) ErrorHandler(trace_id string, err error) *coreJson.APIError {
	if strings.Contains(err.Error(), "context deadline exceeded") {
    	err = erro.ErrTimeout
	} 
	switch err {
	case erro.ErrUpdate:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
	case erro.ErrTransInvalid:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusConflict)
	case erro.ErrInvalidAmount:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusConflict)	
	case erro.ErrBadRequest:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
	case erro.ErrNotFound:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
	case erro.ErrTimeout:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusGatewayTimeout)
	default:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
	}
	return &core_apiError
}

// About add an account
func (h *HttpRouters) AddAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","AddAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	//trace
	span := tracerProvider.Span(ctx, "adapter.api.AddAccount")
	defer span.End()

	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	// prepare body
	account := model.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	defer req.Body.Close()

	//call service
	res, err := h.workerService.AddAccount(ctx, &account)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get an account
func (h *HttpRouters) GetAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.GetAccount")
	defer span.End()
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	account := model.Account{}
	account.AccountID = varID

	// call service
	res, err := h.workerService.GetAccount(ctx, &account)
	if err != nil {
		return h.ErrorHandler(trace_id, err)	
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get an account from PK
func (h *HttpRouters) GetAccountId(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccountId").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	//context
	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.GetAccountId")
	defer span.End()

	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	varIDint, err := strconv.Atoi(varID)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	account := model.Account{}
	account.ID = varIDint

	// call service
	res, err := h.workerService.GetAccountId(ctx, &account)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About update an account
func (h *HttpRouters) UpdateAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","UpdateAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.UpdateAccount")
	defer span.End()
	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	//parameters
	account := model.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID

	// call service
	res, err := h.workerService.UpdateAccount(ctx, &account)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About delete an account
func (h *HttpRouters) DeleteAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","DeleteAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.DeleteAccount")
	defer span.End()
	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	//parameters
	account := model.Account{}
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID

	// call service
	res, err := h.workerService.DeleteAccount(ctx, &account)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About list all person´s account
func (h *HttpRouters) ListAccountPerPerson(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","ListAccountPerPerson").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.ListAccountPerPerson")
	defer span.End()
	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	account := model.Account{}
	account.PersonID = varID

	// call service
	res, err := h.workerService.ListAccountPerPerson(ctx, &account)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}