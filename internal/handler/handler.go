package handler

import (	
	"net/http"
	"context"
	"strconv"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/mux"

	"github.com/go-account/internal/core"
	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/lib"
)

var childLogger = log.With().Str("handler", "handler").Logger()

// Middleware v01
func MiddleWareHandlerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (INICIO)  --------------")
	
		if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
			childLogger.Error().Err(err).Msg("Could not Marshal http headers !!!")
		} else {
			childLogger.Debug().Str("Headers : ", string(reqHeadersBytes) ).Msg("")
		}

		//childLogger.Debug().Str("Method : ", r.Method ).Msg("")
		//childLogger.Debug().Str("URL : ", r.URL.Path ).Msg("")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "max-age=0")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
		
		childLogger.Debug().Msg("---------------------------")
		childLogger.Debug().Str("Jwtid : ", r.Header.Get("Jwtid") ).Msg("")
		childLogger.Debug().Str("RequestId : ", r.Header.Get("requestId") ).Msg("")

		ctx := context.WithValue(r.Context(), "jwt_id", r.Header.Get("Jwtid"))
		ctx = context.WithValue(ctx, "request_id", r.Header.Get("X-Request-Id"))
		childLogger.Debug().Msg("--------------------------")

		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (FIM) ----------------")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware v02 - with decoratorDB
func (h *HttpWorkerAdapter) DecoratorDB(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		childLogger.Debug().Msg("-------------- Decorator - MiddleWareHandlerHeader (INICIO) --------------")
	
		if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
			childLogger.Error().Err(err).Msg("Could not Marshal http headers !!!")
		} else {
			childLogger.Debug().Str("Headers : ", string(reqHeadersBytes) ).Msg("")
		}

		//childLogger.Debug().Str("Method : ", r.Method ).Msg("")
		//childLogger.Debug().Str("URL : ", r.URL.Path ).Msg("")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "max-age=0")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
	
		// If the user was informed then insert it in the session
		if string(r.Header.Get("client-id")) != "" {
			h.workerService.SetSessionVariable(r.Context(),string(r.Header.Get("client-id")))
		} else {
			h.workerService.SetSessionVariable(r.Context(),"NO_INFORMED")
		}

		childLogger.Debug().Msg("-------------- Decorator- MiddleWareHandlerHeader (FIM) ----------------")

		next.ServeHTTP(w, r)
	})
}

func (h *HttpWorkerAdapter) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Health")

	health := true
	json.NewEncoder(rw).Encode(health)
	return
}

func (h *HttpWorkerAdapter) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Live")

	live := true
	json.NewEncoder(rw).Encode(live)
	return
}

func (h *HttpWorkerAdapter) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Header")
	
	json.NewEncoder(rw).Encode(req.Header)
	return
}

