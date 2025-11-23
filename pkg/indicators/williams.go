package indicators

// WilliamsR calculates Williams %R indicator
// Williams %R = (Highest High - Close) / (Highest High - Lowest Low) * -100
func WilliamsR(high, low, close []float64, period int) []float64 {
	if len(high) < period || len(low) < period || len(close) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(high))

	for i := period - 1; i < len(high); i++ {
		// Find highest high and lowest low in the period
		highestHigh := high[i]
		lowestLow := low[i]

		for j := i - period + 1; j <= i; j++ {
			if high[j] > highestHigh {
				highestHigh = high[j]
			}
			if low[j] < lowestLow {
				lowestLow = low[j]
			}
		}

		// Calculate Williams %R
		denominator := highestHigh - lowestLow
		if denominator != 0 {
			result[i] = ((highestHigh - close[i]) / denominator) * -100
		} else {
			result[i] = 0
		}
	}

	return result
}
