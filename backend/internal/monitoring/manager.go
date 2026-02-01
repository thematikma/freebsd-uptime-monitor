package monitoring

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"uptime-monitor/internal/models"
	"uptime-monitor/internal/notifications"

	"github.com/go-ping/ping"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

type Manager struct {
	db                    *sqlx.DB
	cron                  *cron.Cron
	checkers              map[int]*MonitorChecker
	mu                    sync.RWMutex
	shoutrrrManager       *notifications.ShoutrrrManager
	slowResponseThreshold int // in milliseconds
}

type MonitorChecker struct {
	monitor models.Monitor
	cronID  cron.EntryID
	manager *Manager
}

func NewManager(db *sqlx.DB) *Manager {
	return &Manager{
		db:                    db,
		cron:                  cron.New(),
		checkers:              make(map[int]*MonitorChecker),
		shoutrrrManager:       notifications.NewShoutrrrManager(db),
		slowResponseThreshold: 5000, // 5 seconds default
	}
}

func (m *Manager) Start() error {
	// Load existing monitors from database
	if err := m.loadMonitors(); err != nil {
		return fmt.Errorf("failed to load monitors: %v", err)
	}

	m.cron.Start()
	log.Println("Monitor manager started")
	return nil
}

func (m *Manager) Stop() {
	m.cron.Stop()
	log.Println("Monitor manager stopped")
}

func (m *Manager) loadMonitors() error {
	monitors := []models.Monitor{}
	query := "SELECT * FROM monitors WHERE active = ?"

	if err := m.db.Select(&monitors, query, true); err != nil {
		return err
	}

	for _, monitor := range monitors {
		if err := m.AddMonitor(monitor); err != nil {
			log.Printf("Failed to add monitor %s: %v", monitor.Name, err)
		}
	}

	return nil
}

func (m *Manager) AddMonitor(monitor models.Monitor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove existing checker if any
	if checker, exists := m.checkers[monitor.ID]; exists {
		m.cron.Remove(checker.cronID)
	}

	// Create new checker
	checker := &MonitorChecker{
		monitor: monitor,
		manager: m,
	}

	// Schedule checks
	spec := fmt.Sprintf("@every %ds", monitor.Interval)
	cronID, err := m.cron.AddFunc(spec, checker.check)
	if err != nil {
		return fmt.Errorf("failed to schedule monitor: %v", err)
	}

	checker.cronID = cronID
	m.checkers[monitor.ID] = checker

	log.Printf("Added monitor: %s (ID: %d)", monitor.Name, monitor.ID)
	return nil
}

func (m *Manager) RemoveMonitor(monitorID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if checker, exists := m.checkers[monitorID]; exists {
		m.cron.Remove(checker.cronID)
		delete(m.checkers, monitorID)
		log.Printf("Removed monitor ID: %d", monitorID)
	}
}

func (mc *MonitorChecker) check() {
	start := time.Now()
	check := models.MonitorCheck{
		MonitorID: mc.monitor.ID,
		CheckedAt: start,
	}

	var err error
	switch mc.monitor.Type {
	case "http", "https":
		err = mc.checkHTTP(&check)
	case "tcp":
		err = mc.checkTCP(&check)
	case "ping":
		err = mc.checkPing(&check)
	default:
		check.Status = "unknown"
		check.Message = "Unknown monitor type"
	}

	if err != nil {
		check.Status = "down"
		check.Message = err.Error()
	} else if check.Status == "" {
		check.Status = "up"
	}

	check.ResponseTime = int(time.Since(start).Milliseconds())

	// Save check result
	if err := mc.saveCheck(check); err != nil {
		log.Printf("Failed to save check for monitor %d: %v", mc.monitor.ID, err)
	}
}

func (mc *MonitorChecker) checkHTTP(check *models.MonitorCheck) error {
	client := &http.Client{
		Timeout: time.Duration(mc.monitor.Timeout) * time.Second,
	}

	resp, err := client.Get(mc.monitor.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	check.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		check.Status = "up"
		check.Message = "OK"
	} else {
		check.Status = "down"
		check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return nil
}

func (mc *MonitorChecker) checkTCP(check *models.MonitorCheck) error {
	var address string

	// Handle different URL formats for TCP
	if strings.HasPrefix(mc.monitor.URL, "tcp://") {
		u, err := url.Parse(mc.monitor.URL)
		if err != nil {
			return fmt.Errorf("invalid TCP URL: %v", err)
		}
		address = u.Host
	} else {
		// Handle direct host:port format
		address = mc.monitor.URL
	}

	if address == "" {
		return fmt.Errorf("no host:port specified")
	}

	// Validate address format (should be host:port)
	if !strings.Contains(address, ":") {
		return fmt.Errorf("TCP check requires host:port format, got: %s", address)
	}

	// Try to connect
	timeout := time.Duration(mc.monitor.Timeout) * time.Second
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("TCP connection failed to %s: %v", address, err)
	}
	conn.Close()

	check.Status = "up"
	check.Message = fmt.Sprintf("TCP connection successful to %s", address)
	return nil
}

func (mc *MonitorChecker) checkPing(check *models.MonitorCheck) error {
	// Parse URL to get hostname
	u, err := url.Parse(mc.monitor.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	host := u.Host
	if host == "" {
		host = u.Path // For ping://hostname format
	}

	pinger, err := ping.NewPinger(host)
	if err != nil {
		return err
	}
	pinger.SetPrivileged(false) // Use unprivileged mode for FreeBSD compatibility
	pinger.Count = 3
	pinger.Timeout = time.Duration(mc.monitor.Timeout) * time.Second

	err = pinger.Run()
	if err != nil {
		return err
	}

	stats := pinger.Statistics()
	if stats.PacketsRecv == 0 {
		return fmt.Errorf("no packets received")
	}

	check.Status = "up"
	check.Message = fmt.Sprintf("Ping successful, avg RTT: %v", stats.AvgRtt)
	check.ResponseTime = int(stats.AvgRtt.Milliseconds())
	return nil
}

func (mc *MonitorChecker) saveCheck(check models.MonitorCheck) error {
	// Get previous status for notification comparison
	var previousStatus string
	err := mc.manager.db.Get(&previousStatus, `
		SELECT status FROM monitor_checks 
		WHERE monitor_id = ? 
		ORDER BY checked_at DESC 
		LIMIT 1
	`, check.MonitorID)

	if err != nil {
		// No previous checks, assume unknown
		previousStatus = "unknown"
	}

	query := `
		INSERT INTO monitor_checks (monitor_id, status, response_time, status_code, message, checked_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = mc.manager.db.Exec(query,
		check.MonitorID,
		check.Status,
		check.ResponseTime,
		check.StatusCode,
		check.Message,
		check.CheckedAt,
	)

	if err != nil {
		return err
	}

	// Determine which event to send notification for
	event := notifications.DetermineEvent(check.Status, previousStatus, check.ResponseTime, mc.manager.slowResponseThreshold)

	// Send notifications if there's a relevant event
	if event != "" {
		go func() {
			if err := mc.manager.shoutrrrManager.SendMonitorAlert(mc.monitor, check, event, previousStatus); err != nil {
				log.Printf("Failed to send notification for monitor %d: %v", mc.monitor.ID, err)
			}
		}()
	}

	return nil
}
