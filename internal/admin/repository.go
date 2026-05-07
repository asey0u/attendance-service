package admin

import (
	"database/sql"
	"errors"
	"time"

	"github.com/asey0u/attendance-service/internal/auth"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(user *auth.User) error {
	if user.EmployeeID.Int64 == adminEmployeeID && user.Role == "admin" {
		user.EmployeeID = sql.NullInt64{Valid: false}
	} else if user.EmployeeID.Int64 == adminEmployeeID {
		return errors.New("employee_id is required for non-admin users")
	}

	_, err := r.DB.Exec(`
		INSERT INTO users (login, password, employee_id, role_id)
		VALUES ($1, $2, $3,
			(SELECT id FROM roles WHERE name = $4)
		)
	`, user.Login, user.Password, user.EmployeeID, user.Role)

	return err
}

func (r *Repository) GetByLogin(login string) (*auth.User, error) {
	var u auth.User

	err := r.DB.QueryRow(`
		SELECT u.id, u.password, u.employee_id, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.login = $1
	`, login).Scan(&u.ID, &u.Password, &u.EmployeeID, &u.Role)

	if err != nil {
		return nil, err
	}

	u.Login = login
	return &u, nil
}

func (r *Repository) HasAdmin() (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE r.name = $1
	`, "admin").Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) GetUsers(page, limit int) ([]auth.User, int, error) {
	offset := (page - 1) * limit
	return r.ListUsers(limit, offset)
}

func (r *Repository) ListUsers(limit, offset int) ([]auth.User, int, error) {
	var total int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.DB.Query(`
		SELECT u.id, u.login, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		ORDER BY u.id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]auth.User, 0)
	for rows.Next() {
		var u auth.User
		if err := rows.Scan(&u.ID, &u.Login, &u.Role); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

type Employee struct {
	ID        int
	FirstName string
	LastName  string
	Position  string
}

func (r *Repository) ListEmployees(limit, offset int) ([]Employee, int, error) {
	var total int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM employees`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.DB.Query(`
		SELECT id, first_name, last_name, position
		FROM employees
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	employees := make([]Employee, 0)
	for rows.Next() {
		var e Employee
		if err := rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Position); err != nil {
			return nil, 0, err
		}
		employees = append(employees, e)
	}

	return employees, total, nil
}

func (r *Repository) CreateEmployee(firstName, lastName, position string) error {
	_, err := r.DB.Exec(`
		INSERT INTO employees (first_name, last_name, position)
		VALUES ($1, $2, $3)
	`, firstName, lastName, position)

	return err
}

type Attendance struct {
	ID         int
	EmployeeID int
	Date       string
	TimeIn     time.Time
	Status     string
}

func (r *Repository) GetAttendancesByEmployee(employeeID int) ([]Attendance, error) {
	rows, err := r.DB.Query(`
		SELECT id, employee_id, check_in
		FROM attendance
		WHERE employee_id = $1
		ORDER BY check_in DESC
	`, employeeID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Attendance

	for rows.Next() {
		var a Attendance
		err := rows.Scan(&a.ID, &a.EmployeeID, &a.TimeIn)
		if err != nil {
			return nil, err
		}
		a.Date = a.TimeIn.Format("2006-01-02")
		a.Status = "present"
		result = append(result, a)
	}

	return result, nil
}

func (r *Repository) CreateAttendance(employeeID int, t time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO attendance (employee_id, check_in)
		VALUES ($1, $2)
	`, employeeID, t)

	return err
}

func (r *Repository) DeleteAttendance(attendanceID int) error {
	res, err := r.DB.Exec(`DELETE FROM attendance WHERE id = $1`, attendanceID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) DeleteUser(userID int) error {
	res, err := r.DB.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) UpdateUserRole(userID int, role string) error {
	res, err := r.DB.Exec(`
		UPDATE users
		SET role_id = (SELECT id FROM roles WHERE name = $2)
		WHERE id = $1
	`, userID, role)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user or role not found")
	}

	return nil
}

func (r *Repository) GetUserByID(userID int) (*auth.User, error) {
	var u auth.User

	err := r.DB.QueryRow(`
		SELECT u.id, u.login, u.employee_id, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`, userID).Scan(&u.ID, &u.Login, &u.EmployeeID, &u.Role)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
