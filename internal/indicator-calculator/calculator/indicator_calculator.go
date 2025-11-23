package calculator

import (
	"fmt"
	"time"

	"github.com/stock-market-system/pkg/indicators"
	"github.com/stock-market-system/pkg/models"
)

// IndicatorCalculator calculates technical indicators
type IndicatorCalculator struct {
	benchmarkData map[string][]models.StockQuote // Benchmark index data for RS calculation
}

// NewIndicatorCalculator creates a new indicator calculator
func NewIndicatorCalculator() *IndicatorCalculator {
	return &IndicatorCalculator{
		benchmarkData: make(map[string][]models.StockQuote),
	}
}

// SetBenchmarkData sets benchmark index data for relative strength calculations
func (ic *IndicatorCalculator) SetBenchmarkData(indexName string, data []models.StockQuote) {
	ic.benchmarkData[indexName] = data
}

// CalculateAll calculates all technical indicators for a ticker
func (ic *IndicatorCalculator) CalculateAll(quotes []models.StockQuote, exchange models.Exchange, timeframe models.Timeframe) ([]models.TechnicalIndicators, error) {
	if len(quotes) < 200 {
		return nil, fmt.Errorf("insufficient data points: need at least 200, got %d", len(quotes))
	}

	// Extract OHLCV data
	opens, highs, lows, closes, volumes, values := extractOHLCV(quotes)

	// Calculate moving averages
	ema10 := indicators.EMA(closes, 10)
	ema30 := indicators.EMA(closes, 30)
	ema50 := indicators.EMA(closes, 50)
	ema100 := indicators.EMA(closes, 100)
	ema200 := indicators.EMA(closes, 200)
	sma20 := indicators.SMA(closes, 20)
	sma50 := indicators.SMA(closes, 50)

	// Calculate momentum indicators
	rsi14 := indicators.RSI(closes, 14)
	williamsR14 := indicators.WilliamsR(highs, lows, closes, 14)

	// Calculate volume indicators
	volumeSMA20 := indicators.SMA(toFloat64Slice(volumes), 20)
	relativeVolume := indicators.RelativeVolume(toFloat64Slice(volumes), 20)
	vzo14 := indicators.VZO(closes, toFloat64Slice(volumes), 14)
	upDownRatio50 := indicators.UpDownRatio(closes, toFloat64Slice(volumes), 50)

	// Calculate volatility indicators
	atr14 := indicators.ATR(highs, lows, closes, 14)
	volatility15 := indicators.Volatility(closes, 15)
	rmv := indicators.RMV(highs, lows, closes, 15)

	// Calculate Mansfield RS (if benchmark data available)
	var mansfieldRS, mansfieldRSMA []float64
	if benchmarkData, ok := ic.benchmarkData[getBenchmarkName(exchange)]; ok && len(benchmarkData) > 0 {
		benchmarkCloses := extractCloses(benchmarkData)
		mansfieldRS, mansfieldRSMA = indicators.MansfieldRelativeStrength(closes, benchmarkCloses, 52, true)
	} else {
		mansfieldRS = make([]float64, len(closes))
		mansfieldRSMA = make([]float64, len(closes))
	}

	// Calculate MFI
	mfi58 := indicators.MFI(highs, lows, closes, toFloat64Slice(volumes), 58)

	// Calculate Warrior Trend
	warriorTrend, warriorEMARSI, warriorTrendLength := indicators.WarriorTrendIndicator(closes, 90, 33)

	// Calculate Trend Strength (TSD)
	tsdBlue, tsdBlack, tsdTrend, tsdLength := indicators.TrendStrengthIndicator(highs, lows, closes, 10, 21)

	// Calculate Stock Stages
	stockStages := indicators.DetermineStockStages(opens, highs, lows, closes)

	// Calculate AVWAP bands
	dates := extractDates(quotes)
	anchorDates := indicators.GetAnchorDates(time.Now())

	var avwap1Y, avwap1YStd, avwap3Y, avwap3YStd, avwap5Y, avwap5YStd []float64

	if len(anchorDates) >= 3 {
		typicalPrices := make([]float64, len(closes))
		for i := range closes {
			typicalPrices[i] = (highs[i] + lows[i] + closes[i]) / 3.0
		}

		// 1Y AVWAP
		avwapBands1Y := indicators.CalculateAVWAPBands(dates, typicalPrices, toFloat64Slice(volumes), anchorDates[0], []float64{1, 2, 3})
		if avwapBands1Y != nil {
			avwap1Y = avwapBands1Y.AVWAP
			avwap1YStd = avwapBands1Y.StdDev
		}

		// 3Y AVWAP
		avwapBands3Y := indicators.CalculateAVWAPBands(dates, typicalPrices, toFloat64Slice(volumes), anchorDates[1], []float64{1, 2, 3})
		if avwapBands3Y != nil {
			avwap3Y = avwapBands3Y.AVWAP
			avwap3YStd = avwapBands3Y.StdDev
		}

		// 5Y AVWAP
		avwapBands5Y := indicators.CalculateAVWAPBands(dates, typicalPrices, toFloat64Slice(volumes), anchorDates[2], []float64{1, 2, 3})
		if avwapBands5Y != nil {
			avwap5Y = avwapBands5Y.AVWAP
			avwap5YStd = avwapBands5Y.StdDev
		}
	}

	// Build results
	results := make([]models.TechnicalIndicators, 0, len(quotes))

	for i := range quotes {
		ind := models.TechnicalIndicators{
			Ticker:    quotes[i].Ticker,
			Exchange:  quotes[i].Exchange,
			TradeDate: quotes[i].TradeDate,
			Timeframe: timeframe,
			CreatedAt: time.Now(),
		}

		// Moving averages
		if i < len(ema10) {
			ind.EMA10 = ema10[i]
		}
		if i < len(ema30) {
			ind.EMA30 = ema30[i]
		}
		if i < len(ema50) {
			ind.EMA50 = ema50[i]
		}
		if i < len(ema100) {
			ind.EMA100 = ema100[i]
		}
		if i < len(ema200) {
			ind.EMA200 = ema200[i]
		}
		if i < len(sma20) {
			ind.SMA20 = sma20[i]
		}
		if i < len(sma50) {
			ind.SMA50 = sma50[i]
		}

		// Momentum
		if i < len(rsi14) {
			ind.RSI14 = rsi14[i]
		}
		if i < len(williamsR14) {
			ind.WilliamsR14 = williamsR14[i]
		}

		// Volume
		if i < len(volumeSMA20) {
			ind.VolumeSMA20 = uint64(volumeSMA20[i])
		}
		if i < len(relativeVolume) {
			ind.RelativeVolume = relativeVolume[i]
		}
		if i < len(vzo14) {
			ind.VZO14 = vzo14[i]
		}
		if i < len(upDownRatio50) {
			ind.UpDownRatio50 = upDownRatio50[i]
		}

		// Volatility
		if i < len(atr14) {
			ind.ATR14 = atr14[i]
		}
		if i < len(volatility15) {
			ind.Volatility15 = volatility15[i]
		}
		if i < len(rmv) {
			ind.RMV = rmv[i]
		}

		// Relative Strength
		if i < len(mansfieldRS) {
			ind.MansfieldRS = mansfieldRS[i]
		}
		if i < len(mansfieldRSMA) {
			ind.MansfieldRSMA = mansfieldRSMA[i]
		}

		// Money Flow
		if i < len(mfi58) {
			ind.MFI58 = mfi58[i]
		}

		// Warrior Trend
		if i < len(warriorTrend) {
			ind.WarriorTrend = warriorTrend[i]
		}
		if i < len(warriorEMARSI) {
			ind.WarriorEMARSI = warriorEMARSI[i]
		}
		if i < len(warriorTrendLength) {
			ind.WarriorTrendLength = uint32(warriorTrendLength[i])
		}

		// TSD
		if i < len(tsdTrend) {
			ind.TSDDirection = tsdTrend[i]
		}
		if i < len(tsdBlack) {
			ind.TSDLevel = tsdBlack[i]
		}
		if i < len(tsdLength) {
			ind.TSDLength = uint32(tsdLength[i])
		}

		// Stock Stage
		if i < len(stockStages) {
			ind.StockStage = stockStages[i]
		}

		// AVWAP
		if i < len(avwap1Y) {
			ind.AVWAP1Y = avwap1Y[i]
		}
		if i < len(avwap1YStd) {
			ind.AVWAP1YStd = avwap1YStd[i]
		}
		if i < len(avwap3Y) {
			ind.AVWAP3Y = avwap3Y[i]
		}
		if i < len(avwap3YStd) {
			ind.AVWAP3YStd = avwap3YStd[i]
		}
		if i < len(avwap5Y) {
			ind.AVWAP5Y = avwap5Y[i]
		}
		if i < len(avwap5YStd) {
			ind.AVWAP5YStd = avwap5YStd[i]
		}

		results = append(results, ind)
	}

	return results, nil
}

