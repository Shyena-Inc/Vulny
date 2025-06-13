// File: controllers/scan.go
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetAllScans handles GET /api/scans/ and returns all scans (dummy implementation)
func GetAllScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"scans": []interface{}{},
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetScanByID handles GET /api/scans/{id}
func GetScanByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Optionally validate id here (e.g., ObjectID format)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"status":  "pending",
		"message": "Scan details would be fetched from the database here.",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func StartScan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Scan started (stub handler)"}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
