package models

import (
	"time"
)

type Monitor struct {
	ID            int           `json:"id" db:"id"`
	Name          string        `json:"name" db:"name"`
	URL           string        `json:"url" db:"url"`
	Type          string        `json:"type" db:"type"` // http, tcp, ping
	Interval      int           `json:"interval" db:"interval"`
	Timeout       int           `json:"timeout" db:"timeout"`
	MaxRetries    int           `json:"max_retries" db:"max_retries"`
	Active        bool          `json:"active" db:"active"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	LastCheck     *MonitorCheck `json:"last_check,omitempty" db:"-"`
	CurrentStatus string        `json:"current_status,omitempty" db:"-"`
}

type MonitorCheck struct {
	ID           int       `json:"id" db:"id"`
	MonitorID    int       `json:"monitor_id" db:"monitor_id"`
	Status       string    `json:"status" db:"status"`               // up, down, unknown
	ResponseTime int       `json:"response_time" db:"response_time"` // milliseconds
	StatusCode   int       `json:"status_code" db:"status_code"`
	Message      string    `json:"message" db:"message"`
	CheckedAt    time.Time `json:"checked_at" db:"checked_at"`
}

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Alert struct {
	ID        int       `json:"id" db:"id"`
	MonitorID int       `json:"monitor_id" db:"monitor_id"`
	Type      string    `json:"type" db:"type"` // email, webhook, slack
	Target    string    `json:"target" db:"target"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MonitorStats struct {
	MonitorID       int     `json:"monitor_id"`
	UptimePercent   float64 `json:"uptime_percent"`
	TotalChecks     int     `json:"total_checks"`
	SuccessChecks   int     `json:"success_checks"`
	FailedChecks    int     `json:"failed_checks"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// Notification types and structures
type NotificationChannel struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	ShoutrrrURL  string    `json:"shoutrrr_url" db:"shoutrrr_url"` // Shoutrrr URL format
	Events       string    `json:"events" db:"events"`             // JSON array of event types
	Enabled      bool      `json:"enabled" db:"enabled"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationEvent represents the types of events that can trigger notifications
type NotificationEvent string

const (
	EventMonitorUp           NotificationEvent = "monitor_up"
	EventMonitorDown         NotificationEvent = "monitor_down"
	EventResponseSlow        NotificationEvent = "response_slow"
	EventSSLExpiringSoon     NotificationEvent = "ssl_expiring"
	EventRecovery            NotificationEvent = "recovery"
)

// NotificationChannelConfig for frontend
type NotificationChannelConfig struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	ShoutrrrURL string   `json:"shoutrrr_url"`
	Events      []string `json:"events"`
	Enabled     bool     `json:"enabled"`
	MonitorIDs  []int    `json:"monitor_ids,omitempty"`
}

// MonitorNotificationAssoc links monitors to notification channels with specific events
type MonitorNotificationAssoc struct {
	MonitorID int    `json:"monitor_id" db:"monitor_id"`
	ChannelID int    `json:"channel_id" db:"channel_id"`
	Events    string `json:"events" db:"events"` // JSON array, overrides channel default if set
}

// ShoutrrrServiceInfo provides info about supported services
type ShoutrrrServiceInfo struct {
	Name        string `json:"name"`
	URLFormat   string `json:"url_format"`
	Example     string `json:"example"`
	Description string `json:"description"`
}
