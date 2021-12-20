package config

import (
	"grader/pkg/httpserver"
	"grader/pkg/logger"
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	Logger logger.Config     `mapstructure:"log"`
}
