package employee

import (
	"errors"
	"net/http"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListForManager(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	f := filterFromQuery(r)
	if claims.Role == domain.RoleManager {
		f.DepartmentID = claims.DepartmentID
	}

	items, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context(), filterFromQuery(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

type createRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Position     string `json:"position"`
	DepartmentID *int   `json:"department_id,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	id, err := h.svc.Create(r.Context(), &domain.Employee{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Position:     req.Position,
		DepartmentID: req.DepartmentID,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err = h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type assignDeptRequest struct {
	DepartmentID *int `json:"department_id"`
}

func (h *Handler) AssignDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req assignDeptRequest
	if err = httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err = h.svc.AssignDepartment(r.Context(), id, req.DepartmentID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func filterFromQuery(r *http.Request) domain.EmployeeFilter {
	q := r.URL.Query()
	return domain.EmployeeFilter{
		FirstName: q.Get("first_name"),
		LastName:  q.Get("last_name"),
		Position:  q.Get("position"),
	}
}
