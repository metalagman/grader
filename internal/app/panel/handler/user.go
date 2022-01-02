package handler

import (
	"grader/internal/app/panel/storage"
	"grader/pkg/apperr"
	"grader/pkg/session"
	"grader/pkg/templates"
	"net/http"
)

type UserHandler struct {
	tmpl    *templates.Templates
	session session.Manager
	users   storage.UserRepository
}

func NewUserHandler(tmpl *templates.Templates, session session.Manager, users storage.UserRepository) *UserHandler {
	return &UserHandler{tmpl: tmpl, session: session, users: users}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		h.tmpl.Render(r.Context(), w, "login.html", nil)
		return
	}

	name := r.FormValue("username")
	pass := r.FormValue("password")

	user, err := h.users.ReadByNameAndPassword(ctx, name, pass)
	switch err {
	case nil:
		// all is ok
	case apperr.ErrNotFound:
		http.Error(w, "unauthorized", http.StatusBadRequest)
	default:
		http.Error(w, "internal err", http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	if err := h.session.Create(r.Context(), w, user); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/photos/", http.StatusFound)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	//if r.Method != http.MethodPost {
	//	h.tmpl.Render(r.Context(), w, "register.html", nil)
	//	return
	//}
	//
	//login := r.FormValue("login")
	//pass := r.FormValue("password")
	//
	//if !govalidator.IsEmail(email) {
	//	http.Error(w, "Bad email", http.StatusBadRequest)
	//	return
	//}
	//
	//if !loginRE.MatchString(login) {
	//	http.Error(w, "Bad login", http.StatusBadRequest)
	//	return
	//}
	//
	//user, err := h.UsersRepo.Create(login, email, pass)
	//switch err {
	//case nil:
	//	// all is ok
	//case errUserExists:
	//	http.Error(w, "Looks like user exists", http.StatusBadRequest)
	//default:
	//	log.Println("db err", err)
	//	http.Error(w, "Db err", http.StatusInternalServerError)
	//}
	//if err != nil {
	//	return
	//}
	//
	//h.Sessions.Create(r.Context(), w, user)
	//http.Redirect(w, r, "/photos/", http.StatusFound)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.session.DestroyCurrent(r.Context(), w, r)
	http.Redirect(w, r, "/app/user/login", http.StatusTemporaryRedirect)
}
