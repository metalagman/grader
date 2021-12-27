package config

import (
	"grader/pkg/httpserver"
	"grader/pkg/logger"
	"grader/pkg/queue/amqp"
	"time"
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	DB     DatabaseConfig    `mapstructure:"db"`
	AMQP   amqp.Config       `mapstructure:"amqp"`
	Logger logger.Config     `mapstructure:"log"`
}

type ServerConfig struct {
	Listen       string        `mapstructure:"listen"`
	TimeoutRead  time.Duration `mapstructure:"timeout_read"`
	TimeoutWrite time.Duration `mapstructure:"timeout_write"`
	TimeoutIdle  time.Duration `mapstructure:"timeout_idle"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}
