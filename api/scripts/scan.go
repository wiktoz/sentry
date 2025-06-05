package scripts

import (
	"bytes"
	"encoding/xml"
	"log"
	"os/exec"
	"time"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/models"
)

func RunScripts() {
	cmd := exec.Command("nmap", "-sV", "-oX", "-", "127.0.0.1")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Script failed: %v\nOutput: %s", err, output)
		return
	}

	var nmapRun models.NmapRun
	err = xml.NewDecoder(bytes.NewReader(output)).Decode(&nmapRun)
	if err != nil {
		log.Printf("Failed to parse nmap XML: %v", err)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	res, err := tx.Exec("INSERT INTO scans (created_at) VALUES (?)", time.Now())
	if err != nil {
		log.Printf("Failed to insert scan meta: %v", err)
		_ = tx.Rollback()
		return
	}

	scanID, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to get scan ID: %v", err)
		_ = tx.Rollback()
		return
	}

	for _, host := range nmapRun.Hosts {
		for _, addr := range host.Addresses {
			res, err := tx.Exec("INSERT INTO hosts (scan_id, address, addr_type) VALUES (?, ?, ?)", scanID, addr.Addr, addr.AddrType)
			if err != nil {
				log.Printf("Failed to insert host: %v", err)
				_ = tx.Rollback()
				return
			}
			hostID, err := res.LastInsertId()
			if err != nil {
				log.Printf("Failed to get host ID: %v", err)
				_ = tx.Rollback()
				return
			}

			for _, port := range host.Ports.Port {
				_, err := tx.Exec(
					`INSERT INTO ports (host_id, protocol, port_id, state, service_name) 
					 VALUES (?, ?, ?, ?, ?)`,
					hostID, port.Protocol, port.PortId, port.State.State, port.Service.Name)
				if err != nil {
					log.Printf("Failed to insert port: %v", err)
					_ = tx.Rollback()
					return
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	}
}
