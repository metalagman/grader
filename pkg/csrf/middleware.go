package csrf

import (
	"grader/pkg/session"
	"grader/pkg/token"
	"net/http"
)

func ValidateMiddleware(tm token.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
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
		tk := r.Header.Get("csrf-token")
		if err := tm.Validate(tk, s); err == nil {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, err.Error(), http.StatusForbidden)
	})
}
