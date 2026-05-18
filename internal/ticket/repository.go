package ticket

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

var ErrNotFound = errors.New("ticket not found")

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

func (r *Repository) Create(ctx context.Context, employeeID int, ticketDate time.Time, reason string) (int, error) {
	var id int
	err := r.dbtx(ctx).QueryRowContext(ctx, `
		INSERT INTO tickets(employee_id, ticket_date, reason)
		VALUES ($1, $2, $3)
		RETURNING id
	`, employeeID, ticketDate, reason).Scan(&id)
	return id, err
}

func (r *Repository) Get(ctx context.Context, id int) (*domain.Ticket, error) {
	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT t.id, t.employee_id, e.first_name, e.last_name,
		       t.ticket_date, t.reason, t.status, t.reviewed_by, t.reviewed_at, t.created_at
		FROM tickets t
		JOIN employees e ON e.id = t.employee_id
		WHERE t.id = $1
	`, id)

	t := &domain.Ticket{}
	if err := row.Scan(&t.ID, &t.EmployeeID, &t.FirstName, &t.LastName,
		&t.TicketDate, &t.Reason, &t.Status, &t.ReviewedBy, &t.ReviewedAt, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return t, nil
}

func (r *Repository) employeeDepartment(ctx context.Context, ticketID int) (*int, error) {
	var deptID *int
	err := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT e.department_id
		FROM tickets t
		JOIN employees e ON e.id = t.employee_id
		WHERE t.id = $1
	`, ticketID).Scan(&deptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return deptID, nil
}

func (r *Repository) DepartmentOf(ctx context.Context, ticketID int) (*int, error) {
	return r.employeeDepartment(ctx, ticketID)
}

func (r *Repository) Count(ctx context.Context, f domain.TicketFilter) (int, error) {
	var (
		conds []string
		args  []any
	)
	add := func(c string, v any) {
		args = append(args, v)
		conds = append(conds, strings.Replace(c, "?", "$"+strconv.Itoa(len(args)), 1))
	}
	if f.Status != "" {
		add("t.status = ?", f.Status)
	}
	if f.EmployeeID != nil {
		add("t.employee_id = ?", *f.EmployeeID)
	}
	if f.DepartmentID != nil {
		add("e.department_id = ?", *f.DepartmentID)
	}

	q := `SELECT COUNT(*) FROM tickets t JOIN employees e ON e.id = t.employee_id`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}

	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *Repository) List(ctx context.Context, f domain.TicketFilter) ([]domain.Ticket, error) {
	var (
		conds []string
		args  []any
	)
	add := func(c string, v any) {
		args = append(args, v)
		conds = append(conds, strings.Replace(c, "?", "$"+strconv.Itoa(len(args)), 1))
	}
	if f.Status != "" {
		add("t.status = ?", f.Status)
	}
	if f.EmployeeID != nil {
		add("t.employee_id = ?", *f.EmployeeID)
	}
	if f.DepartmentID != nil {
		add("e.department_id = ?", *f.DepartmentID)
	}

	q := `
		SELECT t.id, t.employee_id, e.first_name, e.last_name,
		       t.ticket_date, t.reason, t.status, t.reviewed_by, t.reviewed_at, t.created_at
		FROM tickets t
		JOIN employees e ON e.id = t.employee_id
	`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY t.created_at DESC"

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

	var out []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err = rows.Scan(&t.ID, &t.EmployeeID, &t.FirstName, &t.LastName,
			&t.TicketDate, &t.Reason, &t.Status, &t.ReviewedBy, &t.ReviewedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}

	return out, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, id int, status string, reviewerUserID int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `
		UPDATE tickets
		SET status = $1, reviewed_by = $2, reviewed_at = now()
		WHERE id = $3 AND status = 'pending'
	`, status, reviewerUserID, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `DELETE FROM tickets WHERE id = $1`, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}
