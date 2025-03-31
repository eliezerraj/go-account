package database

import (
	"context"
	"time"
	"errors"
	
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-account").Str("package","internal.adapter.database").Logger()

var tracerProvider go_core_observ.TracerProvider

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Info().Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

// About create a account
func (w WorkerRepository) AddAccount(ctx context.Context, tx pgx.Tx, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","AddAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.AddAccount")
	defer span.End()

	//Prepare
	var id int
	account.CreatedAt = time.Now()

	// Query Execute
	query := `INSERT INTO account ( account_id, 
									person_id, 
									created_at,
									tenant_id) 
				VALUES($1, $2, $3, $4) RETURNING id`

	row := tx.QueryRow(ctx, query,account.AccountID, 
									account.PersonID,
									account.CreatedAt,
									account.TenantID)
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	// Set PK
	account.ID = id
	return account , nil
}

// About get an account
func (w WorkerRepository) GetAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","GetAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// Trace
	span := tracerProvider.Span(ctx, "database.GetAccount")
	defer span.End()

	// db connection
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Prepare
	res_account := model.Account{}

	// Query and Execute
	query := `SELECT id, 
					account_id, 
					person_id, 
					created_at, 
					updated_at, 
					tenant_id, 
					user_last_update 
				FROM account 
				WHERE account_id =$1`

	rows, err := conn.Query(ctx, query, account.AccountID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_account.ID, 
							&res_account.AccountID, 
							&res_account.PersonID, 
							&res_account.CreatedAt,
							&res_account.UpdatedAt,
							&res_account.TenantID,
							&res_account.UserLastUpdate,
							)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_account, nil
	}
	
	return nil, erro.ErrNotFound
}

// About get all account per person
func (w WorkerRepository) ListAccountPerPerson(ctx context.Context, account *model.Account) (*[]model.Account, error){
	childLogger.Info().Str("func","ListAccountPerPerson").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.ListAccount")
	defer span.End()

	// Prepare
	res_account := model.Account{}
	res_account_list := []model.Account{}

	// db connection
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Query and Execute
	query := `SELECT 	id, 
						account_id, 
						person_id, 
						created_at, 
						updated_at, 
						user_last_update,
						tenant_id 
						FROM account 
						WHERE person_id =$1 order by id desc`

	rows, err := conn.Query(ctx, query, account.PersonID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_account.ID, 
							&res_account.AccountID, 
							&res_account.PersonID, 
							&res_account.CreatedAt,
							&res_account.UpdatedAt,
							&res_account.UserLastUpdate,
							&res_account.TenantID,
						)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_account_list = append(res_account_list, res_account)
	}
	
	return &res_account_list , nil
}

// About update an account
func (w WorkerRepository) UpdateAccount(ctx context.Context, tx pgx.Tx, account *model.Account) (int64, error){
	childLogger.Info().Str("func","UpdateAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.UpdateAccount")
	defer span.End()

	// Prepare
	updateAt := time.Now()
	account.UpdatedAt = &updateAt

	//Query Execute
	query := `Update account
				set person_id = $1, 
					updated_at = $2,
					user_last_update =$3,
					tenant_id = $4
				where account_id = $5 `

	row, err := tx.Exec(ctx, query, account.PersonID,
									account.UpdatedAt,
									account.UserLastUpdate,
									account.TenantID,
									account.AccountID)
	if err != nil {
		return 0, errors.New(err.Error())
	}

	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return row.RowsAffected() , nil
}

// About delete an account
func (w WorkerRepository) DeleteAccount(ctx context.Context, account *model.Account) (bool, error){
	childLogger.Info().Str("func","DeleteAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	span := tracerProvider.Span(ctx, "storage.DeleteAccount")	
	defer span.End()

	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return false, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	query := `Delete from account where id = $1`

	_, err = conn.Exec(ctx, query, account.AccountID)
	if err != nil {
		return false, errors.New(err.Error())
	}
		
	return true , nil
}