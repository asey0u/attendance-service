package admin

type SetupStatusResponse struct {
	Initialized bool `json:"initialized"`
}

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserListItem struct {
	UserID    int    `json:"user_id"`
	Login     string `json:"login"`
	Role      string `json:"role"`
	IsCurrent bool   `json:"is_current"`
}

type UserListResponse struct {
	Users         []UserListItem `json:"users"`
	Page          int            `json:"page"`
	Total         int            `json:"total"`
	CurrentUserID int            `json:"current_user_id"`
}

type CreateUserRequest struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	Role       string `json:"role"`
	EmployeeID int    `json:"employee_id"`
}

type CreateEmployeeRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
}

type EmployeeListItem struct {
	EmployeeID int    `json:"employee_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Position   string `json:"position"`
}

type EmployeeListResponse struct {
	Employees []EmployeeListItem `json:"employees"`
	Page      int                `json:"page"`
	Total     int                `json:"total"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

type CreateAttendanceRequest struct {
	EmployeeID int `json:"employee_id"`
}

type AttendanceListItem struct {
	AttendanceID int    `json:"attendance_id"`
	Date         string `json:"date"`
	TimeIn       string `json:"time_in"`
	Status       string `json:"status"`
}

type AttendanceListResponse struct {
	Attendances []AttendanceListItem `json:"attendances"`
	Page        int                  `json:"page"`
	Total       int                  `json:"total"`
}

type CurrentUserResponse struct {
	UserID     int    `json:"user_id"`
	Login      string `json:"login"`
	Role       string `json:"role"`
	EmployeeID *int   `json:"employee_id,omitempty"`
}
