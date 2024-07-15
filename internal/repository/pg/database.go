package pg

import (
	"context"
	"fmt"
	"time"
	"errors"

	"github.com/go-account/internal/core"
	"github.com/go-account/internal/lib"
	"github.com/go-account/internal/erro"

	"github.com/rs/zerolog/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var childLogger = log.With().Str("repository.pg", "WorkerRepo").Logger()

type DatabasePG interface {
	GetConnection() (*pgxpool.Pool)
}

type DatabasePGServer struct {
	connPool   	*pgxpool.Pool
}

func NewDatabasePGServer(ctx context.Context, databaseRDS *core.DatabaseRDS) (DatabasePG, error) {
	childLogger.Debug().Msg("NewDatabasePGServer")
	
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", 
							databaseRDS.User, 
							databaseRDS.Password, 
							databaseRDS.Host, 
							databaseRDS.Port, 
							databaseRDS.DatabaseName) 
							
	connPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return DatabasePGServer{}, err
	}
	
	err = connPool.Ping(ctx)
	if err != nil {
		return DatabasePGServer{}, err
	}

	return DatabasePGServer{
		connPool: connPool,
	}, nil
}

func (d DatabasePGServer) GetConnection() (*pgxpool.Pool) {
	childLogger.Debug().Msg("GetConnection")
	return d.connPool
}

func (d DatabasePGServer) CloseConnection() {
	childLogger.Debug().Msg("CloseConnection")
	defer d.connPool.Close()
}

type WorkerRepository struct {
	databasePG DatabasePG
}

func NewWorkerRepository(databasePG DatabasePG) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databasePG: databasePG,
	}
}

func (w WorkerRepository) SetSessionVariable(ctx context.Context, userCredential string) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("SetSessionVariable")

	connPool := w.databasePG.GetConnection()
	
	_, err := connPool.Query(ctx, "SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable(ctx context.Context) (string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")

	connPool := w.databasePG.GetConnection()

	var res_balance string
	rows, err := connPool.Query(ctx, "SELECT current_setting('sess.user_credential')" )
	if err != nil {
		childLogger.Error().Err(err).Msg("Prepare statement")
		return "", errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_balance )
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return "", errors.New(err.Error())
        }
		return res_balance, nil
	}

	return "", erro.ErrNotFound
}

func (w WorkerRepository) StartTx(ctx context.Context) (pgx.Tx, error) {
	childLogger.Debug().Msg("StartTx")

	conn := w.databasePG.GetConnection()
	tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, errors.New(err.Error())
    }

	return tx, nil
}

func (w WorkerRepository) List(ctx context.Context, account core.Account) (*[]core.Account, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "repo.List")	
	defer span.End()

	conn := w.databasePG.GetConnection()

	result_query := core.Account{}
	balance_list := []core.Account{}

	query := `SELECT 	id, 
						account_id, 
						person_id, 
						create_at, 
						update_at, 
						tenant_id, 
						user_last_update 
						FROM account 
						WHERE person_id =$1`

	rows, err := conn.Query(ctx, query, account.PersonID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_query.ID, 
							&result_query.AccountID, 
							&result_query.PersonID, 
							&result_query.CreateAt,
							&result_query.UpdateAt,
							&result_query.TenantID,
							&result_query.UserLastUpdate,
						)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		balance_list = append(balance_list, result_query)
	}

	defer rows.Close()
	return &balance_list , nil
}

func (w WorkerRepository) Get(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Get")

	span := lib.Span(ctx, "repo.Get")	
	defer span.End()

	conn := w.databasePG.GetConnection()

	result_query := core.Account{}

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
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &result_query.ID, 
							&result_query.AccountID, 
							&result_query.PersonID, 
							&result_query.CreateAt,
							&result_query.UpdateAt,
							&result_query.TenantID,
							&result_query.UserLastUpdate)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }

		return &result_query, nil
	}
	
	return nil, erro.ErrNotFound
}

func (w WorkerRepository) GetId(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("GetId")

	span := lib.Span(ctx, "repo.GetId")	
	defer span.End()

	conn := w.databasePG.GetConnection()

	result_query := core.Account{}

	query := `SELECT id, 
					account_id, 
					person_id, 
					create_at, 
					update_at, 
					tenant_id, 
					user_last_update 
				FROM account 
				WHERE id =$1`

	rows, err := conn.Query(ctx, query,  account.ID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &result_query.ID, 
							&result_query.AccountID, 
							&result_query.PersonID, 
							&result_query.CreateAt,
							&result_query.UpdateAt,
							&result_query.TenantID,
							&result_query.UserLastUpdate,
							)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &result_query, nil
	}
	
	return nil, erro.ErrNotFound
}

func (w WorkerRepository) Add(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "repo.Add")	
	defer span.End()

	conn := w.databasePG.GetConnection()

	query := `INSERT INTO account ( account_id, 
									person_id, 
									create_at,
									tenant_id,
									user_last_update) 
				VALUES($1, $2, $3, $4, $5) RETURNING id`

	account.CreateAt = time.Now()
	account.TenantID = "NA"

	row := conn.QueryRow(ctx, query, account.AccountID, 
										account.PersonID,
										account.CreateAt,
										account.TenantID,
										account.TenantID)

	var id int
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("QueryRow INSERT")
		return nil, errors.New(err.Error())
	}

	account.ID = id
	return &account , nil
}

func (w WorkerRepository) Update(ctx context.Context, account core.Account) (bool, error){
	childLogger.Debug().Msg("Update")

	span := lib.Span(ctx, "repo.Update")	
	defer span.End()

	conn := w.databasePG.GetConnection()

	query := `Update account
				set person_id = $1, 
					update_at = $2,
					user_last_update =$3,
					tenant_id = $4
				where account_id = $5 `

	account.CreateAt = time.Now()
	account.UserLastUpdate = nil

	row, err := conn.Exec(ctx, query, account.PersonID,
									account.CreateAt,
									account.UserLastUpdate,
									account.TenantID,
									account.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("QueryRow Update")
		return false, errors.New(err.Error())
	}

	rowsAffected := row.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}

func (w WorkerRepository) Delete(ctx context.Context, account core.Account) (bool, error){
	childLogger.Debug().Msg("Delete")

	span := lib.Span(ctx, "repo.Delete")	
	defer span.End()

	conn := w.databasePG.GetConnection()

	query := `Delete from account where id = $1`

	_, err := conn.Exec(ctx, query, account.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}
		
	return true , nil
}

