package main

import (
	"log"
	"net/http"
	"time"

	"github.com/asey0u/attendance-service/internal/appsettings"
	"github.com/asey0u/attendance-service/internal/attendance"
	"github.com/asey0u/attendance-service/internal/auth"
	"github.com/asey0u/attendance-service/internal/config"
	"github.com/asey0u/attendance-service/internal/database"
	"github.com/asey0u/attendance-service/internal/department"
	"github.com/asey0u/attendance-service/internal/employee"
	"github.com/asey0u/attendance-service/internal/server"
	"github.com/asey0u/attendance-service/internal/ticket"
	"github.com/asey0u/attendance-service/internal/web"
)

func init() {
	msk, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		msk = time.FixedZone("MSK", 3*60*60)
	}
	time.Local = msk
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.Open(cfg.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	if err = database.Migrate(db, "migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	signer := auth.NewSigner(cfg.JWTSecret)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, signer)
	authHandler := auth.NewHandler(authSvc)

	empRepo := employee.NewRepository(db)
	empSvc := employee.NewService(empRepo)
	empHandler := employee.NewHandler(empSvc)

	deptRepo := department.NewRepository(db)
	deptSvc := department.NewService(deptRepo)
	deptHandler := department.NewHandler(deptSvc)

	settingsRepo := appsettings.NewRepository(db)
	settingsSvc := appsettings.NewService(settingsRepo)

	attRepo := attendance.NewRepository(db)
	attSvc := attendance.NewService(attRepo, settingsSvc)
	attHandler := attendance.NewHandler(attSvc)

	tickRepo := ticket.NewRepository(db)
	tickSvc := ticket.NewService(tickRepo)
	tickHandler := ticket.NewHandler(tickSvc)

	webHandler, err := web.New(&web.Deps{
		DB:                db,
		Signer:            signer,
		AuthService:       authSvc,
		EmployeeService:   empSvc,
		DepartmentService: deptSvc,
		AttendanceService: attSvc,
		TicketService:     tickSvc,
		SettingsService:   settingsSvc,
	})

	if err != nil {
		log.Fatalf("web: %v", err)
	}

	handler := server.New(&server.Deps{
		Signer:            signer,
		AuthHandler:       authHandler,
		EmployeeHandler:   empHandler,
		DepartmentHandler: deptHandler,
		AttendanceHandler: attHandler,
		TicketHandler:     tickHandler,
		Web:               webHandler,
	})

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("server listening on %s", cfg.Addr())
	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
