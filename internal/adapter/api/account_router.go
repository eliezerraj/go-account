package api

import (
	"encoding/json"
	"net/http"
	"github.com/rs/zerolog/log"
	"github.com/go-account/internal/core/service"
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"
	"github.com/eliezerraj/go-core/coreJson"
	"github.com/gorilla/mux"
)

var childLogger = log.With().Str("adapter", "api.router").Logger()

var core_json coreJson.CoreJson
var core_apiError coreJson.APIError
var core_tools go_core_tools.ToolsCore
var tracerProvider go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	return HttpRouters{
		workerService: workerService,
	}
}

func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
}

func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
}

func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Header")
	
	json.NewEncoder(rw).Encode(req.Header)
}

func (h *HttpRouters) AddAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("AddAccount")

	//trace
	span := tracerProvider.Span(req.Context(), "adapter.api.AddAccount")
	defer span.End()

	// prepare body
	account := model.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(erro.ErrUnmarshal, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	//call service
	res, err := h.workerService.AddAccount(req.Context(), &account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		case erro.ErrTransInvalid:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)
		case erro.ErrInvalidAmount:
			core_apiError = core_apiError.NewAPIError(err, http.StatusConflict)	
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpRouters) GetAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("GetAccount")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetAccount")
	defer span.End()

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
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpRouters) UpdateAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("UpdateAccount")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.UpdateAccount")
	defer span.End()

	//parameters
	account := model.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(erro.ErrUnmarshal, http.StatusBadRequest)
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
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		case erro.ErrUpdate:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpRouters) DeleteAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("DeleteAccount")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.DeleteAccount")
	defer span.End()

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
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpRouters) ListAccountPerPerson(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("ListAccountPerPerson")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.ListAccountPerPerson")
	defer span.End()

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
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}