package service

import (
	"context"
	"github.com/rs/zerolog/log"

	"github.com/go-account/internal/lib"
	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/core"
	"github.com/go-account/internal/repository/pg"
)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepo		 	*pg.WorkerRepository
}

func NewWorkerService( workerRepo *pg.WorkerRepository) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepo:		 	workerRepo,
	}
}
// -----------------------------------------------
func (s WorkerService) SetSessionVariable(ctx context.Context, userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepo.SetSessionVariable(ctx, userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Add")

	span := lib.Span(ctx, "service.Add")
	defer span.End()

	// Create account
	res, err := s.workerRepo.Add(ctx, account)
	if err != nil {
		return nil, err
	}

	// Create the Balance Account
	accountBalance := core.AccountBalance{}
	accountBalance.Amount = 0
	accountBalance.Currency = "BRL"
	accountBalance.AccountID = res.AccountID
	accountBalance.FkAccountID = res.ID
	accountBalance.TenantID = res.TenantID

	_, err = s.AddFundBalanceAccount(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Get(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Get")

	span := lib.Span(ctx, "service.Get")
	defer span.End()

	res, err := s.workerRepo.Get(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Update(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Update")

	span := lib.Span(ctx, "service.Update")
	defer span.End()

	res_account, err := s.workerRepo.Get(ctx, account)
	if err != nil {
		return nil, err
	}

	account.ID = res_account.ID
	isUpdated, err := s.workerRepo.Update(ctx, account)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_account, err = s.workerRepo.Get(ctx, account)
	if err != nil {
		return nil, err
	}
	return res_account, nil
}

func (s WorkerService) Delete(ctx context.Context,account core.Account) (bool, error){
	childLogger.Debug().Msg("Delete")

	span := lib.Span(ctx, "service.Delete")
	defer span.End()

	res_account, err := s.workerRepo.Get(ctx, account)
	if err != nil {
		return false, err
	}

	account.ID = res_account.ID
	isDelete, err := s.workerRepo.Delete(ctx, account)
	if err != nil {
		return false, err
	}
	if (isDelete == false) {
		return false, erro.ErrDelete
	}
	return true, nil
}

func (s WorkerService) List(ctx context.Context, account core.Account) (*[]core.Account, error){
	childLogger.Debug().Msg("List")

	span := lib.Span(ctx, "service.List")
	defer span.End()

	res, err := s.workerRepo.List(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) GetId(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("GetId")

	span := lib.Span(ctx, "service.GetId")
	defer span.End()

	res, err := s.workerRepo.GetId(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}