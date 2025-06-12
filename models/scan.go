// models/scan.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vulnerability struct {
	Type        string `bson:"type" json:"type"`
	Parameter   string `bson:"parameter" json:"parameter"`
	Severity    string `bson:"severity" json:"severity"` // low, medium, high, critical
	Description string `bson:"description" json:"description"`
	Remediation string `bson:"remediation" json:"remediation"`
}

type Scan struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User            primitive.ObjectID `bson:"user" json:"user"`
	TargetURL       string             `bson:"targetURL" json:"targetURL"`
	Config          ScanConfig         `bson:"config" json:"config"`
	Status          string             `bson:"status" json:"status"` // pending, running, completed, cancelled, failed
	Vulnerabilities []Vulnerability    `bson:"vulnerabilities" json:"vulnerabilities"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	CompletedAt     *time.Time         `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
}

type ScanConfig struct {
	Depth   int               `bson:"depth" json:"depth"`
	Headers map[string]string `bson:"headers" json:"headers"`
}