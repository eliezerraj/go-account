package storage

import (
	"context"
	"time"
	"errors"

	"github.com/go-account/internal/repository/pg"
	"github.com/go-account/internal/core"
	"github.com/go-account/internal/lib"
	"github.com/go-account/internal/erro"

	"github.com/rs/zerolog/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var childLogger = log.With().Str("repository.pg", "storage").Logger()

type WorkerRepository struct {
	databasePG pg.DatabasePG
}

func NewWorkerRepository(databasePG pg.DatabasePG) WorkerRepository {
	childLogger.Debug().Msg("NewWorkerRepository")
	return WorkerRepository{
		databasePG: databasePG,
	}
}

func (w WorkerRepository) SetSessionVariable(ctx context.Context, userCredential string) (bool, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("SetSessionVariable")

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return false, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)
	
	_, err = conn.Query(ctx, "SET sess.user_credential to '" + userCredential+ "'")
	if err != nil {
		childLogger.Error().Err(err).Msg("SET SESSION statement ERROR")
		return false, errors.New(err.Error())
	}

	return true, nil
}

func (w WorkerRepository) GetSessionVariable(ctx context.Context) (*string, error) {
	childLogger.Debug().Msg("++++++++++++++++++++++++++++++++")
	childLogger.Debug().Msg("GetSessionVariable")

	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	defer w.databasePG.Release(conn)

	var res_balance string
	rows, err := conn.Query(ctx, "SELECT current_setting('sess.user_credential')" )
	if err != nil {
		childLogger.Error().Err(err).Msg("Prepare statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&res_balance)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &res_balance, nil
	}

	return nil, erro.ErrNotFound
}

func (w WorkerRepository) StartTx(ctx context.Context) (pgx.Tx, *pgxpool.Conn, error) {
	childLogger.Debug().Msg("StartTx")

	span := lib.Span(ctx, "repo.StartTx")
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, nil, errors.New(err.Error())
	}
	span.End()
	
	tx, err := conn.Begin(ctx)
    if err != nil {
        return nil, nil ,errors.New(err.Error())
    }

	return tx, conn, nil
}

func (w WorkerRepository) ReleaseTx(connection *pgxpool.Conn) {
	childLogger.Debug().Msg("ReleaseTx")

	defer connection.Release()
}
//-----------------------------------------------
func (w WorkerRepository) List(ctx context.Context, account *core.Account) (*[]core.Account, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "repo.List")	
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

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

func (w WorkerRepository) Get(ctx context.Context, account *core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Get")

	span := lib.Span(ctx, "repo.Get")	
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

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
							&result_query.UserLastUpdate)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }

		return &result_query, nil
	}
	
	return nil, erro.ErrNotFound
}

func (w WorkerRepository) GetId(ctx context.Context, account *core.Account) (*core.Account, error){
	childLogger.Debug().Msg("GetId")

	span := lib.Span(ctx, "repo.GetId")	
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

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

func (w WorkerRepository) Add(ctx context.Context, account *core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "repo.Add")	
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

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
	return account , nil
}

func (w WorkerRepository) Update(ctx context.Context, account *core.Account) (bool, error){
	childLogger.Debug().Msg("Update")

	span := lib.Span(ctx, "repo.Update")	
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return false, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

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

	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return true , nil
}

