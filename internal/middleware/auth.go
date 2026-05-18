package middleware

import (
	"net/http"
	"strings"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/httpx"
)

func Auth(signer *auth.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := tokenFromRequest(r)
			if token == "" {
				httpx.WriteError(w, httpx.ErrUnauthorized)
				return
			}

			claims, err := signer.Parse(token)
			if err != nil {
				httpx.WriteError(w, httpx.ErrUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}

func tokenFromRequest(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if strings.HasPrefix(h, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		}
	}

	if c, err := r.Cookie(auth.CookieName); err == nil {
		return c.Value
	}

	return ""
}
