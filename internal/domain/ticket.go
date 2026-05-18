package domain

import "time"

const (
	TicketPending  = "pending"
	TicketApproved = "approved"
	TicketDeclined = "declined"
)

type Ticket struct {
	ID         int        `json:"id"`
	EmployeeID int        `json:"employee_id"`
	FirstName  string     `json:"first_name,omitempty"`
	LastName   string     `json:"last_name,omitempty"`
	TicketDate time.Time  `json:"ticket_date"`
	Reason     string     `json:"reason"`
	Status     string     `json:"status"`
	ReviewedBy *int       `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type TicketFilter struct {
	Status       string
	EmployeeID   *int
	DepartmentID *int
	Limit        int
	Offset       int
}
