package attendance

import (
	"encoding/json"
	"net/http"

	"github.com/asey0u/attendance-service/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}

	err := h.service.CheckIn(user.Role, user.EmployeeID)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}

	w.Write([]byte("checked in"))
}

func (h *Handler) MyAttendance(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	data, err := h.service.GetMyAttendance(user.EmployeeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(data)
}
