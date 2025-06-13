// services/scan_worker.go
package services

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"regexp"

	"github.com/Shyena-Inc/Vulny/models"
	"github.com/adjust/rmq"
)

var (
	MongoClient *mongo.Client
	JWTSecret   string
)

func IsValidEmail(email string) bool {
    re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
    return re.MatchString(email)
}




func ScanWorkerProcess(delivery rmq.Delivery) {
	scanIDStr := delivery.Payload()
	log.Printf("Processing scan job id: %s", scanIDStr)

	scanID, err := primitive.ObjectIDFromHex(scanIDStr)
	if err != nil {
		log.Printf("Invalid scan ID %s: %v", scanIDStr, err)
		delivery.Reject()
		return
	}

	if err := runScan(scanID); err != nil {
		log.Printf("Scan job %s failed: %v", scanID.Hex(), err)
		delivery.Reject()
		return
	}

	log.Printf("Scan job %s completed", scanID.Hex())
	delivery.Ack()
}

func runScan(scanID primitive.ObjectID) error {
	scanCollection := MongoClient.Database("vulny").Collection("scans")

	ctx := context.Background()
	// Load scan document
	var scan models.Scan
	err := scanCollection.FindOne(ctx, primitive.M{"_id": scanID}).Decode(&scan)
	if err != nil {
		return errors.New("scan not found: " + err.Error())
	}

	_, err = scanCollection.UpdateOne(ctx, primitive.M{"_id": scanID}, primitive.M{
		"$set": primitive.M{
			"status":      "running",
			"completedAt": nil,
		},
	})
	if err != nil {
		return err
	}

	// TODO: Implement actual scan logic calling plugins:
	// port scanning
	portResults := ScanOpenPorts(scan.TargetURL)

	for _, pr := range portResults {
		if pr.Status == "open" {
			log.Printf("[+] Open Port %d (%s)", pr.Port, pr.Service)
		}
	}
	//subdomain enumeration
	wordlistPath := "wordlist/subdomains.txt"
    subdomains, err := EnumerateSubdomainsFromFileConcurrent(scan.TargetURL, wordlistPath)
   	if err != nil {
		log.Printf("Subdomain enumeration error: %v", err)
	} else {
		for _, sub := range subdomains {
		log.Printf("[+] Found subdomain: %s (%v)", sub.Subdomain, sub.IPs)
		}
	}


	// dir brute forcing
	//vuln check

	// Here - placeholder vulnerability result
	vulnerabilities := []models.Vulnerability{
		{
			Type:        "XSS",
			Parameter:   "search",
			Severity:    "medium",
			Description: "Cross-site scripting vulnerability found in search parameter.",
			Remediation: "Sanitize user input properly to prevent XSS.",
		},
		{
			Type:        "SQLi",
			Parameter:   "id",
			Severity:    "high",
			Description: "SQL Injection vulnerability found in id parameter.",
			Remediation: "Use parameterized queries to avoid SQL injection.",
		},
	}

	// Simulate scan time
	time.Sleep(10 * time.Second)

	// Update scan with completion info
	_, err = scanCollection.UpdateOne(ctx, primitive.M{"_id": scanID}, primitive.M{
		"$set": primitive.M{
			"status":          "completed",
			"completedAt":     time.Now(),
			"vulnerabilities": vulnerabilities,
		},
	})
	if err != nil {
		return err
	}

	return nil
}