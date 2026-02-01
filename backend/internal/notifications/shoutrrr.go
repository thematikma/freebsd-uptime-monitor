package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"uptime-monitor/internal/models"

	"github.com/containrrr/shoutrrr"
	"github.com/jmoiron/sqlx"
)

// ShoutrrrManager handles all Shoutrrr-based notifications
type ShoutrrrManager struct {
	db *sqlx.DB
}

// NewShoutrrrManager creates a new Shoutrrr notification manager
func NewShoutrrrManager(db *sqlx.DB) *ShoutrrrManager {
	return &ShoutrrrManager{
		db: db,
	}
}

// GetSupportedServices returns information about supported Shoutrrr services
func GetSupportedServices() []models.ShoutrrrServiceInfo {
	return []models.ShoutrrrServiceInfo{
		{
			Name:        "Discord",
			URLFormat:   "discord://token@id",
			Example:     "discord://token@webhookid",
			Description: "Discord webhook notifications",
		},
		{
			Name:        "Slack",
			URLFormat:   "slack://token:token@channel",
			Example:     "slack://token-a/token-b/token-c@channel",
			Description: "Slack incoming webhooks",
		},
		{
			Name:        "Telegram",
			URLFormat:   "telegram://token@telegram?chats=@channel",
			Example:     "telegram://123456:ABC-DEF@telegram?chats=@mychannel",
			Description: "Telegram bot notifications",
		},
		{
			Name:        "Email (SMTP)",
			URLFormat:   "smtp://user:pass@host:port/?to=recipient",
			Example:     "smtp://user:pass@smtp.gmail.com:587/?from=sender@gmail.com&to=recipient@example.com",
			Description: "Email via SMTP server",
		},
		{
			Name:        "Pushover",
			URLFormat:   "pushover://shoutrrr:token@user/?devices=device",
			Example:     "pushover://shoutrrr:apitoken@userkey",
			Description: "Pushover push notifications",
		},
		{
			Name:        "Gotify",
			URLFormat:   "gotify://host/token",
			Example:     "gotify://gotify.example.com/tokenvalue",
			Description: "Gotify self-hosted notifications",
		},
		{
			Name:        "Ntfy",
			URLFormat:   "ntfy://user:pass@host/topic",
			Example:     "ntfy://ntfy.sh/mytopic",
			Description: "Ntfy pub/sub notifications",
		},
		{
			Name:        "Matrix",
			URLFormat:   "matrix://user:pass@host/?rooms=!roomid",
			Example:     "matrix://user:pass@matrix.org/?rooms=!roomid:matrix.org",
			Description: "Matrix chat notifications",
		},
		{
			Name:        "Mattermost",
			URLFormat:   "mattermost://user@host/token/channel",
			Example:     "mattermost://user@mattermost.example.com/token/town-square",
			Description: "Mattermost webhook notifications",
		},
		{
			Name:        "Opsgenie",
			URLFormat:   "opsgenie://host/apikey",
			Example:     "opsgenie://api.opsgenie.com/apikey",
			Description: "Opsgenie alert notifications",
		},
		{
			Name:        "Pushbullet",
			URLFormat:   "pushbullet://token",
			Example:     "pushbullet://o.abcdefghijklmnopqrstuvwxyz",
			Description: "Pushbullet push notifications",
		},
		{
			Name:        "Webhook (Generic)",
			URLFormat:   "generic://host/path",
			Example:     "generic://example.com/webhook?template=json",
			Description: "Generic webhook POST requests",
		},
		{
			Name:        "Teams",
			URLFormat:   "teams://group@tenant/altId/groupOwner?host=host",
			Example:     "teams://group@tenant/altId/groupOwner?host=organization.webhook.office.com",
			Description: "Microsoft Teams webhook",
		},
		{
			Name:        "Zulip",
			URLFormat:   "zulip://bot-mail:key@host/?stream=stream&topic=topic",
			Example:     "zulip://bot@zulip.com:botkey@chat.zulip.org/?stream=general&topic=alerts",
			Description: "Zulip chat notifications",
		},
		{
			Name:        "Rocket.Chat",
			URLFormat:   "rocketchat://user:token@host/channel",
			Example:     "rocketchat://user:token@rocket.example.com/general",
			Description: "Rocket.Chat notifications",
		},
	}
}

// ValidateShoutrrrURL tests if a Shoutrrr URL is valid
func (sm *ShoutrrrManager) ValidateShoutrrrURL(url string) error {
	// Create a temporary sender to validate
	_, err := shoutrrr.CreateSender(url)
	if err != nil {
		return fmt.Errorf("invalid Shoutrrr URL: %v", err)
	}
	return nil
}

// SendNotification sends a notification via Shoutrrr
func (sm *ShoutrrrManager) SendNotification(shoutrrrURL, message string) error {
	sender, err := shoutrrr.CreateSender(shoutrrrURL)
	if err != nil {
		return fmt.Errorf("failed to create sender: %v", err)
	}

	errs := sender.Send(message, nil)
	for _, err := range errs {
		if err != nil {
			return fmt.Errorf("failed to send notification: %v", err)
		}
	}
	return nil
}

// SendTestNotification sends a test notification to verify the configuration
func (sm *ShoutrrrManager) SendTestNotification(shoutrrrURL string) error {
	message := "🧪 Test notification from Uptime Monitor - Your notification channel is configured correctly!"
	return sm.SendNotification(shoutrrrURL, message)
}

