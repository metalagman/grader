package app

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"grader/internal/app/grader/config"
	"grader/internal/app/grader/handler"
	"grader/pkg/httpserver"
	"grader/pkg/logger"
	mw "grader/pkg/middleware"
	"grader/pkg/workerpool"
	"runtime"
)

type App struct {
	config  config.Config
	logger  logger.Logger
	stop    chan struct{}
	server  *httpserver.Server
	workers *workerpool.Pool
}

func New(cfg config.Config) (*App, error) {
	l := *logger.Global()

	wp := workerpool.New()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(mw.Log(l))

	ah := handler.NewAssessmentHandler(wp)
	r.Post("/assessments", ah.Check)

	hs, err := httpserver.New(cfg.Server, httpserver.WithHandler(r), httpserver.WithLogger(l.Logger))
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	a := &App{
		config:  cfg,
		logger:  l,
		stop:    make(chan struct{}),
		server:  hs,
		workers: wp,
	}

	wp.Start(runtime.GOMAXPROCS(0) * 2)

	return a, nil
}

func (a *App) Stop() {
	close(a.stop)
	a.server.Stop()
	a.workers.Stop()
}
