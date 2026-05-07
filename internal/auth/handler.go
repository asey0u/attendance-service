package auth

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login      string `json:"login"`
		Password   string `json:"password"`
		EmployeeID int    `json:"employee_id"`
		Role       string `json:"role"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	err := h.service.Register(req.Login, req.Password, req.EmployeeID, req.Role)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	w.Write([]byte("user created"))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	token, err := h.service.Login(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
