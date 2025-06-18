package utils

import (
	"github.com/fatih/color"
)

func PrintBanner() {
	c := color.New(color.FgHiGreen).Add(color.Bold)

	banner := `
              _             
 /\   /\_   _| |_ __  _   _ 
 \ \ / / | | | | '_ \| | | |
  \ V /| |_| | | | | | |_| |
   \_/  \__,_|_|_| |_|\__, |
                      |___/`

	c.Println(banner)
}
func PrintVersion() {
	c := color.New(color.FgHiBlue).Add(color.Bold)
	version := "Vulny v1.0.0"
	c.Println(version)
}
func PrintHelp() {
	c := color.New(color.FgHiYellow).Add(color.Bold)
	helpText := `
Usage: vulny [options]
Options:
  -target <url>       Target URL to scan (e.g., example.com)
  -skip <tools>       Comma-separated list of tools to skip (e.g., host,nmap)
  -update             Check for updates and update the tool
  -no-spinner         Disable loading spinner during scan
  -help               Show this help message
Examples:
  vulny -target example.com
  vulny -target example.com -skip host,nmap
  vulny -update
  vulny -help
`
	c.Println(helpText)
}
 