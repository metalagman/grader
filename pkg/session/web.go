package session

import (
	"context"
	"net/http"
)

type contextKeySession struct{}

func FromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(contextKeySession{}).(*Session)
	if !ok {
		return nil, ErrUnauthorized
	}
	return sess, nil
}

func ContextMiddleware(sm Manager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			s, err := sm.Read(ctx, r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// if session found save it into context
			ctx = context.WithValue(ctx, contextKeySession{}, s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
