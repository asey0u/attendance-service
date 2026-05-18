package ticket

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/httpx"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, claims *auth.Claims, dateStr, reason string) (int, error) {
	if claims.EmployeeID == nil {
		return 0, httpx.NewError(http.StatusBadRequest, "user has no employee profile")
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		return 0, httpx.NewError(http.StatusBadRequest, "reason is required")
	}

	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, httpx.NewError(http.StatusBadRequest, "ticket_date must be YYYY-MM-DD")
	}

	return s.repo.Create(ctx, *claims.EmployeeID, d, reason)
}

func (s *Service) CountMine(ctx context.Context, claims *auth.Claims) (int, error) {
	if claims.EmployeeID == nil {
		return 0, nil
	}

	return s.repo.Count(ctx, domain.TicketFilter{EmployeeID: claims.EmployeeID})
}

func (s *Service) ListMine(ctx context.Context, claims *auth.Claims, limit, offset int) ([]domain.Ticket, error) {
	if claims.EmployeeID == nil {
		return []domain.Ticket{}, nil
	}

	return s.repo.List(ctx, domain.TicketFilter{EmployeeID: claims.EmployeeID, Limit: limit, Offset: offset})
}

func (s *Service) CountForManager(ctx context.Context, claims *auth.Claims, status string) (int, error) {
	f := domain.TicketFilter{Status: status}
	if claims.Role == domain.RoleManager {
		if claims.DepartmentID == nil {
			return 0, nil
		}
		f.DepartmentID = claims.DepartmentID
	}

	return s.repo.Count(ctx, f)
}

func (s *Service) ListForManager(ctx context.Context, claims *auth.Claims, status string, limit, offset int) ([]domain.Ticket, error) {
	f := domain.TicketFilter{Status: status, Limit: limit, Offset: offset}
	if claims.Role == domain.RoleManager {
		if claims.DepartmentID == nil {
			return []domain.Ticket{}, nil
		}
		f.DepartmentID = claims.DepartmentID
	}

	return s.repo.List(ctx, f)
}

func (s *Service) CountAll(ctx context.Context, status string) (int, error) {
	return s.repo.Count(ctx, domain.TicketFilter{Status: status})
}

func (s *Service) ListAll(ctx context.Context, status string, limit, offset int) ([]domain.Ticket, error) {
	return s.repo.List(ctx, domain.TicketFilter{Status: status, Limit: limit, Offset: offset})
}

func (s *Service) Approve(ctx context.Context, claims *auth.Claims, id int) error {
	return s.review(ctx, claims, id, domain.TicketApproved)
}

func (s *Service) Decline(ctx context.Context, claims *auth.Claims, id int) error {
	return s.review(ctx, claims, id, domain.TicketDeclined)
}

func (s *Service) review(ctx context.Context, claims *auth.Claims, id int, status string) error {
	if claims.Role == domain.RoleManager {
		deptID, err := s.repo.DepartmentOf(ctx, id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return httpx.ErrNotFound
			}
			return err
		}
		if claims.DepartmentID == nil || deptID == nil || *claims.DepartmentID != *deptID {
			return httpx.ErrForbidden
		}
	}

	if err := s.repo.UpdateStatus(ctx, id, status, claims.UserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return httpx.NewError(http.StatusConflict, "ticket not found or already reviewed")
		}
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, claims *auth.Claims, id int) error {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return httpx.ErrNotFound
		}
		return err
	}

	switch claims.Role {
	case domain.RoleAdmin:
		// allowed
	case domain.RoleEmployee, domain.RoleManager:
		if claims.EmployeeID == nil || t.EmployeeID != *claims.EmployeeID {
			return httpx.ErrForbidden
		}
		if t.Status != domain.TicketPending {
			return httpx.NewError(http.StatusConflict, "only pending tickets can be deleted")
		}
	default:
		return httpx.ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}
