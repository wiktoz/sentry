package scripts

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wiktoz/sentry/db"
)

type NmapRun struct {
	Hosts []Host `xml:"host"`
}

type Host struct {
	TimedOut  bool      `xml:"timedout,attr"` // <-- added to detect timed-out hosts
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

func filterTargets(targets string) string {
	// MAC address regex (allow part of the string)
	macRegex := regexp.MustCompile(`(?i)^([0-9A-F]{2}:){5}[0-9A-F]{2}$`)

	parts := strings.Fields(targets)
	var filtered []string

	for _, part := range parts {
		if macRegex.MatchString(part) {
			log.Printf("Skipping MAC address in targets: %s", part)
			continue
		}
		filtered = append(filtered, part)
	}

	return strings.Join(filtered, " ")
}

func RunFullScan(scanID int, target string) {
	hosts, err := RunNormalScan(target, scanID)
	if err != nil {
		log.Fatalf("Normal scan failed: %v", err)
	}

	err = RunVulnScan(hosts, scanID)
	if err != nil {
		log.Fatalf("Vulnerability scan failed: %v", err)
	}

	log.Println("Scan completed successfully")
}

func RunNormalScan(target string, scanID int) ([]Host, error) {
	log.Println("Normal Scan started")

	cmd := exec.Command("nmap", "--host-timeout", "30s", "-oX", "-", target)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	log.Println("Reading output")
	var nmapRun NmapRun
	if err := xml.NewDecoder(bytes.NewReader(output)).Decode(&nmapRun); err != nil {
		return nil, err
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	var filteredHosts []Host

	for _, host := range nmapRun.Hosts {
		var ipv4Addr string
		for _, addr := range host.Addresses {
			if addr.AddrType == "ipv4" {
				ipv4Addr = addr.Addr
				break
			}
		}

		if ipv4Addr == "" {
			// Skip hosts without an IPv4 address
			continue
		}

		res, err := tx.Exec(
			"INSERT INTO hosts (scan_id, address, addr_type) VALUES (?, ?, ?)",
			scanID, ipv4Addr, "ipv4",
		)
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

		// Append only hosts with IPv4 to return list
		filteredHosts = append(filteredHosts, host)
	}

	log.Println("Committing")

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return filteredHosts, nil
}

func RunVulnScan(hosts []Host, scanID int) error {
	for _, host := range hosts {
		for _, addr := range host.Addresses {
			if addr.AddrType != "ipv4" {
				continue
			}

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

			args := []string{"-sV", "--script", "vulners", "--host-timeout", "30s", "-oX", "-"}
			args = append(args, strings.Split(target, " ")...)

			log.Println("Running nmap for:", addr.Addr)

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

			var emailBody strings.Builder
			totalVulns := 0

			emailBody.WriteString("<html><body style=\"font-family: sans-serif;\">")

			for _, scannedHost := range nmapRun.Hosts {
				if scannedHost.TimedOut {
					log.Printf("Skipping timed-out host: %+v\n", scannedHost.Addresses)
					continue
				}

				for _, scannedAddr := range scannedHost.Addresses {
					var hostID int64
					err := tx.QueryRow("SELECT id FROM hosts WHERE address = ? AND scan_id = ?", scannedAddr.Addr, scanID).Scan(&hostID)
					if err != nil {
						if err == sql.ErrNoRows {
							log.Printf("Host not found in DB, skipping: %s", scannedAddr.Addr)
							continue
						}
						_ = tx.Rollback()
						return err
					}

					var hostVulnCount int
					hostHeaderWritten := false

					for _, scannedPort := range scannedHost.Ports.Port {
						var portID int64
						err := tx.QueryRow("SELECT id FROM ports WHERE host_id = ? AND port_id = ?", hostID, scannedPort.PortID).Scan(&portID)
						if err != nil {
							if err == sql.ErrNoRows {
								log.Printf("Port not found in DB for host %d, port %d - skipping", hostID, scannedPort.PortID)
								continue
							}
							_ = tx.Rollback()
							return err
						}

						for _, script := range scannedPort.Scripts {
							if script.ID == "vulners" {
								vulns := ParseVulnersOutput(script.Output)
								if len(vulns) > 0 {
									if !hostHeaderWritten {
										emailBody.WriteString(fmt.Sprintf(`<h2>Host: <b>%s</b></h2>`, scannedAddr.Addr))
										hostHeaderWritten = true
									}

									hostVulnCount += len(vulns)
									emailBody.WriteString(fmt.Sprintf(`<p><b>Port: %d</b> - Showing up to 5 vulnerabilities:</p><ul>`, scannedPort.PortID))

									for i, vuln := range vulns {
										if i >= 5 {
											break
										}
										emailBody.WriteString(fmt.Sprintf(
											`<li><span style="color:#d9534f;font-weight:bold;"><a href="%s" style="color:#d9534f;text-decoration:none;">%s</a></span> (Score: %.1f)</li>`,
											vuln.URL, vuln.VulnID, vuln.Score,
										))
									}

									emailBody.WriteString("</ul>")
								}

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

					if hostVulnCount > 0 {
						totalVulns += hostVulnCount
						emailBody.WriteString(fmt.Sprintf(`<p><i>Total vulnerabilities on this host: <b>%d</b></i></p>`, hostVulnCount))
					}
				}
			}

			emailBody.WriteString("</body></html>")

			if err := tx.Commit(); err != nil {
				return err
			}

			if totalVulns > 0 {
				subject := fmt.Sprintf("Vulnerabilities detected on %s (%d total)", addr.Addr, totalVulns)
				if err := SendEmail(subject, emailBody.String()); err != nil {
					log.Printf("Failed to send email: %v", err)
				} else {
					log.Printf("Email sent for host %s", addr.Addr)
				}
			}
		}
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

func StartAutoScan(ctx context.Context) {
	go func() {
		var freq time.Duration
		var ticker *time.Ticker

		// Fetch config and update ticker if freq changed
		updateTicker := func() error {
			cfg, err := db.GetConfig(db.DB)
			if err != nil {
				log.Printf("Error getting scan config: %v", err)
				return err
			}

			newFreq := time.Duration(cfg.ScanFrequency) * time.Second
			if newFreq == freq {
				// no change
				return nil
			}

			// freq changed
			if ticker != nil {
				ticker.Stop()
				ticker = nil
			}

			if cfg.ScanFrequency == 0 {
				freq = 0
				log.Println("Scan frequency is 0, auto scanning disabled.")
				return nil
			}

			freq = newFreq
			ticker = time.NewTicker(freq)
			log.Printf("Scan frequency updated to every %v seconds", freq.Seconds())
			return nil
		}

		// Init ticker once at start
		err := updateTicker()
		if err != nil {
			// on error, fallback freq and ticker disabled
			freq = 0
			ticker = nil
		}

		for {
			if ticker == nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Second):
					_ = updateTicker()
					continue
				}
			}

			select {
			case <-ticker.C:
				cfg, err := db.GetConfig(db.DB)
				if err != nil {
					log.Printf("Error getting scan config: %v", err)
					continue
				}

				if cfg.ScanFrequency == 0 {
					if ticker != nil {
						ticker.Stop()
						ticker = nil
						freq = 0
						log.Println("Scan frequency changed to 0, stopping scans.")
					}
					continue
				}

				targets := cfg.ScanTarget
				if targets == "" {
					log.Println("No targets configured, skipping scan.")
					continue
				}

				res, err := db.DB.Exec("INSERT INTO scans (created_at) VALUES (datetime('now'))")
				if err != nil {
					log.Println("Can't create scan record, skipping scan:", err)
					continue
				}

				scanID, err := res.LastInsertId()
				if err != nil {
					log.Println("No scan ID returned, skipping scan.")
					continue
				}

				RunFullScan(int(scanID), targets)
				_ = updateTicker()

			case <-ctx.Done():
				if ticker != nil {
					ticker.Stop()
				}
				return
			}
		}
	}()
}
