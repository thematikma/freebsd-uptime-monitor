package handlers

import (
	"net/http"
	"strconv"
	"uptime-monitor/internal/auth"
	"uptime-monitor/internal/models"
	"uptime-monitor/internal/monitoring"
	"uptime-monitor/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetupRoutes(router *gin.RouterGroup, db *sqlx.DB, monitorManager *monitoring.Manager, wsHub *websocket.Hub, authService *auth.Service) {
	// Authentication routes
	SetupAuthRoutes(router, db, authService)

	// Monitor routes
	router.GET("/monitors", getMonitors(db))
	router.POST("/monitors", createMonitor(db, monitorManager))
	router.GET("/monitors/:id", getMonitor(db))
	router.PUT("/monitors/:id", updateMonitor(db, monitorManager))
	router.DELETE("/monitors/:id", deleteMonitor(db, monitorManager))

	// Check routes
	router.GET("/monitors/:id/checks", getMonitorChecks(db))
	router.GET("/monitors/:id/stats", getMonitorStats(db))

	// Dashboard routes
	router.GET("/dashboard", getDashboard(db))

	// WebSocket endpoint
	router.GET("/ws", gin.WrapH(wsHub.HandleWebSocket()))
}

func getMonitors(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		monitors := []models.Monitor{}
		if err := db.Select(&monitors, "SELECT * FROM monitors ORDER BY created_at DESC"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Add last check information for each monitor
		for i := range monitors {
			var lastCheck models.MonitorCheck
			err := db.Get(&lastCheck, `
				SELECT * FROM monitor_checks 
				WHERE monitor_id = ? 
				ORDER BY checked_at DESC 
				LIMIT 1
			`, monitors[i].ID)

			if err == nil {
				monitors[i].LastCheck = &lastCheck
				monitors[i].CurrentStatus = lastCheck.Status
			} else {
				monitors[i].CurrentStatus = "unknown"
			}
		}

		c.JSON(http.StatusOK, monitors)
	}
}

func createMonitor(db *sqlx.DB, manager *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var monitor models.Monitor
		if err := c.ShouldBindJSON(&monitor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set defaults
		if monitor.Interval == 0 {
			monitor.Interval = 60
		}
		if monitor.Timeout == 0 {
			monitor.Timeout = 30
		}
		if monitor.MaxRetries == 0 {
			monitor.MaxRetries = 3
		}

		query := `
			INSERT INTO monitors (name, url, type, interval, timeout, max_retries, active)
			VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id
		`

		err := db.QueryRow(query, monitor.Name, monitor.URL, monitor.Type,
			monitor.Interval, monitor.Timeout, monitor.MaxRetries, monitor.Active).Scan(&monitor.ID)
		if err != nil {
			// Fallback for SQLite
			result, err := db.Exec(`
				INSERT INTO monitors (name, url, type, interval, timeout, max_retries, active)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, monitor.Name, monitor.URL, monitor.Type,
				monitor.Interval, monitor.Timeout, monitor.MaxRetries, monitor.Active)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			id, _ := result.LastInsertId()
			monitor.ID = int(id)
		}

		// Add to monitoring manager
		if monitor.Active {
			manager.AddMonitor(monitor)
		}

		c.JSON(http.StatusCreated, monitor)
	}
}

func getMonitor(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		var monitor models.Monitor
		err = db.Get(&monitor, "SELECT * FROM monitors WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
			return
		}

		c.JSON(http.StatusOK, monitor)
	}
}

func updateMonitor(db *sqlx.DB, manager *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		var monitor models.Monitor
		if err := c.ShouldBindJSON(&monitor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		monitor.ID = id
		query := `
			UPDATE monitors 
			SET name = ?, url = ?, type = ?, interval = ?, timeout = ?, max_retries = ?, active = ?
			WHERE id = ?
		`

		_, err = db.Exec(query, monitor.Name, monitor.URL, monitor.Type,
			monitor.Interval, monitor.Timeout, monitor.MaxRetries, monitor.Active, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update monitoring manager
		manager.RemoveMonitor(id)
		if monitor.Active {
			manager.AddMonitor(monitor)
		}

		c.JSON(http.StatusOK, monitor)
	}
}

func deleteMonitor(db *sqlx.DB, manager *monitoring.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		_, err = db.Exec("DELETE FROM monitors WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		manager.RemoveMonitor(id)
		c.JSON(http.StatusOK, gin.H{"message": "Monitor deleted"})
	}
}

func getMonitorChecks(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		limit := c.DefaultQuery("limit", "100")

		checks := []models.MonitorCheck{}
		query := `
			SELECT * FROM monitor_checks 
			WHERE monitor_id = ? 
			ORDER BY checked_at DESC 
			LIMIT ?
		`

		if err := db.Select(&checks, query, id, limit); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, checks)
	}
}

func getMonitorStats(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		var stats models.MonitorStats
		query := `
			SELECT 
				monitor_id,
				COUNT(*) as total_checks,
				SUM(CASE WHEN status = 'up' THEN 1 ELSE 0 END) as success_checks,
				SUM(CASE WHEN status = 'down' THEN 1 ELSE 0 END) as failed_checks,
				AVG(CASE WHEN response_time > 0 THEN response_time END) as avg_response_time,
				(SUM(CASE WHEN status = 'up' THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as uptime_percent
			FROM monitor_checks 
			WHERE monitor_id = ? AND checked_at > datetime('now', '-24 hours')
		`

		err = db.Get(&stats, query, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

func getDashboard(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get monitor count by status
		var dashboard struct {
			TotalMonitors int     `json:"total_monitors"`
			UpMonitors    int     `json:"up_monitors"`
			DownMonitors  int     `json:"down_monitors"`
			AvgUptime     float64 `json:"avg_uptime"`
		}

		// Count total monitors
		db.Get(&dashboard.TotalMonitors, "SELECT COUNT(*) FROM monitors WHERE active = ?", true)

		// Get recent status for each monitor
		query := `
			SELECT 
				m.id,
				COALESCE(latest.status, 'unknown') as status
			FROM monitors m
			LEFT JOIN (
				SELECT 
					monitor_id, 
					status,
					ROW_NUMBER() OVER (PARTITION BY monitor_id ORDER BY checked_at DESC) as rn
				FROM monitor_checks
				WHERE checked_at > datetime('now', '-30 minutes')
			) latest ON m.id = latest.monitor_id AND latest.rn = 1
			WHERE m.active = ?
		`

		rows, err := db.Query(query, true)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id int
				var status string
				if err := rows.Scan(&id, &status); err == nil {
					if status == "up" {
						dashboard.UpMonitors++
					} else if status == "down" {
						dashboard.DownMonitors++
					}
				}
			}
		}

		// Calculate average uptime
		var uptime float64
		uptimeQuery := `
			SELECT AVG(uptime_percent) FROM (
				SELECT 
					monitor_id,
					(SUM(CASE WHEN status = 'up' THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as uptime_percent
				FROM monitor_checks 
				WHERE checked_at > datetime('now', '-24 hours')
				GROUP BY monitor_id
			)
		`
		db.Get(&uptime, uptimeQuery)
		dashboard.AvgUptime = uptime

		c.JSON(http.StatusOK, dashboard)
	}
}