func (w WorkerRepository) Delete(ctx context.Context, account *core.Account) (bool, error){
	childLogger.Debug().Msg("Delete")

	span := lib.Span(ctx, "repo.Delete")	
	defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return false, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

	query := `Delete from account where id = $1`

	_, err = conn.Exec(ctx, query, account.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}
		
	return true , nil
}
//------------------------------------------
func (w WorkerRepository) CreateFundBalanceAccount(ctx context.Context, tx pgx.Tx, accountBalance *core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("CreateFundBalanceAccount")

	span := lib.Span(ctx, "repo.CreateFundBalanceAccount")	
    defer span.End()

	accountBalance.CreateAt = time.Now()
	accountBalance.UserLastUpdate = nil

	query := `INSERT INTO ACCOUNT_BALANCE ( fk_account_id, 
											currency, 
											amount,
											tenant_id,
											create_at,
											user_last_update) 
				VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	row := tx.QueryRow(ctx, query, 	accountBalance.FkAccountID, 
									accountBalance.Currency, 
									accountBalance.Amount, 
									accountBalance.TenantID,
									accountBalance.CreateAt, 
									accountBalance.UserLastUpdate)								
	
	var id int
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}

	accountBalance.ID = id

	return accountBalance , nil
}

func (w WorkerRepository) GetFundBalanceAccount(ctx context.Context, accountBalance *core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("GetFundBalanceAccount")

	span := lib.Span(ctx, "repo.GetFundBalanceAccount")	
    defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

	result_accountBalance := core.AccountBalance{}

	query := `select a.account_id, 
					b.fk_account_id,
					b.currency, 
					b.amount, 
					b.create_at 
				from account a,
					account_balance b
				where account_id = $1
				and a.id = b.fk_account_id`

	rows, err := conn.Query(ctx, query, accountBalance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&result_accountBalance.AccountID,
							&result_accountBalance.FkAccountID, 
							&result_accountBalance.Currency, 
							&result_accountBalance.Amount, 
							&result_accountBalance.CreateAt, 
							)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &result_accountBalance, nil
	}

	//accountBalance.Amount=0	
	return accountBalance, nil
}

func (w WorkerRepository) ListAccountStatementMoviment(ctx context.Context, accountBalance *core.AccountBalance) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("ListAccountStatementMoviment")

	span := lib.Span(ctx, "repo.ListAccountStatementMoviment")	
    defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

	result_accountStatement := core.AccountStatement{}
	accountStatement_list := []core.AccountStatement{}

	query := `select a.account_id,
					a.person_id,
					b.type_charge,
					b.currency,
					b.amount,
					b.charged_at
				from account a,
					account_statement b
				where account_id = $1
				and a.id = b.fk_account_id
				order by charged_at desc
				limit 10 `

	rows, err := conn.Query(ctx, query, accountBalance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&result_accountStatement.AccountID, 
							&result_accountStatement.PersonID, 
							&result_accountStatement.Type, 
							&result_accountStatement.Currency, 
							&result_accountStatement.Amount, 
							&result_accountStatement.ChargeAt )
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		accountStatement_list = append(accountStatement_list, result_accountStatement)
	}

	return &accountStatement_list, nil
}

func (w WorkerRepository) GetFundBalanceAccountStatementMoviment(ctx context.Context, type_charge string, accountBalance *core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("GetFundBalanceAccountStatementMoviment:"+type_charge)

	span := lib.Span(ctx, "repo.GetFundBalanceAccountStatementMoviment")	
    defer span.End()

	span = lib.Span(ctx, "repo.Acquire")
	conn, err := w.databasePG.Acquire(ctx)
	if err != nil {
		childLogger.Error().Err(err).Msg("Erro Acquire")
		return nil, errors.New(err.Error())
	}
	span.End()
	defer w.databasePG.Release(conn)

	result_accountBalance := core.AccountBalance{}

	query := `Select a.id, 
					sum(b.amount)
				from account a,
					account_statement b
				where account_id = $1
				and a.id = b.fk_account_id
				and b.type_charge = $2
				group by a.id`

	rows, err := conn.Query(ctx, query, accountBalance.AccountID, type_charge)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	if rows == nil {
		return nil, erro.ErrNotFound
	}
	for rows.Next() {
		err := rows.Scan( &result_accountBalance.ID,
						  &result_accountBalance.Amount )
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &result_accountBalance, nil
	}

	return nil, erro.ErrNotFound
}

func (w WorkerRepository) AddAccountStatement(ctx context.Context, tx pgx.Tx, accountStatement *core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("AddAccountStatement")

	span := lib.Span(ctx, "repo.AddAccountStatement")	
    defer span.End()

	query := `INSERT INTO account_statement (fk_account_id, 
											type_charge,
											charged_at, 
											currency,
											amount,
											tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

	accountStatement.ChargeAt = time.Now()

	row := tx.QueryRow(ctx, query, accountStatement.FkAccountID,  
									accountStatement.Type,
									accountStatement.ChargeAt,
									accountStatement.Currency,
									accountStatement.Amount,
									accountStatement.TenantID)

	var id int
	if err := row.Scan(&id); err != nil {
		childLogger.Error().Err(err).Msg("QueryRow INSERT")
		return nil, errors.New(err.Error())
	}

	accountStatement.ID = id

	return accountStatement , nil
}

func (w WorkerRepository) CommitTransferFundAccount(ctx context.Context, tx pgx.Tx, uuid string, transfer *core.Transfer) (int64 ,error){
	childLogger.Debug().Msg("CommitTransferFundAccount")

	span := lib.Span(ctx, "repo.CommitTransferFundAccount")	
    defer span.End()

	query := `Update ACCOUNT_BALANCE
					set update_at =$3
				where fk_account_id = $1 
				and transaction_id = $2`

	transfer.TransferAt = time.Now()

	row, err := tx.Exec(ctx, query, transfer.FkAccountIDFrom,  
									uuid,
									transfer.TransferAt)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return 0, errors.New(err.Error())
	}
	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return row.RowsAffected(), nil
}

