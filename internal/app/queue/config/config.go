package config

import (
	"grader/pkg/logger"
	"grader/pkg/queue/amqp"
)

type Config struct {
	App    AppConfig     `mapstructure:"app"`
	AMQP   amqp.Config   `mapstructure:"amqp"`
	Logger logger.Config `mapstructure:"log"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	TopicName string `mapstructure:"topic_name"`
}
