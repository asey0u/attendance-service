package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/asey0u/attendance-service/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func parsePagination(r *http.Request) (int, int) {
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	size := 9
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
		}
	}

	return page, size
}

func decodeJSON(r *http.Request, dst interface{}) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func writeJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, size := parsePagination(r)

	users, total, err := h.service.GetUsers(page, size)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	user := middleware.GetUser(r)
	currentUserID := user.UserID

	response := map[string]interface{}{
		"users":           users,
		"page":            page,
		"total":           total,
		"current_user_id": currentUserID,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login      string `json:"login"`
		Password   string `json:"password"`
		Role       string `json:"role"`
		EmployeeID int    `json:"employee_id"`
	}

	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.Login == "" || req.Password == "" || req.Role == "" || req.EmployeeID <= 0 {
		http.Error(w, "login, password, role and employee_id are required", 400)
		return
	}

	err := h.service.CreateUser(req.Login, req.Password, req.EmployeeID, req.Role)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("user created"))
}

func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	page, size := parsePagination(r)

	employees, total, err := h.service.ListEmployees(page, size)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	items := make([]EmployeeListItem, 0, len(employees))
	for _, e := range employees {
		items = append(items, EmployeeListItem{
			EmployeeID: e.ID,
			FirstName:  e.FirstName,
			LastName:   e.LastName,
			Position:   e.Position,
		})
	}

	response := EmployeeListResponse{
		Employees: items,
		Page:      page,
		Total:     total,
	}

	if err := writeJSON(w, response); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "first_name and last_name are required", 400)
		return
	}

	err := h.service.CreateEmployee(req.FirstName, req.LastName, req.Position)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("employee created"))
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", 400)
		return
	}

	user := middleware.GetUser(r)
	if user.UserID == userID {
		http.Error(w, "cannot delete yourself", 400)
		return
	}

	err = h.service.DeleteUser(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("user deleted"))
}

func (h *Handler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", 400)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.Role == "" {
		http.Error(w, "role is required", 400)
		return
	}

	user := middleware.GetUser(r)
	if user.UserID == userID {
		http.Error(w, "cannot change your own role", 400)
		return
	}

	err = h.service.UpdateUserRole(userID, req.Role)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("user role updated"))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, size := parsePagination(r)

	users, total, err := h.service.ListUsers(page, size)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	user := middleware.GetUser(r)
	currentUserID := user.UserID

	userListItems := make([]UserListItem, 0, len(users))
	for _, u := range users {
		userListItems = append(userListItems, UserListItem{
			UserID:    u.ID,
			Login:     u.Login,
			Role:      u.Role,
			IsCurrent: u.ID == currentUserID,
		})
	}

	response := UserListResponse{
		Users:         userListItems,
		Page:          page,
		Total:         total,
		CurrentUserID: currentUserID,
	}

	if err := writeJSON(w, response); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *Handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	initialized, err := h.service.IsAdminInitialized()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{
		"initialized": initialized,
	})
}

func (h *Handler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password are required", 400)
		return
	}

	initialized, err := h.service.IsAdminInitialized()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if initialized {
		http.Error(w, "admin already initialized", 400)
		return
	}

	err = h.service.CreateAdmin(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("admin added successfully"))
}

func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	userData, err := h.service.GetUserByID(user.UserID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	response := map[string]any{
		"user_id": userData.ID,
		"login":   userData.Login,
		"role":    userData.Role,
	}

	if userData.EmployeeID.Valid {
		empID := int(userData.EmployeeID.Int64)
		response["employee_id"] = empID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idPath := strings.Trim(strings.TrimPrefix(r.URL.Path, "/users/"), "/")
	if idPath == "" {
		http.Error(w, "user id is required", 400)
		return
	}

	userID, err := strconv.Atoi(idPath)
	if err != nil {
		http.Error(w, "invalid user id", 400)
		return
	}

	userData, err := h.service.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "user not found", 404)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}

	response := CurrentUserResponse{
		UserID: userData.ID,
		Login:  userData.Login,
		Role:   userData.Role,
	}
	if userData.EmployeeID.Valid {
		empID := int(userData.EmployeeID.Int64)
		response.EmployeeID = &empID
	}

	if err := writeJSON(w, response); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *Handler) GetAttendancesByEmployee(w http.ResponseWriter, r *http.Request) {
	employeeIDStr := r.URL.Query().Get("employee_id")
	employeeID, err := strconv.Atoi(employeeIDStr)
	if err != nil || employeeID <= 0 {
		http.Error(w, "invalid employee id", 400)
		return
	}

	attendances, err := h.service.GetAttendancesByEmployee(employeeID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	items := make([]AttendanceListItem, 0, len(attendances))
	for _, a := range attendances {
		items = append(items, AttendanceListItem{
			AttendanceID: a.ID,
			Date:         a.Date,
			TimeIn:       a.TimeIn.Format("15:04:05"),
			Status:       a.Status,
		})
	}

	response := AttendanceListResponse{
		Attendances: items,
		Page:        1,
		Total:       len(items),
	}

	if err := writeJSON(w, response); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *Handler) CreateAttendance(w http.ResponseWriter, r *http.Request) {
	var req CreateAttendanceRequest
	if err := decodeJSON(r, &req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if req.EmployeeID <= 0 {
		http.Error(w, "employee_id is required", 400)
		return
	}

	if err := h.service.CreateAttendance(req.EmployeeID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("attendance created"))
}

func (h *Handler) DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	attendanceIDStr := r.URL.Query().Get("attendance_id")
	attendanceID, err := strconv.Atoi(attendanceIDStr)
	if err != nil || attendanceID <= 0 {
		http.Error(w, "invalid attendance id", 400)
		return
	}

	if err := h.service.DeleteAttendance(attendanceID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "attendance not found", 404)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("attendance deleted"))
}
