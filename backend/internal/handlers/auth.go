package handlers

import (
	"net/http"
	"strings"
	"uptime-monitor/internal/auth"
	"uptime-monitor/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type SetupStatus struct {
	NeedsSetup bool `json:"needs_setup"`
	UserCount  int  `json:"user_count"`
}

func SetupAuthRoutes(router *gin.RouterGroup, db *sqlx.DB, authService *auth.Service) {
	// Public routes (no auth required)
	router.GET("/auth/setup-status", getSetupStatus(db))
	router.POST("/auth/setup", setupAdmin(db, authService))
	router.POST("/auth/register", register(db, authService))
	router.POST("/auth/login", login(db, authService))
	router.POST("/auth/logout", logout())

	// Protected routes
	router.GET("/auth/profile", authRequired(authService), getProfile(db))
	router.GET("/auth/users", authRequired(authService), adminRequired(), getUsers(db))
	router.PUT("/auth/users/:id", authRequired(authService), adminRequired(), updateUser(db, authService))
	router.DELETE("/auth/users/:id", authRequired(authService), adminRequired(), deleteUser(db))
}

// getSetupStatus checks if the system needs initial setup (no users exist)
func getSetupStatus(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, SetupStatus{
			NeedsSetup: count == 0,
			UserCount:  count,
		})
	}
}

// setupAdmin creates the first admin user during initial setup
func setupAdmin(db *sqlx.DB, authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if setup is still needed
		var count int
		if err := db.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Setup already completed. Users already exist."})
			return
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Validate password strength
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
			return
		}

		// Hash password
		hashedPassword, err := authService.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Create admin user
		var userID int64
		err = db.QueryRow(`
			INSERT INTO users (username, email, password, role, active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id
		`, req.Username, req.Email, hashedPassword, "admin", true).Scan(&userID)

		if err != nil {
			// Fallback for SQLite
			result, err := db.Exec(`
				INSERT INTO users (username, email, password, role, active)
				VALUES (?, ?, ?, ?, ?)
			`, req.Username, req.Email, hashedPassword, "admin", true)

			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
				}
				return
			}
			userID, _ = result.LastInsertId()
		}

		// Fetch the created user
		var user models.User
		db.Get(&user, "SELECT * FROM users WHERE id = ?", userID)

		// Generate token for immediate login
		token, err := authService.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User created but failed to generate token"})
			return
		}

		user.Password = ""
		c.JSON(http.StatusCreated, LoginResponse{
			Token: token,
			User:  user,
		})
	}
}

// register creates a new user account
func register(db *sqlx.DB, authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if at least one user exists (setup must be completed first)
		var count int
		if err := db.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please complete setup first by creating an admin account"})
			return
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		// Validate password strength
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
			return
		}

		// Hash password
		hashedPassword, err := authService.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Create regular user
		var userID int64
		err = db.QueryRow(`
			INSERT INTO users (username, email, password, role, active)
			VALUES ($1, $2, $3, $4, $5) RETURNING id
		`, req.Username, req.Email, hashedPassword, "user", true).Scan(&userID)

		if err != nil {
			// Fallback for SQLite
			result, err := db.Exec(`
				INSERT INTO users (username, email, password, role, active)
				VALUES (?, ?, ?, ?, ?)
			`, req.Username, req.Email, hashedPassword, "user", true)

			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user: " + err.Error()})
				}
				return
			}
			userID, _ = result.LastInsertId()
		}

		// Fetch the created user
		var user models.User
		db.Get(&user, "SELECT * FROM users WHERE id = ?", userID)

		// Generate token for immediate login
		token, err := authService.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User created but failed to generate token"})
			return
		}

		user.Password = ""
		c.JSON(http.StatusCreated, LoginResponse{
			Token: token,
			User:  user,
		})
	}
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

func getProfile(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			return
		}

		var user models.User
		err := db.Get(&user, "SELECT id, username, email, role, active, created_at, updated_at FROM users WHERE id = ?", userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// getUsers returns all users (admin only)
func getUsers(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		err := db.Select(&users, "SELECT id, username, email, role, active, created_at, updated_at FROM users ORDER BY created_at DESC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

// updateUser updates a user's details (admin only)
func updateUser(db *sqlx.DB, authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Active   *bool  `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Build update query dynamically
		updates := []string{}
		args := []interface{}{}

		if req.Username != "" {
			updates = append(updates, "username = ?")
			args = append(args, req.Username)
		}
		if req.Email != "" {
			updates = append(updates, "email = ?")
			args = append(args, req.Email)
		}
		if req.Password != "" {
			if len(req.Password) < 6 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
				return
			}
			hashedPassword, err := authService.HashPassword(req.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}
			updates = append(updates, "password = ?")
			args = append(args, hashedPassword)
		}
		if req.Role != "" {
			if req.Role != "admin" && req.Role != "user" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'admin' or 'user'"})
				return
			}
			updates = append(updates, "role = ?")
			args = append(args, req.Role)
		}
		if req.Active != nil {
			updates = append(updates, "active = ?")
			args = append(args, *req.Active)
		}

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, userID)

		query := "UPDATE users SET " + strings.Join(updates, ", ") + " WHERE id = ?"
		_, err := db.Exec(query, args...)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

// deleteUser removes a user (admin only)
func deleteUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		currentUserID, _ := c.Get("user_id")

		// Prevent self-deletion
		if userID == currentUserID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
			return
		}

		// Check if this is the last admin
		var adminCount int
		db.Get(&adminCount, "SELECT COUNT(*) FROM users WHERE role = 'admin'")

		var targetUser models.User
		if err := db.Get(&targetUser, "SELECT role FROM users WHERE id = ?", userID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		if targetUser.Role == "admin" && adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last admin user"})
			return
		}

		_, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
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

// Middleware to require admin role
func adminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
