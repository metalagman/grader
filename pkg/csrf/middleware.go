package csrf

import (
	"context"
	"grader/pkg/session"
	"grader/pkg/token"
	"net/http"
	"time"
)

const (
	headerName = "X-CSRF-Token"
	formField  = "csrf_token"
)

type contextKey struct{}

func ValidateMiddleware(tm token.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		s, err := session.FromContext(r.Context())
		// skip if no session
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// validate token
		tk := r.Header.Get(headerName)
		if tk == "" {
			tk = r.FormValue(formField)
		}

		if err := tm.Validate(tk, s); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GenerateMiddleware(tm token.Manager, lifetime time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		s, err := session.FromContext(r.Context())
		// skip if no session
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// generate new and save into context
		tk, err := tm.Issue(s, lifetime)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, contextKey{}, tk)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func FromContext(ctx context.Context) string {
	tok, ok := ctx.Value(contextKey{}).(string)
	if !ok {
		return ""
	}
	return tok
}
