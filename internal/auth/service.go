package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/httpx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo   *Repository
	signer *Signer
}

func NewService(repo *Repository, signer *Signer) *Service {
	return &Service{repo: repo, signer: signer}
}

func (s *Service) Login(ctx context.Context, login, password string) (string, *Claims, error) {
	creds, err := s.repo.findCredentials(ctx, login)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", nil, httpx.NewError(http.StatusUnauthorized, "invalid login or password")
		}
		return "", nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(password)); err != nil {
		return "", nil, httpx.NewError(http.StatusUnauthorized, "invalid login or password")
	}

	claims := &Claims{
		UserID:       creds.UserID,
		Login:        creds.Login,
		Role:         creds.Role,
		EmployeeID:   creds.EmployeeID,
		DepartmentID: creds.DepartmentID,
	}

	token, err := s.signer.Sign(claims)
	if err != nil {
		return "", nil, err
	}

	return token, claims, nil
}

func (s *Service) BootstrapAdmin(ctx context.Context, login, password string) error {
	n, err := s.repo.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return httpx.NewError(http.StatusConflict, "admin already exists")
	}

	if len(login) < 3 || len(password) < 6 {
		return httpx.NewError(http.StatusBadRequest, "login >= 3 chars, password >= 6 chars")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err = s.repo.CreateUser(ctx, login, string(hash), domain.RoleAdmin, nil); err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateUser(ctx context.Context, login, password, role string, employeeID *int) (int, error) {
	if role != domain.RoleAdmin && role != domain.RoleManager && role != domain.RoleEmployee {
		return 0, httpx.NewError(http.StatusBadRequest, "Недопустимая роль")
	}
	if len([]rune(login)) < 3 {
		return 0, httpx.NewError(http.StatusBadRequest, "Логин должен содержать не менее 3 символов")
	}
	if len([]rune(password)) < 6 {
		return 0, httpx.NewError(http.StatusBadRequest, "Пароль должен содержать не менее 6 символов")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	id, err := s.repo.CreateUser(ctx, login, string(hash), role, employeeID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, httpx.NewError(http.StatusConflict, "Пользователь с таким логином уже существует")
		}
		return 0, err
	}

	return id, nil
}

func (s *Service) HasAdmin(ctx context.Context) bool {
	n, err := s.repo.CountAdmins(ctx)
	return err == nil && n > 0
}

func (s *Service) GetUser(ctx context.Context, id int) (*domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context) ([]string, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) CountUsers(ctx context.Context) (int, error) {
	return s.repo.CountUsers(ctx)
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	return s.repo.ListUsers(ctx, limit, offset)
}

func (s *Service) DeleteUser(ctx context.Context, id int) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *Service) UpdateRole(ctx context.Context, id int, role string) error {
	if role != domain.RoleAdmin && role != domain.RoleManager && role != domain.RoleEmployee {
		return httpx.NewError(http.StatusBadRequest, "invalid role")
	}

	return s.repo.UpdateUserRole(ctx, id, role)
}
