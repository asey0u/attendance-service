package department

import (
	"errors"
	"net/http"

	"github.com/asey0u/attendance-service/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

type createRequest struct {
	Name string `json:"name"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	id, err := h.svc.Create(r.Context(), req.Name)
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

type assignRequest struct {
	ManagerUserID *int `json:"manager_user_id"`
}

func (h *Handler) AssignManager(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req assignRequest
	if err = httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err = h.svc.AssignManager(r.Context(), id, req.ManagerUserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
