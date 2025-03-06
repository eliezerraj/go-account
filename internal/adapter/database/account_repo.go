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

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("adapter", "database").Logger()

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Debug().Msg("NewWorkerRepository")

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

func (w WorkerRepository) AddAccount(ctx context.Context, tx pgx.Tx, account *model.Account) (*model.Account, error){
	childLogger.Debug().Msg("AddAccount")

	// trace
	span := tracerProvider.Span(ctx, "database.AddAccount")
	defer span.End()

	//Prepare
	var id int
	account.CreateAt = time.Now()

	// Query Execute
	query := `INSERT INTO account ( account_id, 
									person_id, 
									create_at,
									tenant_id,
									user_last_update) 
				VALUES($1, $2, $3, $4, $5) RETURNING id`

	row := tx.QueryRow(ctx, query,account.AccountID, 
									account.PersonID,
									account.CreateAt,
									account.TenantID,
									account.TenantID)
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	// Set PK
	account.ID = id
	return account , nil
}

func (w WorkerRepository) GetAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Debug().Msg("GetAccount")
	
	// Trace
	span := tracerProvider.Span(ctx, "database.GetAccount")
	defer span.End()

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
					create_at, 
					update_at, 
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
							&res_account.CreateAt,
							&res_account.UpdateAt,
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

func (w WorkerRepository) ListAccountPerPerson(ctx context.Context, account *model.Account) (*[]model.Account, error){
	childLogger.Debug().Msg("ListAccount")
	
	// Trace
	span := tracerProvider.Span(ctx, "database.ListAccount")
	defer span.End()

	// Prepare
	res_account := model.Account{}
	res_account_list := []model.Account{}

	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Query and Execute
	query := `SELECT 	id, 
						account_id, 
						person_id, 
						create_at, 
						update_at, 
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
							&res_account.CreateAt,
							&res_account.UpdateAt,
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

func (w WorkerRepository) UpdateAccount(ctx context.Context, tx pgx.Tx, account *model.Account) (int64, error){
	childLogger.Debug().Msg("UpdateAccount")

	// trace
	span := tracerProvider.Span(ctx, "database.UpdateAccount")
	defer span.End()

	// Prepare
	account.CreateAt = time.Now()
	account.UserLastUpdate = nil

	//Query Execute
	query := `Update account
				set person_id = $1, 
					update_at = $2,
					user_last_update =$3,
					tenant_id = $4
				where account_id = $5 `

	row, err := tx.Exec(ctx, query, account.PersonID,
									account.CreateAt,
									account.UserLastUpdate,
									account.TenantID,
									account.AccountID)
	if err != nil {
		return 0, errors.New(err.Error())
	}

	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return row.RowsAffected() , nil
}

func (w WorkerRepository) DeleteAccount(ctx context.Context, account *model.Account) (bool, error){
	childLogger.Debug().Msg("Delete")

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