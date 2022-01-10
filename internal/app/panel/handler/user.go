package handler

import (
	"grader/internal/app/panel/model"
	"grader/internal/app/panel/storage"
	"grader/pkg/apperr"
	"grader/pkg/httputil"
	"grader/pkg/layout"
	"grader/pkg/logger"
	"grader/pkg/session"
	"net/http"
)

type UserHandler struct {
	layout  *layout.Layout
	session session.Manager
	users   storage.UserRepository
}

func NewUserHandler(l *layout.Layout, s session.Manager, u storage.UserRepository) *UserHandler {
	return &UserHandler{layout: l, session: s, users: u}
}

func (h *UserHandler) Default(w http.ResponseWriter, r *http.Request) {
	h.layout.RenderView(w, r, "template/app/views/default.gohtml", nil)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := logger.Ctx(ctx)

	if r.Method != http.MethodPost {
		h.layout.RenderView(w, r, "template/app/views/login.gohtml", nil)
		return
	}

	in := &struct {
		Username string `validate:"required"`
		Password string `validate:"required"`
	}{
		r.FormValue("login"),
		r.FormValue("password"),
	}

	if !httputil.ValidateData(w, in) {
		return
	}

	user, err := h.users.ReadByNameAndPassword(ctx, in.Username, in.Password)
	switch err {
	case nil:
		// all is ok
	case apperr.ErrNotFound:
		http.Error(w, "Unauthorized", http.StatusBadRequest)
	default:
		l.Error().Err(err).Send()
		http.Error(w, apperr.ErrInternal.Error(), http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	if err := h.session.Create(r.Context(), w, user); err != nil {
		l.Error().Err(err).Send()
		http.Error(w, apperr.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app", http.StatusFound)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := logger.Ctx(ctx)

	if r.Method != http.MethodPost {
		h.layout.RenderView(w, r, "template/app/views/register.gohtml", nil)
		return
	}

	in := &struct {
		username string `validate:"required"`
		password string `validate:"required"`
	}{
		r.FormValue("login"),
		r.FormValue("password"),
	}

	if !httputil.ValidateData(w, in) {
		return
	}

	user, err := h.users.Create(ctx, &model.User{Name: in.username, Password: in.password})
	switch err {
	case nil:
		// all is ok
	case apperr.ErrConflict:
		http.Error(w, "User already exists", http.StatusBadRequest)
	default:
		l.Error().Err(err).Send()
		http.Error(w, apperr.ErrInternal.Error(), http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	if err := h.session.Create(r.Context(), w, user); err != nil {
		l.Error().Err(err).Send()
		http.Error(w, apperr.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/app", http.StatusFound)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.session.DestroyCurrent(r.Context(), w, r)
	http.Redirect(w, r, "/app", http.StatusFound)
}
