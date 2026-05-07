package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = []byte(os.Getenv("JWT-secret"))

type Claims struct {
	UserID     int
	EmployeeID int
	Role       string
	jwt.RegisteredClaims
}

func GenerateToken(user *User) (string, error) {
	claims := Claims{
		UserID:     user.ID,
		EmployeeID: user.EmployeeID,
		Role:       user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtKey)
}
