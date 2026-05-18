package middleware

import (
	"net/http"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/httpx"
)

func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := auth.FromContext(r.Context())
			if c == nil {
				httpx.WriteError(w, httpx.ErrUnauthorized)
				return
			}

			if _, ok := allowed[c.Role]; !ok {
				httpx.WriteError(w, httpx.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