// Helper functions

func extractOHLCV(quotes []models.StockQuote) ([]float64, []float64, []float64, []float64, []uint64, []float64) {
	n := len(quotes)
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	volumes := make([]uint64, n)
	values := make([]float64, n)

	for i, q := range quotes {
		opens[i] = q.Open
		highs[i] = q.High
		lows[i] = q.Low
		closes[i] = q.Close
		volumes[i] = q.Volume
		values[i] = q.Value
	}

	return opens, highs, lows, closes, volumes, values
}

func extractCloses(quotes []models.StockQuote) []float64 {
	closes := make([]float64, len(quotes))
	for i, q := range quotes {
		closes[i] = q.Close
	}
	return closes
}

func extractDates(quotes []models.StockQuote) []time.Time {
	dates := make([]time.Time, len(quotes))
	for i, q := range quotes {
		dates[i] = q.TradeDate
	}
	return dates
}

func toFloat64Slice(volumes []uint64) []float64 {
	result := make([]float64, len(volumes))
	for i, v := range volumes {
		result[i] = float64(v)
	}
	return result
}

func getBenchmarkName(exchange models.Exchange) string {
	switch exchange {
	case models.ExchangeMOEX:
		return "IMOEX"
	case models.ExchangeNASDAQ, models.ExchangeNYSE:
		return "SPX"
	default:
		return "IMOEX"
	}
}
