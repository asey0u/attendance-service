package web

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/asey0u/attendance-service/internal/attendance"
	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/department"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/employee"
	"github.com/asey0u/attendance-service/internal/httpx"
	"github.com/asey0u/attendance-service/internal/ticket"
)

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())

	userCount, err := h.authSvc.CountUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	empCount, err := h.empSvc.Count(r.Context(), domain.EmployeeFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	depts, err := h.deptSvc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pendingTks, err := h.tickSvc.CountAll(r.Context(), domain.TicketPending)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.r.Page(w, http.StatusOK, "admin_dashboard", map[string]any{
		"Title":      "Обзор",
		"NavKey":     "dash",
		"Claims":     claims,
		"UserCount":  userCount,
		"EmpCount":   empCount,
		"DeptCount":  len(depts),
		"PendingTks": pendingTks,
	})
}

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	h.renderAdminUsers(w, r, nil)
}

func (h *Handler) AdminEmployeeSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	emps, err := h.empSvc.List(r.Context(), domain.EmployeeFilter{Search: q, Limit: 10})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.r.Partial(w, "employee_search_results", map[string]any{"Rows": emps})
}

func (h *Handler) renderAdminUsers(w http.ResponseWriter, r *http.Request, formErr error) {
	claims := auth.FromContext(r.Context())

	total, err := h.authSvc.CountUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := httpx.NewPager(httpx.ParsePage(r), total, httpx.BaseURLFrom(r))

	users, err := h.authSvc.ListUsers(r.Context(), pager.PageSize, pager.Offset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roles, err := h.authSvc.ListRoles(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var errMsg string
	if formErr != nil {
		errMsg = formErr.Error()
	}

	status := http.StatusOK
	if formErr != nil {
		status = http.StatusUnprocessableEntity
	}

	var formEmp *domain.Employee
	if empIDStr := r.PostFormValue("employee_id"); empIDStr != "" {
		if empID, err2 := strconv.Atoi(empIDStr); err2 == nil {
			formEmp, _ = h.empSvc.Get(r.Context(), empID)
		}
	}

	h.r.Page(w, status, "admin_users", map[string]any{
		"Title":     "Пользователи",
		"NavKey":    "users",
		"Claims":    claims,
		"Users":     users,
		"Roles":     roles,
		"FormError": errMsg,
		"FormLogin": r.PostFormValue("login"),
		"FormRole":  r.PostFormValue("role"),
		"FormEmpID": r.PostFormValue("employee_id"),
		"FormEmp":   formEmp,
		"Pager":     pager,
	})
}

func (h *Handler) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var empPtr *int
	if v := r.PostFormValue("employee_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			empPtr = &n
		}
	}

	if _, err := h.authSvc.CreateUser(r.Context(),
		r.PostFormValue("login"),
		r.PostFormValue("password"),
		r.PostFormValue("role"),
		empPtr,
	); err != nil {
		h.renderAdminUsers(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *Handler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	claims := auth.FromContext(r.Context())
	if id == claims.UserID {
		http.Error(w, "нельзя изменить собственную роль", http.StatusBadRequest)
		return
	}

	if err = h.authSvc.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	claims := auth.FromContext(r.Context())
	if id == claims.UserID {
		http.Error(w, "нельзя изменить собственную роль", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = h.authSvc.UpdateRole(r.Context(), id, r.PostFormValue("role")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *Handler) AdminEmployees(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())

	total, err := h.empSvc.Count(r.Context(), domain.EmployeeFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := httpx.NewPager(httpx.ParsePage(r), total, httpx.BaseURLFrom(r))

	emps, err := h.empSvc.List(r.Context(), domain.EmployeeFilter{Limit: pager.PageSize, Offset: pager.Offset()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	depts, err := h.deptSvc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.r.Page(w, http.StatusOK, "admin_employees", map[string]any{
		"Title":       "Сотрудники",
		"NavKey":      "emps",
		"Claims":      claims,
		"Rows":        emps,
		"Departments": depts,
		"Pager":       pager,
	})
}

func (h *Handler) AdminCreateEmployee(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var deptPtr *int
	if v := r.PostFormValue("department_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			deptPtr = &n
		}
	}

	if _, err := h.empSvc.Create(r.Context(), &domain.Employee{
		FirstName:    r.PostFormValue("first_name"),
		LastName:     r.PostFormValue("last_name"),
		Position:     r.PostFormValue("position"),
		DepartmentID: deptPtr,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/employees", http.StatusSeeOther)
}

func (h *Handler) AdminDeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err = h.empSvc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, employee.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AdminAssignDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var deptPtr *int
	if v := r.PostFormValue("department_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			deptPtr = &n
		}
	}

	if err = h.empSvc.AssignDepartment(r.Context(), id, deptPtr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/employees", http.StatusSeeOther)
}

func (h *Handler) AdminDepartments(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())

	depts, err := h.deptSvc.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	users, err := h.authSvc.ListUsers(r.Context(), 0, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	managersByDept := make(map[int][]map[string]any)
	for _, u := range users {
		if u.Role == domain.RoleManager && u.DepartmentID != nil {
			deptID := *u.DepartmentID
			managersByDept[deptID] = append(managersByDept[deptID], map[string]any{
				"ID":    u.ID,
				"Login": u.Login,
				"Name":  u.FirstName + " " + u.LastName,
			})
		}
	}

	h.r.Page(w, http.StatusOK, "admin_departments", map[string]any{
		"Title":           "Отделы",
		"NavKey":          "depts",
		"Claims":          claims,
		"Rows":            depts,
		"ManagersByDept":  managersByDept,
	})
}

func (h *Handler) AdminCreateDepartment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.deptSvc.Create(r.Context(), r.PostFormValue("name")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/departments", http.StatusSeeOther)
}

func (h *Handler) AdminDeleteDepartment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err = h.deptSvc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, department.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AdminAssignManager(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var mgrPtr *int
	if v := r.PostFormValue("manager_user_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			mgrPtr = &n
		}
	}

	if err = h.deptSvc.AssignManager(r.Context(), id, mgrPtr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/departments", http.StatusSeeOther)
}

func (h *Handler) AdminAttendance(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	q := r.URL.Query()

	f := domain.AttendanceFilter{
		FirstName: q.Get("first_name"),
		LastName:  q.Get("last_name"),
		Position:  q.Get("position"),
		From:      parseDateParam(r, "from", false),
		To:        parseDateParam(r, "to", true),
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
		"Title":     "Посещаемость",
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
		h.r.Partial(w, "admin_attendance_rows", data)
		return
	}

	h.r.Page(w, http.StatusOK, "admin_attendance", data)
}

func (h *Handler) AdminDeleteAttendance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err = h.attSvc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, attendance.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) AdminTickets(w http.ResponseWriter, r *http.Request) {
	claims := auth.FromContext(r.Context())
	status := r.URL.Query().Get("status")

	total, err := h.tickSvc.CountAll(r.Context(), status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := httpx.NewPager(httpx.ParsePage(r), total, httpx.BaseURLFrom(r))

	rows, err := h.tickSvc.ListAll(r.Context(), status, pager.PageSize, pager.Offset())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.r.Page(w, http.StatusOK, "admin_tickets", map[string]any{
		"Title":  "Заявки",
		"NavKey": "tickets",
		"Claims": claims,
		"Rows":   rows,
		"Status": status,
		"Pager":  pager,
	})
}

func (h *Handler) AdminDeleteTicket(w http.ResponseWriter, r *http.Request) {
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
