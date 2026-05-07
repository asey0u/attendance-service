package main

import (
	"log"
	"net/http"

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

	attRepo := attendance.NewRepository(database)
	attService := attendance.NewService(attRepo)
	attHandler := attendance.NewHandler(attService)

	http.HandleFunc("/login", authHandler.Login)
	http.HandleFunc("/register", authHandler.Register)

	http.Handle("/attendance/check-in",
		middleware.AuthMiddleware(http.HandlerFunc(attHandler.CheckIn)),
	)

	http.Handle("/attendance/me",
		middleware.AuthMiddleware(http.HandlerFunc(attHandler.MyAttendance)),
	)

	log.Println("server started :8080")
	http.ListenAndServe(":8080", nil)
}
