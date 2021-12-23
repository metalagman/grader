package config

import (
	"grader/pkg/httpserver"
	"grader/pkg/logger"
	"grader/pkg/queue/amqp"
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	DB     DatabaseConfig    `mapstructure:"db"`
	AMQP   amqp.Config       `mapstructure:"amqp"`
	Logger logger.Config     `mapstructure:"log"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}
