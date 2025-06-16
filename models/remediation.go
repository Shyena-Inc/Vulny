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
	// Add more remediations here
}