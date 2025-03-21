package service

import(
	"github.com/go-account/internal/adapter/database"
	"github.com/rs/zerolog/log"
)

var childLogger = log.With().Str("component","go-account").Str("package","internal.core.service").Logger()

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