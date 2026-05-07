package auth

import "database/sql"

type User struct {
	ID         int
	Login      string
	Password   string
	EmployeeID sql.NullInt64
	Role       string
}
