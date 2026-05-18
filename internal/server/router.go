package server

import (
	"net/http"

	"github.com/asey0u/attendance-service/internal/attendance"
	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/department"
	"github.com/asey0u/attendance-service/internal/domain"
	"github.com/asey0u/attendance-service/internal/employee"
	"github.com/asey0u/attendance-service/internal/middleware"
	"github.com/asey0u/attendance-service/internal/ticket"
)

type WebMounter interface {
	Mount(mux *http.ServeMux)
}

type Deps struct {
	Signer            *auth.Signer
	AuthHandler       *auth.Handler
	EmployeeHandler   *employee.Handler
	DepartmentHandler *department.Handler
	AttendanceHandler *attendance.Handler
	TicketHandler     *ticket.Handler
	Web               WebMounter
}

func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func New(d *Deps) http.Handler {
	mux := http.NewServeMux()

	authMW := middleware.Auth(d.Signer)
	requireAdmin := middleware.RequireAnyRole(domain.RoleAdmin)
	requireManager := middleware.RequireAnyRole(domain.RoleAdmin, domain.RoleManager)

	authed := func(h http.HandlerFunc) http.Handler {
		return chain(h, authMW)
	}
	managerOrAdmin := func(h http.HandlerFunc) http.Handler {
		return chain(h, authMW, requireManager)
	}
	adminOnly := func(h http.HandlerFunc) http.Handler {
		return chain(h, authMW, requireAdmin)
	}

	mux.HandleFunc("POST /api/login", d.AuthHandler.Login)
	mux.HandleFunc("POST /api/logout", d.AuthHandler.Logout)
	mux.HandleFunc("POST /api/setup/admin", d.AuthHandler.SetupAdmin)

	mux.Handle("GET /api/me", authed(d.AuthHandler.Me))
	mux.Handle("GET /api/departments", authed(d.DepartmentHandler.List))

	mux.Handle("POST /api/attendance/check-in", authed(d.AttendanceHandler.CheckIn))
	mux.Handle("POST /api/attendance/check-out", authed(d.AttendanceHandler.CheckOut))
	mux.Handle("GET /api/attendance/me", authed(d.AttendanceHandler.MyHistory))
	mux.Handle("GET /api/attendance/me/stats", authed(d.AttendanceHandler.MyStats))

	mux.Handle("POST /api/tickets", authed(d.TicketHandler.Create))
	mux.Handle("GET /api/tickets/me", authed(d.TicketHandler.ListMine))
	mux.Handle("DELETE /api/tickets/{id}", authed(d.TicketHandler.Delete))

	mux.Handle("GET /api/manager/employees", managerOrAdmin(d.EmployeeHandler.ListForManager))
	mux.Handle("GET /api/manager/attendance", managerOrAdmin(d.AttendanceHandler.ListForManager))
	mux.Handle("GET /api/manager/tickets", managerOrAdmin(d.TicketHandler.ListForManager))
	mux.Handle("POST /api/manager/tickets/{id}/approve", managerOrAdmin(d.TicketHandler.Approve))
	mux.Handle("POST /api/manager/tickets/{id}/decline", managerOrAdmin(d.TicketHandler.Decline))

	mux.Handle("GET /api/admin/users", adminOnly(d.AuthHandler.AdminListUsers))
	mux.Handle("POST /api/admin/users", adminOnly(d.AuthHandler.AdminCreateUser))
	mux.Handle("DELETE /api/admin/users/{id}", adminOnly(d.AuthHandler.AdminDeleteUser))
	mux.Handle("PATCH /api/admin/users/{id}/role", adminOnly(d.AuthHandler.AdminUpdateRole))

	mux.Handle("GET /api/admin/employees", adminOnly(d.EmployeeHandler.ListAll))
	mux.Handle("POST /api/admin/employees", adminOnly(d.EmployeeHandler.Create))
	mux.Handle("DELETE /api/admin/employees/{id}", adminOnly(d.EmployeeHandler.Delete))
	mux.Handle("PATCH /api/admin/employees/{id}/department", adminOnly(d.EmployeeHandler.AssignDepartment))

	mux.Handle("GET /api/admin/departments", adminOnly(d.DepartmentHandler.List))
	mux.Handle("POST /api/admin/departments", adminOnly(d.DepartmentHandler.Create))
	mux.Handle("DELETE /api/admin/departments/{id}", adminOnly(d.DepartmentHandler.Delete))
	mux.Handle("PATCH /api/admin/departments/{id}/manager", adminOnly(d.DepartmentHandler.AssignManager))

	mux.Handle("GET /api/admin/attendance", adminOnly(d.AttendanceHandler.ListAll))
	mux.Handle("DELETE /api/admin/attendance/{id}", adminOnly(d.AttendanceHandler.Delete))

	mux.Handle("GET /api/admin/tickets", adminOnly(d.TicketHandler.ListAll))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if d.Web != nil {
		d.Web.Mount(mux)
	}

	return chain(mux, middleware.RequestLog, middleware.Recover)
}
