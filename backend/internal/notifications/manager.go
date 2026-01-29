package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"uptime-monitor/internal/models"

	"github.com/jmoiron/sqlx"
)

type NotificationManager struct {
	db *sqlx.DB
}

func NewNotificationManager(db *sqlx.DB) *NotificationManager {
	return &NotificationManager{
		db: db,
	}
}

// SendMonitorAlert sends notifications when a monitor status changes
func (nm *NotificationManager) SendMonitorAlert(monitor models.Monitor, check models.MonitorCheck, previousStatus string) error {
	// Only send notifications on status changes
	if check.Status == previousStatus {
		return nil
	}

	// Get all notification channels
	channels := []models.NotificationChannel{}
	err := nm.db.Select(&channels, `
		SELECT nc.* FROM notification_channels nc
		INNER JOIN monitor_notifications mn ON nc.id = mn.channel_id
		WHERE mn.monitor_id = ? AND nc.enabled = ?
	`, monitor.ID, true)

	if err != nil {
		return fmt.Errorf("failed to get notification channels: %v", err)
	}

	// Send notifications to each channel
	for _, channel := range channels {
		switch channel.Type {
		case "discord":
			err := nm.sendDiscordNotification(channel, monitor, check, previousStatus)
			if err != nil {
				fmt.Printf("Failed to send Discord notification: %v\n", err)
			}
		case "webhook":
			err := nm.sendWebhookNotification(channel, monitor, check, previousStatus)
			if err != nil {
				fmt.Printf("Failed to send webhook notification: %v\n", err)
			}
		}
	}

	return nil
}

func (nm *NotificationManager) sendDiscordNotification(channel models.NotificationChannel, monitor models.Monitor, check models.MonitorCheck, previousStatus string) error {
	var config models.DiscordWebhookConfig
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse Discord config: %v", err)
	}

	// Determine color based on status
	color := 0x6b7280 // gray for unknown
	if check.Status == "up" {
		color = 0x10b981 // green
	} else if check.Status == "down" {
		color = 0xef4444 // red
	}

	// Create status change message
	statusEmoji := "🟢"
	if check.Status == "down" {
		statusEmoji = "🔴"
	} else if check.Status == "unknown" {
		statusEmoji = "🟡"
	}

	title := fmt.Sprintf("%s Monitor Alert: %s", statusEmoji, monitor.Name)
	description := fmt.Sprintf("Status changed from **%s** to **%s**\n\nURL: %s\nMessage: %s",
		previousStatus, check.Status, monitor.URL, check.Message)

	if check.ResponseTime > 0 {
		description += fmt.Sprintf("\nResponse Time: %dms", check.ResponseTime)
	}

	message := models.DiscordMessage{
		Username: config.Username,
		Embeds: []models.DiscordEmbed{
			{
				Title:       title,
				Description: description,
				Color:       color,
				Timestamp:   check.CheckedAt.Format(time.RFC3339),
			},
		},
	}

	if config.AvatarURL != "" {
		message.Avatar = config.AvatarURL
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %v", err)
	}

	resp, err := http.Post(config.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Discord webhook returned status: %d", resp.StatusCode)
	}

	return nil
}

func (nm *NotificationManager) sendWebhookNotification(channel models.NotificationChannel, monitor models.Monitor, check models.MonitorCheck, previousStatus string) error {
	// Basic webhook payload
	payload := map[string]interface{}{
		"monitor_id":      monitor.ID,
		"monitor_name":    monitor.Name,
		"monitor_url":     monitor.URL,
		"status":          check.Status,
		"previous_status": previousStatus,
		"message":         check.Message,
		"response_time":   check.ResponseTime,
		"checked_at":      check.CheckedAt,
		"timestamp":       time.Now().Unix(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %v", err)
	}

	// Parse webhook config (should contain URL)
	var config map[string]string
	if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
		return fmt.Errorf("failed to parse webhook config: %v", err)
	}

	webhookURL, exists := config["url"]
	if !exists {
		return fmt.Errorf("webhook URL not configured")
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	return nil
}
