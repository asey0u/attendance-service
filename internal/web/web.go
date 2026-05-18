package web

import (
	"database/sql"
	"net/http"

	"github.com/asey0u/attendance-service/internal/appsettings"
	"github.com/asey0u/attendance-service/internal/attendance"
	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/department"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/employee"
	"github.com/asey0u/attendance-service/internal/ticket"
)

type Handler struct {
	db          *sql.DB
	r           *Renderer
	signer      *auth.Signer
	authSvc     *auth.Service
	empSvc      *employee.Service
	deptSvc     *department.Service
	attSvc      *attendance.Service
	tickSvc     *ticket.Service
	settingsSvc *appsettings.Service
}

type Deps struct {
	DB                *sql.DB
	Signer            *auth.Signer
	AuthService       *auth.Service
	EmployeeService   *employee.Service
	DepartmentService *department.Service
	AttendanceService *attendance.Service
	TicketService     *ticket.Service
	SettingsService   *appsettings.Service
}

func New(d *Deps) (*Handler, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, err
	}
	return &Handler{
		db:          d.DB,
		r:           renderer,
		signer:      d.Signer,
		authSvc:     d.AuthService,
		empSvc:      d.EmployeeService,
		deptSvc:     d.DepartmentService,
		attSvc:      d.AttendanceService,
		tickSvc:     d.TicketService,
		settingsSvc: d.SettingsService,
	}, nil
}

func (h *Handler) Mount(mux *http.ServeMux) {
	auth := func(fn http.HandlerFunc) http.Handler {
		return requireAuth(h.signer, h.roleDB(fn))
	}
	role := func(fn http.HandlerFunc, roles ...string) http.Handler {
		return requireAuth(h.signer, h.roleDB(h.requireRole(fn, roles...)))
	}

	mux.HandleFunc("GET /setup", h.SetupPage)
	mux.HandleFunc("POST /setup", h.SetupSubmit)
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.LoginSubmit)
	mux.HandleFunc("POST /logout", h.LogoutSubmit)
	mux.HandleFunc("GET /", h.Root)

	mux.Handle("GET /me", auth(h.EmployeeDashboard))
	mux.Handle("GET /me/attendance", auth(h.EmployeeAttendance))
	mux.Handle("GET /me/tickets", auth(h.EmployeeTickets))
	mux.Handle("POST /me/check-in", auth(h.CheckIn))
	mux.Handle("POST /me/check-out", auth(h.CheckOut))
	mux.Handle("POST /me/tickets", auth(h.EmployeeCreateTicket))
	mux.Handle("DELETE /me/tickets/{id}", auth(h.EmployeeDeleteTicket))

	mgrRoles := []string{domain.RoleManager, domain.RoleAdmin}
	mux.Handle("GET /manager", role(h.ManagerDashboard, mgrRoles...))
	mux.Handle("GET /manager/attendance", role(h.ManagerAttendance, mgrRoles...))
	mux.Handle("GET /manager/tickets", role(h.ManagerTickets, mgrRoles...))
	mux.Handle("GET /manager/employees", role(h.ManagerEmployees, mgrRoles...))
	mux.Handle("POST /manager/tickets/{id}/approve", role(h.ManagerApproveTicket, mgrRoles...))
	mux.Handle("POST /manager/tickets/{id}/decline", role(h.ManagerDeclineTicket, mgrRoles...))
	mux.Handle("GET /manager/settings", role(h.SettingsPage, mgrRoles...))
	mux.Handle("POST /manager/settings", role(h.SettingsSave, mgrRoles...))

	mux.Handle("GET /admin", role(h.AdminDashboard, domain.RoleAdmin))
	mux.Handle("GET /admin/users", role(h.AdminUsers, domain.RoleAdmin))
	mux.Handle("GET /admin/users/employee-search", role(h.AdminEmployeeSearch, domain.RoleAdmin))
	mux.Handle("POST /admin/users", role(h.AdminCreateUser, domain.RoleAdmin))
	mux.Handle("DELETE /admin/users/{id}", role(h.AdminDeleteUser, domain.RoleAdmin))
	mux.Handle("POST /admin/users/{id}/role", role(h.AdminUpdateRole, domain.RoleAdmin))

	mux.Handle("GET /admin/employees", role(h.AdminEmployees, domain.RoleAdmin))
	mux.Handle("POST /admin/employees", role(h.AdminCreateEmployee, domain.RoleAdmin))
	mux.Handle("DELETE /admin/employees/{id}", role(h.AdminDeleteEmployee, domain.RoleAdmin))
	mux.Handle("POST /admin/employees/{id}/department", role(h.AdminAssignDepartment, domain.RoleAdmin))

	mux.Handle("GET /admin/departments", role(h.AdminDepartments, domain.RoleAdmin))
	mux.Handle("POST /admin/departments", role(h.AdminCreateDepartment, domain.RoleAdmin))
	mux.Handle("DELETE /admin/departments/{id}", role(h.AdminDeleteDepartment, domain.RoleAdmin))
	mux.Handle("POST /admin/departments/{id}/manager", role(h.AdminAssignManager, domain.RoleAdmin))

	mux.Handle("GET /admin/attendance", role(h.AdminAttendance, domain.RoleAdmin))
	mux.Handle("DELETE /admin/attendance/{id}", role(h.AdminDeleteAttendance, domain.RoleAdmin))

	mux.Handle("GET /admin/tickets", role(h.AdminTickets, domain.RoleAdmin))
	mux.Handle("DELETE /admin/tickets/{id}", role(h.AdminDeleteTicket, domain.RoleAdmin))
}
