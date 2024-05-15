package service

import (
	"context"
	"github.com/rs/zerolog/log"

	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/core"
	"github.com/go-account/internal/repository/postgre"
	"github.com/aws/aws-xray-sdk-go/xray"

)

var childLogger = log.With().Str("service", "service").Logger()

type WorkerService struct {
	workerRepository 	*postgre.WorkerRepository
}

func NewWorkerService(workerRepository *postgre.WorkerRepository) *WorkerService{
	childLogger.Debug().Msg("NewWorkerService")

	return &WorkerService{
		workerRepository:	workerRepository,
	}
}
// -----------------------------------------------
func (s WorkerService) SetSessionVariable(ctx context.Context, userCredential string) (bool, error){
	childLogger.Debug().Msg("SetSessionVariable")

	res, err := s.workerRepository.SetSessionVariable(ctx, userCredential)
	if err != nil {
		return false, err
	}

	return res, nil
}

func (s WorkerService) Add(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Add")

	_, root := xray.BeginSubsegment(ctx, "Service.Add")
	defer func() {
		root.Close(nil)
	}()

	// Create account
	res, err := s.workerRepository.Add(ctx, account)
	if err != nil {
		return nil, err
	}

	// Get ID account
	res, err = s.workerRepository.Get(ctx, account)
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

	_, root := xray.BeginSubsegment(ctx, "Service.Get")
	defer func() {
		root.Close(nil)
	}()

	res, err := s.workerRepository.Get(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) Update(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("Update")

	_, root := xray.BeginSubsegment(ctx, "Service.Update")
	defer func() {
		root.Close(nil)
	}()

	res_account, err := s.workerRepository.Get(ctx, account)
	if err != nil {
		return nil, err
	}

	account.ID = res_account.ID
	isUpdated, err := s.workerRepository.Update(ctx, account)
	if err != nil {
		return nil, err
	}
	if (isUpdated == false) {
		return nil, erro.ErrUpdate
	}

	res_account, err = s.workerRepository.Get(ctx, account)
	if err != nil {
		return nil, err
	}
	return res_account, nil
}

func (s WorkerService) Delete(ctx context.Context,account core.Account) (bool, error){
	childLogger.Debug().Msg("Delete")

	_, root := xray.BeginSubsegment(ctx, "Service.Delete")
	defer func() {
		root.Close(nil)
	}()

	res_account, err := s.workerRepository.Get(ctx,account)
	if err != nil {
		return false, err
	}

	account.ID = res_account.ID
	isDelete, err := s.workerRepository.Delete(ctx, account)
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

	_, root := xray.BeginSubsegment(ctx, "Service.List")
	defer root.Close(nil)

	res, err := s.workerRepository.List(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s WorkerService) GetId(ctx context.Context, account core.Account) (*core.Account, error){
	childLogger.Debug().Msg("GetId")

	_, root := xray.BeginSubsegment(ctx, "Service.GetId")
	defer func() {
		root.Close(nil)
	}()

	res, err := s.workerRepository.GetId(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}