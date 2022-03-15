package workerpool

import (
	"context"
	"fmt"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/rs/zerolog"
	"time"
)

func AddRecovery(job Job) Job {
	return func(ctx context.Context) (retErr error) {
		defer func() {
			p := recover()
			if p != nil {
				retErr = fmt.Errorf("panic recovered: %v", p)
				zerolog.Ctx(ctx).Error().Err(retErr).Msg("Panic recovered")
			}
		}()
		return job(ctx)
	}
}

func AddLogger(job Job, l zerolog.Logger) Job {
	return func(ctx context.Context) error {
		ctx = l.WithContext(ctx)
		return job(ctx)
	}
}

func AddRetry(job Job, strategies ...strategy.Strategy) Job {
	return func(ctx context.Context) error {
		return retry.Retry(
			func(attempt uint) error {
				l := zerolog.Ctx(ctx).With().Uint("attempt", attempt).Logger()
				return AddLogger(job, l)(ctx)
			},
			strategies...,
		)
	}
}

func AddTimeout(job Job, timeout time.Duration) Job {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return job(ctx)
	}
}

func AddPostRun(job Job, hook func(err error)) Job {
	return func(ctx context.Context) error {
		err := job(ctx)
		hook(err)
		return err
	}
}
