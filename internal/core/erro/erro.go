package erro

import (
	"errors"
)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrUpdate			= errors.New("update unsuccessful")
	ErrInsert 			= errors.New("insert data error")
	ErrUnmarshal 		= errors.New("unmarshal json error")
	ErrUnauthorized 	= errors.New("not authorized")
	ErrServer		 	= errors.New("server identified error")
	ErrHTTPForbiden		= errors.New("forbiden request")
	ErrTransInvalid		= errors.New("transaction invalid")
	ErrInvalidAmount	= errors.New("invalid amount for this transaction type")
	ErrTimeout			= errors.New("timeout: context deadline exceeded.")
)