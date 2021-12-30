package config

import (
	"grader/pkg/aws"
	"grader/pkg/httpserver"
	"grader/pkg/logger"
	"grader/pkg/queue/amqp"
)

type Config struct {
	App      AppConfig         `mapstructure:"app"`
	Server   httpserver.Config `mapstructure:"server"`
	DB       DatabaseConfig    `mapstructure:"db"`
	AMQP     amqp.Config       `mapstructure:"amqp"`
	Logger   logger.Config     `mapstructure:"log"`
	AWS      aws.Config        `mapstructure:"aws"`
	Redis    RedisConfig       `mapstructure:"redis"`
	Security SecurityConfig    `mapstructure:"security"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type SecurityConfig struct {
	SecretKey string `mapstructure:"secret_key"`
}
