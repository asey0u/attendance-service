package domain

type Department struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	ManagerID   *int    `json:"manager_id,omitempty"`
	ManagerName *string `json:"manager_name,omitempty"`
}
