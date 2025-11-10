package model

import (
	"time"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

type AppServer struct {
	InfoPod 		*InfoPod 					`json:"info_pod"`
	Server     		*Server     				`json:"server"`
	ConfigOTEL		*go_core_observ.ConfigOTEL	`json:"otel_config"`
	DatabaseConfig	*go_core_pg.DatabaseConfig  `json:"database"`
	CacheConfig		*CacheConfig				`json:"cache_config"`
}

type CacheConfig struct {
	Host		string `json:"host"`
	Username	string `json:"username"`
	Password	string `json:"password"`
}

type InfoPod struct {
	PodName				string 	`json:"pod_name"`
	ApiVersion			string 	`json:"version"`
	OSPID				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	AvailabilityZone 	string 	`json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	AccountID			string `json:"account_id,omitempty"`
}

type Server struct {
	Port 			int `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
}

type MessageRouter struct {
	Message			string `json:"message"`
}

type Account struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
}

type AccountStatement struct {
	ID				int			`json:"id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	ChargedAt		time.Time 	`json:"charged_at,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
	TransactionID	*string  	`json:"transaction_id,omitempty"`
	Obs				string  	`json:"obs,omitempty"`
}

type MovimentAccount struct {
	AccountBalance					*AccountBalance		`json:"account_balance,omitempty"`
	AccountBalanceStatementCredit	float64			`json:"account_balance_statement_credit,omitempty"`
	AccountBalanceStatementDebit	float64			`json:"account_balance_statement_debit,omitempty"`
	AccountBalanceStatementTotal	float64			`json:"account_balance_debit.debit_total,omitempty"`
	AccountStatement				*[]AccountStatement	`json:"account_statement,omitempty"`
}

type Transfer struct {
	ID				int			`json:"id,omitempty"`
	AccountFrom		AccountBalance	`json:"account_from,omitempty"`
	AccountTo		AccountBalance	`json:"account_to,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`
	TransferAt		time.Time 	`json:"transfer_at,omitempty"`
	Type			string  	`json:"type_charge,omitempty"`
	Status			string  	`json:"status,omitempty"`
}

type AccountBalance struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	FkAccountID		int			`json:"fk_account_id,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64 	`json:"amount"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
	JwtId			*string  	`json:"jwt_id,omitempty"`
	RequestId		*string  	`json:"request_id,omitempty"`
	TransactionID	*string  	`json:"transaction_id,omitempty"`
	TenantID		string  	`json:"tenant_id,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`
}