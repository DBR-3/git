package indicators

// MansfieldRelativeStrength calculates the Mansfield Relative Strength indicator
// stockPrices: stock close prices
// benchmarkPrices: benchmark index close prices (e.g., IMOEX, S&P500)
// maLength: moving average length (default: 52)
// originalStyle: use original Mansfield RS calculation (default: true)
func MansfieldRelativeStrength(stockPrices, benchmarkPrices []float64, maLength int, originalStyle bool) ([]float64, []float64) {
	if len(stockPrices) != len(benchmarkPrices) || len(stockPrices) < maLength {
		return make([]float64, 0), make([]float64, 0)
	}

	n := len(stockPrices)
	result := make([]float64, n)
	zeroLineMA := make([]float64, n)

	// Calculate relative strength (stock / benchmark * 100)
	stockDividedByBenchmark := make([]float64, n)
	for i := 0; i < n; i++ {
		if benchmarkPrices[i] != 0 {
			stockDividedByBenchmark[i] = (stockPrices[i] / benchmarkPrices[i]) * 100
		}
	}

	// Calculate moving average (zero line)
	zeroLineMA = SMA(stockDividedByBenchmark, maLength)

	// Calculate Mansfield RS
	if originalStyle {
		// Original Mansfield RS: ((RS / MA) - 1) * 100
		for i := 0; i < n; i++ {
			if zeroLineMA[i] != 0 {
				result[i] = ((stockDividedByBenchmark[i] / zeroLineMA[i]) - 1) * 100
			}
		}
	} else {
		// Just relative strength
		result = stockDividedByBenchmark
	}

	return result, zeroLineMA
}

// MansfieldRSSignal determines the signal based on Mansfield RS
type MansfieldRSSignal string

const (
	RSCrossedZeroUp          MansfieldRSSignal = "Линия 0 пробита ВВЕРХ"
	RSCrossedZeroDown        MansfieldRSSignal = "Линия 0 пробита ВНИЗ"
	RSGrowing3DaysPositive   MansfieldRSSignal = "Рост RS 3+ дн. RS>0"
	RSGrowing3DaysNegative   MansfieldRSSignal = "Рост RS 3+ дл. RS<0"
	RSDeclining3DaysPositive MansfieldRSSignal = "Снижение RS 3+ дн. RS>0"
	RSDeclining3DaysNegative MansfieldRSSignal = "Снижение RS 3+ дн. RS<0"
	RSNoSignal               MansfieldRSSignal = ""
)

// DetectMansfieldSignal detects Mansfield RS signals
// Returns the signal type for each data point
func DetectMansfieldSignal(mansfieldRS []float64) []MansfieldRSSignal {
	if len(mansfieldRS) < 4 {
		return make([]MansfieldRSSignal, 0)
	}

	signals := make([]MansfieldRSSignal, len(mansfieldRS))

	for i := 3; i < len(mansfieldRS); i++ {
		current := mansfieldRS[i]
		prev := mansfieldRS[i-1]
		prev2 := mansfieldRS[i-2]
		prev3 := mansfieldRS[i-3]

		// Check for zero line cross
		if current > 0 && prev < 0 {
			signals[i] = RSCrossedZeroUp
		} else if current < 0 && prev > 0 {
			signals[i] = RSCrossedZeroDown
		} else if current < 0 && current > prev && prev > prev2 {
			// RS growing for 3+ days, RS < 0
			signals[i] = RSGrowing3DaysNegative
		} else if current > 0 && current > prev && prev > prev2 {
			// RS growing for 3+ days, RS > 0
			signals[i] = RSGrowing3DaysPositive
		} else if current > 0 && current < prev && prev < prev2 {
			// RS declining for 3+ days, RS > 0
			signals[i] = RSDeclining3DaysPositive
		} else if current < 0 && current < prev && prev < prev2 {
			// RS declining for 3+ days, RS < 0
			signals[i] = RSDeclining3DaysNegative
		} else {
			signals[i] = RSNoSignal
		}
	}

	return signals
}

// MansfieldRSMetrics holds Mansfield RS metrics at a specific point
type MansfieldRSMetrics struct {
	Current    float64
	Prev1Day   float64
	Prev2Days  float64
	Prev3Days  float64
	Prev4Days  float64
	Signal     MansfieldRSSignal
	IsPositive bool
	Trend      string // "Growing", "Declining", "Flat"
}

// GetMansfieldMetrics calculates Mansfield RS metrics for the latest data point
func GetMansfieldMetrics(mansfieldRS []float64) *MansfieldRSMetrics {
	if len(mansfieldRS) < 5 {
		return nil
	}

	idx := len(mansfieldRS) - 1
	metrics := &MansfieldRSMetrics{
		Current:    mansfieldRS[idx],
		Prev1Day:   mansfieldRS[idx-1],
		Prev2Days:  mansfieldRS[idx-2],
		Prev3Days:  mansfieldRS[idx-3],
		Prev4Days:  mansfieldRS[idx-4],
		IsPositive: mansfieldRS[idx] > 0,
	}

	// Determine trend
	if metrics.Current > metrics.Prev1Day && metrics.Prev1Day > metrics.Prev2Days {
		metrics.Trend = "Growing"
	} else if metrics.Current < metrics.Prev1Day && metrics.Prev1Day < metrics.Prev2Days {
		metrics.Trend = "Declining"
	} else {
		metrics.Trend = "Flat"
	}

	// Detect signal
	signals := DetectMansfieldSignal(mansfieldRS)
	metrics.Signal = signals[idx]

	return metrics
}

// MansfieldRank ranks stocks by their Mansfield RS
// Returns indices sorted by Mansfield RS (descending)
func MansfieldRank(mansfieldRS [][]float64) []int {
	type tickerRS struct {
		idx          int
		latestRS     float64
		avgRecentRS  float64 // Average of last 5 values
	}

	rankings := make([]tickerRS, len(mansfieldRS))

	for i, rs := range mansfieldRS {
		if len(rs) == 0 {
			rankings[i] = tickerRS{idx: i, latestRS: -9999, avgRecentRS: -9999}
			continue
		}

		latest := rs[len(rs)-1]

		// Calculate average of recent RS values
		sum := 0.0
		count := 0
		for j := len(rs) - 1; j >= 0 && count < 5; j-- {
			sum += rs[j]
			count++
		}
		avg := sum / float64(count)

		rankings[i] = tickerRS{idx: i, latestRS: latest, avgRecentRS: avg}
	}

	// Sort by average recent RS (descending)
	// Using bubble sort for simplicity (can be optimized with sort.Slice)
	for i := 0; i < len(rankings)-1; i++ {
		for j := 0; j < len(rankings)-i-1; j++ {
			if rankings[j].avgRecentRS < rankings[j+1].avgRecentRS {
				rankings[j], rankings[j+1] = rankings[j+1], rankings[j]
			}
		}
	}

	// Extract indices
	result := make([]int, len(rankings))
	for i, r := range rankings {
		result[i] = r.idx
	}

	return result
}
