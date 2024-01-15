package service

import (
	"context"

	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/core"
	"github.com/aws/aws-xray-sdk-go/xray"

)

func (s WorkerService) AddFundBalanceAccount(ctx context.Context, accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("AddFundBalanceAccount")

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

	// Try to update the account_balance
	res, err := s.workerRepository.UpdateFundBalanceAccount(ctx, tx, accountBalance)
	if err != nil {
		return nil, err
	}

	// If the account_balance so it created one
	if res == 0 {
		res_create, err := s.workerRepository.CreateFundBalanceAccount(ctx, tx, accountBalance)
		if err != nil {
			return nil, err
		}
		return res_create, nil
	}

	return &accountBalance, nil
}

func (s WorkerService) GetFundBalanceAccount(ctx context.Context, accountBalance core.AccountBalance) (interface{} , error){
	childLogger.Debug().Msg("GetFundBalanceAccount")

	_, root := xray.BeginSubsegment(ctx, "Service.GetFundBalanceAccount")
	defer root.Close(nil)

	account := core.Account{}
	account.AccountID = accountBalance.AccountID
	_, err := s.workerRepository.Get(ctx, account)
	if err != nil {
		return nil, err
	}

	res_fundAccountBalance, err := s.workerRepository.GetFundBalanceAccount(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	return &res_fundAccountBalance, nil
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

	res_accountBalanceStatementCredit, err := s.workerRepository.GetFundBalanceAccountStatementMoviment(ctx, "CREDIT", accountBalance)
	if err != nil {
		if (err != erro.ErrNotFound){
			return nil, err
		}
	}

	res_accountBalanceStatementDebit, err := s.workerRepository.GetFundBalanceAccountStatementMoviment(ctx, "DEBIT", accountBalance)
	if err != nil {
		if (err != erro.ErrNotFound){
			return nil, err
		}
	}

	res_list_accountBalance, err := s.workerRepository.ListAccountStatementMoviment(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	movimentAccount := core.MovimentAccount{}
	movimentAccount.AccountBalance = res_accountBalance
	if (res_accountBalanceStatementCredit != nil){
		movimentAccount.AccountBalanceStatementCredit = res_accountBalanceStatementCredit.Amount
	}
	if (res_accountBalanceStatementDebit != nil){
		movimentAccount.AccountBalanceStatementDebit = res_accountBalanceStatementDebit.Amount
	}
	if (res_list_accountBalance != nil){
		movimentAccount.AccountStatement = res_list_accountBalance
	}

	return &movimentAccount, nil
}

func (s WorkerService) TransferFundAccount(ctx context.Context, transfer core.Transfer) (interface{} , error){
	childLogger.Debug().Msg("TransferFundAccount")
	//childLogger.Debug().Interface("=>transfer : ", transfer).Msg("")

	_, root := xray.BeginSubsegment(ctx, "Service.TransferFundAccount")
	
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

	// Debit the fund
	if (transfer.Type != "TRANSFER") {
		return nil, erro.ErrTransaction
	}
	transfer.Amount = (transfer.Amount * -1)
	res, uuid ,err := s.workerRepository.TransferFundAccount(ctx, tx, transfer)
	if err != nil {
		return nil, err
	}
	if res == 0 {
		return nil, erro.ErrUpdate
	}

	accountStatementFrom := core.AccountStatement{}
	accountStatementFrom.AccountID = transfer.AccountIDFrom
	accountStatementFrom.FkAccountID = transfer.FkAccountIDFrom
	accountStatementFrom.Type = "DEBIT"
	accountStatementFrom.Currency = transfer.Currency
	accountStatementFrom.Amount = transfer.Amount
	_, err = s.workerRepository.AddAccountStatement(ctx, tx, accountStatementFrom)
	if err != nil {
		return nil, err
	}

	// Add the fund
	accountBalance := core.AccountBalance{}
	accountBalance.Amount = (transfer.Amount * -1)
	accountBalance.Currency = transfer.Currency
	accountBalance.FkAccountID = transfer.FkAccountIDTo

	res, err = s.workerRepository.UpdateFundBalanceAccount(ctx, tx, accountBalance)
	if err != nil {
		return nil, err
	}
	if res == 0 {
		return nil, erro.ErrUpdate
	}
	accountStatementTo := core.AccountStatement{}
	accountStatementTo.AccountID = transfer.AccountIDTo
	accountStatementTo.FkAccountID = transfer.FkAccountIDTo
	accountStatementTo.Type = "CREDIT"
	accountStatementTo.Currency = transfer.Currency
	accountStatementTo.Amount = (transfer.Amount * -1)
	_, err = s.workerRepository.AddAccountStatement(ctx, tx, accountStatementTo)
	if err != nil {
		return nil, err
	}

	res, err = s.workerRepository.CommitTransferFundAccount(ctx, tx, uuid, transfer)
	if err != nil {
		return nil, err
	}
	if res == 0 {
		return nil, erro.ErrUpdate
	}

	return &transfer, nil
}