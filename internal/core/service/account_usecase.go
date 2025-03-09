package service

import(
	"context"

	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

func (s *WorkerService) AddAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Debug().Msg("AddAccount")
	childLogger.Debug().Interface("account: ", account).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.AddAccount")
	
	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// Add the account
	res, err := s.workerRepository.AddAccount(ctx, tx, account)
	if err != nil {
		return nil, err
	}

	// Create the Balance Account
	accountBalance := model.AccountBalance{}
	accountBalance.Amount = 0
	accountBalance.Currency = "BRL"
	accountBalance.AccountID = res.AccountID
	accountBalance.FkAccountID = res.ID
	accountBalance.TenantID = res.TenantID

	// Try to update the account_balance
	res_update, err := s.workerRepository.UpdateAccountBalance(ctx, tx, &accountBalance)
	if err != nil {
		return nil, err
	}

	// If the account_balance so it created one
	if res_update == 0 {
		_, err = s.workerRepository.AddAccountBalance(ctx, tx, &accountBalance)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *WorkerService) UpdateAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Debug().Msg("UpdateAccount")
	childLogger.Debug().Interface("account: ", account).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.UpdateAccount")
	
	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	// Get account (check if exists)
	res, err := s.workerRepository.GetAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	// Update the account
	res_update, err := s.workerRepository.UpdateAccount(ctx, tx, account)
	if err != nil {
		return nil, err
	}
	if (res_update == 0) {
		return nil, erro.ErrUpdate
	}

	return res, nil
}

func (s *WorkerService) DeleteAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Debug().Msg("DeleteAccount")
	childLogger.Debug().Interface("account: ", account).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.UpdateAccount")
	defer span.End()
	
	// Get account (check if exists)
	res, err := s.workerRepository.GetAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	// Delete the account
	_, err = s.workerRepository.DeleteAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *WorkerService) GetAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Debug().Msg("GetAccount")
	childLogger.Debug().Interface("account: ", account).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.GetAccount")
	defer span.End()
	
	// Get account
	res, err := s.workerRepository.GetAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *WorkerService) ListAccountPerPerson(ctx context.Context, account *model.Account) (*[]model.Account, error){
	childLogger.Debug().Msg("ListAccountPerPerson")
	childLogger.Debug().Interface("account: ", account).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.ListAccountPerPerson")
	defer span.End()
	
	// List account
	res, err := s.workerRepository.ListAccountPerPerson(ctx, account)
	if err != nil {
		return nil, err
	}
	return res, nil
}