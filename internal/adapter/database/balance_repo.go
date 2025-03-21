package database

import (
	"context"
	"time"
	"errors"

	"github.com/go-account/internal/core/erro"	
	"github.com/go-account/internal/core/model"
	"github.com/jackc/pgx/v5"
)

// About add a account balance
func (w WorkerRepository) AddAccountBalance(ctx context.Context, tx pgx.Tx, accountBalance *model.AccountBalance) (*model.AccountBalance, error){
	childLogger.Info().Str("func","AddAccountBalance").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	//Trace
	span := tracerProvider.Span(ctx, "database.AddAccountBalance")
	defer span.End()

	// Prepare 
	var id int
	accountBalance.CreateAt = time.Now()
	accountBalance.UserLastUpdate = nil

	// Query and Execute
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
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	// Set PK
	accountBalance.ID = id

	return accountBalance , nil
}

// About update the account balance
func (w WorkerRepository) UpdateAccountBalance(ctx context.Context, tx pgx.Tx, accountBalance *model.AccountBalance) (int64, error){
	childLogger.Info().Str("func","UpdateAccountBalance").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// Trace
	span := tracerProvider.Span(ctx, "database.UpateAccountBalance")
	defer span.End()

	// Prepare
	updateAt := time.Now()
	accountBalance.UpdateAt = &updateAt

	// Query and Execute
	query := `Update ACCOUNT_BALANCE
				set amount = amount + $1, 
					update_at = $2,
					request_id = $3,
					jwt_id	= $4,
					transaction_id =$6
			where fk_account_id = $5 `

	row, err := tx.Exec(ctx, query, accountBalance.Amount,  
									accountBalance.UpdateAt,
									accountBalance.RequestId,
									accountBalance.JwtId,
									accountBalance.FkAccountID,
									accountBalance.TransactionID)
	if err != nil {
		return 0, errors.New(err.Error())
	}

	childLogger.Debug().Int("rowsAffected : ",int(row.RowsAffected())).Msg("")

	return row.RowsAffected() , nil
}

// About get all account balance
func (w WorkerRepository) GetAccountBalance(ctx context.Context, accountBalance *model.AccountBalance) (*model.AccountBalance, error){
	childLogger.Info().Str("func","GetAccountBalance").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.GetAccountBalance")
	defer span.End()

	// db connetion
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Prepare
	res_accountBalance := model.AccountBalance{}

	// Query and Execute
	query := `select a.account_id, 
					b.fk_account_id,
					b.currency, 
					b.amount, 
					b.create_at 
				from account a,
					account_balance b
				where account_id = $1 and a.id = b.fk_account_id`

	rows, err := conn.Query(ctx, query, accountBalance.AccountID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( &res_accountBalance.AccountID,
							&res_accountBalance.FkAccountID, 
							&res_accountBalance.Currency, 
							&res_accountBalance.Amount, 
							&res_accountBalance.CreateAt,
		) 
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_accountBalance, nil
	}
	
	return accountBalance, nil
}

// About get the sum of all stalments of a account
func (w WorkerRepository) GetSumAccountBalance(ctx context.Context, type_charge string, accountBalance *model.AccountBalance) (*model.AccountBalance, error){
	childLogger.Info().Str("func","GetSumAccountBalance").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.GetSumAccountBalance")
	defer span.End()

	// db connection
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Prepare
	res_accountBalance := model.AccountBalance{}

	// Query and Execute
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
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	if rows == nil {
		return nil, erro.ErrNotFound
	}

	for rows.Next() {
		err := rows.Scan( &res_accountBalance.ID,
							&res_accountBalance.Amount,
		) 
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_accountBalance, nil
	}
	
	return nil, erro.ErrNotFound
}

// About list all account stalements from account
func (w WorkerRepository) ListAccountBalance(ctx context.Context, accountBalance *model.AccountBalance) (*[]model.AccountStatement, error){
	childLogger.Info().Str("func","ListAccountBalance").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.ListAccountBalance")
	defer span.End()

	// db connection
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Prepare
	res_accountStatement := model.AccountStatement{}
	res_accountStatement_list := []model.AccountStatement{}

	// Query and Execute
	query := `select a.account_id,
					a.person_id,
					b.type_charge,
					b.currency,
					b.amount,
					b.charged_at,
					b.transaction_id
				from account a,
					account_statement b
				where account_id = $1 and a.id = b.fk_account_id
				order by charged_at desc
				limit 10 `

	rows, err := conn.Query(ctx, query, accountBalance.AccountID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_accountStatement.AccountID, 
							&res_accountStatement.PersonID, 
							&res_accountStatement.Type, 
							&res_accountStatement.Currency, 
							&res_accountStatement.Amount, 
							&res_accountStatement.ChargeAt,
							&res_accountStatement.TransactionID,
		)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		res_accountStatement_list = append(res_accountStatement_list, res_accountStatement)
	}

	return &res_accountStatement_list, nil
}

// About create a uuid transaction
func (w WorkerRepository) GetTransactionUUID(ctx context.Context) (*string, error){
	childLogger.Info().Str("func","GetTransactionUUID").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()
	
	// Trace
	span := tracerProvider.Span(ctx, "database.GetTransactionUUID")
	defer span.End()

	//db connection
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// Prepare
	var uuid string

	// Query and Execute
	query := `SELECT uuid_generate_v4()`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&uuid) 
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &uuid, nil
	}
	
	return &uuid, nil
}