package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"uptime-monitor/internal/models"
	"uptime-monitor/internal/notifications"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

var shoutrrrManager *notifications.ShoutrrrManager

func SetupNotificationRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	// Initialize Shoutrrr manager
	shoutrrrManager = notifications.NewShoutrrrManager(db)

	// Notification channel routes
	router.GET("/notifications/channels", getNotificationChannels(db))
	router.POST("/notifications/channels", createNotificationChannel(db))
	router.GET("/notifications/channels/:id", getNotificationChannel(db))
	router.PUT("/notifications/channels/:id", updateNotificationChannel(db))
	router.DELETE("/notifications/channels/:id", deleteNotificationChannel(db))

	// Test notification
	router.POST("/notifications/channels/:id/test", testNotificationChannel(db))
	router.POST("/notifications/test", testShoutrrrURL())

	// Validate Shoutrrr URL
	router.POST("/notifications/validate", validateShoutrrrURL())

	// Get supported services
	router.GET("/notifications/services", getSupportedServices())

	// Get available events
	router.GET("/notifications/events", getAvailableEvents())

	// Monitor-notification associations
	router.GET("/monitors/:id/notifications", getMonitorNotifications(db))
	router.POST("/monitors/:id/notifications", addMonitorNotification(db))
	router.PUT("/monitors/:id/notifications", updateMonitorNotifications(db))
	router.DELETE("/monitors/:monitor_id/notifications/:channel_id", removeMonitorNotification(db))
}

// GetShoutrrrManager returns the Shoutrrr manager for use in monitoring
func GetShoutrrrManager() *notifications.ShoutrrrManager {
	return shoutrrrManager
}

func getSupportedServices() gin.HandlerFunc {
	return func(c *gin.Context) {
		services := notifications.GetSupportedServices()
		c.JSON(http.StatusOK, services)
	}
}

func getAvailableEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		events := []map[string]string{
			{"id": "monitor_up", "name": "Monitor Up", "description": "When a monitor comes online (first check)"},
			{"id": "monitor_down", "name": "Monitor Down", "description": "When a monitor goes offline"},
			{"id": "recovery", "name": "Recovery", "description": "When a monitor recovers from down state"},
			{"id": "response_slow", "name": "Slow Response", "description": "When response time exceeds threshold"},
			{"id": "ssl_expiring", "name": "SSL Expiring", "description": "When SSL certificate is about to expire"},
		}
		c.JSON(http.StatusOK, events)
	}
}

func validateShoutrrrURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			URL string `json:"url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := shoutrrrManager.ValidateShoutrrrURL(req.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"valid": true})
	}
}

func testShoutrrrURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			URL string `json:"url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := shoutrrrManager.SendTestNotification(req.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Test notification sent successfully"})
	}
}

func getNotificationChannels(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		channels := []models.NotificationChannel{}
		if err := db.Select(&channels, "SELECT * FROM notification_channels ORDER BY name"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, channels)
	}
}

func getNotificationChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
			return
		}

		var channel models.NotificationChannel
		err = db.Get(&channel, "SELECT * FROM notification_channels WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification channel not found"})
			return
		}

		c.JSON(http.StatusOK, channel)
	}
}

func createNotificationChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string   `json:"name" binding:"required"`
			ShoutrrrURL string   `json:"shoutrrr_url" binding:"required"`
			Events      []string `json:"events"`
			Enabled     bool     `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate Shoutrrr URL
		if err := shoutrrrManager.ValidateShoutrrrURL(req.ShoutrrrURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Shoutrrr URL: " + err.Error()})
			return
		}

		// Default events if not specified
		if len(req.Events) == 0 {
			req.Events = []string{"monitor_up", "monitor_down", "recovery"}
		}

		eventsJSON, err := json.Marshal(req.Events)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode events"})
			return
		}

		var channel models.NotificationChannel
		channel.Name = req.Name
		channel.ShoutrrrURL = req.ShoutrrrURL
		channel.Events = string(eventsJSON)
		channel.Enabled = req.Enabled

		// Try with RETURNING clause first (PostgreSQL)
		err = db.QueryRow(`
			INSERT INTO notification_channels (name, shoutrrr_url, events, enabled)
			VALUES ($1, $2, $3, $4) RETURNING id
		`, channel.Name, channel.ShoutrrrURL, channel.Events, channel.Enabled).Scan(&channel.ID)

		if err != nil {
			// Fallback for SQLite
			result, err := db.Exec(`
				INSERT INTO notification_channels (name, shoutrrr_url, events, enabled)
				VALUES (?, ?, ?, ?)
			`, channel.Name, channel.ShoutrrrURL, channel.Events, channel.Enabled)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			id, _ := result.LastInsertId()
			channel.ID = int(id)
		}

		c.JSON(http.StatusCreated, channel)
	}
}

func updateNotificationChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
			return
		}

		var req struct {
			Name        string   `json:"name" binding:"required"`
			ShoutrrrURL string   `json:"shoutrrr_url" binding:"required"`
			Events      []string `json:"events"`
			Enabled     bool     `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate Shoutrrr URL
		if err := shoutrrrManager.ValidateShoutrrrURL(req.ShoutrrrURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Shoutrrr URL: " + err.Error()})
			return
		}

		// Default events if not specified
		if len(req.Events) == 0 {
			req.Events = []string{"monitor_up", "monitor_down", "recovery"}
		}

		eventsJSON, err := json.Marshal(req.Events)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode events"})
			return
		}

		_, err = db.Exec(`
			UPDATE notification_channels 
			SET name = ?, shoutrrr_url = ?, events = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, req.Name, req.ShoutrrrURL, string(eventsJSON), req.Enabled, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return updated channel
		var channel models.NotificationChannel
		db.Get(&channel, "SELECT * FROM notification_channels WHERE id = ?", id)
		c.JSON(http.StatusOK, channel)
	}
}

func deleteNotificationChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
			return
		}

		// First delete all monitor associations
		db.Exec("DELETE FROM monitor_notifications WHERE channel_id = ?", id)

		// Then delete the channel
		_, err = db.Exec("DELETE FROM notification_channels WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification channel deleted"})
	}
}

func testNotificationChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
			return
		}

		var channel models.NotificationChannel
		err = db.Get(&channel, "SELECT * FROM notification_channels WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification channel not found"})
			return
		}

		if err := shoutrrrManager.SendTestNotification(channel.ShoutrrrURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Test notification sent successfully"})
	}
}

func getMonitorNotifications(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		// Get channels with their association info
		type ChannelWithAssoc struct {
			models.NotificationChannel
			AssocEvents *string `db:"assoc_events"`
		}

		channels := []ChannelWithAssoc{}
		query := `
			SELECT nc.*, mn.events as assoc_events 
			FROM notification_channels nc
			INNER JOIN monitor_notifications mn ON nc.id = mn.channel_id
			WHERE mn.monitor_id = ?
		`

		if err := db.Select(&channels, query, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Convert to response format
		var result []map[string]interface{}
		for _, ch := range channels {
			events := ch.Events
			if ch.AssocEvents != nil && *ch.AssocEvents != "" {
				events = *ch.AssocEvents
			}
			result = append(result, map[string]interface{}{
				"id":           ch.ID,
				"name":         ch.Name,
				"shoutrrr_url": ch.ShoutrrrURL,
				"events":       events,
				"enabled":      ch.Enabled,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}

func addMonitorNotification(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		monitorID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		var req struct {
			ChannelID int      `json:"channel_id" binding:"required"`
			Events    []string `json:"events"` // Optional: override channel events for this monitor
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var eventsJSON *string
		if len(req.Events) > 0 {
			jsonBytes, _ := json.Marshal(req.Events)
			jsonStr := string(jsonBytes)
			eventsJSON = &jsonStr
		}

		_, err = db.Exec(`
			INSERT OR REPLACE INTO monitor_notifications (monitor_id, channel_id, events)
			VALUES (?, ?, ?)
		`, monitorID, req.ChannelID, eventsJSON)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Notification added to monitor"})
	}
}

func updateMonitorNotifications(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		monitorID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		var req struct {
			ChannelIDs []int `json:"channel_ids"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Start transaction
		tx, err := db.Beginx()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Delete existing associations
		_, err = tx.Exec("DELETE FROM monitor_notifications WHERE monitor_id = ?", monitorID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Add new associations
		for _, channelID := range req.ChannelIDs {
			_, err = tx.Exec(`
				INSERT INTO monitor_notifications (monitor_id, channel_id)
				VALUES (?, ?)
			`, monitorID, channelID)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Monitor notifications updated"})
	}
}

func removeMonitorNotification(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		monitorID, err := strconv.Atoi(c.Param("monitor_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		channelID, err := strconv.Atoi(c.Param("channel_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
			return
		}

		_, err = db.Exec(`
			DELETE FROM monitor_notifications 
			WHERE monitor_id = ? AND channel_id = ?
		`, monitorID, channelID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification removed from monitor"})
	}
}
