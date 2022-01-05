package layout

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

type ViewData map[string]interface{}
type ViewDataFunc func(context.Context, ViewData) (ViewData, error)

type Layout struct {
	fs   fs.FS
	tmpl *template.Template

	dataFunc ViewDataFunc
}

func NewLayout(tmplFS fs.FS, layoutFile string, dataFunc ViewDataFunc) (*Layout, error) {
	tmpl, err := template.ParseFS(tmplFS, layoutFile)
	if err != nil {
		return nil, fmt.Errorf("new layout: %w", err)
	}

	return &Layout{
		fs:   tmplFS,
		tmpl: tmpl,
	}, nil
}

func (l *Layout) View(viewFile ...string) (*View, error) {
	tmpl, err := l.tmpl.Clone()
	if err != nil {
		return nil, fmt.Errorf("template clone: %w", err)
	}

	tmpl, err = tmpl.ParseFS(l.fs, viewFile...)
	if err != nil {
		return nil, fmt.Errorf("view parse: %w", err)
	}

	return &View{tmpl}, nil
}

type View struct {
	tmpl *template.Template
}

func (t *View) Render(w http.ResponseWriter, data map[string]interface{}) error {
	if err := t.tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("template execute: %w", err)
	}

	return nil
}

func (l *Layout) RenderView(w http.ResponseWriter, r *http.Request, viewName string, data map[string]interface{}) {
	view, err := l.View(viewName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if l.dataFunc != nil {
		ctx := r.Context()
		data, err = l.dataFunc(ctx, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := view.Render(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Templates struct {
	fs           fs.FS
	tmpl         *template.Template
	tokenManager token.Manager
}

func NewTemplates(assets fs.FS, tm token.Manager, patterns ...string) (*Templates, error) {
	tmpl, err := template.ParseFS(assets, patterns...)
	if err != nil {
		return nil, fmt.Errorf("new templates: %w", err)
	}
	return &Templates{
		fs:           assets,
		tmpl:         tmpl,
		tokenManager: tm,
	}, nil
}

func (tpl *Templates) Render(ctx context.Context, w http.ResponseWriter, tplName string, data map[string]interface{}) {
	l := logger.Ctx(ctx)

	if data == nil {
		data = make(map[string]interface{}, 3)
	}

	data["View"] = tplName

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

	err = tpl.tmpl.ExecuteTemplate(w, tplName, data)
	if err != nil {
		l.Error().Err(err).Send()
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
