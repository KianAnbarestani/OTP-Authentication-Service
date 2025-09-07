package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"OTP-Authenticate-Service/internal/repos"
	"OTP-Authenticate-Service/internal/services"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	otpStore    services.OTPStore
	rateLimiter *services.RateLimiter
	userRepo    *repos.UserRepo
	authSrv     *services.AuthService
	otpTTL      time.Duration
}

func NewHandler(otp services.OTPStore, rl *services.RateLimiter, ur *repos.UserRepo, as *services.AuthService, otpTTL time.Duration) *Handler {
	return &Handler{otpStore: otp, rateLimiter: rl, userRepo: ur, authSrv: as, otpTTL: otpTTL}
}

type reqPhone struct {
	Phone string `json:"phone" binding:"required,e164" example:"+14165551234"`
}

// User represents a user in the system
type User struct {
	ID           uint      `json:"id" example:"1"`
	Phone        string    `json:"phone" example:"+14165551234"`
	RegisteredAt time.Time `json:"registered_at" example:"2025-01-07T12:34:56Z"`
}

// RequestOTP handles OTP request
// @Summary Request OTP
// @Description Generate and send OTP to the provided phone number
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body reqPhone true "Phone number in E.164 format"
// @Success 200 {object} map[string]string "OTP generated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 429 {object} map[string]interface{} "Rate limit exceeded"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/request-otp [post]
func (h *Handler) RequestOTP(c *gin.Context) {
	var r reqPhone
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()

	allowed, rem, err := h.rateLimiter.Allow(ctx, r.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded", "remaining": rem})
		return
	}

	otp, _ := services.GenerateOTP(6)
	_ = h.otpStore.Set(ctx, r.Phone, otp, h.otpTTL)

	fmt.Printf("OTP for %s → %s\n", maskPhone(r.Phone), otp)
	c.JSON(http.StatusOK, gin.H{"message": "OTP generated. Check console logs."})
}

type verifyReq struct {
	Phone string `json:"phone" binding:"required" example:"+14165551234"`
	OTP   string `json:"otp" binding:"required" example:"123456"`
}

// VerifyOTP verifies OTP and registers/logs in user
// @Summary Verify OTP
// @Description Verify OTP and authenticate user, returns JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body verifyReq true "Phone number and OTP"
// @Success 200 {object} map[string]string "Authentication successful"
// @Failure 400 {object} map[string]string "Invalid request or OTP"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/verify-otp [post]
func (h *Handler) VerifyOTP(c *gin.Context) {
	var r verifyReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := c.Request.Context()

	stored, err := h.otpStore.Get(ctx, r.Phone)
	if err != nil || stored != r.OTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired otp"})
		return
	}

	_ = h.otpStore.Delete(ctx, r.Phone)

	user, err := h.userRepo.CreateIfNotExist(ctx, r.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authSrv.GenerateToken(user.ID, user.Phone, 30*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetUserByID returns a user by ID
// @Summary Get user by ID
// @Description Retrieve a specific user by their ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User "User found"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	id := uint(id64)

	user, err := h.userRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers returns paginated list of users
// @Summary List users
// @Description Get paginated list of users with optional search
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by phone number"
// @Success 200 {object} map[string]interface{} "Users list with pagination"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	users, total, err := h.userRepo.List(c.Request.Context(), page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func maskPhone(p string) string {
	if len(p) <= 4 {
		return "****"
	}
	n := len(p)
	return p[:2] + "****" + p[n-2:]
}
