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

func AuthMiddleware(sm Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, err := sm.Read(ctx, r)
		if err != nil {
			http.Error(w, "No auth", http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, contextKeySession{}, s)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
