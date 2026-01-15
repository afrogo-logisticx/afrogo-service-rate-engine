package engine

import "testing"

func TestComputeRatePerParcel(t *testing.T) {
	const base = 30.00
	const kmFactor = 0.125
	const minRate = 25.00
	const maxRate = 80.00

	tests := []struct{
		dist float64
		want float64
	}{
		{10, 31.25},
		{0, 30.00},
		{1000, 80.00}, // cap
	}

	for _, tt := range tests {
		got := ComputeRatePerParcel(base, kmFactor, minRate, maxRate, tt.dist)
		if got != tt.want {
			t.Fatalf("distance=%v, want %v, got %v", tt.dist, tt.want, got)
		}
	}
}
