package types

// Scraper Arg types
type ScrapeArgs struct {
	Concurrency int
	Ports       []string
	Timeout     int
	PortList    string
	Help        bool
	Input       string
	Debug       bool
	OutputFile  string
}

// Result Types
type CertificateInfo struct {
	IP           string `json:"ip"`
	Organization string `json:"organization"`
	CommonName   string `json:"commonName"`
	SAN          string `json:"san"`
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
