package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/afrogo-logisticx/afrogo-service-rate-engine/pkg/config"
	"github.com/afrogo-logisticx/afrogo-service-rate-engine/pkg/ledger"
	"github.com/afrogo-logisticx/afrogo-service-rate-engine/pkg/snapshot"
	"github.com/google/uuid"
)

type QuoteRequest struct {
	RouteID           string                 `json:"routeId"`
	PlannedDistanceKm float64                `json:"plannedDistanceKm"`
	ParcelCount       int                    `json:"parcelCount"`
	DriverID          string                 `json:"driverId,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type QuoteResponse struct {
	RatePerParcelZAR float64 `json:"ratePerParcelZar"`
	TotalPayoutZAR   float64 `json:"totalPayoutZar"`
	ConfigVersion    string  `json:"configVersion"`
	Trace            string  `json:"trace"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func parsePersist(r *http.Request) bool {
	// Check header first
	if v := r.Header.Get("X-Persist"); v != "" {
		b, _ := strconv.ParseBool(v)
		return b
	}
	// Check query param
	if v := r.URL.Query().Get("persist"); v != "" {
		b, _ := strconv.ParseBool(v)
		return b
	}
	return false
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	var req QuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// Basic validation
	if req.PlannedDistanceKm < 0 {
		http.Error(w, "plannedDistanceKm must be >= 0", http.StatusBadRequest)
		return
	}
	if req.ParcelCount <= 0 {
		http.Error(w, "parcelCount must be >= 1", http.StatusBadRequest)
		return
	}

	cfg := config.Load()

	ratePerParcel := cfg.BaseRate + cfg.KmFactor*req.PlannedDistanceKm
	if ratePerParcel < cfg.MinRate {
		ratePerParcel = cfg.MinRate
	}
	if ratePerParcel > cfg.MaxRate {
		ratePerParcel = cfg.MaxRate
	}
	total := ratePerParcel * float64(req.ParcelCount)

	// Deterministic trace: sha256 of canonical input JSON
	b, _ := json.Marshal(req)
	h := sha256.Sum256(b)

	resp := QuoteResponse{
		RatePerParcelZAR: round2(ratePerParcel),
		TotalPayoutZAR:   round2(total),
		ConfigVersion:    cfg.Version,
		Trace:            fmt.Sprintf("%x", h[:]),
	}

	// Persist snapshot + ledger if requested
	if parsePersist(r) {
		// Construct snapshot object
		snapObj := map[string]interface{}{
			"routeId":           req.RouteID,
			"driverId":          req.DriverID,
			"plannedDistanceKm": req.PlannedDistanceKm,
			"parcelCount":       req.ParcelCount,
			"metadata":          req.Metadata,
			"inputHash":         fmt.Sprintf("%x", h[:]),
			"configVersion":     cfg.Version,
			"serviceVersion":    "rate-engine-v1.0.0",
			"createdAt":         time.Now().UTC().Format(time.RFC3339),
		}
		path, err := snapshot.SaveLocally(req.RouteID, snapObj)
		if err != nil {
			http.Error(w, "failed to save snapshot", http.StatusInternalServerError)
			return
		}

		// Compute cents
		rateCents := int64(round2(ratePerParcel) * 100)
		totalCents := int64(round2(total) * 100)

		// Ledger entry
		entry := ledger.Entry{
			ID:                uuid.New().String(),
			RouteID:           req.RouteID,
			DriverID:          req.DriverID,
			ConfigVersion:     cfg.Version,
			ServiceVersion:    "rate-engine-v1.0.0",
			RateCents:         rateCents,
			TotalCents:        totalCents,
			ParcelCount:       req.ParcelCount,
			PlannedDistanceKm: req.PlannedDistanceKm,
			SnapshotPath:      path,
			InputHash:         fmt.Sprintf("%x", h[:]),
			OutputHash:        fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf(\"%d\", totalCents)))[:]),
		}
		if err := ledger.Append(entry); err != nil {
			http.Error(w, "failed to append ledger", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100.0
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/quote", quoteHandler)

	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("AfroGo Rate Engine listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
