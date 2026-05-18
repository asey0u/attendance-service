package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/httpx"
	"github.com/asey0u/attendance-service/internal/ticket"
)

func (h *Handler) EmployeeDashboard(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	h.renderDashboard(w, r, claims, "")
}

func (h *Handler) renderDashboard(w http.ResponseWriter, r *http.Request, claims *auth.Claims, errMsg string) {
	var (
		today *domain.Attendance
		stats domain.AttendanceStats
	)

	if claims.EmployeeID != nil {
		var err error
		today, err = h.attSvc.OpenSession(r.Context(), *claims.EmployeeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if today == nil {
			today, _ = h.attSvc.TodaySession(r.Context(), *claims.EmployeeID)
		}
		stats, err = h.attSvc.MyStats(r.Context(), *claims.EmployeeID, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	navKey := "dash"
	if claims.Role == domain.RoleManager {
		navKey = "me"
	}

	h.r.Page(w, http.StatusOK, "employee_dashboard", map[string]any{
		"Title":  "Главная",
		"NavKey": navKey,
		"Claims": claims,
		"Today":  today,
		"Stats":  stats,
		"Error":  errMsg,
	})
}

func (h *Handler) EmployeeAttendance(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())

	var (
		rows  []domain.Attendance
		stats domain.AttendanceStats
		pager httpx.Pager
	)

	if claims.EmployeeID != nil {
		total, err := h.attSvc.CountByEmployee(r.Context(), *claims.EmployeeID, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pager = httpx.NewPager(httpx.ParsePage(r), total, httpx.BaseURLFrom(r))

		rows, err = h.attSvc.MyHistory(r.Context(), *claims.EmployeeID, nil, nil, pager.PageSize, pager.Offset())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stats, err = h.attSvc.MyStats(r.Context(), *claims.EmployeeID, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.r.Page(w, http.StatusOK, "employee_attendance", map[string]any{
		"Title":  "Моя посещаемость",
		"NavKey": "att",
		"Claims": claims,
		"Rows":   rows,
		"Stats":  stats,
		"Pager":  pager,
	})
}

func (h *Handler) EmployeeTickets(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())

	rows, err := h.tickSvc.ListMine(r.Context(), claims, 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.r.Page(w, http.StatusOK, "employee_tickets", map[string]any{
		"Title":  "Мои заявки",
		"NavKey": "tickets",
		"Claims": claims,
		"Rows":   rows,
	})
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims.EmployeeID == nil {
		http.Redirect(w, r, "/me", http.StatusSeeOther)
		return
	}

	if _, err := h.attSvc.CheckIn(r.Context(), *claims.EmployeeID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/me", http.StatusSeeOther)
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	if claims.EmployeeID == nil {
		http.Redirect(w, r, "/me", http.StatusSeeOther)
		return
	}

	if _, err := h.attSvc.CheckOut(r.Context(), *claims.EmployeeID); err != nil {
		h.dashboardWithError(w, r, claims, "Ошибка отметки ухода: "+err.Error())
		return
	}

	http.Redirect(w, r, "/me", http.StatusSeeOther)
}

func (h *Handler) dashboardWithError(w http.ResponseWriter, r *http.Request, claims *auth.Claims, errMsg string) {
	h.renderDashboard(w, r, claims, errMsg)
}

func (h *Handler) EmployeeCreateTicket(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims := auth.FromContext(r.Context())
	_, err := h.tickSvc.Create(r.Context(), claims,
		r.PostFormValue("ticket_date"),
		r.PostFormValue("reason"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/me/tickets", http.StatusSeeOther)
}

func (h *Handler) EmployeeDeleteTicket(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	claims := auth.FromContext(r.Context())
	if err = h.tickSvc.Delete(r.Context(), claims, id); err != nil {
		if errors.Is(err, ticket.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func todayStr() string {
	return nowFunc().Format("2006-01-02")
}
