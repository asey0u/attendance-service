package appsettings

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/asey0u/attendance-service/internal/httpx"
)

const KeyLateThreshold = "late_threshold"
const DefaultLateThreshold = "09:30"

var timeRx = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LateThreshold(ctx context.Context) string {
	v, err := s.repo.Get(ctx, KeyLateThreshold)
	if err != nil || v == "" {
		return DefaultLateThreshold
	}
	return v
}

func (s *Service) SetLateThreshold(ctx context.Context, value string) error {
	if !timeRx.MatchString(value) {
		return httpx.NewError(http.StatusBadRequest,
			fmt.Sprintf("недопустимый формат времени %q — используйте ЧЧ:ММ", value))
	}
	return s.repo.Set(ctx, KeyLateThreshold, value)
}

