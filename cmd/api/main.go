package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// QuoteRequest is the input payload for a payout quote.
type QuoteRequest struct {
	RouteID           string                 `json:"routeId"`
	PlannedDistanceKm float64                `json:"plannedDistanceKm"`
	ParcelCount       int                    `json:"parcelCount"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// QuoteResponse is the output payload for a payout quote.
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

	// Bootstrap parameters (production: load from SSM/config)
	const baseRate = 30.00
	const kmFactor = 0.125
	const minRate = 25.00
	const maxRate = 80.00

	ratePerParcel := baseRate + kmFactor*req.PlannedDistanceKm
	if ratePerParcel < minRate {
		ratePerParcel = minRate
	}
	if ratePerParcel > maxRate {
		ratePerParcel = maxRate
	}
	total := ratePerParcel * float64(req.ParcelCount)

	// Deterministic trace: sha256 of canonical input JSON
	b, _ := json.Marshal(req)
	h := sha256.Sum256(b)

	resp := QuoteResponse{
		RatePerParcelZAR: round2(ratePerParcel),
		TotalPayoutZAR:   round2(total),
		ConfigVersion:    "model_c_v1",
		Trace:            fmt.Sprintf("%x", h[:]),
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
