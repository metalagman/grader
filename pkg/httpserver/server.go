package httpserver

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Listen       string        `mapstructure:"listen"`
	TimeoutRead  time.Duration `mapstructure:"timeout_read"`
	TimeoutWrite time.Duration `mapstructure:"timeout_write"`
	TimeoutIdle  time.Duration `mapstructure:"timeout_idle"`
}

type Server struct {
	listener net.Listener
	server   *http.Server
	handler  http.Handler
	logger   zerolog.Logger
}

type Option func(*Server)

func WithHandler(h http.Handler) Option {
	return func(server *Server) {
		server.handler = h
	}
}

func WithLogger(l zerolog.Logger) Option {
	return func(server *Server) {
		server.logger = l
	}
}

func New(cfg Config, opts ...Option) (*Server, error) {
	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	hs := &http.Server{
		ReadTimeout:  cfg.TimeoutRead,
		WriteTimeout: cfg.TimeoutWrite,
		IdleTimeout:  cfg.TimeoutIdle,
	}

	s := &Server{
		logger:   log.Logger,
		listener: ln,
		server:   hs,
	}

	for _, o := range opts {
		o(s)
	}

	go func() {
		s.logger.Info().Str("listen", cfg.Listen).Msg("Listening incoming connections")
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal().Err(err).Send()
		}
	}()

	return s, nil
}

func (s *Server) Stop() {
	const ShutdownTimeout = 5 * time.Second
	ctxShutdown, ctxCancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer ctxCancel()

	if err := s.server.Shutdown(ctxShutdown); err != nil {
		s.logger.Error().Err(fmt.Errorf("server shutdown: %w", err)).Send()
	}
}
