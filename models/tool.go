package models

import "github.com/fatih/color"

// Tool represents a scanning tool
type Tool struct {
	Name        string
	Description string
	Command     string
	Enabled     bool
	Resp        ToolResponse
	Status      ToolStatus
}

// ToolResponse defines the vulnerability response
type ToolResponse struct {
	Message    string
	Severity   string
	RemedIndex int
}

// ToolStatus defines the tool's status check
type ToolStatus struct {
	CheckString  string
	Inverted     bool
	ProcLevel    string
	EstTime      string
	ID           string
	BadResponses []string
}

// Colors for terminal output
var (
	Red    = color.New(color.FgRed).SprintFunc()
	Green  = color.New(color.FgGreen).SprintFunc()
	Blue   = color.New(color.FgBlue).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Reset  = color.New(color.Reset).SprintFunc()
)

// Tools is a slice of defined tools
var Tools = []Tool{
	{
		Name:        "host",
		Description: "Host - Checks for existence of IPV6 address",
		Command:     "host %s",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Does not have an IPv6 Address. It is good to have one.",
			Severity:   "i",
			RemedIndex: 1,
		},
		Status: ToolStatus{
			CheckString:  "has IPv6",
			Inverted:     true,
			ProcLevel:    Green("●"),
			EstTime:      "< 15s",
			ID:           "ipv6",
			BadResponses: []string{"not found", "has IPv6"},
		},
	},
	{
		Name:        "nmap",
		Description: "Nmap - Full TCP port scan with service and OS detection",
		Command:     "nmap -p- -sS -sV -O -Pn %s",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Full port scan complete. Analyze detected services and possible misconfigurations.",
			Severity:   "m",
			RemedIndex: 4,
		},
		Status: ToolStatus{
			CheckString:  "open",
			Inverted:     false,
			ProcLevel:    Yellow("●"),
			EstTime:      "≈ 5-10m",
			ID:           "nmapfull",
			BadResponses: []string{"Failed to resolve", "Host seems down"},
		},
	},
	
	{
		Name:        "wpscan",
		Description: "WPScan - Checks for WordPress vulnerabilities",
		Command:     "wpscan --url https://%s --no-banner --disable-tls-checks",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "WordPress vulnerabilities or misconfigurations detected.",
			Severity:   "m",
			RemedIndex: 3,
		},
		Status: ToolStatus{
			CheckString:  "Vulnerable",
			Inverted:     false,
			ProcLevel:    Yellow("●"),
			EstTime:      "< 5m",
			ID:           "wpscan",
			BadResponses: []string{"not a WordPress"},
		},
	},
	{
		Name:        "joomscan",
		Description: "JoomScan - Checks for Joomla vulnerabilities",
		Command:     "joomscan -u https://%s --silent",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Joomla vulnerabilities or misconfigurations detected.",
			Severity:   "m",
			RemedIndex: 4,
		},
		Status: ToolStatus{
			CheckString:  "Vulnerability",
			Inverted:     false,
			ProcLevel:    Yellow("●"),
			EstTime:      "< 5m",
			ID:           "joomscan",
			BadResponses: []string{"not a Joomla"},
		},
	},
	{
		Name:        "droopescan",
		Description: "Droopescan - Checks for Drupal vulnerabilities",
		Command:     "droopescan scan drupal -u https://%s",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Drupal vulnerabilities or misconfigurations detected.",
			Severity:   "m",
			RemedIndex: 5,
		},
		Status: ToolStatus{
			CheckString:  "vulnerability",
			Inverted:     false,
			ProcLevel:    Yellow("●"),
			EstTime:      "< 5m",
			ID:           "droopescan",
			BadResponses: []string{"not a Drupal"},
		},
	},
	{
		Name:        "sslscan",
		Description: "SSLScan - Checks for SSL vulnerabilities (HEARTBLEED, POODLE)",
		Command:     "sslscan %s:443",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "SSL vulnerabilities detected (e.g., HEARTBLEED, POODLE).",
			Severity:   "h",
			RemedIndex: 6,
		},
		Status: ToolStatus{
			CheckString:  "VULNERABLE",
			Inverted:     false,
			ProcLevel:    Red("●"),
			EstTime:      "< 1m",
			ID:           "sslscan",
			BadResponses: []string{"Connection refused"},
		},
	},
	{
		Name:        "amass",
		Description: "Amass - Sub-domain brute-forcing",
		Command:     "amass enum -d %s -timeout 5",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Sub-domains discovered, potential for further reconnaissance.",
			Severity:   "i",
			RemedIndex: 7,
		},
		Status: ToolStatus{
			CheckString:  "Subdomain Name",
			Inverted:     false,
			ProcLevel:    Blue("●"),
			EstTime:      "< 10m",
			ID:           "amass",
			BadResponses: []string{"No subdomains found"},
		},
	},
	{
		Name:        "nikto",
		Description: "Nikto - Checks for XSS, SQLi, and server misconfigurations",
		Command:     "nikto -h https://%s -Tuning x",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Potential XSS, SQLi, or server misconfigurations detected.",
			Severity:   "m",
			RemedIndex: 8,
		},
		Status: ToolStatus{
			CheckString:  "vulnerability",
			Inverted:     false,
			ProcLevel:    Yellow("●"),
			EstTime:      "< 5m",
			ID:           "nikto",
			BadResponses: []string{"0 items checked"},
		},
	},
	// Placeholder for DNS Zone Transfers (e.g., Fierce)
	{
		Name:        "fierce",
		Description: "Fierce - DNS Zone Transfer (Placeholder)",
		Command:     "fierce -dns %s",
		Enabled:     false,
		Resp: ToolResponse{
			Message:    "DNS Zone Transfer possible, exposing internal records.",
			Severity:   "h",
			RemedIndex: 9,
		},
		Status: ToolStatus{
			CheckString:  "Zone transfer",
			Inverted:     false,
			ProcLevel:    Red("●"),
			EstTime:      "< 1m",
			ID:           "fierce",
			BadResponses: []string{"transfer failed"},
		},
	},
	// Placeholder for Slow-Loris DoS
	{
		Name:        "slowloris",
		Description: "Slow-Loris - DoS Attack (Placeholder)",
		Command:     "slowloris %s",
		Enabled:     false,
		Resp: ToolResponse{
			Message:    "Server vulnerable to Slow-Loris DoS attack.",
			Severity:   "h",
			RemedIndex: 10,
		},
		Status: ToolStatus{
			CheckString:  "connection timeout",
			Inverted:     false,
			ProcLevel:    Red("●"),
			EstTime:      "< 5m",
			ID:           "slowloris",
			BadResponses: []string{"connection refused"},
		},
	},
	{
		Name:        "dalfox",
		Description: "Dalfox - Fast and powerful XSS scanning tool",
		Command:     "dalfox url %s --skip-bav --only-poc",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Dalfox detected potential reflected/stored XSS vulnerabilities. Manual verification recommended.",
			Severity:   "h",
			RemedIndex: 6,
		},
		Status: ToolStatus{
			CheckString:  "[POC]",
			Inverted:     false,
			ProcLevel:    Red("●"),
			EstTime:      "< 2m",
			ID:           "dalfox-xss",
			BadResponses: []string{"Error", "no target"},
		},
	},
	
	{
		Name:        "wafw00f",
		Description: "WAFW00F - Web Application Firewall detection",
		Command:     "wafw00f https://%s",
		Enabled:     true,
		Resp: ToolResponse{
			Message:    "Web Application Firewall detected, may block attacks.",
			Severity:   "i",
			RemedIndex: 14,
		},
		Status: ToolStatus{
			CheckString:  "WAF",
			Inverted:     false,
			ProcLevel:    Blue("●"),
			EstTime:      "< 1m",
			ID:           "wafw00f",
			BadResponses: []string{"No WAF"},
		},
	},
}

// VulInfo formats severity level
func VulInfo(severity string) string {
	switch severity {
	case "c":
		return Red("critical") + Reset()
	case "h":
		return Red("high") + Reset()
	case "m":
		return Yellow("medium") + Reset()
	case "l":
		return Blue("low") + Reset()
	default:
		return Green("info") + Reset()
	}
}