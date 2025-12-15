package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"hr-portal-backend/internal/attendance"
	"hr-portal-backend/internal/auth"
	"hr-portal-backend/internal/db"
	"hr-portal-backend/internal/messaging"
	"hr-portal-backend/internal/rbac"
	"hr-portal-backend/internal/requests"
	"hr-portal-backend/internal/user"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:pass@localhost:5432/hr_portal?sslmode=disable"
	}

	sqlDB, err := db.Open(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "changeme-secret"
	}
	jwtMgr := auth.NewJWTManager(jwtSecret)

	// wiring user repo + service + handler
	userRepo := user.NewRepository(sqlDB)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	// auth handler (pakai service yg sama)
	authHandler := auth.NewHandler(userSvc, jwtMgr)

	// RBAC handler for menus and permissions
	rbacRepo := rbac.NewRepository(sqlDB)
	rbacHandler := rbac.NewHandler(rbacRepo)

	// Messaging handler
	messagingRepo := messaging.NewRepository(sqlDB)
	messagingHandler := messaging.NewHandler(messagingRepo, userRepo)

	// Requests handler
	requestsRepo := requests.NewRepository(sqlDB)
	requestsSvc := requests.NewService(requestsRepo)
	requestsHandler := requests.NewHandler(requestsSvc)

	// Attendance handler
	attRepo := attendance.NewRepository(sqlDB)
	attSvc := attendance.NewService(attRepo)
	attHandler := attendance.NewHandler(attSvc)

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://127.0.0.1:5173,http://localhost:4173,http://127.0.0.1:4173,http://localhost:5174,http://127.0.0.1:5174",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	api := app.Group("/api")

	// ========= Public routes (tanpa JWT) =========
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/logout", authHandler.Logout)

	// ========= Protected routes (wajib JWT) =========
	protected := api.Group("/", auth.JWTMiddleware(jwtMgr))

	// employees CRUD
	protected.Get("/employees", userHandler.ListEmployees)
	protected.Get("/employees/next-code", userHandler.GetNextEmployeeCode)
	protected.Post("/employees", userHandler.CreateEmployee)
	protected.Put("/employees/:id", userHandler.UpdateEmployee)
	protected.Delete("/employees/:id", userHandler.DeleteEmployee)
	protected.Delete("/employees/by-code/:code", userHandler.DeleteEmployeeByCode)
	// Hard delete endpoints
	protected.Delete("/employees/:id/hard", userHandler.HardDeleteEmployee)
	protected.Delete("/employees/by-code/:code/hard", userHandler.HardDeleteEmployeeByCode)

	// RBAC: menus and permissions
	protected.Get("/me/menus", rbacHandler.GetMyMenus)
	protected.Get("/me/permissions", rbacHandler.GetMyPermissions)

	// User profile
	protected.Get("/me", userHandler.GetMyProfile)
	protected.Put("/me", userHandler.UpdateMyProfile)

	// Messaging & Announcements
	protected.Get("/inbox", messagingHandler.GetInbox)
	protected.Post("/inbox", messagingHandler.SendMessage)
	protected.Put("/inbox/:id/read", messagingHandler.MarkMessageRead)
	protected.Delete("/inbox/:id", messagingHandler.DeleteMessage)
	protected.Get("/announcements", messagingHandler.GetAnnouncements)
	protected.Post("/announcements/:id/read", messagingHandler.MarkAnnouncementRead)
	protected.Delete("/announcements/:id", messagingHandler.DeleteAnnouncement)

	// Requests (Leave, Overtime)
	protected.Post("/requests", requestsHandler.CreateRequest)
	protected.Get("/requests/my", requestsHandler.GetMyRequests)
	protected.Get("/requests/approvals", requestsHandler.GetPendingRequests)
	protected.Post("/requests/:id/approve", requestsHandler.ApproveRequest)
	protected.Post("/requests/:id/reject", requestsHandler.RejectRequest)
	protected.Get("/requests/summary", requestsHandler.GetSummary)
	protected.Get("/requests/summary/my", requestsHandler.GetMySummary)
	protected.Get("/requests/processed", requestsHandler.GetProcessedByMonth)
	protected.Get("/requests/processed/export", requestsHandler.ExportProcessedByMonth)
	// Attendance
	protected.Post("/attendance/checkin", attHandler.Checkin)
	protected.Post("/attendance/checkout", attHandler.Checkout)
	protected.Get("/attendance/summary", attHandler.GetSummary)
	protected.Get("/attendance/list", attHandler.GetList)

	log.Println("Listening on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
