package scripts

import (
	"bytes"
	"context"
	"encoding/xml"
	"html"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/wiktoz/sentry/db"
	"github.com/wiktoz/sentry/models"
)

type NmapRun struct {
	Hosts []Host `xml:"host"`
}

type Host struct {
	Addresses []Address `xml:"address"`
	Ports     Ports     `xml:"ports"`
}

type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type Ports struct {
	Port []Port `xml:"port"`
}

type Port struct {
	Protocol string   `xml:"protocol,attr"`
	PortID   int      `xml:"portid,attr"`
	State    State    `xml:"state"`
	Service  Service  `xml:"service"`
	Scripts  []Script `xml:"script"`
}

type State struct {
	State string `xml:"state,attr"`
}

type Service struct {
	Name string `xml:"name,attr"`
}

type Script struct {
	ID     string `xml:"id,attr"`
	Output string `xml:"output,attr"`
}

type Vulnerability struct {
	VulnID      string
	Score       float64
	URL         string
	Description string
}

func RunFullScan(scanID int, target string) {
	hosts, err := RunNormalScan(target, scanID)
	if err != nil {
		log.Fatalf("Normal scan failed: %v", err)
	}

	err = RunVulnScan(hosts)
	if err != nil {
		log.Fatalf("Vulnerability scan failed: %v", err)
	}

	log.Println("Scan completed successfully")
}

func RunNormalScan(target string, scanID int) ([]Host, error) {
	cmd := exec.Command("nmap", "-p-", "-T4", "--host-timeout 30s", "-oX", "-", target)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var nmapRun NmapRun
	if err := xml.NewDecoder(bytes.NewReader(output)).Decode(&nmapRun); err != nil {
		return nil, err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	for _, host := range nmapRun.Hosts {
		for _, addr := range host.Addresses {
			res, err := tx.Exec("INSERT INTO hosts (scan_id, address, addr_type) VALUES (?, ?, ?)", scanID, addr.Addr, addr.AddrType)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			hostID, _ := res.LastInsertId()

			for _, port := range host.Ports.Port {
				_, err := tx.Exec(
					`INSERT INTO ports (host_id, protocol, port_id, state, service_name)
					 VALUES (?, ?, ?, ?, ?)`,
					hostID, port.Protocol, port.PortID, port.State.State, port.Service.Name,
				)
				if err != nil {
					_ = tx.Rollback()
					return nil, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return nmapRun.Hosts, nil
}

func RunVulnScan(hosts []Host) error {
	var targets []string
	for _, host := range hosts {
		for _, addr := range host.Addresses {
			var ports []string
			for _, port := range host.Ports.Port {
				if port.State.State == "open" {
					ports = append(ports, strconv.Itoa(port.PortID))
				}
			}
			if len(ports) == 0 {
				continue
			}
			target := addr.Addr + " -p " + strings.Join(ports, ",")
			targets = append(targets, target)
		}
	}

	if len(targets) == 0 {
		return nil
	}

	args := []string{"-sV", "--script", "vulners", "-T4", "--max-retries", "5", "--host-timeout", "30s", "-oX", "-"}
	args = append(args, targets...)

	cmd := exec.Command("nmap", args...)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var nmapRun NmapRun
	if err := xml.NewDecoder(bytes.NewReader(output)).Decode(&nmapRun); err != nil {
		return err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	for _, host := range nmapRun.Hosts {
		for _, addr := range host.Addresses {
			var hostID int64
			err := tx.QueryRow("SELECT id FROM hosts WHERE address = ?", addr.Addr).Scan(&hostID)
			if err != nil {
				_ = tx.Rollback()
				return err
			}

			for _, port := range host.Ports.Port {
				var portID int64
				err := tx.QueryRow("SELECT id FROM ports WHERE host_id = ? AND port_id = ?", hostID, port.PortID).Scan(&portID)
				if err != nil {
					_ = tx.Rollback()
					return err
				}

				for _, script := range port.Scripts {
					if script.ID == "vulners" {
						vulns := ParseVulnersOutput(script.Output)
						for _, vuln := range vulns {
							_, err := tx.Exec(
								`INSERT INTO vulnerabilities (port_id, vuln_id, score, url, description)
								 VALUES (?, ?, ?, ?, ?)`,
								portID, vuln.VulnID, vuln.Score, vuln.URL, vuln.Description,
							)
							if err != nil {
								_ = tx.Rollback()
								return err
							}
						}
					}
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func ParseVulnersOutput(rawOutput string) []Vulnerability {
	cleaned := html.UnescapeString(rawOutput)
	var vulns []Vulnerability

	lines := strings.Split(cleaned, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "cpe:") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}

		vulnID := fields[0]
		score, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}
		url := fields[2]

		desc := ""
		if len(fields) > 3 {
			desc = fields[3]
		}

		vulns = append(vulns, Vulnerability{
			VulnID:      vulnID,
			Score:       score,
			URL:         url,
			Description: desc,
		})
	}
	return vulns
}

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

func StartAutoScan(ctx context.Context) {
	go func() {
		var freq time.Duration
		var ticker *time.Ticker

		setTicker := func() {
			cfg, err := db.GetConfig(db.DB)
			if err != nil {
				log.Printf("Error getting scan config: %v", err)
				cfg.ScanFrequency = 300 // fallback in seconds
			}

			newFreq := time.Duration(cfg.ScanFrequency) * time.Second

			if freq != newFreq {
				freq = newFreq
				if ticker != nil {
					ticker.Stop()
				}
				ticker = time.NewTicker(freq)
				log.Printf("Scan frequency set to %v", freq)
			}
		}

		setTicker() // initial

		checkFreqTicker := time.NewTicker(30 * time.Second)
		defer checkFreqTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				if ticker != nil {
					ticker.Stop()
				}
				return
			case <-checkFreqTicker.C:
				setTicker()
			case <-ticker.C:
				log.Println("Auto scan triggered")
				RunScripts()
			}
		}
	}()
}