func (w WorkerRepository) UpdateFundBalanceAccount(ctx context.Context, tx pgx.Tx, accountBalance *core.AccountBalance) (int64, error){
	childLogger.Debug().Msg("UpdateFundBalanceAccount")
	childLogger.Debug().Interface("==>>accountBalance : ", accountBalance).Msg("")

	span := lib.Span(ctx, "repo.UpdateFundBalanceAccount")	
    defer span.End()

	query := `Update ACCOUNT_BALANCE
				set amount = amount + $1, 
					update_at = $2,
					request_id = $3,
					jwt_id	= $4
			where fk_account_id = $5 `

	updateAt := time.Now()
	accountBalance.UpdateAt = &updateAt

	row, err := tx.Exec(ctx, query, accountBalance.Amount,  
									accountBalance.UpdateAt,
									accountBalance.RequestId,
									accountBalance.JwtId,
									accountBalance.FkAccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Update statement")
		return 0, errors.New(err.Error())
	}

	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return row.RowsAffected() , nil
}

func (w WorkerRepository) TransferFundAccount(ctx context.Context, tx pgx.Tx, transfer *core.Transfer) (int64, string ,error){
	childLogger.Debug().Msg("TransferFundAccount")

	span := lib.Span(ctx, "repo.TransferFundAccount")	
    defer span.End()

	query := `SELECT uuid_generate_v4()`

	rows_uuid, err := tx.Query(ctx, query)
	if err != nil {
		childLogger.Error().Err(err).Msg("ERROR QueryContext UUID")
		return 0,"" ,errors.New(err.Error())
	}

	var uuid string
	for rows_uuid.Next() {
		err := rows_uuid.Scan( &uuid )
		if err != nil {
			childLogger.Error().Err(err).Msg("Erro Scan rows_uuid")
			return 0, uuid ,errors.New(err.Error())
        }
	}

	query = `Update ACCOUNT_BALANCE
				set transaction_id = $2,
					update_at =$3,
					amount = amount + $4
			where fk_account_id = $1 `

	transfer.TransferAt = time.Now()

	row, err := tx.Exec(ctx, query, transfer.FkAccountIDFrom,
									uuid,  
									transfer.TransferAt,
									transfer.Amount)
	if err != nil {
		childLogger.Error().Err(err).Msg("Update statement")
		return 0, uuid, errors.New(err.Error())
	}

	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return row.RowsAffected(), uuid, nil
}