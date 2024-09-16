package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/handlers"
	"github.com/xtommas/challenge-hetmo/internal/middleware"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
	"github.com/xtommas/challenge-hetmo/internal/validator"
)

func runMigrations(dbUrl string, logger echo.Logger) {
	m, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		logger.Fatalf("Could not initialize migration: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatalf("Could not apply migrations: %v", err)
	}
	logger.Info("Migrations applied successfully")
}

func createInitialAdminUser(db *sql.DB, logger echo.Logger) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = true").Scan(&count)
	if err != nil {
		logger.Fatal("Failed to check admin user existence:", err)
	}

	if count == 0 {
		adminUsername := os.Getenv("ADMIN_USERNAME")
		adminPassword := os.Getenv("ADMIN_PASSWORD")
		if adminUsername == "" || adminPassword == "" {
			logger.Fatal("Admin credentials not provided in .env file")
		}

		// Create the admin user
		user := &models.User{
			Username: adminUsername,
			IsAdmin:  true,
		}

		// Hash the password
		if err := user.SetPassword(adminPassword); err != nil {
			logger.Fatal("Failed to hash admin password:", err)
		}

		// Insert the admin user into the database
		userRepo := repositories.UserRepository{DB: db}
		if err := userRepo.Create(user); err != nil {
			logger.Fatal("Failed to create initial admin user:", err)
		}

		logger.Info("Admin user created successfully")
	} else {
		logger.Info("Admin user already exists")
	}
}

func main() {
	e := echo.New()

	// Load .env
	err := godotenv.Load()
	if err != nil {
		e.Logger.Fatal("Error loading .env file")
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		"db",
		"5432",
		os.Getenv("POSTGRES_DB"))

	// Run migrations
	runMigrations(dbURL, e.Logger)

	// Connect to the database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	// Create initial admin user if it doesn't exist
	createInitialAdminUser(db, e.Logger)

	// Custom validator
	e.Validator = validator.NewCustomValidator()

	// Initialize repositories
	eventRepo := &repositories.EventRepository{DB: db}
	userRepo := &repositories.UserRepository{DB: db}
	userEventRepo := &repositories.UserEventRepository{DB: db}

	// Public routes
	e.POST("/register", handlers.Register(userRepo))
	e.POST("/login", handlers.Login(userRepo))

	// Authenticated routes
	r := e.Group("/api/v1")
	r.Use(middleware.JWTMiddleware)

	r.GET("/events", handlers.GetAllEvents(eventRepo))
	r.GET("/events/:id", handlers.GetEvent(eventRepo))
	r.POST("/events", middleware.AdminOnly(handlers.CreateEvent(eventRepo)))
	r.DELETE("/events/:id", middleware.AdminOnly(handlers.DeleteEvent(eventRepo)))
	r.PATCH("/events/:id", middleware.AdminOnly(handlers.UpdateEvent(eventRepo)))
	r.POST("/events/:id/signup", handlers.SignUpForEvent(userEventRepo))
	r.GET("/user/events", handlers.GetUserEvents(userEventRepo))

	e.Logger.Fatal(e.Start(":8080"))
}
