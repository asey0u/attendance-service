package ticket

import (
	"net/http"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	TicketDate string `json:"ticket_date"`
	Reason     string `json:"reason"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	claims := auth.FromContext(r.Context())
	id, err := h.svc.Create(r.Context(), claims, req.TicketDate, req.Reason)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListMine(r.Context(), auth.FromContext(r.Context()), 0, 0)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) ListForManager(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	items, err := h.svc.ListForManager(r.Context(), auth.FromContext(r.Context()), status, 0, 0)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	items, err := h.svc.ListAll(r.Context(), status, 0, 0)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, true)
}

func (h *Handler) Decline(w http.ResponseWriter, r *http.Request) {
	h.review(w, r, false)
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request, approve bool) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	claims := auth.FromContext(r.Context())
	if approve {
		err = h.svc.Approve(r.Context(), claims, id)
	} else {
		err = h.svc.Decline(r.Context(), claims, id)
	}
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	status := "approved"
	if !approve {
		status = "declined"
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParseIntParam(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err = h.svc.Delete(r.Context(), auth.FromContext(r.Context()), id); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
