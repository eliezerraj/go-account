package postgre

import (
	"context"
	"time"
	"errors"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/go-account/internal/core"
	"github.com/go-account/internal/erro"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (w WorkerRepository) CreateFundBalanceAccount(ctx context.Context,  tx *sql.Tx, accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("CreateFundBalanceAccount")

	_, root := xray.BeginSubsegment(ctx, "Repository.CreateFundBalanceAccount")
	defer func() {
		root.Close(nil)
	}()

	stmt, err := tx.Prepare(`INSERT INTO ACCOUNT_BALANCE ( 	fk_account_id, 
															currency, 
															amount,
															tenant_id,
															create_at,
															user_last_update) 
									VALUES($1, $2, $3, $4, $5, $6) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}

	_, err = stmt.ExecContext(	ctx,	
								accountBalance.FkAccountID, 
								accountBalance.Currency,
								accountBalance.Amount,
								accountBalance.TenantID,
								time.Now(),
								"NA")
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}

	defer stmt.Close()
	return &accountBalance , nil
}

func (w WorkerRepository) UpdateFundBalanceAccount(ctx context.Context, tx *sql.Tx, accountBalance core.AccountBalance) (int64, error){
	childLogger.Debug().Msg("UpdateFundBalanceAccount")
	//childLogger.Debug().Interface("==>>accountBalance : ", accountBalance).Msg("")

	_, root := xray.BeginSubsegment(ctx, "Repository.UpdateFundBalanceAccount")
	defer func() {
		root.Close(nil)
	}()

	stmt, err := tx.Prepare(`Update ACCOUNT_BALANCE
									set amount = amount + $1, 
										update_at = $2
								where fk_account_id = $3 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("Update statement")
		return 0, errors.New(err.Error())
	}

	result, err := stmt.ExecContext(ctx,	
									accountBalance.Amount, 
									time.Now(),
									accountBalance.FkAccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return 0, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	defer stmt.Close()
	return rowsAffected , nil
}

func (w WorkerRepository) GetFundBalanceAccount(ctx context.Context, accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("GetFundBalanceAccount")

	_, root := xray.BeginSubsegment(ctx, "Repository.GetFundBalanceAccount")
	defer func() {
		root.Close(nil)
	}()

	client := w.databaseHelper.GetConnection()

	result_accountBalance := core.AccountBalance{}
	rows, err := client.QueryContext(ctx, `select a.account_id, b.fk_account_id ,b.currency , b.amount, b.create_at 
											from account a,
												account_balance b
											where account_id = $1
											and a.id = b.fk_account_id`, accountBalance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}

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

	defer rows.Close()
	accountBalance.Amount=0
	return &accountBalance, nil
}

func (w WorkerRepository) ListAccountStatementMoviment(ctx context.Context, accountBalance core.AccountBalance) (*[]core.AccountStatement, error){
	childLogger.Debug().Msg("ListAccountStatementMoviment")
	_, root := xray.BeginSubsegment(ctx, "Repository.ListAccountStatementMoviment")
	defer func() {
		root.Close(nil)
	}()

	client := w.databaseHelper.GetConnection()

	result_accountStatement := core.AccountStatement{}
	accountStatement_list := []core.AccountStatement{}

	rows, err := client.QueryContext(ctx, `select 	a.account_id,
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
											limit 10 `, accountBalance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_accountStatement.AccountID, 
							&result_accountStatement.PersonID, 
							&result_accountStatement.Type, 
							&result_accountStatement.Currency, 
							&result_accountStatement.Amount, 
							&result_accountStatement.ChargeAt, 
							)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		accountStatement_list = append(accountStatement_list, result_accountStatement)
	}

	defer rows.Close()
	return &accountStatement_list, nil
}

func (w WorkerRepository) GetFundBalanceAccountStatementMoviment(ctx context.Context, type_charge string , accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("GetFundBalanceAccountStatementMoviment:"+type_charge)

	_, root := xray.BeginSubsegment(ctx, "Repository.GetFundBalanceAccountStatementMoviment:"+type_charge)
	defer func() {
		root.Close(nil)
	}()

	client := w.databaseHelper.GetConnection()

	result_accountBalance := core.AccountBalance{}
	rows, err := client.QueryContext(ctx, `Select a.id, 
												sum(b.amount)
											from account a,
												account_statement b
											where account_id = $1
											and a.id = b.fk_account_id
											and b.type_charge = $2
											group by a.id`, accountBalance.AccountID, type_charge)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}

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

	defer rows.Close()
	return nil, erro.ErrNotFound
}

