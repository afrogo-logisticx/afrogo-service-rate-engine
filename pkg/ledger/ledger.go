package ledger

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry is a single ledger row (dev/local representation).
type Entry struct {
	ID                string  `json:"id,omitempty"`
	RouteID           string  `json:"route_id"`
	DriverID          string  `json:"driver_id,omitempty"`
	ConfigVersion     string  `json:"config_version"`
	ServiceVersion    string  `json:"service_version,omitempty"`
	RateCents         int64   `json:"rate_cents"`
	TotalCents        int64   `json:"total_cents"`
	ParcelCount       int     `json:"parcel_count"`
	PlannedDistanceKm float64 `json:"planned_distance_km"`
	SnapshotPath      string  `json:"snapshot_path"`
	InputHash         string  `json:"input_hash"`
	OutputHash        string  `json:"output_hash"`
	CreatedAt         string  `json:"created_at"`
}

var mu sync.Mutex
var ledgerFile = "ledger/payouts_ledger.ndjson"

// Append writes the entry as a newline-delimited JSON line.
func Append(e Entry) error {
	mu.Lock()
	defer mu.Unlock()

	if err := os.MkdirAll("ledger", 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(ledgerFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if e.CreatedAt == "" {
		e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}
