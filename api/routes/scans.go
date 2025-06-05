package routes

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/helpers"
	"github.com/wiktoz/sentry/models"
	"github.com/wiktoz/sentry/scripts"
)

func GetScanById(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/scan-get/")
	scanID, err := strconv.Atoi(idStr)

	if err != nil || scanID <= 0 {
		http.Error(w, "invalid scan id", http.StatusBadRequest)
		return
	}

	scanData, err := getScanData(scanID)
	if err == sql.ErrNoRows {
		http.Error(w, "scan not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	resp := models.ScanResponse{Scan: scanData}
	helpers.WriteJSON(w, resp)
}

func GetScans(w http.ResponseWriter, r *http.Request) {
	// Query all scan IDs ordered by date desc
	rows, err := db.DB.Query("SELECT id FROM scans ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var scans []models.ScanData
	for rows.Next() {
		var scanID int
		if err := rows.Scan(&scanID); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		scanData, err := getScanData(scanID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		scans = append(scans, scanData)
	}

	helpers.WriteJSON(w, map[string]any{"scans": scans})
}

func RunScan(w http.ResponseWriter, r *http.Request) {
	go scripts.RunScripts()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"scripts started"}`))
}

func getScanData(scanID int) (models.ScanData, error) {
	var createdAt string
	err := db.DB.QueryRow("SELECT created_at FROM scans WHERE id = ?", scanID).Scan(&createdAt)
	if err != nil {
		return models.ScanData{}, err
	}

	hostRows, err := db.DB.Query("SELECT id, address FROM hosts WHERE scan_id = ?", scanID)
	if err != nil {
		return models.ScanData{}, err
	}
	defer hostRows.Close()

	var hosts []models.HostData

	for hostRows.Next() {
		var hostID int
		var address string
		if err := hostRows.Scan(&hostID, &address); err != nil {
			return models.ScanData{}, err
		}

		portRows, err := db.DB.Query("SELECT port_id, service_name, protocol, state FROM ports WHERE host_id = ?", hostID)
		if err != nil {
			return models.ScanData{}, err
		}

		var ports []models.PortData
		for portRows.Next() {
			var p models.PortData
			if err := portRows.Scan(&p.PortNum, &p.ServiceName, &p.Protocol, &p.State); err != nil {
				portRows.Close()
				return models.ScanData{}, err
			}
			ports = append(ports, p)
		}
		portRows.Close()

		hosts = append(hosts, models.HostData{
			Address: address,
			Ports:   ports,
		})
	}

	return models.ScanData{
		Date:  createdAt,
		Hosts: hosts,
	}, nil
}
