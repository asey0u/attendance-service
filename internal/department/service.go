package department

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

func (s *Service) List(ctx context.Context) ([]domain.Department, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, httpx.NewError(http.StatusBadRequest, "name is required")
	}
	return s.repo.Create(ctx, name)
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) AssignManager(ctx context.Context, id int, managerUserID *int) error {
	return s.repo.AssignManager(ctx, id, managerUserID)
}
