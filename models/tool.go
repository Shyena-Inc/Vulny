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
)

// Tools is a slice of defined tools (extend as needed)
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
			ProcLevel:    Green("●"), // Low
			EstTime:      "< 15s",
			ID:           "ipv6",
			BadResponses: []string{"not found", "has IPv6"},
		},
	},
	// Add more tools here
}

// VulInfo formats severity level
func VulInfo(severity string) string {
	switch severity {
	case "c":
		return Red(" critical ")
	case "h":
		return Red(" high ")
	case "m":
		return Yellow(" medium ")
	case "l":
		return Blue(" low ")
	default:
		return Green(" info ")
	}
}