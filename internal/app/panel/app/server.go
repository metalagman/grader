package app

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"grader/internal/app/panel/config"
	"net/http"
	"time"
)

func (a *App) startServer(cfg config.ServerConfig) {
	l := a.logger

	a.server = &http.Server{
		Addr:         cfg.Listen,
		ReadTimeout:  cfg.TimeoutRead,
		WriteTimeout: cfg.TimeoutWrite,
		IdleTimeout:  cfg.TimeoutIdle,
		Handler:      a.router(),
	}

	go func() {
		l.Info().Str("listen", cfg.Listen).Msg("Listening incoming connections")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Send()
		}
	}()
}

func (a *App) stopServer() {
	l := a.logger

	ctxShutdown, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	if err := a.server.Shutdown(ctxShutdown); err != nil {
		l.Error().Err(fmt.Errorf("server shutdown failed: %w", err)).Send()
	}
}

func (a *App) router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	return r
}
