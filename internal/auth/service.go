package auth

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Register(login, password string, employeeID int, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Login:      login,
		Password:   string(hash),
		EmployeeID: sql.NullInt64{Int64: int64(employeeID), Valid: true},
		Role:       role,
	}

	return s.repo.Create(user)
}

func (s *Service) Login(login, password string) (string, error) {
	user, err := s.repo.GetByLogin(login)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	return GenerateToken(user)
}
