package web

import (
	"net/http"
)

func (h *Handler) SetupPage(w http.ResponseWriter, r *http.Request) {
	if h.authSvc.HasAdmin(r.Context()) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.r.Page(w, http.StatusOK, "setup", map[string]any{
		"Title": "Настройка системы",
	})
}

func (h *Handler) SetupSubmit(w http.ResponseWriter, r *http.Request) {
	if h.authSvc.HasAdmin(r.Context()) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderSetup(w, "Некорректная форма")
		return
	}

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")

	if err := h.authSvc.BootstrapAdmin(r.Context(), login, password); err != nil {
		h.renderSetup(w, err.Error())
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) renderSetup(w http.ResponseWriter, errMsg string) {
	h.r.Page(w, http.StatusUnprocessableEntity, "setup", map[string]any{
		"Title": "Настройка системы",
		"Error": errMsg,
	})
}
