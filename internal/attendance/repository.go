package attendance

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/asey0u/attendance-service/internal/database"
	"github.com/asey0u/attendance-service/internal/domain"
)

var (
	ErrNotFound      = errors.New("attendance not found")
	ErrAlreadyOpen   = errors.New("already checked in today")
	ErrNoOpenSession = errors.New("no open check-in to close")
)

type Repository struct {
	db database.DBTX
}

func NewRepository(db database.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbtx(ctx context.Context) database.DBTX {
	if db := database.FromContext(ctx); db != nil {
		return db
	}
	return r.db
}

func (r *Repository) OpenSessionToday(ctx context.Context, employeeID int) (*domain.Attendance, error) {
	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT id, employee_id, check_in, check_out
		FROM attendance
		WHERE employee_id = $1
		  AND check_out IS NULL
		ORDER BY check_in DESC
		LIMIT 1
	`, employeeID)

	var a domain.Attendance
	if err := row.Scan(&a.ID, &a.EmployeeID, &a.CheckIn, &a.CheckOut); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &a, nil
}

func (r *Repository) TodaySession(ctx context.Context, employeeID int, threshold string) (*domain.Attendance, error) {
	now := time.Now().Local()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT id, employee_id, check_in, check_out
		FROM attendance
		WHERE employee_id = $1
		  AND check_in >= $2
		  AND check_in <  $3
		ORDER BY check_in DESC
		LIMIT 1
	`, employeeID, startOfDay.UTC(), endOfDay.UTC())

	var a domain.Attendance
	if err := row.Scan(&a.ID, &a.EmployeeID, &a.CheckIn, &a.CheckOut); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	a.Status = domain.DeriveStatus(a.CheckIn, threshold)
	return &a, nil
}

func (r *Repository) CheckIn(ctx context.Context, employeeID int, t time.Time, threshold string) (*domain.Attendance, error) {
	existing, err := r.OpenSessionToday(ctx, employeeID)
	if err == nil {
		existing.Status = domain.DeriveStatus(existing.CheckIn, threshold)
		return existing, ErrAlreadyOpen
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	var id int
	if err = r.dbtx(ctx).QueryRowContext(ctx, `
		INSERT INTO attendance(employee_id, check_in)
		VALUES ($1, $2)
		RETURNING id
	`, employeeID, t).Scan(&id); err != nil {
		return nil, err
	}

	return &domain.Attendance{
		ID:         id,
		EmployeeID: employeeID,
		CheckIn:    t,
		Status:     domain.DeriveStatus(t, threshold),
	}, nil
}

func (r *Repository) CheckOut(ctx context.Context, employeeID int, t time.Time, threshold string) (*domain.Attendance, error) {
	var sessionID int
	err := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT id FROM attendance
		WHERE employee_id = $1 AND check_out IS NULL
		ORDER BY check_in DESC
		LIMIT 1
	`, employeeID).Scan(&sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoOpenSession
		}
		return nil, err
	}

	if _, err = r.dbtx(ctx).ExecContext(ctx, `
		UPDATE attendance SET check_out = $1 WHERE id = $2
	`, t, sessionID); err != nil {
		return nil, err
	}

	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT id, employee_id, check_in, check_out
		FROM attendance WHERE id = $1
	`, sessionID)

	var a domain.Attendance
	if err = row.Scan(&a.ID, &a.EmployeeID, &a.CheckIn, &a.CheckOut); err != nil {
		return nil, err
	}

	a.Status = domain.DeriveStatus(a.CheckIn, threshold)
	return &a, nil
}

func (r *Repository) CountPresentToday(ctx context.Context, deptID *int) (int, error) {
	q := `SELECT COUNT(*) FROM v_attendance_today`
	var args []any
	if deptID != nil {
		args = append(args, *deptID)
		q += ` WHERE department_id = $1`
	}

	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *Repository) StatsForEmployee(ctx context.Context, employeeID int, from, to *time.Time) (domain.AttendanceStats, error) {
	row := r.dbtx(ctx).QueryRowContext(ctx,
		`SELECT days_present, days_late, total_hours, avg_hours
		 FROM fn_employee_stats($1, $2, $3)`,
		employeeID, from, to,
	)

	var s domain.AttendanceStats
	var totalHours, avgHours float64
	if err := row.Scan(&s.DaysPresent, &s.DaysLate, &totalHours, &avgHours); err != nil {
		return domain.AttendanceStats{}, err
	}
	s.TotalHours = totalHours
	s.AverageHours = avgHours
	return s, nil
}

