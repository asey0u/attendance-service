package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey = contextKey("user")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "missing token", 401)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims := &auth.Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return auth.JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", 401)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(r *http.Request) *auth.Claims {
	user := r.Context().Value(UserContextKey)
	if user == nil {
		return nil
	}
	return user.(*auth.Claims)
}
