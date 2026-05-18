package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenTTL = 24 * time.Hour

type Claims struct {
	UserID       int    `json:"uid"`
	Login        string `json:"login"`
	Role         string `json:"role"`
	EmployeeID   *int   `json:"eid,omitempty"`
	DepartmentID *int   `json:"did,omitempty"`
	jwt.RegisteredClaims
}

type ctxKey struct{}

var claimsCtxKey = ctxKey{}

func WithClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsCtxKey, c)
}

func FromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsCtxKey).(*Claims)

	return c
}

type Signer struct {
	secret []byte
}

func NewSigner(secret []byte) *Signer {
	return &Signer{secret: secret}
}

func (s *Signer) Sign(c *Claims) (string, error) {
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	return tok.SignedString(s.secret)
}

func (s *Signer) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})

	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
