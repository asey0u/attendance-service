package auth

type User struct {
	ID         int
	Login      string
	Password   string
	EmployeeID int
	Role       string
}
