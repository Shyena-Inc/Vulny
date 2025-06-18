package main

import (
	"fmt"
	"os"

	"github.com/Shyena-Inc/Vulny/cmd"
	"github.com/Shyena-Inc/Vulny/models"
	"github.com/Shyena-Inc/Vulny/report"
	"github.com/Shyena-Inc/Vulny/scanner"
	"github.com/Shyena-Inc/Vulny/utils"
)

func main() {
	// Parse command-line arguments
	args := cmd.ParseArgs()
	if args.Help || (!args.Update && args.Target == "") {
		fmt.Println("Vulny - The Multi-Tool Web Vulnerability Scanner")
		fmt.Println("Usage:")
		fmt.Println("\tvulny -target example.com")
		fmt.Println("\tvulny -target example.com -skip host,nmap")
		fmt.Println("\tvulny -update")
		os.Exit(0)
	}

	if args.Update {
		fmt.Println("🔄 Checking for updates...")
		err := utils.UpdateBinary()
		if err != nil {
			fmt.Printf("❌ Update failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Vulny updated successfully! Run the command again.")
		os.Exit(0)
	}


	// Normalize target URL
	target := utils.NormalizeURL(args.Target)

	// Perform scan
	vulnerabilities, totalElapsed, skippedChecks := scanner.Scan(target, args.Skip, args.NoSpinner)

	// Generate report
	toolCount := len(models.Tools)
	report.GenerateReport(target, vulnerabilities, toolCount, skippedChecks, totalElapsed)
}