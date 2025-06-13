package services

import (
	"fmt"
	"net"
	"sort"
	"time"
)
 
type PortResult struct {
	Port    int    `json:"port"`
	Service string `json:"service"`
	Status  string `json:"status"`
}
 
var commonPorts = map[int]string{
	21:   "FTP",
	22:   "SSH",
	23:   "Telnet",
	25:   "SMTP",
	53:   "DNS",
	80:   "HTTP",
	110:  "POP3",
	143:  "IMAP",
	443:  "HTTPS",
	3306: "MySQL",
	5432: "PostgreSQL",
	6379: "Redis",
	8080: "HTTP-Alt",
}

 
func ScanOpenPorts(host string) []PortResult {
	var results []PortResult
	timeout := 2 * time.Second

	for port, service := range commonPorts {
		address := net.JoinHostPort(host, fmt.Sprintf("%d", port))

		conn, err := net.DialTimeout("tcp", address, timeout)

		status := "closed"
		if err == nil {
			status = "open"
			conn.Close()
		}

		results = append(results, PortResult{
			Port:    port,
			Service: service,
			Status:  status,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results
}
