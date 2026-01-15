package engine

// ComputeRatePerParcel computes the ZAR rate per parcel for a given distance.
func ComputeRatePerParcel(baseRate, kmFactor, minRate, maxRate, distanceKm float64) float64 {
	r := baseRate + kmFactor*distanceKm
	if r < minRate {
		return minRate
	}
	if r > maxRate {
		return maxRate
	}
	return round2(r)
}

func round2(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100.0
}
