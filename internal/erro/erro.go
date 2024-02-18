package erro

import (
	"errors"

)

var (
	ErrNotFound 		= errors.New("Item não encontrado")
	ErrInsert 			= errors.New("Erro na inserção do dado")
	ErrUpdate			= errors.New("Erro no update do dado")
	ErrDelete 			= errors.New("Erro no Delete")
	ErrUnmarshal 		= errors.New("Erro na conversão do JSON")
	ErrUnauthorized 	= errors.New("Erro de autorização")
	ErrTransaction		= errors.New("Type of Transaction invalid !!!")
	ErrConvStrint		= errors.New("The field must be numeric !!!")
)
