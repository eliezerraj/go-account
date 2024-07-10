package postgre

import (
	"context"
	"time"
	"errors"

	_ "github.com/lib/pq"

	"github.com/go-account/internal/core"
	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/lib"
)

func (w WorkerRepository) Add(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "repo.Account.Add")	
	defer span.End()

	client := w.databaseHelper.GetConnection()

	stmt, err := client.Prepare(`INSERT INTO account ( 	account_id, 
														person_id, 
														create_at,
														tenant_id,
														user_last_update) 
									VALUES($1, $2, $3, $4, $5) `)
	if err != nil {
		childLogger.Error().Err(err).Msg("INSERT statement")
		return nil, errors.New(err.Error())
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(	ctx,	
								account.AccountID, 
								account.PersonID,
								time.Now(),
								account.TenantID,
								"NA")
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return nil, errors.New(err.Error())
	}

	return &account , nil
}

func (w WorkerRepository) Get(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Get")

	span := lib.Span(ctx, "repo.Account.Get")	
	defer span.End()

	client := w.databaseHelper.GetConnection()

	result_query := core.Account{}

	//firstQueryStart  := time.Now()
	rows, err := client.QueryContext(ctx, `SELECT id, account_id, person_id, create_at, update_at, tenant_id, user_last_update FROM account WHERE account_id =$1`, account.AccountID)
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
		//firstQueryEnd := time.Now()
		//fmt.Println(fmt.Sprintf(" ==== > query took %s",  firstQueryEnd.Sub(firstQueryStart).String()))

		return &result_query, nil
	}
	
	return nil, erro.ErrNotFound
}

func (w WorkerRepository) Update(ctx context.Context, account core.Account) (bool, error){
	childLogger.Debug().Msg("Update...")
	//childLogger.Debug().Interface("account : ", account).Msg("account")

	span := lib.Span(ctx, "repo.Account.Update")	
	defer span.End()

	client := w.databaseHelper.GetConnection()

	stmt, err := client.Prepare(`Update account
									set person_id = $1, 
										update_at = $2,
										user_last_update =$3,
										tenant_id = $4
								where account_id = $5 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("UPDATE statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx,	
									account.PersonID,
									time.Now(),
									"NA",
									account.TenantID,
									account.AccountID,
								)
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}

	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")

	return true , nil
}

func (w WorkerRepository) Delete(ctx context.Context, account core.Account) (bool, error){
	childLogger.Debug().Msg("Delete")

	span := lib.Span(ctx, "repo.Account.Delete")	
	defer span.End()

	client := w.databaseHelper.GetConnection()

	stmt, err := client.Prepare(`Delete from account where id = $1 `)
	if err != nil {
		childLogger.Error().Err(err).Msg("DELETE statement")
		return false, errors.New(err.Error())
	}

	result, err := stmt.ExecContext(ctx,account.ID )
	if err != nil {
		childLogger.Error().Err(err).Msg("Exec statement")
		return false, errors.New(err.Error())
	}
	defer stmt.Close()
	
	rowsAffected, _ := result.RowsAffected()
	childLogger.Debug().Int("rowsAffected : ",int(rowsAffected)).Msg("")
	
	return true , nil
}

func (w WorkerRepository) List(ctx context.Context, account core.Account) (*[]core.Account, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "repo.Account.List")	
	defer span.End()

	client:= w.databaseHelper.GetConnection()
	
	result_query := core.Account{}
	balance_list := []core.Account{}
	rows, err := client.QueryContext(ctx, `SELECT id, account_id, person_id, create_at, update_at, tenant_id, user_last_update FROM account WHERE person_id =$1`, account.PersonID)
	if err != nil {
		childLogger.Error().Err(err).Msg("SELECT statement")
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

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

	return &balance_list , nil
}

func (w WorkerRepository) GetId(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("GetId")

	span := lib.Span(ctx, "repo.Account.GetId")	
	defer span.End()

	client := w.databaseHelper.GetConnection()

	result_query := core.Account{}
	rows, err := client.QueryContext(ctx, `SELECT id, account_id, person_id, create_at, update_at, tenant_id, user_last_update FROM account WHERE id =$1`, account.ID)
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
