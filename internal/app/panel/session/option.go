package session

import "time"

type Option func(*Redis)

func WithIssuer(v string) Option {
	return func(s *Redis) {
		s.issuer = v
	}
}

func WithRedisKeyPrefix(v string) Option {
	return func(s *Redis) {
		s.redisKeyPrefix = v
	}
}

func WithTokenLifetime(v time.Duration) Option {
	return func(s *Redis) {
		s.tokenLifetime = v
	}
}