func (h *HttpWorkerAdapter) Add( rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Add")

	span := lib.Span(req.Context(), "handler.Add")
	defer span.End()

	account := core.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrUnmarshal.Error())
        return
    }

	res, err := h.workerService.Add(req.Context(), account)
	if err != nil {
		switch err {
		default:
			rw.WriteHeader(400)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) Get(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Get")

	span := lib.Span(req.Context(), "handler.Get")
	defer span.End()
	
	vars := mux.Vars(req)
	varID := vars["id"]

	account := core.Account{}
	account.AccountID = varID
	
	res, err := h.workerService.Get(req.Context(), account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			rw.WriteHeader(404)
			json.NewEncoder(rw).Encode(err.Error())
			return
		default:
			rw.WriteHeader(500)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) GetId(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("GetId")
	
	span := lib.Span(req.Context(), "handler.GetId")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	i, err := strconv.Atoi(varID)
	if err != nil{
		rw.WriteHeader(400)
		json.NewEncoder(rw).Encode(erro.ErrConvStrint.Error())
		return
	}

	account := core.Account{}
	account.ID = i
	
	res, err := h.workerService.GetId(req.Context(), account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			rw.WriteHeader(404)
			json.NewEncoder(rw).Encode(err.Error())
			return
		default:
			rw.WriteHeader(500)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) Update(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Update")
	
	span := lib.Span(req.Context(), "handler.Update")
	defer span.End()

	account := core.Account{}
	err := json.NewDecoder(req.Body).Decode(&account)
    if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrUnmarshal.Error())
        return
    }
	
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID

	res, err := h.workerService.Update(req.Context(), account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			rw.WriteHeader(404)
			json.NewEncoder(rw).Encode(err.Error())
			return
		default:
			rw.WriteHeader(500)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) Delete(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Delete")

	span := lib.Span(req.Context(), "handler.Delete")
	defer span.End()

	account := core.Account{}
	vars := mux.Vars(req)
	varID := vars["id"]
	account.AccountID = varID
	
	res, err := h.workerService.Delete(req.Context(), account)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			rw.WriteHeader(404)
			json.NewEncoder(rw).Encode(err.Error())
			return
		default:
			rw.WriteHeader(500)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) List(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("List")
	
	span := lib.Span(req.Context(), "handler.List")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	account := core.Account{}
	account.PersonID = varID
	
	res, err := h.workerService.List(req.Context(), account)
	if err != nil {
		switch err {
		default:
			rw.WriteHeader(500)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

//-------------------------

func (h *HttpWorkerAdapter) AddFundBalanceAccount(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("AddFundBalanceAccount")

	span := lib.Span(req.Context(), "handler.AddFundBalanceAccount")
	defer span.End()

	accountBalance := core.AccountBalance{}
	err := json.NewDecoder(req.Body).Decode(&accountBalance)
    if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrUnmarshal.Error())
        return
    }

	childLogger.Debug().Interface("===>jwtid: ", req.Context().Value("jwtid")).Msg("")
	childLogger.Debug().Interface("===>request_id: ", req.Context().Value("request_id")).Msg("")

	if (req.Context().Value("jwt_id") != nil) {
		jwtid := req.Context().Value("jwt_id").(string)
		accountBalance.JwtId = &jwtid
	}

	if (req.Context().Value("request_id") != nil) {
		request_id := req.Context().Value("request_id").(string)
		accountBalance.RequestId = &request_id
	}

	res, err := h.workerService.AddFundBalanceAccount(req.Context(), accountBalance)
	if err != nil {
		switch err {
		default:
			rw.WriteHeader(400)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) GetMovimentBalanceAccount( rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("GetMovimentBalanceAccount")

	span := lib.Span(req.Context(), "handler.GetMovimentBalanceAccount")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	accountBalance := core.AccountBalance{}
	accountBalance.AccountID = varID

	res, err := h.workerService.GetMovimentBalanceAccount(req.Context(), accountBalance)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			rw.WriteHeader(404)
			json.NewEncoder(rw).Encode(err.Error())
			return
		default:
			rw.WriteHeader(400)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) GetFundBalanceAccount( rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("GetFundBalanceAccount")

	span := lib.Span(req.Context(), "handler.GetFundBalanceAccount")
	defer span.End()

	vars := mux.Vars(req)
	varID := vars["id"]

	accountBalance := core.AccountBalance{}
	accountBalance.AccountID = varID

	res, err := h.workerService.GetFundBalanceAccount(req.Context(), accountBalance)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			rw.WriteHeader(404)
			json.NewEncoder(rw).Encode(err.Error())
			return
		default:
			rw.WriteHeader(400)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}

func (h *HttpWorkerAdapter) TransferFundAccount( rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("TransferFundAccount")

	span := lib.Span(req.Context(), "handler.TransferFundAccount")
	defer span.End()

	transfer := core.Transfer{}
	err := json.NewDecoder(req.Body).Decode(&transfer)
    if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrUnmarshal.Error())
        return
    }

	res, err := h.workerService.TransferFundAccount(req.Context(), transfer)
	if err != nil {
		switch err {
		default:
			rw.WriteHeader(400)
			json.NewEncoder(rw).Encode(err.Error())
			return
		}
	}

	json.NewEncoder(rw).Encode(res)
	return
}