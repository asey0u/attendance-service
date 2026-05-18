package domain

const (
	RoleAdmin    = "admin"
	RoleManager  = "manager"
	RoleEmployee = "employee"
)

type User struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	Role         string `json:"role"`
	EmployeeID   *int   `json:"employee_id,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Position     string `json:"position,omitempty"`
	DepartmentID *int   `json:"department_id,omitempty"`
}
