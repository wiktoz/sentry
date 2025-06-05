package models

// Data structures for JSON responses

type ScanData struct {
	Date  string     `json:"date"`
	Hosts []HostData `json:"hosts"`
}

type HostData struct {
	Address string     `json:"address"`
	Ports   []PortData `json:"ports"`
}

type PortData struct {
	PortNum     int    `json:"port_num"`
	ServiceName string `json:"service_name"`
	Protocol    string `json:"protocol"`
	State       string `json:"state"`
}
