package templates

import (
	"context"
	"fmt"
	"grader/pkg/logger"
	"grader/pkg/session"
	"grader/pkg/token"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

type Templates struct {
	tmpl         *template.Template
	tokenManager token.Manager
}

func NewTemplates(assets fs.FS, tm token.Manager) (*Templates, error) {
	tmpl, err := template.ParseFS(assets, "web/template/app/*.html")
	if err != nil {
		return nil, fmt.Errorf("new templates: %w", err)
	}
	return &Templates{
		tmpl:         tmpl,
		tokenManager: tm,
	}, nil
}

func (tpl *Templates) Render(ctx context.Context, w http.ResponseWriter, tmplName string, data map[string]interface{}) {
	l := logger.Ctx(ctx)
	if data == nil {
		data = make(map[string]interface{}, 3)
	}

	s, err := session.FromContext(ctx)
	if err == nil {
		data["Authorized"] = true
		data["Session"] = s

		tk, err := tpl.tokenManager.Issue(s, 24*time.Hour)
		if err != nil {
			l.Error().Err(err).Send()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		data["CSRFToken"] = tk
	} else {
		data["Authorized"] = false
	}

	err = tpl.tmpl.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
