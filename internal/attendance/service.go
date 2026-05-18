package attendance

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/asey0u/attendance-service/internal/appsettings"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/httpx"
)

type Service struct {
	repo     *Repository
	settings *appsettings.Service
}

func NewService(repo *Repository, settings *appsettings.Service) *Service {
	return &Service{repo: repo, settings: settings}
}

func (s *Service) threshold(ctx context.Context) string {
	return s.settings.LateThreshold(ctx)
}

func (s *Service) OpenSession(ctx context.Context, employeeID int) (*domain.Attendance, error) {
	a, err := s.repo.OpenSessionToday(ctx, employeeID)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	a.Status = domain.DeriveStatus(a.CheckIn, s.threshold(ctx))
	return a, nil
}

func (s *Service) TodaySession(ctx context.Context, employeeID int) (*domain.Attendance, error) {
	return s.repo.TodaySession(ctx, employeeID, s.threshold(ctx))
}

func (s *Service) CheckIn(ctx context.Context, employeeID int) (*domain.Attendance, error) {
	a, err := s.repo.CheckIn(ctx, employeeID, time.Now().UTC(), s.threshold(ctx))
	if err != nil && !errors.Is(err, ErrAlreadyOpen) {
		return nil, err
	}

	return a, nil
}

func (s *Service) CheckOut(ctx context.Context, employeeID int) (*domain.Attendance, error) {
	a, err := s.repo.CheckOut(ctx, employeeID, time.Now().UTC(), s.threshold(ctx))
	if errors.Is(err, ErrNoOpenSession) {
		return nil, httpx.NewError(http.StatusBadRequest, "нет незакрытого прихода на сегодня")
	}

	return a, err
}

func (s *Service) CountByEmployee(ctx context.Context, employeeID int, from, to *time.Time) (int, error) {
	return s.repo.CountByEmployee(ctx, employeeID, from, to)
}

func (s *Service) MyHistory(ctx context.Context, employeeID int, from, to *time.Time, limit, offset int) ([]domain.Attendance, error) {
	return s.repo.ListByEmployee(ctx, employeeID, from, to, s.threshold(ctx), limit, offset)
}

func (s *Service) MyStats(ctx context.Context, employeeID int, from, to *time.Time) (domain.AttendanceStats, error) {
	rows, err := s.repo.ListByEmployee(ctx, employeeID, from, to, s.threshold(ctx), 0, 0)
	if err != nil {
		return domain.AttendanceStats{}, err
	}

	var stats domain.AttendanceStats
	for _, a := range rows {
		stats.DaysPresent++
		if a.Status == domain.AttendanceStatusLate {
			stats.DaysLate++
		}
		if a.CheckOut != nil {
			stats.TotalHours += a.CheckOut.Sub(a.CheckIn).Hours()
		}
	}

	if stats.DaysPresent > 0 {
		stats.AverageHours = stats.TotalHours / float64(stats.DaysPresent)
	}

	return stats, nil
}

func (s *Service) CountFiltered(ctx context.Context, f domain.AttendanceFilter) (int, error) {
	return s.repo.CountFiltered(ctx, f)
}

func (s *Service) ListFiltered(ctx context.Context, f domain.AttendanceFilter) ([]domain.Attendance, error) {
	return s.repo.ListFiltered(ctx, f, s.threshold(ctx))
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
