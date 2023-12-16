package service

import (
	"context"

	"github.com/go-account/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (s WorkerService) AddFundBalanceAccount(ctx context.Context, accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("AddFundBalanceAccount")

	childLogger.Debug().Interface("=>accountBalance : ", accountBalance).Msg("")

	_, root := xray.BeginSubsegment(ctx, "Service.AddFundBalanceAccount")
	
	tx, err := s.workerRepository.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
		root.Close(nil)
	}()

	res, err := s.workerRepository.UpdateFundBalanceAccount(ctx, tx, accountBalance)
	if err != nil {
		return nil, err
	}

	if res == 0 {
		res_create, err := s.workerRepository.CreateFundBalanceAccount(ctx, tx, accountBalance)
		if err != nil {
			return nil, err
		}
		return res_create, nil
	}

	return &accountBalance, nil
}

func (s WorkerService) GetMovimentBalanceAccount(ctx context.Context, accountBalance core.AccountBalance) (interface{} , error){
	childLogger.Debug().Msg("GetMovimentBalanceAccount")
	childLogger.Debug().Interface("=>accountBalance : ", accountBalance).Msg("")

	_, root := xray.BeginSubsegment(ctx, "Service.GetMovimentBalanceAccount")
	defer root.Close(nil)

	account := core.Account{}
	account.AccountID = accountBalance.AccountID
	_, err := s.workerRepository.Get(ctx, account)
	if err != nil {
		return nil, err
	}

	res_accountBalance, err := s.workerRepository.GetFundBalanceAccount(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	res_list_accountBalance, err := s.workerRepository.ListAccountStatementMoviment(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	movimentAccount := core.MovimentAccount{}
	movimentAccount.AccountBalance = res_accountBalance
	movimentAccount.AccountStatement = res_list_accountBalance

	return &movimentAccount, nil
}