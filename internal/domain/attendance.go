package domain

import (
	"fmt"
	"time"
)

const (
	AttendanceStatusPresent = "present"
	AttendanceStatusLate    = "late"
	AttendanceStatusAbsent  = "absent"
)

type Attendance struct {
	ID         int        `json:"id"`
	EmployeeID int        `json:"employee_id"`
	FirstName  string     `json:"first_name,omitempty"`
	LastName   string     `json:"last_name,omitempty"`
	Position   string     `json:"position,omitempty"`
	CheckIn    time.Time  `json:"check_in"`
	CheckOut   *time.Time `json:"check_out,omitempty"`
	Status     string     `json:"status"`
}

type AttendanceFilter struct {
	FirstName    string
	LastName     string
	Position     string
	DepartmentID *int
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

type AttendanceStats struct {
	DaysPresent  int     `json:"days_present"`
	DaysLate     int     `json:"days_late"`
	TotalHours   float64 `json:"total_hours"`
	AverageHours float64 `json:"average_hours"`
}

// DeriveStatus returns "present" or "late" based on the configurable threshold.
// threshold is "HH:MM" (e.g. "09:30"). Falls back to "09:30" on parse error.
func DeriveStatus(checkIn time.Time, threshold string) string {
	var h, m int
	if _, err := fmt.Sscanf(threshold, "%d:%d", &h, &m); err != nil {
		h, m = 9, 30
	}
	cutoff := time.Date(checkIn.Year(), checkIn.Month(), checkIn.Day(), h, m, 0, 0, checkIn.Location())
	if checkIn.After(cutoff) {
		return AttendanceStatusLate
	}
	return AttendanceStatusPresent
}
