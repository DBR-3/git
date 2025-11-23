package indicators

import "math"

// RSI calculates the Relative Strength Index
// RSI = 100 - (100 / (1 + RS))
// where RS = Average Gain / Average Loss
func RSI(prices []float64, period int) []float64 {
	if len(prices) < period+1 {
		return make([]float64, 0)
	}

	result := make([]float64, len(prices))

	// Calculate initial gains and losses
	gains := make([]float64, len(prices))
	losses := make([]float64, len(prices))

	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains[i] = change
			losses[i] = 0
		} else {
			gains[i] = 0
			losses[i] = math.Abs(change)
		}
	}

	// Calculate first average gain and loss using SMA
	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate first RSI
	if avgLoss == 0 {
		result[period] = 100
	} else if avgGain == 0 {
		result[period] = 0
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	// Calculate subsequent RSI values using Wilder's smoothing
	for i := period + 1; i < len(prices); i++ {
		avgGain = (avgGain*(float64(period)-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*(float64(period)-1) + losses[i]) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else if avgGain == 0 {
			result[i] = 0
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result
}

// RMA calculates the Running Moving Average (Wilder's smoothing)
func RMA(source []float64, length int) []float64 {
	if len(source) < length {
		return make([]float64, 0)
	}

	alpha := 1.0 / float64(length)
	result := make([]float64, len(source))

	// First value is simple average
	sum := 0.0
	for i := 0; i < length; i++ {
		sum += source[i]
	}
	result[length-1] = sum / float64(length)

	// Subsequent values use Wilder's smoothing
	for i := length; i < len(source); i++ {
		if math.IsNaN(result[i-1]) {
			// If previous value is NaN, calculate simple average
			sum := 0.0
			start := i - length + 1
			if start < 0 {
				start = 0
			}
			for j := start; j <= i; j++ {
				sum += source[j]
			}
			result[i] = sum / float64(i-start+1)
		} else {
			result[i] = alpha*source[i] + (1-alpha)*result[i-1]
		}
	}

	return result
}

// WarriorTrendIndicator calculates the Warrior Trend Indicator
// Returns: trend direction, ema_rsi, trend length
func WarriorTrendIndicator(closes []float64, longTrendLength, shortTrendLength int) ([]string, []float64, []int) {
	if len(closes) < longTrendLength+shortTrendLength {
		return []string{}, []float64{}, []int{}
	}

	// Calculate price changes
	changes := make([]float64, len(closes))
	for i := 1; i < len(closes); i++ {
		changes[i] = closes[i] - closes[i-1]
	}

	// Calculate max change (gains) and min change (losses)
	maxChanges := make([]float64, len(closes))
	minChanges := make([]float64, len(closes))
	for i := 0; i < len(changes); i++ {
		if changes[i] > 0 {
			maxChanges[i] = changes[i]
		}
		if changes[i] < 0 {
			minChanges[i] = -changes[i]
		}
	}

	// Calculate RMA for gains and losses
	up := RMA(maxChanges, longTrendLength)
	down := RMA(minChanges, longTrendLength)

	// Calculate RSI
	rsi := make([]float64, len(closes))
	for i := 0; i < len(up); i++ {
		if down[i] == 0 {
			rsi[i] = 100
		} else if up[i] == 0 {
			rsi[i] = 0
		} else {
			rsi[i] = 100 - (100 / (1 + up[i]/down[i]))
		}
	}

	// Calculate EMA of RSI
	emaRSI := EMA(rsi, shortTrendLength)

	// Determine trend direction
	trend := make([]string, len(closes))
	trendLength := make([]int, len(closes))

	for i := 1; i < len(emaRSI); i++ {
		if emaRSI[i] >= emaRSI[i-1] {
			trend[i] = "Trend is UP"
		} else {
			trend[i] = "Trend is DOWN"
		}

		// Calculate trend length
		if i > 0 && trend[i] == trend[i-1] {
			trendLength[i] = trendLength[i-1] + 1
		} else {
			trendLength[i] = 1
		}
	}

	return trend, emaRSI, trendLength
}
