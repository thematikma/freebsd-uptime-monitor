package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Monitor  MonitorConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Type     string // "sqlite" or "postgres"
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret string
}

type MonitorConfig struct {
	CheckInterval int // seconds
	Timeout       int // seconds
	MaxRetries    int
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Database: getEnv("DB_NAME", "uptime_monitor.db"),
			Username: getEnv("DB_USER", ""),
			Password: getEnv("DB_PASS", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-this"),
		},
		Monitor: MonitorConfig{
			CheckInterval: getEnvInt("MONITOR_INTERVAL", 60),
			Timeout:       getEnvInt("MONITOR_TIMEOUT", 30),
			MaxRetries:    getEnvInt("MONITOR_RETRIES", 3),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
