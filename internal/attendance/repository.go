package attendance

import (
	"database/sql"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(employeeID int, t time.Time) error {
	_, err := r.DB.Exec(`
		INSERT INTO attendance (employee_id, check_in)
		VALUES ($1, $2)
	`, employeeID, t)

	return err
}

func (r *Repository) GetByEmployee(employeeID int) ([]Attendance, error) {
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
