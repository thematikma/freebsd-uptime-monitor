package handlers

import (
	"net/http"
	"uptime-monitor/internal/auth"
	"uptime-monitor/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func SetupAuthRoutes(router *gin.RouterGroup, db *sqlx.DB, authService *auth.Service) {
	// Authentication routes
	router.POST("/auth/login", login(db, authService))
	router.POST("/auth/logout", logout())
	router.GET("/auth/profile", authRequired(authService), getProfile())

	// Initialize default admin user if it doesn't exist
	initDefaultUser(db, authService)
}

func login(db *sqlx.DB, authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Find user by username
		var user models.User
		err := db.Get(&user, "SELECT * FROM users WHERE username = ? AND active = ?", req.Username, true)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Verify password
		if !authService.CheckPassword(req.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate JWT token
		token, err := authService.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Return user info without password
		user.Password = ""
		response := LoginResponse{
			Token: token,
			User:  user,
		}

		c.JSON(http.StatusOK, response)
	}
}

func logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For JWT tokens, logout is handled client-side by removing the token
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func getProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by authRequired middleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// Middleware to require authentication
func authRequired(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", (*claims)["user_id"])
		c.Set("username", (*claims)["username"])
		c.Set("role", (*claims)["role"])

		c.Next()
	}
}

// Initialize default admin user if no users exist
func initDefaultUser(db *sqlx.DB, authService *auth.Service) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users")
	if err != nil || count > 0 {
		return // Users already exist or error occurred
	}

	// Hash the default password
	hashedPassword, err := authService.HashPassword("password")
	if err != nil {
		return // Failed to hash password
	}

	// Create default admin user
	_, err = db.Exec(`
		INSERT INTO users (username, email, password, role, active) 
		VALUES (?, ?, ?, ?, ?)
	`, "admin", "admin@example.com", hashedPassword, "admin", true)

	if err == nil {
		// Log success but don't fail if we can't create user
		println("Created default admin user: admin/password")
	}
}
