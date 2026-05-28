package department

import (
	"context"
	"errors"

	"github.com/asey0u/attendance-service/internal/database"
	"github.com/asey0u/attendance-service/internal/domain"
)

var ErrNotFound = errors.New("department not found")

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

func (r *Repository) List(ctx context.Context) ([]domain.Department, error) {
	rows, err := r.dbtx(ctx).QueryContext(ctx, `
		SELECT d.id, d.name, d.manager_id,
		       CASE WHEN d.manager_id IS NULL THEN NULL
		            ELSE (
		                COALESCE(e.first_name, '') || ' ' || COALESCE(e.last_name, '')
		            )
		       END AS manager_name
		FROM departments d
		LEFT JOIN users u ON u.id = d.manager_id
		LEFT JOIN employees e ON e.id = u.employee_id
		ORDER BY d.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Department
	for rows.Next() {
		var d domain.Department
		if err = rows.Scan(&d.ID, &d.Name, &d.ManagerID, &d.ManagerName); err != nil {
			return nil, err
		}
		out = append(out, d)
	}

	return out, rows.Err()
}

func (r *Repository) Create(ctx context.Context, name string) (int, error) {
	var id int
	err := r.dbtx(ctx).QueryRowContext(ctx,
		`INSERT INTO departments(name) VALUES ($1) RETURNING id`, name).Scan(&id)
	return id, err
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	tx, err := database.BeginTx(ctx, r.dbtx(ctx))
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM departments WHERE id = $1`, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return tx.Commit()
}

func (r *Repository) AssignManager(ctx context.Context, id int, managerUserID *int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx,
		`UPDATE departments SET manager_id = $1 WHERE id = $2`, managerUserID, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}
