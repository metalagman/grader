package app

import (
	"fmt"
	"grader/internal/app/grader/config"
	"grader/pkg/httpserver"
	"grader/pkg/logger"
)

type App struct {
	config config.Config
	logger logger.Logger
	stop   chan struct{}
	server *httpserver.Server
}

func New(cfg config.Config) (*App, error) {
	hs, err := httpserver.New(cfg.Server)
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	a := &App{
		config: cfg,
		logger: *logger.Global(),
		stop:   make(chan struct{}),
		server: hs,
	}

	go func() {
		<-a.stop
		hs.Stop()
	}()

	return a, nil
}

func (a *App) Stop() {
	close(a.stop)
}
