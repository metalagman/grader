package workerpool

import (
	"context"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type Job func(ctx context.Context) error

type Pool struct {
	wg     *sync.WaitGroup
	logger zerolog.Logger

	jobs chan Job
	stop chan struct{}

	DefaultContext func() context.Context
}

type PoolOption func(service *Pool)

func New(opts ...PoolOption) *Pool {
	s := &Pool{
		wg:     &sync.WaitGroup{},
		logger: log.Logger,

		DefaultContext: func() context.Context {
			return context.Background()
		},
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func WithLogger(zl zerolog.Logger) PoolOption {
	return func(service *Pool) {
		service.logger = zl
	}
}

func (s *Pool) Start(numWorkers int) {
	s.logger.Info().Int("worker_num", numWorkers).Msg("Starting workers")
	s.stop = make(chan struct{})
	s.jobs = make(chan Job)
	s.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer s.wg.Done()
			l := s.logger.With().Int("worker_id", workerID).Logger()
			for {
				select {
				case <-s.stop:
					return
				case job, ok := <-s.jobs:
					if !ok {
						// if jobs channel was closed
						return
					}
					id := xid.New()
					now := time.Now()
					ll := l.With().Str("job_id", id.String()).Logger()
					ll.Debug().Msg("Running job")
					if err := AddLogger(AddRecovery(job), ll)(s.DefaultContext()); err != nil {
						ll.Error().Err(err).Msg("Error running job")
					}
					ll.Info().Dur("job_duration", time.Since(now)).Msg("Done running job")
				}
			}
		}(i)
	}

	s.logger.Info().Msg("Done starting workers")
}

func (s *Pool) Stop() {
	s.logger.Info().Msg("Shutting down workers")
	close(s.stop)
	s.wg.Wait()
	close(s.jobs)
	s.logger.Info().Msg("Done shutting down workers")
}

func (s *Pool) Run(job Job) {
	s.jobs <- job
}
