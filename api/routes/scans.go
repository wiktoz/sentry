package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/helpers"
	"github.com/wiktoz/sentry/models"
	"github.com/wiktoz/sentry/scripts"
)

func GetScanById(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/scan/")
	scanID, err := strconv.Atoi(idStr)

	if err != nil || scanID <= 0 {
		http.Error(w, "Invalid scan ID", http.StatusBadRequest)
		return
	}

	scanData, err := getScanData(scanID)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, scanData)
}

func GetScans(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id FROM scans ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var scans []models.ScanData

	for rows.Next() {
		var scanID int
		if err := rows.Scan(&scanID); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		scanData, err := getScanData(scanID)
		if err != nil {
			// Log error and skip problematic scans instead of failing the whole response
			log.Printf("error loading scan ID %d: %v", scanID, err)
			continue
		}

		scans = append(scans, scanData)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, scans)
}

func RunScan(w http.ResponseWriter, r *http.Request) {
	// Insert new scan record with current timestamp
	res, err := db.DB.Exec("INSERT INTO scans (created_at) VALUES (datetime('now'))")
	if err != nil {
		http.Error(w, "failed to create scan", http.StatusInternalServerError)
		return
	}

	scanID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "failed to retrieve scan ID", http.StatusInternalServerError)
		return
	}

	cfg, err := db.GetConfig(db.DB)
	if err != nil {
		http.Error(w, "Error getting scan config", http.StatusInternalServerError)
		return
	}

	targets := cfg.ScanTarget
	if targets == "" {
		http.Error(w, "No scan targets", http.StatusInternalServerError)
		return
	}

	// Start the scan in background, passing scanID as int
	go scripts.RunFullScan(int(scanID), targets)

	// Return the scan ID immediately as JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"status":"scan started", "scan_id": %d}`, scanID)))
}

func getScanData(scanID int) (models.ScanData, error) {
	var scan models.ScanData

	// Fetch scan metadata
	err := db.DB.QueryRow(`SELECT id, created_at FROM scans WHERE id = ?`, scanID).Scan(&scan.ID, &scan.Date)
	if err != nil {
		return models.ScanData{}, err
	}

	// Fetch hosts for the scan
	hostRows, err := db.DB.Query(`SELECT id, address FROM hosts WHERE scan_id = ?`, scanID)
	if err != nil {
		return models.ScanData{}, err
	}
	defer hostRows.Close()

	for hostRows.Next() {
		var host models.HostData
		var hostID int

		if err := hostRows.Scan(&hostID, &host.Address); err != nil {
			return models.ScanData{}, err
		}

		ports, err := fetchPortsWithVulns(hostID)
		if err != nil {
			return models.ScanData{}, err
		}

		host.Ports = ports
		scan.Hosts = append(scan.Hosts, host)
	}

	if err := hostRows.Err(); err != nil {
		return models.ScanData{}, err
	}

	return scan, nil
}

func fetchPortsWithVulns(hostID int) ([]models.PortData, error) {
	portRows, err := db.DB.Query(`
		SELECT id, port_id, service_name, protocol, state 
		FROM ports 
		WHERE host_id = ?`, hostID)
	if err != nil {
		return nil, err
	}
	defer portRows.Close()

	var ports []models.PortData

	for portRows.Next() {
		var p models.PortData
		var portID int

		if err := portRows.Scan(&portID, &p.PortNum, &p.ServiceName, &p.Protocol, &p.State); err != nil {
			return nil, err
		}

		vulnRows, err := db.DB.Query(`
			SELECT vuln_id, description, score, url 
			FROM vulnerabilities 
			WHERE port_id = ?`, portID)
		if err != nil {
			return nil, err
		}

		for vulnRows.Next() {
			var v models.VulnerabilityData
			if err := vulnRows.Scan(&v.CVE, &v.Description, &v.Score, &v.URL); err != nil {
				vulnRows.Close()
				return nil, err
			}
			p.Vulnerabilities = append(p.Vulnerabilities, v)
		}
		vulnRows.Close()

		ports = append(ports, p)
	}

	if err := portRows.Err(); err != nil {
		return nil, err
	}

	return ports, nil
}
