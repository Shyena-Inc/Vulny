package scanner

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/Shyena-Inc/Vulny/models"
	"github.com/fatih/color"
)

// Scan performs vulnerability scanning
func Scan(target string, skip []string, noSpinner bool) ([]string, time.Duration, int) {
	fmt.Printf("\n%s[Preliminary Scan Phase Initiated... Loaded %d vulnerability checks.]%s\n", models.Blue, len(models.Tools), models.Reset)

	var vulnerabilities []string
	var totalElapsed time.Duration
	var skippedChecks int

	spinner := NewSpinner("Scanning...", 100*time.Millisecond)
	if !noSpinner {
		spinner.Start()
	}

	for i, tool := range models.Tools {
		// Check if tool is skipped or unavailable
		if contains(skip, tool.Name) || !checkTool(strings.Split(tool.Command, " ")[0]) {
			models.Tools[i].Enabled = false
			fmt.Printf("\n[%s%s] Deploying %d/%d | %s%s%s\n", tool.Status.ProcLevel, tool.Status.EstTime, i+1, len(models.Tools), models.Blue, tool.Description, models.Reset)
			fmt.Println(models.Yellow("Scanning Tool Unavailable. Skipping Test..."))
			skippedChecks++
			continue
		}

		fmt.Printf("\n[%s%s] Deploying %d/%d | %s%s%s\n", tool.Status.ProcLevel, tool.Status.EstTime, i+1, len(models.Tools), models.Blue, tool.Description, models.Reset)
		tempFile := fmt.Sprintf("/tmp/rapidscan_temp_%s", tool.Status.ID)

		output, elapsed, err := runTool(tool.Command, target, tempFile)
		totalElapsed += elapsed

		if err != nil && err != exec.ErrNotFound {
			if !noSpinner {
				spinner.Stop()
			}
			fmt.Printf("\n%sScan Interrupted in %v%s\n", models.Blue, elapsed, models.Reset)
			fmt.Println(models.Yellow("Test Skipped. Press Ctrl+C to quit RapidScan."))
			skippedChecks++
			if !noSpinner {
				spinner.Start()
			}
			continue
		}

		if !noSpinner {
			spinner.Stop()
		}
		fmt.Printf("\n%sScan Completed in %v%s\n", models.Blue, elapsed, models.Reset)

		// Check for vulnerabilities
		isVulnerable := false
		if tool.Status.Inverted {
			if !strings.Contains(strings.ToLower(output), strings.ToLower(tool.Status.CheckString)) {
				isVulnerable = true
			}
		} else {
			if strings.Contains(strings.ToLower(output), strings.ToLower(tool.Status.CheckString)) {
				isVulnerable = true
			}
		}

		for _, badResp := range tool.Status.BadResponses {
			if strings.Contains(output, badResp) {
				isVulnerable = false
				break
			}
		}

		if isVulnerable {
			vulnerabilities = append(vulnerabilities, fmt.Sprintf("%s*%s", tool.Status.ID, tool.Description))
			printVulInfo(tool)
		}

		if !noSpinner {
			spinner.Start()
		}
	}

	if !noSpinner {
		spinner.Stop()
	}
	fmt.Println(models.Blue("[ Preliminary Scan Phase Completed. ]"))
	return vulnerabilities, totalElapsed, skippedChecks
}

// checkTool verifies if a tool is available
func checkTool(toolName string) bool {
	cmd := exec.Command("which", toolName)
	return cmd.Run() == nil
}

// runTool executes a tool and captures output
func runTool(cmdStr string, target string, tempFile string) (string, time.Duration, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(cmdStr, target)+" > "+tempFile+" 2>&1")
	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	output, errRead := ioutil.ReadFile(tempFile)
	if errRead != nil {
		return "", elapsed, errRead
	}
	return string(output), elapsed, err
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// printVulInfo displays vulnerability information
func printVulInfo(tool models.Tool) {
	fmt.Printf("%sVulnerability Threat Level%s\n", color.New(color.Bold).SprintFunc(), models.Reset)
	fmt.Printf("\t%s %s%s\n", models.VulInfo(tool.Resp.Severity), models.Yellow(tool.Resp.Message), models.Reset)
	fmt.Printf("%sVulnerability Definition%s\n", color.New(color.Bold).SprintFunc(), models.Reset)
	fmt.Printf("\t%s%s%s\n", models.Red(models.Remediations[tool.Resp.RemedIndex-1].Description), models.Reset, models.Reset)
	fmt.Printf("%sVulnerability Remediation%s\n", color.New(color.Bold).SprintFunc(), models.Reset)
	fmt.Printf("\t%s%s%s\n", models.Green(models.Remediations[tool.Resp.RemedIndex-1].Remediation), models.Reset, models.Reset)
}