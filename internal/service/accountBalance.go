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