package attendance

import "time"

type Attendance struct {
	ID         int
	EmployeeID int
	Date       string
	TimeIn     time.Time
	Status     string
}
