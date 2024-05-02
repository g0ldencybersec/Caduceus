package types

import "net"

// Scraper Arg types
type ScrapeArgs struct {
	Concurrency    int
	Ports          []string
	Timeout        int
	PortList       string
	Help           bool
	Input          string
	Debug          bool
	JsonOutput     bool
	PrintWildcards bool
	PrintStats     bool
}

// Result Types
type CertificateInfo struct {
	OriginIP         string   `json:"originip"`
	Organization     []string `json:"org"`
	OrganizationUnit []string `json:"orgunit"`
	CommonName       string   `json:"commonName"`
	SAN              []string `json:"san"`
	Domains          []string `json:"domains"`
	Emails           []string `json:"emails"`
	IPAddrs          []net.IP `json:"ips"`
}

type Result struct {
	IP          string
	Hit         bool
	Timeout     bool
	Error       error
	Certificate *CertificateInfo
}

// Stats Types
type Stats struct {
	hits   int
	misses int
	total  int
}
