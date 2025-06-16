package models

// Remediation holds vulnerability remediation info
type Remediation struct {
	Index       int
	Description string
	Remediation string
}

// Remediations is a slice of remediation info
var Remediations = []Remediation{
	{
		Index:       1,
		Description: "Not a vulnerability, just an informational alert. The host does not have IPv6 support...",
		Remediation: "It is recommended to implement IPv6. More information on how to implement IPv6 can be found from this resource...",
	},
	{
		Index:       2,
		Description: "Commonly opened ports detected by Nmap scan.",
		Remediation: "Review open ports and close unnecessary ones. Use a firewall to restrict access to sensitive services. Perform a detailed Nmap scan for further analysis.",
	},
	{
		Index:       3,
		Description: "WordPress vulnerabilities or misconfigurations detected by WPScan.",
		Remediation: "Update WordPress core, plugins, and themes to the latest versions. Disable unused plugins and enforce strong passwords. Review WPScan report for specific issues.",
	},
	{
		Index:       4,
		Description: "Joomla vulnerabilities or misconfigurations detected by JoomScan.",
		Remediation: "Update Joomla core and extensions. Harden configuration by disabling unused components and securing admin access. Review JoomScan report for details.",
	},
	{
		Index:       5,
		Description: "Drupal vulnerabilities or misconfigurations detected by Droopescan.",
		Remediation: "Update Drupal core and modules. Apply security patches and restrict file permissions. Review Droopescan report for specific vulnerabilities.",
	},
	{
		Index:       6,
		Description: "SSL vulnerabilities detected (e.g., HEARTBLEED, POODLE).",
		Remediation: "Update OpenSSL to the latest version. Disable SSLv3 and weak ciphers. Enable TLS 1.2 or higher. Retest with sslscan or testssl.sh.",
	},
	{
		Index:       7,
		Description: "Sub-domains discovered by Amass, increasing attack surface.",
		Remediation: "Review discovered sub-domains for misconfigurations or unintended exposure. Secure or remove unnecessary sub-domains.",
	},
	{
		Index:       8,
		Description: "Potential XSS, SQLi, or server misconfigurations detected by Nikto.",
		Remediation: "Sanitize user inputs to prevent XSS and SQLi. Update server software and apply secure configurations. Review Nikto report for specific issues.",
	},
	{
		Index:       9,
		Description: "DNS Zone Transfer possible, exposing internal DNS records.",
		Remediation: "Restrict DNS Zone Transfers to authorized servers only. Configure DNS server to deny unauthorized AXFR requests.",
	},
	{
		Index:       10,
		Description: "Server vulnerable to Slow-Loris DoS attack.",
		Remediation: "Implement rate-limiting and connection timeouts on the web server. Use a WAF or load balancer to mitigate DoS attacks.",
	},
}