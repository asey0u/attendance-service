package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/asey0u/attendance-service/internal/database"
	"github.com/asey0u/attendance-service/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

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

type credentials struct {
	UserID       int
	Login        string
	PasswordHash string
	Role         string
	EmployeeID   *int
	DepartmentID *int
}

func (r *Repository) findCredentials(ctx context.Context, login string) (*credentials, error) {
	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT u.id, u.login, u.password, r.name, u.employee_id, e.department_id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		LEFT JOIN employees e ON e.id = u.employee_id
		WHERE u.login = $1
	`, login)

	var c credentials
	if err := row.Scan(&c.UserID, &c.Login, &c.PasswordHash, &c.Role, &c.EmployeeID, &c.DepartmentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	row := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT u.id, u.login, r.name, u.employee_id,
		       COALESCE(e.first_name, ''), COALESCE(e.last_name, ''),
		       COALESCE(e.position, ''), e.department_id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		LEFT JOIN employees e ON e.id = u.employee_id
		WHERE u.id = $1
	`, id)

	var u domain.User
	if err := row.Scan(&u.ID, &u.Login, &u.Role, &u.EmployeeID,
		&u.FirstName, &u.LastName, &u.Position, &u.DepartmentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *Repository) ListRoles(ctx context.Context) ([]string, error) {
	rows, err := r.dbtx(ctx).QueryContext(ctx, `SELECT name FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}

	return out, rows.Err()
}

func (r *Repository) CountAdmins(ctx context.Context) (int, error) {
	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE r.name = 'admin'
	`).Scan(&n)
	return n, err
}

func (r *Repository) CreateUser(ctx context.Context, login, passwordHash, roleName string, employeeID *int) (int, error) {
	var id int
	err := r.dbtx(ctx).QueryRowContext(ctx, `
		INSERT INTO users(login, password, employee_id, role_id)
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4))
		RETURNING id
	`, login, passwordHash, employeeID, roleName).Scan(&id)
	return id, err
}

func (r *Repository) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := r.dbtx(ctx).QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *Repository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	q := `
		SELECT u.id, u.login, r.name, u.employee_id,
		       COALESCE(e.first_name, ''), COALESCE(e.last_name, ''),
		       COALESCE(e.position, ''), e.department_id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		LEFT JOIN employees e ON e.id = u.employee_id
		ORDER BY u.id
	`
	var args []any
	if limit > 0 {
		args = append(args, limit, offset)
		q += " LIMIT $1 OFFSET $2"
	}

	rows, err := r.dbtx(ctx).QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err = rows.Scan(&u.ID, &u.Login, &u.Role, &u.EmployeeID,
			&u.FirstName, &u.LastName, &u.Position, &u.DepartmentID); err != nil {
			return nil, err
		}
		out = append(out, u)
	}

	return out, rows.Err()
}

func (r *Repository) DeleteUser(ctx context.Context, id int) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) UpdateUserRole(ctx context.Context, id int, roleName string) error {
	res, err := r.dbtx(ctx).ExecContext(ctx, `
		UPDATE users
		SET role_id = (SELECT id FROM roles WHERE name = $1)
		WHERE id = $2
	`, roleName, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}

	return nil
}
