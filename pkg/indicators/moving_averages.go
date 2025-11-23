package indicators

// SMA calculates Simple Moving Average
func SMA(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(data))

	// Calculate first SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate subsequent SMAs
	for i := period; i < len(data); i++ {
		sum = sum - data[i-period] + data[i]
		result[i] = sum / float64(period)
	}

	return result
}

// EMA calculates Exponential Moving Average
func EMA(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, 0)
	}

	result := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)

	// First EMA is SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate subsequent EMAs
	for i := period; i < len(data); i++ {
		result[i] = (data[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}

// CalculateAllEMAs calculates multiple EMAs at once
func CalculateAllEMAs(closes []float64) map[int][]float64 {
	periods := []int{10, 30, 50, 100, 200}
	result := make(map[int][]float64)

	for _, period := range periods {
		result[period] = EMA(closes, period)
	}

	return result
}

// CalculateAllSMAs calculates multiple SMAs at once
func CalculateAllSMAs(data []float64) map[int][]float64 {
	periods := []int{20, 50}
	result := make(map[int][]float64)

	for _, period := range periods {
		result[period] = SMA(data, period)
	}

	return result
}
