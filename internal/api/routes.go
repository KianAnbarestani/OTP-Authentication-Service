package api

import (
	"time"

	"OTP-Authenticate-Service/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine, h *Handler, jwtSecret string) {
	// Public auth endpoints
	r.POST("/auth/request-otp", h.RequestOTP)
	r.POST("/auth/verify-otp", h.VerifyOTP)

	// JWT-protected endpoints
	auth := r.Group("/", middleware.JWTMiddleware([]byte(jwtSecret)))
	{
		auth.GET("/users/:id", h.GetUserByID)
		auth.GET("/users", h.ListUsers)
	}

	// Health check
	r.GET("/health", healthCheck)

	// Swagger docs route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// healthCheck returns the health status of the service
// @Summary Health check
// @Description Check if the service is running
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Service is healthy"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "time": time.Now()})
}
