package web

import (
	"context"
	"net/http"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/database"
	"github.com/asey0u/attendance-service/internal/domain"
)

func (h *Handler) roleDB(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.FromContext(r.Context())
		if claims == nil {
			next.ServeHTTP(w, r)
			return
		}

		conn, err := h.db.Conn(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		dbRole := "attendance_employee"
		switch claims.Role {
		case domain.RoleAdmin:
			dbRole = "attendance_admin"
		case domain.RoleManager:
			dbRole = "attendance_manager"
		}

		if _, err = conn.ExecContext(r.Context(), "SET ROLE "+dbRole); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.ExecContext(context.Background(), "RESET ROLE")

		next.ServeHTTP(w, r.WithContext(database.WithDB(r.Context(), conn)))
	})
}

func requireAuth(signer *auth.Signer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if c, err := r.Cookie(auth.CookieName); err == nil {
			token = c.Value
		}
		if token == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		claims, err := signer.Parse(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
	})
}

func (h *Handler) requireRole(next http.Handler, allowed ...string) http.Handler {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		allowedSet[a] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := auth.FromContext(r.Context())
		if c == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if _, ok := allowedSet[c.Role]; !ok {
			h.renderForbidden(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) renderForbidden(w http.ResponseWriter, r *http.Request) {
	h.r.Page(w, http.StatusForbidden, "error", map[string]any{
		"Title":   "Доступ запрещён",
		"Claims":  auth.FromContext(r.Context()),
		"Code":    403,
		"Message": "У вас нет доступа к этой странице.",
	})
}
