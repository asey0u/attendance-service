package web

import (
	"net/http"
	"time"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/domain"
)

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.r.Page(w, http.StatusNotFound, "error", map[string]any{
			"Title":   "Не найдено",
			"Code":    404,
			"Message": "Страница не найдена.",
		})
		return
	}

	if !h.authSvc.HasAdmin(r.Context()) {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	claims := h.claimsOrNil(r)
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, landingFor(claims.Role), http.StatusSeeOther)
}

func (h *Handler) claimsOrNil(r *http.Request) *auth.Claims {
	c, err := r.Cookie(auth.CookieName)
	if err != nil {
		return nil
	}

	claims, err := h.signer.Parse(c.Value)
	if err != nil {
		return nil
	}

	return claims
}

func landingFor(role string) string {
	switch role {
	case domain.RoleAdmin:
		return "/admin"
	case domain.RoleManager:
		return "/manager"
	default:
		return "/me"
	}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if !h.authSvc.HasAdmin(r.Context()) {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	if claims := h.claimsOrNil(r); claims != nil {
		http.Redirect(w, r, landingFor(claims.Role), http.StatusSeeOther)
		return
	}

	h.r.Page(w, http.StatusOK, "login", map[string]any{
		"Title": "Вход",
	})
}

func (h *Handler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.loginError(w, "", "", "Некорректная форма")
		return
	}

	login := r.PostFormValue("login")
	pass := r.PostFormValue("password")

	token, claims, err := h.authSvc.Login(r.Context(), login, pass)
	if err != nil {
		h.loginError(w, login, "", "Неверный логин или пароль")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(auth.TokenTTL),
	})

	http.Redirect(w, r, landingFor(claims.Role), http.StatusSeeOther)
}

func (h *Handler) loginError(w http.ResponseWriter, login, password, msg string) {
	h.r.Page(w, http.StatusUnauthorized, "login", map[string]any{
		"Title":    "Вход",
		"Error":    msg,
		"Login":    login,
		"Password": password,
	})
}

func (h *Handler) LogoutSubmit(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