func (w WorkerRepository) TransferFundAccount(ctx context.Context, tx *sql.Tx, transfer core.Transfer) (int64, string ,error){
	childLogger.Debug().Msg("TransferFundAccount")

	_, root := xray.BeginSubsegment(ctx, "Repository.TransferFundAccount")
	defer func() {
		root.Close(nil)
	}()

	rows_uuid, err := tx.QueryContext(ctx, "SELECT uuid_generate_v4()")
	if err != nil {
		childLogger.Error().Err(err).Msg("ERROR QueryContext UUID")
		return 0,"" ,errors.New(err.Error())
	}
	var uuid string
	for rows_uuid.Next() {
		err := rows_uuid.Scan( &uuid )
		if err != nil {
			childLogger.Error().Err(err).Msg("Erro Scan rows_uuid")
			return 0, "" ,errors.New(err.Error())
        }
	}

	stmt, err := tx.Prepare(`Update ACCOUNT_BALANCE
								set transaction_id = $2,
									update_at =$3,
									amount = amount + $4
								where fk_account_id = $1 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("Update statement")
		return 0, "", errors.New(err.Error())
	}

	result, err := stmt.ExecContext(ctx,
									transfer.FkAccountIDFrom,	
									uuid, 
									time.Now(),
									transfer.Amount)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return 0, "",errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	defer stmt.Close()
	return rowsAffected, uuid ,nil
}

func (w WorkerRepository) AddAccountStatement(ctx context.Context, tx *sql.Tx ,credit core.AccountStatement) (*core.AccountStatement, error){
	childLogger.Debug().Msg("AddAccountStatement")

	_, root := xray.BeginSubsegment(ctx, "Repository.AddAccountStatement")
	defer func() {
		root.Close(nil)
	}()

	stmt, err := tx.Prepare(`INSERT INTO account_statement ( 	fk_account_id, 
																type_charge,
																charged_at, 
																currency,
																amount,
																tenant_id) 
									VALUES($1, $2, $3, $4, $5, $6) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}

	_, err = stmt.ExecContext(	ctx,
								credit.FkAccountID, 
								credit.Type,
								time.Now(),
								credit.Currency,
								credit.Amount,
								credit.TenantID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}

	credit.ChargeAt = time.Now()

	defer stmt.Close()
	return &credit , nil
}

func (w WorkerRepository) CommitTransferFundAccount(ctx context.Context, tx *sql.Tx, uuid string ,transfer core.Transfer) (int64 ,error){
	childLogger.Debug().Msg("CommitTransferFundAccount")

	_, root := xray.BeginSubsegment(ctx, "Repository.CommitTransferFundAccount")
	defer func() {
		root.Close(nil)
	}()

	stmt, err := tx.Prepare(`Update ACCOUNT_BALANCE
								set update_at =$3
							where fk_account_id = $1 
							and transaction_id = $2`)
	if err != nil {
		childLogger.Error().Err(err).Msg("Update statement")
		return 0, errors.New(err.Error())
	}

	result, err := stmt.ExecContext(ctx,
									transfer.FkAccountIDFrom,	
									uuid, 
									time.Now(),)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return 0, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	defer stmt.Close()
	return rowsAffected ,nil
}