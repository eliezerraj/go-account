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
	rows, err := client.QueryContext(ctx, `select b.currency , b.amount 
											from account a,
												account_balance b,
											where account_id = $1
											and a.id = b.fk_account_id`, accountBalance.AccountID)
	if err != nil {
		childLogger.Error().Err(err).Msg("Query statement")
		return nil, errors.New(err.Error())
	}

	for rows.Next() {
		err := rows.Scan( 	&result_accountBalance.Currency, 
							&result_accountBalance.Amount, 
							)
		if err != nil {
			childLogger.Error().Err(err).Msg("Scan statement")
			return nil, errors.New(err.Error())
        }
		return &result_accountBalance, nil
	}

	defer rows.Close()
	return nil, erro.ErrNotFound
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
											order by charged_at desc`, accountBalance.AccountID)
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