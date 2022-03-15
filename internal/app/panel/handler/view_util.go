package handler

import (
	"context"
	"grader/internal/app/panel/pkg/auth"
	"grader/pkg/csrf"
	"grader/pkg/layout"
	"grader/pkg/session"
	"grader/pkg/token"
)

func ViewDataFunc(tm token.Manager) layout.ViewDataFunc {
	return func(ctx context.Context, data layout.ViewData) (layout.ViewData, error) {
		if data == nil {
			data = make(layout.ViewData, 3)
		}

		s, err := session.FromContext(ctx)
		if err == nil {
			data["Authorized"] = true
			data["Session"] = s
			data["CSRFToken"] = csrf.FromContext(ctx)
		} else {
			data["Authorized"] = false
		}

		u, err := auth.UserFromContext(ctx)
		if err == nil {
			data["CurrentUser"] = u
		}

		return data, nil
	}
}
