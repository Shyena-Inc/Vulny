package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Shyena-Inc/Vulny/middlewares"
	"github.com/Shyena-Inc/Vulny/models"
	"github.com/adjust/rmq/v4"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	scanCollection *mongo.Collection
	scanQueue      rmq.Queue
)

// SetScanCollection sets the MongoDB scan collection
func SetScanCollection(db *mongo.Database) {
	scanCollection = db.Collection("scans")
}

// SetScanQueue sets the RMQ scan queue
func SetScanQueue(queue rmq.Queue) {
	scanQueue = queue
}

// scanRequest defines the request body for starting a scan
type scanRequest struct {
	Target   string            `json:"target"`
	ScanType string            `json:"scan_type"`
	Config   *models.ScanConfig `json:"config,omitempty"`
}

// scanResponse defines the response for scan operations
type scanResponse struct {
	Message string     `json:"message"`
	Scan    scanData   `json:"scan,omitempty"`
	Scans   []scanData `json:"scans,omitempty"`
}

type scanData struct {
	ID              string                `json:"id"`
	TargetURL       string                `json:"targetURL"`
	ScanType        string                `json:"scan_type"`
	Config          models.ScanConfig     `json:"config"`
	Status          string                `json:"status"`
	Vulnerabilities []models.Vulnerability `json:"vulnerabilities"`
	Ports           []models.PortResult   `json:"ports,omitempty"`
	Subdomains      []models.SubdomainResult `json:"subdomains,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
	CompletedAt     *time.Time            `json:"completedAt,omitempty"`
}

// StartScan handles POST /api/scans
func StartScan(w http.ResponseWriter, r *http.Request) {
	user := middlewares.UserFromRequest(r)
	if user == nil {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.Target = strings.TrimSpace(req.Target)
	req.ScanType = strings.TrimSpace(req.ScanType)
	if req.Target == "" || req.ScanType == "" {
		sendError(w, "Target and scan_type are required", http.StatusBadRequest)
		return
	}
	if !isValidScanType(req.ScanType) {
		sendError(w, "Invalid scan_type", http.StatusBadRequest)
		return
	}

	// Set default config if not provided
	config := models.ScanConfig{Depth: 1, Headers: map[string]string{}}
	if req.Config != nil {
		config = *req.Config
		if config.Headers == nil {
			config.Headers = map[string]string{}
		}
	}

	scan := models.Scan{
		User:      user.ID,
		TargetURL: req.Target,
		Config:    config,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := scanCollection.InsertOne(ctx, scan)
	if err != nil {
		log.Printf("Failed to create scan: %v", err)
		sendError(w, "Failed to create scan", http.StatusInternalServerError)
		return
	}

	scanID := res.InsertedID.(primitive.ObjectID)
	scan.ID = scanID

	jobPayload, err := json.Marshal(map[string]interface{}{
		"scan_id":   scanID.Hex(),
		"target":    req.Target,
		"scan_type": req.ScanType,
		"user_id":   user.ID.Hex(),
		"config":    config,
	})
	if err != nil {
		log.Printf("Failed to marshal scan job: %v", err)
		sendError(w, "Failed to queue scan", http.StatusInternalServerError)
		return
	}

	if err := scanQueue.Publish(string(jobPayload)); err != nil {
		log.Printf("Failed to publish scan job: %v", err)
		sendError(w, "Failed to queue scan", http.StatusInternalServerError)
		return
	}

	resp := scanResponse{
		Message: "Scan started",
		Scan: scanData{
			ID:              scanID.Hex(),
			TargetURL:       scan.TargetURL,
			ScanType:        req.ScanType,
			Config:          scan.Config,
			Status:          scan.Status,
			Vulnerabilities: scan.Vulnerabilities,
			Ports:           scan.Ports,
			Subdomains:      scan.Subdomains,
			CreatedAt:       scan.CreatedAt,
			CompletedAt:     scan.CompletedAt,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetAllScans handles GET /api/scans
func GetAllScans(w http.ResponseWriter, r *http.Request) {
	user := middlewares.UserFromRequest(r)
	if user == nil {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := scanCollection.Find(ctx, bson.M{"user": user.ID})
	if err != nil {
		log.Printf("Failed to query scans: %v", err)
		sendError(w, "Failed to query scans", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var scans []models.Scan
	if err := cursor.All(ctx, &scans); err != nil {
		log.Printf("Failed to decode scans: %v", err)
		sendError(w, "Failed to decode scans", http.StatusInternalServerError)
		return
	}

	scanDataList := make([]scanData, len(scans))
	for i, scan := range scans {
		scanType := ""
		if scan.Config.Depth > 0 {
			scanType = "custom" // Placeholder, adjust based on scan_type
		}
		scanDataList[i] = scanData{
			ID:              scan.ID.Hex(),
			TargetURL:       scan.TargetURL,
			ScanType:        scanType,
			Config:          scan.Config,
			Status:          scan.Status,
			Vulnerabilities: scan.Vulnerabilities,
			Ports:           scan.Ports,
			Subdomains:      scan.Subdomains,
			CreatedAt:       scan.CreatedAt,
			CompletedAt:     scan.CompletedAt,
		}
	}

	resp := scanResponse{
		Message: "Scans retrieved",
		Scans:   scanDataList,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetScanByID handles GET /api/scans/{id}
func GetScanByID(w http.ResponseWriter, r *http.Request) {
	user := middlewares.UserFromRequest(r)
	if user == nil {
		sendError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	scanID, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		sendError(w, "Invalid scan ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var scan models.Scan
	err = scanCollection.FindOne(ctx, bson.M{"_id": scanID, "user": user.ID}).Decode(&scan)
	if err == mongo.ErrNoDocuments {
		sendError(w, "Scan not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Failed to query scan: %v", err)
		sendError(w, "Failed to query scan", http.StatusInternalServerError)
		return
	}

	resp := scanResponse{
		Message: "Scan retrieved",
		Scan: scanData{
			ID:              scan.ID.Hex(),
			TargetURL:       scan.TargetURL,
			ScanType:        func() string { if scan.Config.Depth > 0 { return "custom" } else { return "" } }(), // Placeholder
			Config:          scan.Config,
			Status:          scan.Status,
			Vulnerabilities: scan.Vulnerabilities,
			Ports:           scan.Ports,
			Subdomains:      scan.Subdomains,
			CreatedAt:       scan.CreatedAt,
			CompletedAt:     scan.CompletedAt,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// isValidScanType checks if the scan type is supported
func isValidScanType(scanType string) bool {
	validTypes := map[string]struct{}{
		"port_scan":      {},
		"subdomain_enum": {},
	}
	_, ok := validTypes[scanType]
	return ok
}
