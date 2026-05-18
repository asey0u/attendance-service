package domain

type Employee struct {
	ID             int     `json:"id"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Position       string  `json:"position"`
	DepartmentID   *int    `json:"department_id,omitempty"`
	DepartmentName *string `json:"department_name,omitempty"`
}

type EmployeeFilter struct {
	FirstName    string
	LastName     string
	Position     string
	Search       string
	DepartmentID *int
	Limit        int
	Offset       int
}
