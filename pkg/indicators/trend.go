package indicators

import "math"

// TrendStrengthIndicator calculates the Trend Strength (TSD) indicator
// Returns: direction, blue trend line, black trend line, trend, trend length
func TrendStrengthIndicator(high, low, close []float64, n1, n2 int) ([]float64, []float64, []string, []int) {
	if len(close) < n1+n2 {
		return []float64{}, []float64{}, []string{}, []int{}
	}

	// Calculate average price (HLC3)
	ap := make([]float64, len(close))
	for i := 0; i < len(close); i++ {
		ap[i] = (high[i] + low[i] + close[i]) / 3.0
	}

	// Calculate ESA (Exponential Smoothed Average)
	esa := EMA(ap, n1)

	// Calculate d (absolute difference)
	d := make([]float64, len(close))
	for i := 0; i < len(close); i++ {
		d[i] = math.Abs(ap[i] - esa[i])
	}

	// Calculate EMA of d
	dEMA := EMA(d, n1)

	// Calculate CI (Commodity Index)
	ci := make([]float64, len(close))
	for i := 0; i < len(close); i++ {
		if dEMA[i] != 0 {
			ci[i] = (ap[i] - esa[i]) / (0.015 * dEMA[i])
		}
	}

	// Calculate TCI (True Commodity Index) - blue trend line
	blueTr := EMA(ci, n2)

	// Calculate black trend line (SMA of blue trend)
	blackTr := SMA(blueTr, 4)

	// Determine trend direction
	trend := make([]string, len(close))
	trendLength := make([]int, len(close))

	for i := 1; i < len(close); i++ {
		if blueTr[i] >= blackTr[i] {
			trend[i] = "Рост"
		} else {
			trend[i] = "Снижение"
		}

		// Calculate trend length
		if i > 0 && trend[i] == trend[i-1] {
			trendLength[i] = trendLength[i-1] + 1
		} else {
			trendLength[i] = 1
		}
	}

	return blueTr, blackTr, trend, trendLength
}

// CountSequential counts sequential occurrences of the same value
func CountSequential(values []string) []int {
	result := make([]int, len(values))
	counter := 1

	for i := 0; i < len(values); i++ {
		if i > 0 {
			if values[i] == values[i-1] {
				counter++
			} else {
				counter = 1
			}
		}
		result[i] = counter
	}

	return result
}

// StockStage represents Wyckoff stock stage
type StockStage struct {
	Stage     string
	MA10      float64
	MA20      float64
	MA50      float64
	Price     float64
	PrevStage string
}

// DetermineStockStages calculates Wyckoff stock stages
func DetermineStockStages(open, high, low, close []float64) []string {
	if len(close) < 50 {
		return make([]string, 0)
	}

	result := make([]string, len(close))

	// Calculate moving averages
	ma50 := SMA(close, 50)
	ma20 := SMA(close, 20)
	ema10 := EMA(close, 10)

	for i := 50; i < len(close); i++ {
		if i < 1 {
			continue
		}

		price := close[i]
		prevPrice := close[i-1]
		prevMA50 := ma50[i-1]
		prevMA20 := ma20[i-1]
		prevEMA10 := ema10[i-1]

		// Stage logic
		prevStage := ""
		if i > 0 {
			prevStage = result[i-1]
		}

		// 1. Basing (1A, 1B): Sideways movement, MA50 flat, price near MA50
		ma50Change := math.Abs(ma50[i]-prevMA50) / ma50[i]
		priceMA50Diff := math.Abs(price-ma50[i]) / ma50[i]

		if ma50Change < 0.01 && priceMA50Diff < 0.05 {
			if ma20[i] > ma50[i] && ema10[i] > ma20[i] {
				result[i] = "Basing (1B)"
			} else {
				result[i] = "Basing (1A)"
			}
		} else if ma50[i] > prevMA50 && ma20[i] > ma50[i] && ema10[i] > ma20[i] && price > ema10[i] {
			// 2. Advancing (2A, 2B, 2C)
			if prevStage == "Basing (1A)" || prevStage == "Basing (1B)" {
				result[i] = "Advancing (2A)"
			} else {
				// Check if price is near recent high
				recentHigh := close[i]
				for j := i - 10; j < i; j++ {
					if j >= 0 && close[j] > recentHigh {
						recentHigh = close[j]
					}
				}
				if price > recentHigh*0.95 {
					result[i] = "Advancing (2C)"
				} else {
					result[i] = "Advancing (2B)"
				}
			}
		} else if ema10[i] < prevEMA10 && ma20[i] < prevMA20 && ema10[i] < ma20[i] {
			// 3. Distribution (3A, 3B)
			if prevStage[:9] == "Advancing" {
				result[i] = "Distribution (3A)"
			} else {
				result[i] = "Distribution (3B)"
			}
		} else if ma50[i] < prevMA50 && ma20[i] < ma50[i] && ema10[i] < ma20[i] && price < ema10[i] {
			// 4. Declining (4A, 4B, 4C)
			if prevStage[:12] == "Distribution" {
				result[i] = "Declining (4A)"
			} else {
				// Check if price is near recent low
				recentLow := close[i]
				for j := i - 10; j < i; j++ {
					if j >= 0 && close[j] < recentLow {
						recentLow = close[j]
					}
				}
				if price < recentLow*1.05 {
					result[i] = "Declining (4C)"
				} else {
					result[i] = "Declining (4B)"
				}
			}
		} else {
			// Continue previous stage if no clear signal
			if i > 0 && result[i-1] != "" {
				result[i] = result[i-1]
			}
		}
	}

	return result
}

// DetectStageChange detects if stage has changed
func DetectStageChange(stages []string) []bool {
	result := make([]bool, len(stages))

	for i := 1; i < len(stages); i++ {
		if stages[i] != stages[i-1] {
			result[i] = true
		}
	}

	return result
}
