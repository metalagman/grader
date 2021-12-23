package config

import (
	"grader/pkg/aws"
	"grader/pkg/httpserver"
	"grader/pkg/logger"
	"grader/pkg/queue/amqp"
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	DB     DatabaseConfig    `mapstructure:"db"`
	AMQP   amqp.Config       `mapstructure:"amqp"`
	Logger logger.Config     `mapstructure:"log"`
	AWS    aws.Config        `mapstructure:"aws"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}
