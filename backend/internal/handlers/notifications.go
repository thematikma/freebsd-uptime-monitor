package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"uptime-monitor/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func SetupNotificationRoutes(router *gin.RouterGroup, db *sqlx.DB) {
	// Notification channel routes
	router.GET("/notifications/channels", getNotificationChannels(db))
	router.POST("/notifications/channels", createNotificationChannel(db))
	router.PUT("/notifications/channels/:id", updateNotificationChannel(db))
	router.DELETE("/notifications/channels/:id", deleteNotificationChannel(db))

	// Monitor-notification associations
	router.GET("/monitors/:id/notifications", getMonitorNotifications(db))
	router.POST("/monitors/:id/notifications", addMonitorNotification(db))
	router.DELETE("/notifications/:monitor_id/:channel_id", removeMonitorNotification(db))
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

func createNotificationChannel(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var channel models.NotificationChannel
		if err := c.ShouldBindJSON(&channel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate configuration based on type
		if err := validateNotificationConfig(channel.Type, channel.Config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		query := `
			INSERT INTO notification_channels (name, type, config, enabled)
			VALUES (?, ?, ?, ?) RETURNING id
		`

		err := db.QueryRow(query, channel.Name, channel.Type, channel.Config, channel.Enabled).Scan(&channel.ID)
		if err != nil {
			// Fallback for SQLite
			result, err := db.Exec(`
				INSERT INTO notification_channels (name, type, config, enabled)
				VALUES (?, ?, ?, ?)
			`, channel.Name, channel.Type, channel.Config, channel.Enabled)

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

		var channel models.NotificationChannel
		if err := c.ShouldBindJSON(&channel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate configuration
		if err := validateNotificationConfig(channel.Type, channel.Config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		channel.ID = id
		_, err = db.Exec(`
			UPDATE notification_channels 
			SET name = ?, type = ?, config = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, channel.Name, channel.Type, channel.Config, channel.Enabled, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

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

		_, err = db.Exec("DELETE FROM notification_channels WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Notification channel deleted"})
	}
}

func getMonitorNotifications(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID"})
			return
		}

		channels := []models.NotificationChannel{}
		query := `
			SELECT nc.* FROM notification_channels nc
			INNER JOIN monitor_notifications mn ON nc.id = mn.channel_id
			WHERE mn.monitor_id = ?
		`

		if err := db.Select(&channels, query, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, channels)
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
			ChannelID int `json:"channel_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(`
			INSERT INTO monitor_notifications (monitor_id, channel_id)
			VALUES (?, ?)
		`, monitorID, req.ChannelID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Notification added to monitor"})
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

func validateNotificationConfig(notificationType, config string) error {
	switch notificationType {
	case "discord":
		var discordConfig models.DiscordWebhookConfig
		if err := json.Unmarshal([]byte(config), &discordConfig); err != nil {
			return err
		}
		if discordConfig.WebhookURL == "" {
			return errors.New("webhook_url is required for Discord notifications")
		}
	case "webhook":
		var webhookConfig map[string]string
		if err := json.Unmarshal([]byte(config), &webhookConfig); err != nil {
			return err
		}
		if webhookConfig["url"] == "" {
			return errors.New("url is required for webhook notifications")
		}
	}
	return nil
}
