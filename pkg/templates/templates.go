package templates

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"
)

type MyTemplate struct {
	Tmpl *template.Template
}

func NewTemplates(assets fs.FS) (*MyTemplate, error) {
	tmpl, err := template.ParseFS(assets, "*.html")
	if err != nil {
		return nil, fmt.Errorf("new templates: %w", err)
	}
	return &MyTemplate{
		Tmpl: tmpl,
	}, nil
}

func (tpl *MyTemplate) Render(ctx context.Context, w http.ResponseWriter, tmplName string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{}, 3)
	}

	sess, err := session.SessionFromContext(ctx)
	if err == nil {
		data["Authorized"] = true
		data["Session"] = sess

		token, err := tpl.Tokens.Create(sess, time.Now().Add(24*time.Hour).Unix())
		if err != nil {
			log.Println("csrf token creation error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		data["CSRFToken"] = token
	} else {
		data["Authorized"] = false
	}

	err = tpl.Tmpl.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		log.Println("cant execute template", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
