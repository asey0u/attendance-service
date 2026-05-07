package main

import (
	"log"
	"net/http"
	"os"

	"github.com/asey0u/attendance-service/internal/admin"
	"github.com/asey0u/attendance-service/internal/attendance"
	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/db"
	"github.com/asey0u/attendance-service/internal/middleware"
)

func main() {
	database := db.Init()

	authRepo := auth.NewRepository(database)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	adminRepo := admin.NewRepository(database)
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)

	attRepo := attendance.NewRepository(database)
	attService := attendance.NewService(attRepo)
	attHandler := attendance.NewHandler(attService)

	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/setup/status", adminHandler.SetupStatus)
	http.HandleFunc("/setup/admin", adminHandler.CreateAdmin)

	http.Handle("/attendance/check-in",
		middleware.AuthMiddleware(http.HandlerFunc(attHandler.CheckIn)),
	)

	http.Handle("/attendance/me",
		middleware.AuthMiddleware(http.HandlerFunc(attHandler.MyAttendance)),
	)

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/users", adminHandler.ListUsers)
	adminMux.HandleFunc("/users/create", adminHandler.CreateUser)
	adminMux.HandleFunc("/users/delete", adminHandler.DeleteUser)
	adminMux.HandleFunc("/users/update-role", adminHandler.UpdateUserRole)
	adminMux.HandleFunc("/users/", adminHandler.GetUserByID)
	adminMux.HandleFunc("/employees", adminHandler.ListEmployees)
	adminMux.HandleFunc("/employees/create", adminHandler.CreateEmployee)
	adminMux.HandleFunc("/attendances", adminHandler.GetAttendancesByEmployee)
	adminMux.HandleFunc("/attendances/create", adminHandler.CreateAttendance)
	adminMux.HandleFunc("/attendances/delete", adminHandler.DeleteAttendance)
	adminMux.HandleFunc("/me", adminHandler.GetCurrentUser)

	http.Handle("/admin/", http.StripPrefix("/admin", middleware.AuthMiddleware(middleware.AdminMiddleware(adminMux))))

	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("APP_PORT")

	log.Println("server started :", appPort)
	http.ListenAndServe(appHost+":"+appPort, nil)
}
