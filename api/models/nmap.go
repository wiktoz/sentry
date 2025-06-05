package models

import (
	"encoding/xml"
)

type NmapRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts   []Host   `xml:"host"`
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
	Protocol string  `xml:"protocol,attr"`
	PortId   int     `xml:"portid,attr"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}

type State struct {
	State string `xml:"state,attr"`
}

type Service struct {
	Name string `xml:"name,attr"`
}
