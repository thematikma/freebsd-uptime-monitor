package main

import (
	"log"
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

	// Initialize monitoring system
	monitorManager := monitoring.NewManager(db)
	go monitorManager.Start()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Setup router
	router := gin.Default()

	// API routes
	api := router.Group("/api/v1")
	handlers.SetupRoutes(api, db, monitorManager, wsHub)

	// Static files for frontend
	router.Static("/assets", "./frontend/dist/assets")
	router.StaticFile("/", "./frontend/dist/index.html")
	router.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
