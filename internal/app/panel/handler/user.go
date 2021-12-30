package handler

import (
	"grader/internal/app/panel/session"
	"grader/internal/app/panel/storage"
	"grader/pkg/apperr"
	"log"
	"net/http"
)

type UserHandler struct {
	session session.Manager
	users   storage.UserRepository
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		h.Tmpl.Render(r.Context(), w, "login.html", map[string]interface{}{
			"VKAuthURL": VK_AUTH_URL,
		})
		return
	}

	name := r.FormValue("username")
	pass := r.FormValue("password")

	user, err := h.users.ReadByNameAndPassword(ctx, name, pass)
	switch err {
	case nil:
		// all is ok
	case apperr.ErrNotFound:
		http.Error(w, "No user", http.StatusBadRequest)
	default:
		http.Error(w, "Db err", http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	h.Sessions.Create(r.Context(), w, user)
	http.Redirect(w, r, "/photos/", http.StatusFound)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.Tmpl.Render(r.Context(), w, "reg.html", nil)
		return
	}

	login := r.FormValue("login")
	pass := r.FormValue("password")
	email := r.FormValue("email")

	if !govalidator.IsEmail(email) {
		http.Error(w, "Bad email", http.StatusBadRequest)
		return
	}

	if !loginRE.MatchString(login) {
		http.Error(w, "Bad login", http.StatusBadRequest)
		return
	}

	user, err := h.UsersRepo.Create(login, email, pass)
	switch err {
	case nil:
		// all is ok
	case errUserExists:
		http.Error(w, "Looks like user exists", http.StatusBadRequest)
	default:
		log.Println("db err", err)
		http.Error(w, "Db err", http.StatusInternalServerError)
	}
	if err != nil {
		return
	}

	h.Sessions.Create(r.Context(), w, user)
	http.Redirect(w, r, "/photos/", http.StatusFound)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Sessions.DestroyCurrent(r.Context(), w, r)
	http.Redirect(w, r, "/user/login", http.StatusFound)
}
