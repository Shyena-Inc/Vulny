package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Shyena-Inc/Vulny/models"
	"github.com/Ullaakut/nmap/v2"
)

// ScanOpenPorts scans ports on the host using nmap
func ScanOpenPorts(ctx context.Context, host string, config models.ScanConfig) ([]models.PortResult, error) {
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}

	// Determine port range based on config.Depth
	var ports string
	switch config.Depth {
	case 0, 1:
		ports = "21,22,23,25,53,80,110,143,443,3306,5432,6379,8080" // Common ports
	case 2:
		ports = "1-1000" // Top 1000 ports
	default:
		ports = "1-65535" // Full range (slow)
	}

	// Set a timeout for the scan context
	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(host),
		nmap.WithPorts(ports),
		nmap.WithServiceInfo(),
		nmap.WithContext(scanCtx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nmap scanner: %v", err)
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, fmt.Errorf("nmap scan failed: %v", err)
	}
	if len(warnings) > 0 {
		log.Printf("nmap warnings for %s: %v", host, warnings)
	}

	var results []models.PortResult
	for _, h := range result.Hosts {
		if h.Status.State != "up" {
			log.Printf("Host %s is down", host)
			continue
		}
		for _, port := range h.Ports {
			service := port.Service.Name
			if service == "" {
				service = strings.ToUpper(port.Protocol)
			}
			results = append(results, models.PortResult{
				Port:    int(port.ID),
				Service: service,
				Status:  string(port.State.State),
			})
		}
	}

	return results, nil
}
