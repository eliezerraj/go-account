package api

import (
	"fmt"
	"net/http"
	"strconv"
	"reflect"	
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/go-account/internal/core/service"
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"

	"github.com/eliezerraj/go-core/coreJson"
	"github.com/gorilla/mux"
)

var childLogger = log.With().Str("component", "go-account").Str("package", "internal.adapter.api").Logger()

var core_json coreJson.CoreJson
var core_apiError coreJson.APIError
var core_tools go_core_tools.ToolsCore
var tracerProvider go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
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

// About add an account
func (h *HttpRouters) AddAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","AddAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	//trace
	span := tracerProvider.Span(req.Context(), "adapter.api.AddAccount")
	defer span.End()

	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	// prepare body
	account := model.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	//call service
	res, err := h.workerService.AddAccount(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		case erro.ErrTransInvalid:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusConflict)
		case erro.ErrInvalidAmount:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusConflict)	
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get an account
func (h *HttpRouters) GetAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetAccount")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	account := model.Account{}
	account.AccountID = varID

	// call service
	res, err := h.workerService.GetAccount(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get an account from PK
func (h *HttpRouters) GetAccountId(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccountId").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetAccountId")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	varIDint, err := strconv.Atoi(varID)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	account := model.Account{}
	account.ID = varIDint

	// call service
	res, err := h.workerService.GetAccountId(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About update an account
func (h *HttpRouters) UpdateAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","UpdateAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.UpdateAccount")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	//parameters
	account := model.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
		return &core_apiError
    }
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID

	// call service
	res, err := h.workerService.UpdateAccount(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		case erro.ErrUpdate:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About delete an account
func (h *HttpRouters) DeleteAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","DeleteAccount").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.DeleteAccount")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	//parameters
	account := model.Account{}
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID

	// call service
	res, err := h.workerService.DeleteAccount(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About list all person´s account
func (h *HttpRouters) ListAccountPerPerson(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","ListAccountPerPerson").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.ListAccountPerPerson")
	defer span.End()
	trace_id := fmt.Sprintf("%v",req.Context().Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	account := model.Account{}
	account.PersonID = varID

	// call service
	res, err := h.workerService.ListAccountPerPerson(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}