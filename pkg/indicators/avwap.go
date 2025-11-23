package indicators

import (
	"math"
	"time"
)

// AVWAPBands represents AVWAP with standard deviation bands
type AVWAPBands struct {
	AVWAP     []float64
	StdDev    []float64
	Band1Up   []float64
	Band1Down []float64
	Band2Up   []float64
	Band2Down []float64
	Band3Up   []float64
	Band3Down []float64
}

// CalculateAVWAPBands calculates Anchored VWAP with bands from anchor point
// dates: slice of dates
// prices: typical prices (or close prices)
// volumes: volume data
// anchorDate: the date to anchor from
// multipliers: standard deviation multipliers for bands (default: 1, 2, 3)
func CalculateAVWAPBands(dates []time.Time, prices, volumes []float64, anchorDate time.Time, multipliers []float64) *AVWAPBands {
	if len(dates) != len(prices) || len(prices) != len(volumes) {
		return nil
	}

	// Find anchor index
	anchorIdx := -1
	for i, date := range dates {
		if date.Equal(anchorDate) || date.After(anchorDate) {
			anchorIdx = i
			break
		}
	}

	if anchorIdx == -1 {
		return nil
	}

	n := len(prices)
	result := &AVWAPBands{
		AVWAP:     make([]float64, n),
		StdDev:    make([]float64, n),
		Band1Up:   make([]float64, n),
		Band1Down: make([]float64, n),
		Band2Up:   make([]float64, n),
		Band2Down: make([]float64, n),
		Band3Up:   make([]float64, n),
		Band3Down: make([]float64, n),
	}

	// Default multipliers if not provided
	if len(multipliers) == 0 {
		multipliers = []float64{1, 2, 3}
	}

	// Calculate cumulative values from anchor point
	for i := anchorIdx; i < n; i++ {
		// Calculate cumulative volume * price
		cumulativeVolumePrice := 0.0
		cumulativeVolume := 0.0

		for j := anchorIdx; j <= i; j++ {
			cumulativeVolumePrice += volumes[j] * prices[j]
			cumulativeVolume += volumes[j]
		}

		// Calculate AVWAP
		if cumulativeVolume > 0 {
			result.AVWAP[i] = cumulativeVolumePrice / cumulativeVolume
		}

		// Calculate volume-weighted standard deviation
		if i > anchorIdx {
			weightedVariance := 0.0
			for j := anchorIdx; j <= i; j++ {
				diff := prices[j] - result.AVWAP[i]
				weightedVariance += (diff * diff * volumes[j])
			}
			if cumulativeVolume > 0 {
				result.StdDev[i] = math.Sqrt(weightedVariance / cumulativeVolume)
			}
		}

		// Calculate bands
		if len(multipliers) >= 1 {
			result.Band1Up[i] = result.AVWAP[i] + (result.StdDev[i] * multipliers[0])
			result.Band1Down[i] = result.AVWAP[i] - (result.StdDev[i] * multipliers[0])
		}
		if len(multipliers) >= 2 {
			result.Band2Up[i] = result.AVWAP[i] + (result.StdDev[i] * multipliers[1])
			result.Band2Down[i] = result.AVWAP[i] - (result.StdDev[i] * multipliers[1])
		}
		if len(multipliers) >= 3 {
			result.Band3Up[i] = result.AVWAP[i] + (result.StdDev[i] * multipliers[2])
			result.Band3Down[i] = result.AVWAP[i] - (result.StdDev[i] * multipliers[2])
		}
	}

	return result
}

// CalculateMultipleAVWAP calculates AVWAP for multiple anchor points
// Useful for 1Y, 3Y, 5Y AVWAP
func CalculateMultipleAVWAP(dates []time.Time, prices, volumes []float64, anchorDates []time.Time) map[string]*AVWAPBands {
	result := make(map[string]*AVWAPBands)

	for i, anchorDate := range anchorDates {
		label := ""
		switch i {
		case 0:
			label = "1Y"
		case 1:
			label = "3Y"
		case 2:
			label = "5Y"
		default:
			label = anchorDate.Format("2006-01-02")
		}

		bands := CalculateAVWAPBands(dates, prices, volumes, anchorDate, []float64{1, 2, 3})
		if bands != nil {
			result[label] = bands
		}
	}

	return result
}

// GetAnchorDates returns anchor dates for current year, 2 years ago, and 4 years ago
func GetAnchorDates(currentDate time.Time) []time.Time {
	currentYear := currentDate.Year()

	return []time.Time{
		time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC),         // Current year start
		time.Date(currentYear-2, 1, 1, 0, 0, 0, 0, time.UTC),       // 2 years ago
		time.Date(currentYear-4, 1, 1, 0, 0, 0, 0, time.UTC),       // 4 years ago
	}
}

// AVWAPPosition calculates the position of price relative to AVWAP bands
// Returns percentage distance from AVWAP
func AVWAPPosition(price float64, avwap float64) float64 {
	if avwap == 0 {
		return 0
	}
	return ((price / avwap) - 1) * 100
}

// IsAboveAVWAP checks if price is above AVWAP
func IsAboveAVWAP(price, avwap float64) bool {
	return price > avwap
}

// GetBandLevel determines which AVWAP band the price is in
// Returns: -3, -2, -1, 0, 1, 2, 3 (0 means between -1 and +1)
func GetBandLevel(price float64, bands *AVWAPBands, idx int) int {
	if idx >= len(bands.AVWAP) {
		return 0
	}

	avwap := bands.AVWAP[idx]

	if price >= bands.Band3Up[idx] {
		return 3
	} else if price >= bands.Band2Up[idx] {
		return 2
	} else if price >= bands.Band1Up[idx] {
		return 1
	} else if price <= bands.Band3Down[idx] {
		return -3
	} else if price <= bands.Band2Down[idx] {
		return -2
	} else if price <= bands.Band1Down[idx] {
		return -1
	}

	// Price is between -1 and +1 bands (near AVWAP)
	if price > avwap {
		return 1
	} else if price < avwap {
		return -1
	}
	return 0
}
