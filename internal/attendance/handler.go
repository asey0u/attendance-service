package attendance

import (
	"errors"
	"net/http"
	"time"

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

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims.EmployeeID == nil {
		httpx.WriteError(w, httpx.NewError(http.StatusBadRequest, "user has no employee profile"))
		return
	}

	a, err := h.svc.CheckIn(r.Context(), *claims.EmployeeID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, a)
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims.EmployeeID == nil {
		httpx.WriteError(w, httpx.NewError(http.StatusBadRequest, "user has no employee profile"))
		return
	}

	a, err := h.svc.CheckOut(r.Context(), *claims.EmployeeID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, a)
}

func (h *Handler) MyHistory(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims.EmployeeID == nil {
		httpx.JSON(w, http.StatusOK, []domain.Attendance{})
		return
	}

	from, to := parseDateRange(r)
	items, err := h.svc.MyHistory(r.Context(), *claims.EmployeeID, from, to, 0, 0)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) MyStats(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims.EmployeeID == nil {
		httpx.JSON(w, http.StatusOK, domain.AttendanceStats{})
		return
	}

	from, to := parseDateRange(r)
	stats, err := h.svc.MyStats(r.Context(), *claims.EmployeeID, from, to)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, stats)
}

func (h *Handler) ListForManager(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	f := attendanceFilterFromQuery(r)
	if claims.Role == domain.RoleManager {
		f.DepartmentID = claims.DepartmentID
	}

	items, err := h.svc.ListFiltered(r.Context(), f)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListFiltered(r.Context(), attendanceFilterFromQuery(r))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, items)
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

func attendanceFilterFromQuery(r *http.Request) domain.AttendanceFilter {
	q := r.URL.Query()
	from, to := parseDateRange(r)
	return domain.AttendanceFilter{
		FirstName: q.Get("first_name"),
		LastName:  q.Get("last_name"),
		Position:  q.Get("position"),
		From:      from,
		To:        to,
	}
}

func parseDateRange(r *http.Request) (*time.Time, *time.Time) {
	q := r.URL.Query()
	var from, to *time.Time
	if s := q.Get("from"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			from = &t
		}
	}
	if s := q.Get("to"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			to = &t
		}
	}
	return from, to
}
