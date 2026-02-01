package database

import (
	"fmt"
	"uptime-monitor/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func Initialize(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	var dsn string
	var driver string

	switch cfg.Type {
	case "postgres":
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)
	case "sqlite":
		driver = "sqlite3"
		dsn = cfg.Database
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db, cfg.Type); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return db, nil
}

func runMigrations(db *sqlx.DB, dbType string) error {
	var schema string

	if dbType == "postgres" {
		schema = postgresSchema
	} else {
		schema = sqliteSchema
	}

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Run migrations for existing tables
	return migrateNotificationChannels(db, dbType)
}

// migrateNotificationChannels handles migration from old schema to new Shoutrrr-based schema
func migrateNotificationChannels(db *sqlx.DB, dbType string) error {
	// Check if shoutrrr_url column exists
	var hasColumn bool
	if dbType == "postgres" {
		err := db.Get(&hasColumn, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'notification_channels' AND column_name = 'shoutrrr_url'
			)
		`)
		if err != nil {
			return err
		}
	} else {
		// SQLite - check pragma
		rows, err := db.Query("PRAGMA table_info(notification_channels)")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var cid int
			var name, colType string
			var notNull, pk int
			var defaultValue interface{}
			if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
				return err
			}
			if name == "shoutrrr_url" {
				hasColumn = true
				break
			}
		}
	}

	// If column doesn't exist, we need to migrate the table
	if !hasColumn {
		if dbType == "sqlite" {
			// SQLite requires recreating the table
			_, err := db.Exec(`
				-- Create new table with correct schema
				CREATE TABLE IF NOT EXISTS notification_channels_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					shoutrrr_url TEXT NOT NULL DEFAULT '',
					events TEXT DEFAULT '["monitor_up","monitor_down","recovery"]',
					enabled BOOLEAN DEFAULT true,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);

				-- Copy existing data (if any compatible columns exist)
				INSERT OR IGNORE INTO notification_channels_new (id, name, enabled, created_at, updated_at)
				SELECT id, name, enabled, created_at, updated_at FROM notification_channels;

				-- Drop old table
				DROP TABLE notification_channels;

				-- Rename new table
				ALTER TABLE notification_channels_new RENAME TO notification_channels;
			`)
			if err != nil {
				return fmt.Errorf("failed to migrate notification_channels table: %v", err)
			}
		} else {
			// PostgreSQL - add columns
			_, err := db.Exec(`
				ALTER TABLE notification_channels 
				ADD COLUMN IF NOT EXISTS shoutrrr_url TEXT NOT NULL DEFAULT '',
				ADD COLUMN IF NOT EXISTS events TEXT DEFAULT '["monitor_up","monitor_down","recovery"]';
			`)
			if err != nil {
				return fmt.Errorf("failed to migrate notification_channels table: %v", err)
			}
		}
	}

	return nil
}

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'user',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'http',
    interval INTEGER DEFAULT 60,
    timeout INTEGER DEFAULT 30,
    max_retries INTEGER DEFAULT 3,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitor_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    monitor_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    response_time INTEGER,
    status_code INTEGER,
    message TEXT,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    monitor_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    target TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    shoutrrr_url TEXT NOT NULL,
    events TEXT DEFAULT '["monitor_up","monitor_down","recovery"]',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitor_notifications (
    monitor_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    events TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (monitor_id, channel_id),
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_monitor_checks_monitor_id ON monitor_checks(monitor_id);
CREATE INDEX IF NOT EXISTS idx_monitor_checks_checked_at ON monitor_checks(checked_at);
`

const postgresSchema = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'http',
    interval INTEGER DEFAULT 60,
    timeout INTEGER DEFAULT 30,
    max_retries INTEGER DEFAULT 3,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitor_checks (
    id SERIAL PRIMARY KEY,
    monitor_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    response_time INTEGER,
    status_code INTEGER,
    message TEXT,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    monitor_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    target TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    shoutrrr_url TEXT NOT NULL,
    events TEXT DEFAULT '["monitor_up","monitor_down","recovery"]',
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS monitor_notifications (
    monitor_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    events TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (monitor_id, channel_id),
    FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    FOREIGN KEY (channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_monitor_checks_monitor_id ON monitor_checks(monitor_id);
CREATE INDEX IF NOT EXISTS idx_monitor_checks_checked_at ON monitor_checks(checked_at);
`