// SendMonitorAlert sends notifications when a monitor event occurs
func (sm *ShoutrrrManager) SendMonitorAlert(monitor models.Monitor, check models.MonitorCheck, event models.NotificationEvent, previousStatus string) error {
	// Get all notification channels associated with this monitor that have this event enabled
	channels, err := sm.getChannelsForMonitorEvent(monitor.ID, event)
	if err != nil {
		return fmt.Errorf("failed to get notification channels: %v", err)
	}

	if len(channels) == 0 {
		log.Printf("No notification channels configured for monitor %d and event %s", monitor.ID, event)
		return nil
	}

	// Build the notification message
	message := sm.buildMessage(monitor, check, event, previousStatus)

	// Send to all channels
	var lastErr error
	for _, channel := range channels {
		if !channel.Enabled {
			continue
		}

		err := sm.SendNotification(channel.ShoutrrrURL, message)
		if err != nil {
			log.Printf("Failed to send notification to channel %s: %v", channel.Name, err)
			lastErr = err
		} else {
			log.Printf("Notification sent to channel %s for monitor %s", channel.Name, monitor.Name)
		}
	}

	return lastErr
}

// getChannelsForMonitorEvent retrieves channels that should be notified for a specific event
func (sm *ShoutrrrManager) getChannelsForMonitorEvent(monitorID int, event models.NotificationEvent) ([]models.NotificationChannel, error) {
	// First, get channels directly associated with this monitor
	channels := []models.NotificationChannel{}

	query := `
		SELECT nc.* FROM notification_channels nc
		INNER JOIN monitor_notifications mn ON nc.id = mn.channel_id
		WHERE mn.monitor_id = ? AND nc.enabled = ?
	`

	err := sm.db.Select(&channels, query, monitorID, true)
	if err != nil {
		return nil, err
	}

	// Filter channels that have this event enabled
	var filteredChannels []models.NotificationChannel
	for _, channel := range channels {
		if sm.channelHasEvent(channel, event) {
			filteredChannels = append(filteredChannels, channel)
		}
	}

	return filteredChannels, nil
}

// channelHasEvent checks if a notification channel has a specific event enabled
func (sm *ShoutrrrManager) channelHasEvent(channel models.NotificationChannel, event models.NotificationEvent) bool {
	if channel.Events == "" {
		// If no events specified, default to up/down events
		return event == models.EventMonitorUp || event == models.EventMonitorDown
	}

	var events []string
	if err := json.Unmarshal([]byte(channel.Events), &events); err != nil {
		log.Printf("Failed to parse events for channel %s: %v", channel.Name, err)
		return false
	}

	for _, e := range events {
		if e == string(event) {
			return true
		}
	}
	return false
}

// buildMessage creates a formatted notification message
func (sm *ShoutrrrManager) buildMessage(monitor models.Monitor, check models.MonitorCheck, event models.NotificationEvent, previousStatus string) string {
	var emoji, title string

	switch event {
	case models.EventMonitorUp:
		emoji = "✅"
		title = "Monitor UP"
	case models.EventMonitorDown:
		emoji = "🔴"
		title = "Monitor DOWN"
	case models.EventResponseSlow:
		emoji = "🐢"
		title = "Slow Response"
	case models.EventSSLExpiringSoon:
		emoji = "🔐"
		title = "SSL Certificate Expiring"
	case models.EventRecovery:
		emoji = "🔄"
		title = "Monitor Recovered"
	default:
		emoji = "ℹ️"
		title = "Monitor Alert"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %s: %s\n", emoji, title, monitor.Name))
	sb.WriteString(fmt.Sprintf("URL: %s\n", monitor.URL))

	if previousStatus != "" && previousStatus != check.Status {
		sb.WriteString(fmt.Sprintf("Status: %s → %s\n", previousStatus, check.Status))
	} else {
		sb.WriteString(fmt.Sprintf("Status: %s\n", check.Status))
	}

	if check.ResponseTime > 0 {
		sb.WriteString(fmt.Sprintf("Response Time: %dms\n", check.ResponseTime))
	}

	if check.Message != "" {
		sb.WriteString(fmt.Sprintf("Message: %s\n", check.Message))
	}

	sb.WriteString(fmt.Sprintf("Checked: %s", check.CheckedAt.Format("2006-01-02 15:04:05 MST")))

	return sb.String()
}

// DetermineEvent determines what notification event should be triggered based on status change
func DetermineEvent(currentStatus, previousStatus string, responseTime int, slowThreshold int) models.NotificationEvent {
	// Status changed from down to up
	if previousStatus == "down" && currentStatus == "up" {
		return models.EventRecovery
	}

	// Status changed from up to down
	if previousStatus == "up" && currentStatus == "down" {
		return models.EventMonitorDown
	}

	// Monitor is up but was unknown before
	if previousStatus == "unknown" && currentStatus == "up" {
		return models.EventMonitorUp
	}

	// Monitor went down from unknown
	if previousStatus == "unknown" && currentStatus == "down" {
		return models.EventMonitorDown
	}

	// Response time is slow (only if up)
	if currentStatus == "up" && slowThreshold > 0 && responseTime > slowThreshold {
		return models.EventResponseSlow
	}

	return ""
}
