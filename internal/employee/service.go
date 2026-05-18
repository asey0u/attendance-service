package employee

import (
	"context"
	"net/http"
	"strings"

	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/httpx"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Count(ctx context.Context, f domain.EmployeeFilter) (int, error) {
	return s.repo.Count(ctx, f)
}

func (s *Service) List(ctx context.Context, f domain.EmployeeFilter) ([]domain.Employee, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) Get(ctx context.Context, id int) (*domain.Employee, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, e *domain.Employee) (int, error) {
	e.FirstName = strings.TrimSpace(e.FirstName)
	e.LastName = strings.TrimSpace(e.LastName)
	e.Position = strings.TrimSpace(e.Position)
	if e.FirstName == "" || e.LastName == "" || e.Position == "" {
		return 0, httpx.NewError(http.StatusBadRequest, "first_name, last_name, position are required")
	}

	return s.repo.Create(ctx, e)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) AssignDepartment(ctx context.Context, id int, departmentID *int) error {
	return s.repo.AssignDepartment(ctx, id, departmentID)
}
