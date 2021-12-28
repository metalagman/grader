package app

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	_ "github.com/lib/pq"
	"grader/internal/app/panel/config"
	"grader/pkg/aws"
	"grader/pkg/httpserver"
	"grader/pkg/logger"
	mw "grader/pkg/middleware"
	"grader/pkg/queue"
	"grader/pkg/queue/amqp"
	"grader/pkg/workerpool"
)

type App struct {
	config  config.Config
	logger  logger.Logger
	stop    chan struct{}
	queue   queue.Queue
	workers *workerpool.Pool
	server  *httpserver.Server
	s3      *aws.S3
}

func New(cfg config.Config) (*App, error) {
	l := *logger.Global()

	wp := workerpool.New()

	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	//if err := migrate.Up(db); err != nil {
	//	return nil, fmt.Errorf("migrate up: %w", err)
	//}

	conn, err := rabbitmq.Dial(cfg.AMQP.DSN)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	sendCh, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %w", err)
	}

	// init amqp dep
	q, err := amqp.New(cfg.AMQP)
	if err != nil {
		return nil, fmt.Errorf("amqp: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(mw.Log(l))

	hs, err := httpserver.New(cfg.Server, r, httpserver.WithLogger(l.Logger))
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	s3, err := aws.NewS3(cfg.AWS)
	if err != nil {
		return nil, fmt.Errorf("s3: %w", err)
	}

	a := &App{
		config:  cfg,
		logger:  l,
		stop:    make(chan struct{}),
		queue:   q,
		server:  hs,
		workers: wp,
		s3:      s3,
	}

	go func() {
		<-a.stop
		q.Stop()
		_ = sendCh.Close()
		_ = conn.Close()
	}()

	return a, nil
}

func (a *App) Stop() {
	close(a.stop)
	a.server.Stop()
	a.workers.Stop()
}
