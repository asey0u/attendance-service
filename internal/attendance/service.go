package attendance

import (
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CheckIn(role string, employeeID int) error {

	if role != "employee" && role != "admin" {
		return errors.New("forbidden")
	}

	return s.repo.Create(employeeID, time.Now())
}

func (s *Service) GetMyAttendance(employeeID int) ([]Attendance, error) {
	return s.repo.GetByEmployee(employeeID)
}
