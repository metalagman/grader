package app

import (
	"fmt"
	"grader/internal/app/queue/config"
	"grader/internal/app/queue/pkg/sender"
	"grader/pkg/logger"
	"grader/pkg/queue"
	"grader/pkg/queue/amqp"
	"grader/pkg/workerpool"
	"runtime"
)

type App struct {
	config  config.Config
	logger  logger.Logger
	stop    chan struct{}
	workers *workerpool.Pool
	queue   queue.Queue
	Sender  *sender.Sender
}

func New(cfg config.Config) (*App, error) {
	l := *logger.Global()

	wp := workerpool.New()

	// init amqp dep
	q, err := amqp.New(cfg.AMQP)
	if err != nil {
		return nil, fmt.Errorf("amqp: %w", err)
	}

	sndr, err := sender.New(q, cfg.App.TopicName)
	if err != nil {
		return nil, fmt.Errorf("sender: %w", err)
	}

	a := &App{
		config:  cfg,
		logger:  l,
		stop:    make(chan struct{}),
		workers: wp,
		queue:   q,

		Sender: sndr,
	}

	a.workers.Start(runtime.GOMAXPROCS(0) * 2)

	go func() {
		<-a.stop
		q.Stop()
	}()

	return a, nil
}

func (a *App) Stop() {
	close(a.stop)
	a.workers.Stop()
}
