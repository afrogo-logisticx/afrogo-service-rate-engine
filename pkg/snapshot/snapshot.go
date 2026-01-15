package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveLocally saves a snapshot as an indented JSON file under ./snapshots.
// The filename pattern: <timestamp>_<routeId>_<uuid>.json (we use timestamp+route for determinism).
func SaveLocally(routeID string, payload interface{}) (string, error) {
	dir := "snapshots"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	t := time.Now().UTC().Format("20060102T150405Z")
	filename := fmt.Sprintf("%s_%s.json", t, sanitize(routeID))
	path := filepath.Join(dir, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	return path, nil
}

// sanitize reduces characters unsafe for filenames (simple implementation)
func sanitize(s string) string {
	if s == "" {
		return "route"
	}
	// replace spaces and slashes
	out := ""
	for _, r := range s {
		switch r {
		case '/', '\\', ' ', ':':
			out += "_"
		default:
			out += string(r)
		}
	}
	return out
}
