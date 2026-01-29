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
