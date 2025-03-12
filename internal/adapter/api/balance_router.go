package api

import (
	"net/http"
	"encoding/json"

	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"
	"github.com/gorilla/mux"
)

func (h *HttpRouters) GetAccountBalance(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("GetAccountBalance")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetAccountBalance")
	defer span.End()

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	accountBalance := model.AccountBalance{}
	accountBalance.AccountID = varID

	// call service
	res, err := h.workerService.GetAccountBalance(req.Context(), &accountBalance)
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

func (h *HttpRouters) AddAccountBalance(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("AddAccountBalance")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.AddAccountBalance")
	defer span.End()

	//parameters
	accountBalance := model.AccountBalance{}
	err := json.NewDecoder(req.Body).Decode(&accountBalance)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, http.StatusBadRequest)
		return &core_apiError
    }

	// Add jwt-id and request-id injected by authorizer
	childLogger.Debug().Interface("===>jwtid: ", req.Context().Value("jwt_id")).Msg("")
	childLogger.Debug().Interface("===>request_id: ", req.Context().Value("request_id")).Msg("")

	if (req.Context().Value("jwt_id") != nil) {
		jwtid := req.Context().Value("jwt_id").(string)
		accountBalance.JwtId = &jwtid
	}
	if (req.Context().Value("request_id") != nil) {
		request_id := req.Context().Value("request_id").(string)
		accountBalance.RequestId = &request_id
	}

	// call service
	res, err := h.workerService.AddAccountBalance(req.Context(), &accountBalance)
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

func (h *HttpRouters) GetMovimentAccountBalance(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("GetMovimentAccountBalance")

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetMovimentAccountBalance")
	defer span.End()

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	accountBalance := model.AccountBalance{}
	accountBalance.AccountID = varID

	// call service
	res, err := h.workerService.GetMovimentAccountBalance(req.Context(), &accountBalance)
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
