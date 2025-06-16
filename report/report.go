package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shyena-Inc/Vulny/models"
)

// GenerateReport creates vulnerability report and debug log
func GenerateReport(target string, vulnerabilities []string, totalChecks, skippedChecks int, totalElapsed time.Duration) {
	fmt.Println(models.Blue("[ Report Generation Phase Initiated. ]") + models.Reset())

	currentDate := time.Now().Format("2006-01-02")
	vulReport := fmt.Sprintf("vulny.vul.%s.%s", target, currentDate)
	debugLog := fmt.Sprintf("vulny.dbg.%s.%s", target, currentDate)

	if len(vulnerabilities) == 0 {
		fmt.Println(models.Green("No Vulnerabilities Detected.") + models.Reset())
	} else {
		f, err := os.Create(vulReport)
		if err != nil {
			fmt.Printf("%sError creating report file: %v%s\n", models.Red(""), err, models.Reset())
			os.Exit(1)
		}
		defer f.Close()

		for _, vul := range vulnerabilities {
			parts := strings.Split(vul, "*")
			f.WriteString(parts[1] + "\n------------------------\n\n")
			data, _ := os.ReadFile(fmt.Sprintf("/tmp/vulny_temp_%s", parts[0]))
			f.Write(data)
			f.WriteString("\n\n")
		}
		fmt.Printf("\tComplete Vulnerability Report for %s%s%s named %s%s%s is available.\n", models.Blue(""), target, models.Reset(), models.Green(""), vulReport, models.Reset())
	}

	// Debug log
	f, err := os.Create(debugLog)
	if err != nil {
		fmt.Printf("%sError creating debug log: %v%s\n", models.Red(""), err, models.Reset())
		os.Exit(1)
	}
	defer f.Close()

	for _, tool := range models.Tools {
		if !tool.Enabled {
			continue
		}
		data, err := os.ReadFile(fmt.Sprintf("/tmp/vulny_temp_%s", tool.Status.ID))
		if err == nil {
			f.WriteString(tool.Description + "\n------------------------\n\n")
			f.Write(data)
			f.WriteString("\n\n")
		}
	}

	fmt.Printf("\tTotal Number of Vulnerability Checks        : %s%d%s\n", models.Green(""), totalChecks, models.Reset())
	fmt.Printf("\tTotal Number of Vulnerability Checks Skipped: %s%d%s\n", models.Yellow(""), skippedChecks, models.Reset())
	fmt.Printf("\tTotal Number of Vulnerabilities Detected    : %s%d%s\n", models.Red(""), len(vulnerabilities), models.Reset())
	fmt.Printf("\tTotal Time Elapsed for the Scan             : %s%v%s\n", models.Blue(""), totalElapsed, models.Reset())
	fmt.Printf("\tDebug log available at %s%s%s\n", models.Blue(""), debugLog, models.Reset())
	fmt.Println(models.Blue("[ Report Generation Phase Completed. ]") + models.Reset())

	// Clean up temporary files
	files, _ := filepath.Glob("/tmp/vulny_temp_*")
	for _, file := range files {
		os.Remove(file)
	}
}