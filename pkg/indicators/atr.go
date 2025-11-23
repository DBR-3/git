package indicators

import "math"

// ATR calculates Average True Range
func ATR(high, low, close []float64, period int) []float64 {
	if len(high) < period || len(low) < period || len(close) < period {
		return make([]float64, 0)
	}

	// Calculate True Range
	tr := make([]float64, len(high))
	tr[0] = high[0] - low[0]

	for i := 1; i < len(high); i++ {
		highLow := high[i] - low[i]
		highClose := math.Abs(high[i] - close[i-1])
		lowClose := math.Abs(low[i] - close[i-1])

		tr[i] = math.Max(highLow, math.Max(highClose, lowClose))
	}

	// Calculate ATR using Wilder's smoothing
	return RMA(tr, period)
}

// RMV calculates Relative Market Volatility
func RMV(high, low, close []float64, lookbackPeriod int) []float64 {
	if len(high) < 144 || len(low) < 144 || len(close) < 144 {
		return make([]float64, 0)
	}

	result := make([]float64, len(high))

	// Short-term ATRs
	shortATR1 := ATR(high, low, close, 3)
	shortATR2 := ATR(high, low, close, 5)
	shortATR3 := ATR(high, low, close, 8)

	// Long-term ATRs
	longATR1 := ATR(high, low, close, 55)
	longATR2 := ATR(high, low, close, 89)
	longATR3 := ATR(high, low, close, 144)

	// Calculate averages
	shortAvg := make([]float64, len(high))
	longAvg := make([]float64, len(high))
	combinedATR := make([]float64, len(high))

	for i := 0; i < len(high); i++ {
		if i < 144 {
			continue
		}

		shortAvg[i] = (shortATR1[i] + shortATR2[i] + shortATR3[i]) / 3.0
		longAvg[i] = (longATR1[i] + longATR2[i] + longATR3[i]) / 3.0
		combinedATR[i] = (shortAvg[i] + longAvg[i]) / 2.0
	}

	// Calculate RMV
	for i := lookbackPeriod; i < len(high); i++ {
		// Find highest and lowest combined ATR over lookback period
		highest := combinedATR[i]
		lowest := combinedATR[i]

		for j := i - lookbackPeriod; j < i; j++ {
			if combinedATR[j] > highest {
				highest = combinedATR[j]
			}
			if combinedATR[j] < lowest && combinedATR[j] > 0 {
				lowest = combinedATR[j]
			}
		}

		// Calculate RMV
		denominator := highest - lowest
		if denominator > 0.001 {
			result[i] = ((combinedATR[i] - lowest) / denominator) * 100
		} else {
			result[i] = 0
		}
	}

	return result
}

// Volatility calculates price volatility (standard deviation of returns)
func Volatility(prices []float64, period int) []float64 {
	if len(prices) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(prices))

	// Calculate returns
	returns := make([]float64, len(prices))
	for i := 1; i < len(prices); i++ {
		if prices[i-1] != 0 {
			returns[i] = (prices[i] - prices[i-1]) / prices[i-1]
		}
	}

	// Calculate rolling standard deviation
	for i := period; i < len(prices); i++ {
		// Calculate mean
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += returns[j]
		}
		mean := sum / float64(period)

		// Calculate variance
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := returns[j] - mean
			variance += diff * diff
		}
		variance /= float64(period)

		result[i] = math.Sqrt(variance)
	}

	return result
}
