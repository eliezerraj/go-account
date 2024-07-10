package service

import (
	"context"

	"github.com/go-account/internal/lib"
	"github.com/go-account/internal/erro"
	"github.com/go-account/internal/core"
)

func (s WorkerService) AddFundBalanceAccount(ctx context.Context, accountBalance core.AccountBalance) (*core.AccountBalance, error){
	childLogger.Debug().Msg("AddFundBalanceAccount")

	span := lib.Span(ctx, "service.AddFundBalanceAccount")
	
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
		span.End()
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

	span := lib.Span(ctx, "service.GetFundBalanceAccount")
	span.End()

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
	//childLogger.Debug().Interface("=>accountBalance : ", accountBalance).Msg("")

	span := lib.Span(ctx, "service.GetMovimentBalanceAccount")
	span.End()

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
	if (res_accountBalanceStatementCredit != nil) && (res_accountBalanceStatementDebit != nil){
		movimentAccount.AccountBalanceStatementTotal = res_accountBalanceStatementCredit.Amount + res_accountBalanceStatementDebit.Amount
	}
	if (res_list_accountBalance != nil){
		movimentAccount.AccountStatement = res_list_accountBalance
	}

	return &movimentAccount, nil
}

func (s WorkerService) TransferFundAccount(ctx context.Context, transfer core.Transfer) (interface{} , error){
	childLogger.Debug().Msg("TransferFundAccount")
	//childLogger.Debug().Interface("=>transfer : ", transfer).Msg("")

	span := lib.Span(ctx, "service.TransferFundAccount")
	span.End()

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
		span.End()
	}()

	// Debit the fund
	if (transfer.Type != "TRANSFER") {
		err = erro.ErrTransaction 
		return nil, err
	}

	transfer.Amount = (transfer.Amount * -1)
	res, uuid ,err := s.workerRepository.TransferFundAccount(ctx, tx, transfer)
	if err != nil {
		return nil, err
	}
	if res == 0 {
		err = erro.ErrUpdate
		return nil, err
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
		err = erro.ErrUpdate
		return nil, err
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
		err = erro.ErrUpdate
		return nil, err
	}

	return &transfer, nil
}