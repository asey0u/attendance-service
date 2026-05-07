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
		INSERT INTO attendance (employee_id, date, time_in, status)
		VALUES ($1, $2, $3, $4)
	`, employeeID, t.Format("2006-01-02"), t, "present")

	return err
}

func (r *Repository) GetByEmployee(employeeID int) ([]Attendance, error) {
	rows, err := r.DB.Query(`
		SELECT attendance_id, employee_id, date, time_in, status
		FROM attendance
		WHERE employee_id = $1
		ORDER BY date DESC
	`, employeeID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Attendance

	for rows.Next() {
		var a Attendance
		err := rows.Scan(&a.ID, &a.EmployeeID, &a.Date, &a.TimeIn, &a.Status)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}

	return result, nil
}
