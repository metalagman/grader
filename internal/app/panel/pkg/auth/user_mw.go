package auth

import (
	"context"
	"github.com/google/uuid"
	"grader/internal/app/panel/storage"
	"grader/internal/pkg/model"
	"grader/pkg/apperr"
	"grader/pkg/logger"
	"grader/pkg/session"
	"net/http"
)

type contextKeyUser struct{}

func UserFromContext(ctx context.Context) (*model.User, error) {
	sess, ok := ctx.Value(contextKeyUser{}).(*model.User)
	if !ok {
		return nil, apperr.ErrUnauthorized
	}
	return sess, nil
}

func ContextMiddleware(u storage.UserRepository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			l := logger.Ctx(ctx)

			s, err := session.FromContext(ctx)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			id, err := uuid.Parse(s.UserID)
			if err != nil {
				l.Error().Err(err).Send()
				next.ServeHTTP(w, r)
				return
			}

			user, err := u.Read(ctx, id)
			if err != nil {
				l.Error().Err(err).Send()
				next.ServeHTTP(w, r)
				return
			}

			ctx = context.WithValue(ctx, contextKeyUser{}, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, err := UserFromContext(ctx)
			if err != nil {
				http.Error(w, "No session", http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, contextKeyUser{}, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
