package employee

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/asey0u/attendance-service/internal/database"
	"github.com/asey0u/attendance-service/internal/domain"
)

var ErrNotFound = errors.New("employee not found")

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

func buildEmployeeConds(f domain.EmployeeFilter) ([]string, []any) {
	var conds []string
	var args []any

	if f.Search != "" {
		s := "%" + strings.ToLower(f.Search) + "%"
		args = append(args, s, s, s)
		n := len(args)
		conds = append(conds,
			"(LOWER(e.first_name) LIKE $"+strconv.Itoa(n-2)+
				" OR LOWER(e.last_name) LIKE $"+strconv.Itoa(n-1)+
				" OR LOWER(e.position) LIKE $"+strconv.Itoa(n)+")")
	} else {
		if f.FirstName != "" {
			args = append(args, "%"+strings.ToLower(f.FirstName)+"%")
			conds = append(conds, "LOWER(e.first_name) LIKE $"+strconv.Itoa(len(args)))
		}
		if f.LastName != "" {
			args = append(args, "%"+strings.ToLower(f.LastName)+"%")
			conds = append(conds, "LOWER(e.last_name) LIKE $"+strconv.Itoa(len(args)))
		}
		if f.Position != "" {
			args = append(args, "%"+strings.ToLower(f.Position)+"%")
			conds = append(conds, "LOWER(e.position) LIKE $"+strconv.Itoa(len(args)))
		}
	}
	if f.DepartmentID != nil {
		args = append(args, *f.DepartmentID)
		conds = append(conds, "e.department_id = $"+strconv.Itoa(len(args)))
	}
	return conds, args
}

func (r *Repository) Count(ctx context.Context, f domain.EmployeeFilter) (int, error) {
	conds, args := buildEmployeeConds(f)

	q := `SELECT COUNT(*) FROM employees e`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}

	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *Repository) List(ctx context.Context, f domain.EmployeeFilter) ([]domain.Employee, error) {
	conds, args := buildEmployeeConds(f)

	q := `
		SELECT e.id, e.first_name, e.last_name, e.position, e.department_id, d.name
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
	`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY e.last_name, e.first_name"

	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += " LIMIT $" + strconv.Itoa(len(args))
		args = append(args, f.Offset)
		q += " OFFSET $" + strconv.Itoa(len(args))
	}

	rows, err := r.dbtx(ctx).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Employee
	for rows.Next() {
		var e domain.Employee
		if err = rows.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Position, &e.DepartmentID, &e.DepartmentName); err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	return out, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id int) (*domain.Employee, error) {
	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT e.id, e.first_name, e.last_name, e.position, e.department_id, d.name
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.id = $1
	`, id)

	var e domain.Employee
	if err := row.Scan(&e.ID, &e.FirstName, &e.LastName, &e.Position, &e.DepartmentID, &e.DepartmentName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &e, nil
}

func (r *Repository) Create(ctx context.Context, e *domain.Employee) (int, error) {
	var id int
	err := r.dbtx(ctx).QueryRowContext(ctx, `
		INSERT INTO employees(first_name, last_name, position, department_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, e.FirstName, e.LastName, e.Position, e.DepartmentID).Scan(&id)
	return id, err
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) AssignDepartment(ctx context.Context, id int, departmentID *int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `UPDATE employees SET department_id = $1 WHERE id = $2`, departmentID, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}
