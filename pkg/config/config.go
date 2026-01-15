package config

import (
	"os"
	"strconv"
)

// RateConfig holds the rate engine parameters.
type RateConfig struct {
	BaseRate float64
	KmFactor float64
	MinRate  float64
	MaxRate  float64
	Version  string
}

// getEnvFloat reads a float64 from env or returns default.
func getEnvFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// Load returns a RateConfig sourced from environment variables.
// Production: this will be replaced/extended with SSM/parameter-store loading.
func Load() *RateConfig {
	cfg := &RateConfig{
		BaseRate: getEnvFloat("BASE_RATE_ZAR", 30.00),
		KmFactor: getEnvFloat("KM_FACTOR_ZAR", 0.125),
		MinRate:  getEnvFloat("MIN_RATE_PER_PARCEL_ZAR", 25.00),
		MaxRate:  getEnvFloat("MAX_RATE_PER_PARCEL_ZAR", 80.00),
		Version:  os.Getenv("CONFIG_VERSION"),
	}
	if cfg.Version == "" {
		cfg.Version = "model_c_v1"
	}
	return cfg
}
