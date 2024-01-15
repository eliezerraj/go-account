package handler

import (	
	"net/http"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/gorilla/mux"

	"github.com/go-account/internal/core"
	"github.com/go-account/internal/erro"
	
)

var childLogger = log.With().Str("handler", "handler").Logger()

// Middleware v01
func MiddleWareHandlerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (INICIO)  --------------")
	
		/*if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
			childLogger.Error().Err(err).Msg("Could not Marshal http headers !!!")
		} else {
			//childLogger.Debug().Str("Headers : ", string(reqHeadersBytes) ).Msg("")
		}*/

		childLogger.Debug().Str("Method : ", r.Method ).Msg("")
		//childLogger.Debug().Str("URL : ", r.URL.Path ).Msg("")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers","Content-Type,access-control-allow-origin, access-control-allow-headers")
		//log.Println(r.Header.Get("Host"))
		//log.Println(r.Header.Get("User-Agent"))
		//log.Println(r.Header.Get("X-Forwarded-For"))

		childLogger.Debug().Msg("-------------- MiddleWareHandlerHeader (FIM) ----------------")

		next.ServeHTTP(w, r)
	})
}

// Middleware v02 - with decoratorDB
func (h *HttpWorkerAdapter) DecoratorDB(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		childLogger.Debug().Msg("-------------- Decorator - MiddleWareHandlerHeader (INICIO) --------------")
	
		/*if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
			childLogger.Error().Err(err).Msg("Could not Marshal http headers !!!")
		} else {
			childLogger.Debug().Str("Headers : ", string(reqHeadersBytes) ).Msg("")
		}*/

		//childLogger.Debug().Str("Method : ", r.Method ).Msg("")
		//childLogger.Debug().Str("URL : ", r.URL.Path ).Msg("")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
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

func (h *HttpWorkerAdapter) Update(rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("Update")

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

func (h *HttpWorkerAdapter) AddFundBalanceAccount( rw http.ResponseWriter, req *http.Request) {
	childLogger.Debug().Msg("AddFundBalanceAccount")

	accountBalance := core.AccountBalance{}
	err := json.NewDecoder(req.Body).Decode(&accountBalance)
    if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(erro.ErrUnmarshal.Error())
        return
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