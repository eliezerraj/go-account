package postgre

import (
	"context"
	"time"
	"errors"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/go-account/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (w WorkerRepository) CreateFundBalanceAccount(ctx context.Context,  tx *sql.Tx, accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("CreateFundBalanceAccount")

	_, root := xray.BeginSubsegment(ctx, "SQL.CreateFundBalanceAccount")
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

	childLogger.Debug().Interface("==>>accountBalance : ", accountBalance).Msg("")

	_, root := xray.BeginSubsegment(ctx, "SQL.UpdateFundBalanceAccount")
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