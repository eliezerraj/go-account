package core

import (
	"time"

)

type Account struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	CreateAt		time.Time 	`json:"create_at,omitempty"`
	UpdateAt		*time.Time 	`json:"update_at,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
}

type AccountBalance struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount"`
	TenantID		string  	`json:"tenant_id,omitempty"`
	CreateAt		time.Time 	`json:"create_at,omitempty"`
	UpdateAt		*time.Time 	`json:"update_at,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
}

type AccountStatement struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	ChargeAt		time.Time 	`json:"charged_at,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
}

type MovimentAccount struct {
	AccountBalance				*AccountBalance		`json:"account_balance,omitempty"`
	AccountBalanceStatementCredit	float64			`json:"account_balance_statement_credit,omitempty"`
	AccountBalanceStatementDebit	float64			`json:"account_balance_statement_debit,omitempty"`
	AccountBalanceStatementTotal	float64			`json:"account_balance_debit.debit_total,omitempty"`
	AccountStatement				*[]AccountStatement	`json:"account_statement,omitempty"`
}

type Transfer struct {
	ID				int			`json:"id,omitempty"`
	AccountIDFrom	string		`json:"account_id_from,omitempty"`
	FkAccountIDFrom	int			`json:"fk_account_id_from,omitempty"`
	TransferAt		time.Time 	`json:"transfer_at,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	Status			string  	`json:"status,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	AccountIDTo		string		`json:"account_id_to,omitempty"`
	FkAccountIDTo	int			`json:"fk_account_id_to,omitempty"`
}