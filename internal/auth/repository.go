package auth

import "database/sql"

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetByLogin(login string) (*User, error) {
	var u User

	err := r.DB.QueryRow(`
		SELECT u.user_id, u.password, u.employee_id, r.name
		FROM users u
		JOIN roles r ON u.role_id = r.role_id
		WHERE u.login = $1
	`, login).Scan(&u.ID, &u.Password, &u.EmployeeID, &u.Role)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *Repository) Create(user User) error {
	_, err := r.DB.Exec(`
		INSERT INTO users (login, password, employee_id, role_id)
		VALUES ($1, $2, $3,
			(SELECT role_id FROM roles WHERE name = $4)
		)
	`, user.Login, user.Password, user.EmployeeID, user.Role)

	return err
}
