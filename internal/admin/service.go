package admin

import (
	"database/sql"
	"errors"
	"time"

	"github.com/asey0u/attendance-service/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

const adminEmployeeID = -1

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateAdmin(login, password string) error {
	exists, err := s.repo.HasAdmin()
	if err != nil {
		return err
	}
	if exists {
		return errors.New("admin already initialized")
	}

	return s.CreateUser(login, password, adminEmployeeID, "admin")
}

func (s *Service) CreateUser(login, password string, employeeID int, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &auth.User{
		Login:      login,
		Password:   string(hash),
		EmployeeID: sql.NullInt64{Int64: int64(employeeID), Valid: true},
		Role:       role,
	}

	return s.repo.Create(user)
}

func (s *Service) GetUsers(page, limit int) ([]auth.User, int, error) {
	return s.repo.GetUsers(page, limit)
}

func (s *Service) DeleteUser(userID int) error {
	return s.repo.DeleteUser(userID)
}

func (s *Service) UpdateUserRole(userID int, role string) error {
	if role != "admin" && role != "employee" {
		return errors.New("invalid role")
	}
	return s.repo.UpdateUserRole(userID, role)
}

func (s *Service) CreateEmployee(firstName, lastName, position string) error {
	return s.repo.CreateEmployee(firstName, lastName, position)
}

func (s *Service) ListEmployees(page, size int) ([]Employee, int, error) {
	offset := (page - 1) * size
	return s.repo.ListEmployees(size, offset)
}

func (s *Service) IsAdminInitialized() (bool, error) {
	return s.repo.HasAdmin()
}

func (s *Service) ListUsers(page, size int) ([]auth.User, int, error) {
	offset := (page - 1) * size
	return s.repo.ListUsers(size, offset)
}

func (s *Service) GetUserByID(userID int) (*auth.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *Service) GetAttendancesByEmployee(employeeID int) ([]Attendance, error) {
	return s.repo.GetAttendancesByEmployee(employeeID)
}

func (s *Service) CreateAttendance(employeeID int) error {
	return s.repo.CreateAttendance(employeeID, time.Now())
}

func (s *Service) DeleteAttendance(attendanceID int) error {
	return s.repo.DeleteAttendance(attendanceID)
}
