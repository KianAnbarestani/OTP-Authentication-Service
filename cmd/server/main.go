// @title OTP Authentication Service API
// @version 1.0
// @description A backend service for OTP-based authentication with user management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"os"
	"time"

	_ "OTP-Authenticate-Service/docs" // generated docs package

	"github.com/gin-gonic/gin"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"OTP-Authenticate-Service/internal/api"
	"OTP-Authenticate-Service/internal/models"
	"OTP-Authenticate-Service/internal/repos"
	"OTP-Authenticate-Service/internal/services"
)

func main() {
	// Load environment variables
	dsn := os.Getenv("DATABASE_DSN")
	jwtSecret := os.Getenv("JWT_SECRET")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPass := os.Getenv("REDIS_PASS")
	otpTTL := 120 * time.Second

	// Initialize PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	// Auto-migrate User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("failed to migrate db: %v", err)
	}

	// Initialize repositories
	userRepo := repos.NewUserRepo(db)

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
	})

	// Initialize services
	otpStore := services.NewRedisOTPStore(rdb) // or services.NewInMemoryOTPStore()
	rateLimiter := services.NewRateLimiter(rdb, 3, 10*time.Minute)
	authSrv := services.NewAuthService(jwtSecret)

	// Initialize handler
	handler := api.NewHandler(otpStore, rateLimiter, userRepo, authSrv, otpTTL)

	// Initialize Gin
	router := gin.Default()

	// Register routes
	api.RegisterRoutes(router, handler, jwtSecret)

	// Run server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
