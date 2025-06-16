package main

import (
	"fmt"
	"os"

	"github.com/Shyena-Inc/Vulny/cmd"
	"github.com/Shyena-Inc/Vulny/report"
	"github.com/Shyena-Inc/Vulny/scanner"
	"github.com/Shyena-Inc/Vulny/utils"
)

func main() {
	// Parse command-line arguments
	args := cmd.ParseArgs()
	if args.Help || (!args.Update && args.Target == "") {
		fmt.Println("RapidScan - The Multi-Tool Web Vulnerability Scanner")
		fmt.Println("Usage:")
		fmt.Println("\trapidscan -target example.com")
		fmt.Println("\trapidscan -target example.com -skip host,nmap")
		fmt.Println("\trapidscan -update")
		os.Exit(0)
	}

	if args.Update {
		fmt.Println("RapidScan is updating... Please wait.")
		// Placeholder for update logic
		fmt.Println("Update functionality not implemented in this version.")
		os.Exit(0)
	}

	// Normalize target URL
	target := utils.NormalizeURL(args.Target)

	// Perform scan
	vulnerabilities, totalElapsed, skippedChecks := scanner.Scan(target, args.Skip, args.NoSpinner)

	// Generate report
	report.GenerateReport(target, vulnerabilities, len(scanner.Tools), skippedChecks, totalElapsed)
}