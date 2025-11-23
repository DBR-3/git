package indicators

import "math"

// MFI calculates Money Flow Index
func MFI(high, low, close, volume []float64, period int) []float64 {
	if len(high) < period || len(low) < period || len(close) < period || len(volume) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(high))

	// Calculate typical price
	typicalPrice := make([]float64, len(high))
	for i := 0; i < len(high); i++ {
		typicalPrice[i] = (high[i] + low[i] + close[i]) / 3.0
	}

	// Calculate raw money flow
	rawMoneyFlow := make([]float64, len(high))
	for i := 0; i < len(high); i++ {
		rawMoneyFlow[i] = typicalPrice[i] * volume[i]
	}

	// Calculate change in typical price
	changeInTypicalPrice := make([]float64, len(high))
	for i := 1; i < len(high); i++ {
		changeInTypicalPrice[i] = typicalPrice[i] - typicalPrice[i-1]
	}

	// Calculate MFI
	for i := period; i < len(high); i++ {
		positiveFlow := 0.0
		negativeFlow := 0.0

		for j := i - period + 1; j <= i; j++ {
			if changeInTypicalPrice[j] <= 0 {
				positiveFlow += rawMoneyFlow[j]
			}
			if changeInTypicalPrice[j] >= 0 {
				negativeFlow += rawMoneyFlow[j]
			}
		}

		// Calculate MFI using the helper function
		var mfi float64
		if negativeFlow == 0 {
			mfi = 100
		} else if positiveFlow == 0 {
			mfi = 0
		} else {
			moneyRatio := positiveFlow / negativeFlow
			mfi = 100.0 - (100.0 / (1.0 + moneyRatio))
		}

		// Adjust as per the formula: (mfi - 50) * 3
		result[i] = (mfi - 50) * 3
	}

	return result
}

// VZO calculates Volume Zone Oscillator
func VZO(close, volume []float64, period int) []float64 {
	if len(close) < period || len(volume) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(close))

	// Calculate volume direction
	volumeDirection := make([]float64, len(close))
	for i := 1; i < len(close); i++ {
		if close[i] > close[i-1] {
			volumeDirection[i] = volume[i]
		} else {
			volumeDirection[i] = -volume[i]
		}
	}

	// Calculate EMA of volume direction
	vzoVolume := EMA(volumeDirection, period)
	totalVolume := EMA(volume, period)

	// Calculate VZO
	for i := 0; i < len(close); i++ {
		if totalVolume[i] != 0 {
			result[i] = 100 * vzoVolume[i] / totalVolume[i]
		}
	}

	return result
}

// UpDownRatio calculates the ratio of up volume to down volume
func UpDownRatio(close, volume []float64, period int) []float64 {
	if len(close) < period || len(volume) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(close))
	upVol := make([]float64, len(close))
	dnVol := make([]float64, len(close))

	// Calculate up and down volumes
	for i := 1; i < len(close); i++ {
		if close[i] > close[i-1] {
			upVol[i] = volume[i]
		} else if close[i] < close[i-1] {
			dnVol[i] = volume[i]
		}
	}

	// Calculate rolling sums
	sumUpVol := SMA(upVol, period)
	sumDnVol := SMA(dnVol, period)

	// Calculate ratio
	for i := 0; i < len(close); i++ {
		if sumDnVol[i] != 0 {
			result[i] = sumUpVol[i] / sumDnVol[i]
		}
	}

	return result
}

// AccumulationDistribution calculates Accumulation/Distribution indicator
func AccumulationDistribution(open, high, low, close, value []float64, trendPeriods int) []float64 {
	if len(open) < trendPeriods {
		return make([]float64, 0)
	}

	result := make([]float64, len(open))
	mfMultiplier := make([]float64, len(open))
	accDist := make([]float64, len(open))

	// Calculate Money Flow Multiplier and Accumulation/Distribution
	for i := 0; i < len(open); i++ {
		if high[i] != low[i] {
			mfm := ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i])
			ac := mfm * value[i]
			mfMultiplier[i] = mfm
			accDist[i] = ac
		}
	}

	// Calculate EMA of accumulation/distribution
	accDistEMA := EMA(accDist, trendPeriods)

	copy(result, accDistEMA)
	return result
}

// RelativeVolume calculates relative volume compared to average
func RelativeVolume(volume []float64, period int) []float64 {
	if len(volume) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(volume))
	avgVolume := SMA(volume, period)

	for i := 0; i < len(volume); i++ {
		if avgVolume[i] != 0 {
			result[i] = volume[i] / avgVolume[i]
		}
	}

	return result
}

// VolumeWeightedStdDev calculates volume-weighted standard deviation
func VolumeWeightedStdDev(prices, volumes []float64, vwap float64, startIdx, endIdx int) float64 {
	if endIdx <= startIdx {
		return 0
	}

	weightedSumSquaredDiff := 0.0
	totalVolume := 0.0

	for i := startIdx; i <= endIdx; i++ {
		diff := prices[i] - vwap
		weightedSumSquaredDiff += diff * diff * volumes[i]
		totalVolume += volumes[i]
	}

	if totalVolume == 0 {
		return 0
	}

	variance := weightedSumSquaredDiff / totalVolume
	return math.Sqrt(variance)
}
