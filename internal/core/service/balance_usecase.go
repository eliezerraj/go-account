package service

import(
	"context"
	"github.com/go-account/internal/core/erro"
	"github.com/go-account/internal/core/model"
)

func (s *WorkerService) GetAccountBalance(ctx context.Context, accountBalance *model.AccountBalance) (*model.AccountBalance, error){
	childLogger.Debug().Msg("GetAccountBalance")
	childLogger.Debug().Interface("accountBalance: ", accountBalance).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.GetAccountBalance")
	defer span.End()
	
	// Check if account exists
	account := model.Account{}
	account.AccountID = accountBalance.AccountID
	_, err := s.workerRepository.GetAccount(ctx, &account)
	if err != nil {
		return nil, err
	}

	// Get account balance
	res_accountBalance, err := s.workerRepository.GetAccountBalance(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	return res_accountBalance, nil
}

func (s *WorkerService) AddAccountBalance(ctx context.Context, accountBalance *model.AccountBalance) (*model.AccountBalance, error){
	childLogger.Debug().Msg("AddAccountBalance")
	childLogger.Debug().Interface("accountBalance: ", accountBalance).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.AddAccountBalance")
	defer span.End()
	
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

	// Update the account_balance
	res_update, err := s.workerRepository.UpdateAccountBalance(ctx, tx, accountBalance)
	if err != nil {
		return nil, err
	}

	// If the account_balance doesnt exist created one
	if res_update == 0 {
		res_create, err := s.workerRepository.AddAccountBalance(ctx, tx, accountBalance)
		if err != nil {
			return nil, err
		}
		return res_create, nil
	}
	return accountBalance, nil
}

func (s *WorkerService) GetMovimentAccountBalance(ctx context.Context, accountBalance *model.AccountBalance) (*model.MovimentAccount, error){
	childLogger.Debug().Msg("GetMovimentAccountBalance")
	childLogger.Debug().Interface("accountBalance: ", accountBalance).Msg("")

	// Trace
	span := tracerProvider.Span(ctx, "service.GetMovimentAccountBalance")
	defer span.End()
	
	// Check if account exists
	account := model.Account{}
	account.AccountID = accountBalance.AccountID
	_, err := s.workerRepository.GetAccount(ctx, &account)
	if err != nil {
		return nil, err
	}

	// Get account balance
	res_accountBalance, err := s.workerRepository.GetAccountBalance(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	// Get all credits
	res_accountBalanceStatementCredit, err := s.workerRepository.GetSumAccountBalance(ctx, "CREDIT", accountBalance)
	if err != nil {
		if (err != erro.ErrNotFound){
			return nil, err
		}
	}

	// Get all debits
	res_accountBalanceStatementDebit, err := s.workerRepository.GetSumAccountBalance(ctx, "DEBIT", accountBalance)
	if err != nil {
		if (err != erro.ErrNotFound){
			return nil, err
		}
	}

	res_list_accountBalance, err := s.workerRepository.ListAccountBalance(ctx, accountBalance)
	if err != nil {
		return nil, err
	}

	movimentAccount := model.MovimentAccount{}
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