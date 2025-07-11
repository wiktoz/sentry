package db

import (
	"database/sql"

	"github.com/wiktoz/sentry/models"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

const Schema = `
DROP TABLE IF EXISTS vulnerabilities;
DROP TABLE IF EXISTS ports;
DROP TABLE IF EXISTS hosts;
DROP TABLE IF EXISTS scans;
DROP TABLE IF EXISTS config;

CREATE TABLE IF NOT EXISTS scans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME
);

CREATE TABLE IF NOT EXISTS hosts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id INTEGER,
    address TEXT,
    addr_type TEXT,
    FOREIGN KEY(scan_id) REFERENCES scans(id)
);

CREATE TABLE IF NOT EXISTS ports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    host_id INTEGER,
    protocol TEXT,
    port_id INTEGER,
    state TEXT,
    service_name TEXT,
    FOREIGN KEY(host_id) REFERENCES hosts(id)
);

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    scan_frequency INTEGER NOT NULL,
    email TEXT NOT NULL,
    scan_targets TEXT NOT NULL
);

INSERT INTO config (id, scan_frequency, email, scan_targets)
VALUES (1, 300, 'admin@example.com', '192.168.1.0/24');

CREATE TABLE vulnerabilities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    port_id INTEGER NOT NULL,
    vuln_id TEXT NOT NULL,
    score REAL,
    url TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (port_id) REFERENCES ports(id) ON DELETE CASCADE
);
`

func GetConfig(db *sql.DB) (models.Config, error) {
	var cfg models.Config
	err := db.QueryRow("SELECT scan_frequency, email, scan_targets FROM config WHERE id = 1").
		Scan(&cfg.ScanFrequency, &cfg.Email, &cfg.ScanTarget)
	return cfg, err
}

func SaveConfig(db *sql.DB, cfg models.Config) error {
	_, err := db.Exec(`
		UPDATE config SET scan_frequency = ?, email = ?, scan_targets = ? WHERE id = 1
	`, cfg.ScanFrequency, cfg.Email, cfg.ScanTarget)
	return err
}
