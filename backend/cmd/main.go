package main

import (
	"log"
	"strings"
	"uptime-monitor/internal/auth"
	"uptime-monitor/internal/config"
	"uptime-monitor/internal/database"
	"uptime-monitor/internal/handlers"
	"uptime-monitor/internal/monitoring"
	"uptime-monitor/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize authentication service
	authService := auth.NewService(cfg.Auth.JWTSecret)

	// Initialize monitoring system
	monitorManager := monitoring.NewManager(db)
	go monitorManager.Start()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Setup router
	router := gin.Default()

	// CORS and headers middleware for proxy support
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Handle CORS for development
		if origin := c.GetHeader("Origin"); origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api/v1")
	handlers.SetupRoutes(api, db, monitorManager, wsHub, authService)

	// Debug endpoint to test backend connectivity
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Backend is running"})
	})

	// Custom handler for _app directory with explicit MIME types
	router.GET("/_app/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		fullPath := "./frontend/dist/_app" + filepath

		// Log the request for debugging
		log.Printf("Serving _app file: %s -> %s", filepath, fullPath)

		// Set MIME type based on file extension
		if strings.HasSuffix(filepath, ".js") {
			c.Header("Content-Type", "application/javascript; charset=utf-8")
			c.Header("X-Content-Type-Options", "nosniff")
		} else if strings.HasSuffix(filepath, ".css") {
			c.Header("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(filepath, ".json") {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}

		c.File(fullPath)
	})

	// Static files for other assets
	router.Static("/assets", "./frontend/dist/assets")

	// Serve favicon and other root assets
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// Serve index.html for root and handle SPA routing
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./frontend/dist/index.html")
	})

	// Fallback for SPA routing - serve index.html for any non-API routes
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Don't serve index.html for API routes
		if strings.HasPrefix(path, "/api/") {
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}

		// Don't serve index.html for static asset paths
		if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/_app/") {
			c.JSON(404, gin.H{"error": "Static file not found"})
			return
		}

		// Serve index.html for SPA routes
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./frontend/dist/index.html")
	})

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
