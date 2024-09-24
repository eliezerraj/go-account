package controller

import (	
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/mux"

	"github.com/go-account/internal/service"
	"github.com/go-account/internal/core"
	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/lib"
)

var childLogger = log.With().Str("handler", "controller").Logger()

type HttpWorkerAdapter struct {
	workerService 	*service.WorkerService
}

func NewHttpWorkerAdapter(workerService *service.WorkerService) HttpWorkerAdapter {
	childLogger.Debug().Msg("NewHttpWorkerAdapter")

	return HttpWorkerAdapter{
		workerService: workerService,
	}
}

type APIError struct {
	StatusCode	int  `json:"statusCode"`
	Msg			string `json:"msg"`
}

func (e APIError) Error() string {
	return e.Msg
}

func NewAPIError(statusCode int, err error) APIError {
	return APIError{
		StatusCode: statusCode,
		Msg:		err.Error(),
	}
}

func WriteJSON(rw http.ResponseWriter, code int, v any) error{
	rw.WriteHeader(code)
	return json.NewEncoder(rw).Encode(v)
}

func (h *HttpWorkerAdapter) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
}

func (h *HttpWorkerAdapter) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
}

func (h *HttpWorkerAdapter) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Header")
	
	json.NewEncoder(rw).Encode(req.Header)
}

func (h *HttpWorkerAdapter) Add( rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("Add")

	span := lib.Span(req.Context(), "handler.Add")
	defer span.End()

	account := core.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		apiError := NewAPIError(http.StatusBadRequest, erro.ErrUnmarshal)
		return apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.Add(req.Context(), &account)
	if err != nil {
		var apiError APIError
		switch err {
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) Get(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("Get")

	span := lib.Span(req.Context(), "handler.Get")
	defer span.End()
	
	account := core.Account{}
	vars := mux.Vars(req)
	varID := vars["id"]

	account.AccountID = varID
	
	res, err := h.workerService.Get(req.Context(), &account)
	if err != nil {
		var apiError APIError
		switch err {
			case erro.ErrNotFound:
				apiError = NewAPIError(http.StatusNotFound, err)
			default:
				apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) GetId(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("GetId")
	
	span := lib.Span(req.Context(), "handler.GetId")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	i, err := strconv.Atoi(varID)
	if err != nil{
		apiError := NewAPIError(http.StatusBadRequest, erro.ErrConvStrint)
		return apiError
	}

	account := core.Account{}
	account.ID = i
	
	res, err := h.workerService.GetId(req.Context(), &account)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) Update(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("Update")
	
	span := lib.Span(req.Context(), "handler.Update")
	defer span.End()

	account := core.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		apiError := NewAPIError(http.StatusBadRequest, erro.ErrUnmarshal)
		return apiError
    }
	
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID

	res, err := h.workerService.Update(req.Context(), &account)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) Delete(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("Delete")

	span := lib.Span(req.Context(), "handler.Delete")
	defer span.End()

	account := core.Account{}
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID
	
	res, err := h.workerService.Delete(req.Context(), &account)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) List(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("List")
	
	span := lib.Span(req.Context(), "handler.List")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	account := core.Account{}
	account.PersonID = varID
	
	res, err := h.workerService.List(req.Context(), &account)
	if err != nil {
		var apiError APIError
		switch err {
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

//-------------------------

func (h *HttpWorkerAdapter) AddFundBalanceAccount(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("AddFundBalanceAccount")

	span := lib.Span(req.Context(), "handler.AddFundBalanceAccount")
	defer span.End()

	accountBalance := core.AccountBalance{}
	err := json.NewDecoder(req.Body).Decode(&accountBalance)
    if err != nil {
		apiError := NewAPIError(http.StatusBadRequest, erro.ErrUnmarshal)
		return apiError
    }

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

	res, err := h.workerService.AddFundBalanceAccount(req.Context(), &accountBalance)
	if err != nil {
		var apiError APIError
		switch err {
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) GetMovimentBalanceAccount( rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("GetMovimentBalanceAccount")

	span := lib.Span(req.Context(), "handler.GetMovimentBalanceAccount")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	accountBalance := core.AccountBalance{}
	accountBalance.AccountID = varID

	res, err := h.workerService.GetMovimentBalanceAccount(req.Context(), &accountBalance)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) GetFundBalanceAccount( rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("GetFundBalanceAccount")

	span := lib.Span(req.Context(), "handler.GetFundBalanceAccount")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	accountBalance := core.AccountBalance{}
	accountBalance.AccountID = varID

	res, err := h.workerService.GetFundBalanceAccount(req.Context(), &accountBalance)
	if err != nil {
		var apiError APIError
		switch err {
		case erro.ErrNotFound:
			apiError = NewAPIError(http.StatusNotFound, err)
		default:
			apiError = NewAPIError(http.StatusInternalServerError, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}

func (h *HttpWorkerAdapter) TransferFundAccount( rw http.ResponseWriter, req *http.Request) error {
	childLogger.Debug().Msg("TransferFundAccount")

	span := lib.Span(req.Context(), "handler.TransferFundAccount")
	defer span.End()

	transfer := core.Transfer{}
	err := json.NewDecoder(req.Body).Decode(&transfer)
    if err != nil {
		apiError := NewAPIError(http.StatusBadRequest, erro.ErrUnmarshal)
		return apiError
    }

	res, err := h.workerService.TransferFundAccount(req.Context(), &transfer)
	if err != nil {
		var apiError APIError
		switch err {
		default:
			apiError = NewAPIError(http.StatusNotFound, err)
		}
		return apiError
	}

	return WriteJSON(rw, http.StatusOK, res)
}