func (r *Repository) CountByEmployee(ctx context.Context, employeeID int, from, to *time.Time) (int, error) {
	var (
		conds = []string{"employee_id = $1"}
		args  = []any{employeeID}
	)
	if from != nil {
		args = append(args, *from)
		conds = append(conds, "check_in >= $"+strconv.Itoa(len(args)))
	}
	if to != nil {
		args = append(args, *to)
		conds = append(conds, "check_in <= $"+strconv.Itoa(len(args)))
	}

	q := `SELECT COUNT(*) FROM attendance WHERE ` + strings.Join(conds, " AND ")

	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *Repository) ListByEmployee(ctx context.Context, employeeID int, from, to *time.Time, threshold string, limit, offset int) ([]domain.Attendance, error) {
	var (
		conds = []string{"employee_id = $1"}
		args  = []any{employeeID}
	)
	if from != nil {
		args = append(args, *from)
		conds = append(conds, "check_in >= $"+strconv.Itoa(len(args)))
	}
	if to != nil {
		args = append(args, *to)
		conds = append(conds, "check_in <= $"+strconv.Itoa(len(args)))
	}

	q := `SELECT id, employee_id, check_in, check_out FROM attendance WHERE ` +
		strings.Join(conds, " AND ") + ` ORDER BY check_in DESC`

	if limit > 0 {
		args = append(args, limit)
		q += " LIMIT $" + strconv.Itoa(len(args))
		args = append(args, offset)
		q += " OFFSET $" + strconv.Itoa(len(args))
	} else {
		q += " LIMIT 500"
	}

	rows, err := r.dbtx(ctx).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Attendance
	for rows.Next() {
		var a domain.Attendance
		if err = rows.Scan(&a.ID, &a.EmployeeID, &a.CheckIn, &a.CheckOut); err != nil {
			return nil, err
		}
		a.Status = domain.DeriveStatus(a.CheckIn, threshold)
		out = append(out, a)
	}

	return out, rows.Err()
}

func (r *Repository) CountFiltered(ctx context.Context, f domain.AttendanceFilter) (int, error) {
	var (
		conds []string
		args  []any
	)
	add := func(c string, v any) {
		args = append(args, v)
		conds = append(conds, strings.Replace(c, "?", "$"+strconv.Itoa(len(args)), 1))
	}

	if f.FirstName != "" {
		add("LOWER(e.first_name) LIKE ?", "%"+strings.ToLower(f.FirstName)+"%")
	}
	if f.LastName != "" {
		add("LOWER(e.last_name) LIKE ?", "%"+strings.ToLower(f.LastName)+"%")
	}
	if f.Position != "" {
		add("LOWER(e.position) LIKE ?", "%"+strings.ToLower(f.Position)+"%")
	}
	if f.DepartmentID != nil {
		add("e.department_id = ?", *f.DepartmentID)
	}
	if f.From != nil {
		add("a.check_in >= ?", *f.From)
	}
	if f.To != nil {
		add("a.check_in <= ?", *f.To)
	}

	q := `SELECT COUNT(*) FROM attendance a JOIN employees e ON e.id = a.employee_id`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}

	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *Repository) ListFiltered(ctx context.Context, f domain.AttendanceFilter, threshold string) ([]domain.Attendance, error) {
	var (
		conds []string
		args  []any
	)
	add := func(c string, v any) {
		args = append(args, v)
		conds = append(conds, strings.Replace(c, "?", "$"+strconv.Itoa(len(args)), 1))
	}

	if f.FirstName != "" {
		add("LOWER(e.first_name) LIKE ?", "%"+strings.ToLower(f.FirstName)+"%")
	}
	if f.LastName != "" {
		add("LOWER(e.last_name) LIKE ?", "%"+strings.ToLower(f.LastName)+"%")
	}
	if f.Position != "" {
		add("LOWER(e.position) LIKE ?", "%"+strings.ToLower(f.Position)+"%")
	}
	if f.DepartmentID != nil {
		add("e.department_id = ?", *f.DepartmentID)
	}
	if f.From != nil {
		add("a.check_in >= ?", *f.From)
	}
	if f.To != nil {
		add("a.check_in <= ?", *f.To)
	}

	q := `
		SELECT a.id, a.employee_id, a.check_in, a.check_out,
		       e.first_name, e.last_name, e.position
		FROM attendance a
		JOIN employees e ON e.id = a.employee_id
	`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY a.check_in DESC"

	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += " LIMIT $" + strconv.Itoa(len(args))
		args = append(args, f.Offset)
		q += " OFFSET $" + strconv.Itoa(len(args))
	} else {
		q += " LIMIT 500"
	}

	rows, err := r.dbtx(ctx).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Attendance
	for rows.Next() {
		var a domain.Attendance
		if err = rows.Scan(&a.ID, &a.EmployeeID, &a.CheckIn, &a.CheckOut,
			&a.FirstName, &a.LastName, &a.Position); err != nil {
			return nil, err
		}
		a.Status = domain.DeriveStatus(a.CheckIn, threshold)
		out = append(out, a)
	}

	return out, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `DELETE FROM attendance WHERE id = $1`, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}
