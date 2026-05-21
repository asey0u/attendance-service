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

func (h *Handler) ManagerDashboard(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())

	pending, err := h.tickSvc.ListForManager(r.Context(), claims, domain.TicketPending, 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var deptID *int
	if claims.Role == domain.RoleManager {
		deptID = claims.DepartmentID
	}
	presentToday, err := h.attSvc.CountPresentToday(r.Context(), deptID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	empFilter := domain.EmployeeFilter{}
	if claims.Role == domain.RoleManager {
		empFilter.DepartmentID = claims.DepartmentID
	}
	emps, err := h.empSvc.List(r.Context(), empFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var myToday *domain.Attendance
	if claims.EmployeeID != nil {
		myToday, _ = h.attSvc.OpenSession(r.Context(), *claims.EmployeeID)
		if myToday == nil {
			myToday, _ = h.attSvc.TodaySession(r.Context(), *claims.EmployeeID)
		}
	}

	h.r.Page(w, http.StatusOK, "manager_dashboard", map[string]any{
		"Title":          "Обзор руководителя",
		"NavKey":         "dash",
		"Claims":         claims,
		"PendingCount":   len(pending),
		"PresentToday":   presentToday,
		"TeamSize":       len(emps),
		"PendingTickets": pending,
		"MyToday":        myToday,
	})
}

func (h *Handler) ManagerAttendance(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	q := r.URL.Query()

	f := domain.AttendanceFilter{
		FirstName: q.Get("first_name"),
		LastName:  q.Get("last_name"),
		Position:  q.Get("position"),
		From:      parseDateParam(r, "from", false),
		To:        parseDateParam(r, "to", true),
	}
	if claims.Role == domain.RoleManager {
		f.DepartmentID = claims.DepartmentID
	}

	total, err := h.attSvc.CountFiltered(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := httpx.NewPager(httpx.ParsePage(r), total, httpx.BaseURLFrom(r))
	f.Limit = pager.PageSize
	f.Offset = pager.Offset()

	rows, err := h.attSvc.ListFiltered(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":     "Посещаемость команды",
		"NavKey":    "att",
		"Claims":    claims,
		"Rows":      rows,
		"Pager":     pager,
		"FirstName": f.FirstName,
		"LastName":  f.LastName,
		"Position":  f.Position,
		"From":      q.Get("from"),
		"To":        q.Get("to"),
		"HXRequest": r.Header.Get("HX-Request") == "true",
	}

	if r.Header.Get("HX-Request") == "true" {
		h.r.Partial(w, "attendance_rows", data)
		return
	}

	h.r.Page(w, http.StatusOK, "manager_attendance", data)
}

func (h *Handler) ManagerTickets(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	status := r.URL.Query().Get("status")
	if status == "" {
		status = domain.TicketPending
	}

	total, err := h.tickSvc.CountForManager(r.Context(), claims, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := httpx.NewPager(httpx.ParsePage(r), total, httpx.BaseURLFrom(r))

	rows, err := h.tickSvc.ListForManager(r.Context(), claims, status, pager.PageSize, pager.Offset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.r.Page(w, http.StatusOK, "manager_tickets", map[string]any{
		"Title":  "Заявки команды",
		"NavKey": "tickets",
		"Claims": claims,
		"Rows":   rows,
		"Status": status,
		"Pager":  pager,
	})
}

func (h *Handler) ManagerEmployees(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	f := domain.EmployeeFilter{}
	if claims.Role == domain.RoleManager {
		f.DepartmentID = claims.DepartmentID
	}

	rows, err := h.empSvc.List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.r.Page(w, http.StatusOK, "manager_employees", map[string]any{
		"Title":  "Команда",
		"NavKey": "team",
		"Claims": claims,
		"Rows":   rows,
	})
}

func (h *Handler) ManagerApproveTicket(w http.ResponseWriter, r *http.Request) {
	h.managerReviewTicket(w, r, true)
}

func (h *Handler) ManagerDeclineTicket(w http.ResponseWriter, r *http.Request) {
	h.managerReviewTicket(w, r, false)
}

func (h *Handler) managerReviewTicket(w http.ResponseWriter, r *http.Request, approve bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	claims := auth.FromContext(r.Context())
	if approve {
		err = h.tickSvc.Approve(r.Context(), claims, id)
	} else {
		err = h.tickSvc.Decline(r.Context(), claims, id)
	}
	if err != nil {
		if errors.Is(err, ticket.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
