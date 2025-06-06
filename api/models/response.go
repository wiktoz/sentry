package models

// Data structures for JSON responses

type ScanData struct {
	ID    int        `json:"id"`
	Date  string     `json:"date"`
	Hosts []HostData `json:"hosts"`
}

type HostData struct {
	Address string     `json:"address"`
	Ports   []PortData `json:"ports"`
}

type VulnerabilityData struct {
	CVE         string  `json:"cve"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	URL         string  `json:"url"`
}

type PortData struct {
	PortNum         int                 `json:"port_num"`
	ServiceName     string              `json:"service_name"`
	Protocol        string              `json:"protocol"`
	State           string              `json:"state"`
	Vulnerabilities []VulnerabilityData `json:"vulnerabilities"`
}

type Config struct {
	ScanFrequency int    `json:"scan_frequency"`
	Email         string `json:"email"`
	ScanTarget    string `json:"target"`
}
