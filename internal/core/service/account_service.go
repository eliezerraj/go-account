package service

import(
	"time"
	"context"
	"github.com/rs/zerolog/log"

	"github.com/go-account/internal/adapter/database"
	"github.com/go-account/internal/core/model"
	"github.com/go-account/internal/core/erro"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var childLogger = log.With().Str("component","go-account").Str("package","internal.core.service").Logger()
var tracerProvider go_core_observ.TracerProvider
var apiService go_core_api.ApiService

type WorkerService struct {
	workerRepository *database.WorkerRepository
}

// About new worker service
func NewWorkerService(workerRepository *database.WorkerRepository) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		workerRepository: workerRepository,
	}
}

// About handle/convert http status code
func (s *WorkerService) Stat(ctx context.Context) (go_core_pg.PoolStats){
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	return s.workerRepository.Stat(ctx)
}

// About add an account
func (s *WorkerService) AddAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","AddAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("account", account).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.AddAccount")
	
	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)
	
	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		span.End()
	}()

	// Add the account
	res, err := s.workerRepository.AddAccount(ctx, tx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About update an account
func (s *WorkerService) UpdateAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","UpdateAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("account", account).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.UpdateAccount")
	
	// Get the database connection
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)

	// Handle the transaction
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
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

// About delete an account
func (s *WorkerService) DeleteAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","DeleteAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("account", account).Send()

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

// About get an account
func (s *WorkerService) GetAccount(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","GetAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("account", account).Send()

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

// About get an account
func (s *WorkerService) GetAccountId(ctx context.Context, account *model.Account) (*model.Account, error){
	childLogger.Info().Str("func","GetAccountId").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("account", account).Send()

	// Trace
	span := tracerProvider.Span(ctx, "service.GetAccountId")
	defer span.End()
	
	time.Sleep(0 * time.Second) // just for test
	// Get account
	res, err := s.workerRepository.GetAccountId(ctx, account)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// About list all person´s account
func (s *WorkerService) ListAccountPerPerson(ctx context.Context, account *model.Account) (*[]model.Account, error){
	childLogger.Info().Str("func","ListAccountPerPerson").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("account", account).Send()

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