package app

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"grader/internal/app/panel/config"
	"grader/pkg/logger"
	"grader/pkg/queue"
	"grader/pkg/queue/amqp"
	"net/http"
)

type App struct {
	config config.Config
	logger logger.Logger
	stop   chan struct{}
	queue  queue.Queue
	server *http.Server
}

func New(cfg config.Config) (*App, error) {
	db, err := sql.Open("mysql", cfg.DB.DSN)
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

	a := &App{
		config: cfg,
		logger: *logger.Global(),
		stop:   make(chan struct{}),
		queue:  q,
	}

	a.startServer(cfg.Server)

	go func() {
		<-a.stop
		q.Stop()
		_ = sendCh.Close()
		_ = conn.Close()
	}()

	return a, nil
}

func (a *App) Stop() {
	a.stopServer()
	close(a.stop)
}
