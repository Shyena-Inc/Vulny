package cmd

import (
	"flag"
	"strings"
)

// Args holds command-line arguments
type Args struct {
	Help      bool
	Update    bool
	Skip      []string
	NoSpinner bool
	Target    string
}

// ParseArgs parses command-line arguments
func ParseArgs() Args {
	help := flag.Bool("help", false, "Show help message and exit")
	update := flag.Bool("update", false, "Update Vulny")
	skip := flag.String("skip", "", "Comma-separated list of tools to skip")
	nospinner := flag.Bool("nospinner", false, "Disable the idle loader/spinner")
	target := flag.String("target", "", "URL to scan")
	flag.Parse()

	return Args{
		Help:      *help,
		Update:    *update,
		Skip:      strings.Split(*skip, ","),
		NoSpinner: *nospinner,
		Target:    *target,
	}
}