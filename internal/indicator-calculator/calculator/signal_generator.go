package calculator

import (
	"fmt"
	"time"

	"github.com/stock-market-system/pkg/models"
)

// SignalGenerator generates trading signals based on technical indicators
type SignalGenerator struct{}

// NewSignalGenerator creates a new signal generator
func NewSignalGenerator() *SignalGenerator {
	return &SignalGenerator{}
}

// GenerateSignals generates trading signals for a set of technical indicators
func (sg *SignalGenerator) GenerateSignals(indicators []models.TechnicalIndicators) []models.TradingSignal {
	if len(indicators) < 3 {
		return nil
	}

	signals := make([]models.TradingSignal, 0)
	lastIdx := len(indicators) - 1

	// Get current and previous data points
	current := indicators[lastIdx]
	prev1 := indicators[lastIdx-1]
	prev2 := indicators[lastIdx-2]

	// 1. Mansfield RS Signals
	mansfieldSignals := sg.detectMansfieldSignals(current, prev1, prev2)
	signals = append(signals, mansfieldSignals...)

	// 2. Stage Change Signals
	stageSignals := sg.detectStageChanges(current, prev1)
	signals = append(signals, stageSignals...)

	// 3. Trend Change Signals
	trendSignals := sg.detectTrendChanges(current, prev1)
	signals = append(signals, trendSignals...)

	// 4. Volume Spike Signals
	volumeSignals := sg.detectVolumeSpikes(current)
	signals = append(signals, volumeSignals...)

	// 5. RSI Signals
	rsiSignals := sg.detectRSISignals(current)
	signals = append(signals, rsiSignals...)

	// 6. Moving Average Crossovers
	maCrossSignals := sg.detectMACrossover(current, prev1)
	signals = append(signals, maCrossSignals...)

	// 7. Williams %R Signals
	williamsSignals := sg.detectWilliamsRSignals(current)
	signals = append(signals, williamsSignals...)

	return signals
}

// detectMansfieldSignals detects Mansfield RS signal patterns
func (sg *SignalGenerator) detectMansfieldSignals(current, prev1, prev2 models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	// Zero cross from below (bullish)
	if prev1.MansfieldRS < 0 && current.MansfieldRS >= 0 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "Mansfield RS",
			Value:       current.MansfieldRS,
			Threshold:   0.0,
			Description: "Mansfield RS crossed above zero (relative strength improving)",
			Confidence:  0.75,
			CreatedAt:   time.Now(),
		})
	}

	// Zero cross from above (bearish)
	if prev1.MansfieldRS > 0 && current.MansfieldRS <= 0 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "Mansfield RS",
			Value:       current.MansfieldRS,
			Threshold:   0.0,
			Description: "Mansfield RS crossed below zero (relative weakness)",
			Confidence:  0.75,
			CreatedAt:   time.Now(),
		})
	}

	// 3-day positive pattern (strong bullish)
	if prev2.MansfieldRS < prev1.MansfieldRS && prev1.MansfieldRS < current.MansfieldRS && current.MansfieldRS > 0 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "Mansfield RS",
			Value:       current.MansfieldRS,
			Threshold:   prev2.MansfieldRS,
			Description: "3-day positive Mansfield RS pattern (strong uptrend)",
			Confidence:  0.85,
			CreatedAt:   time.Now(),
		})
	}

	// 3-day negative pattern (strong bearish)
	if prev2.MansfieldRS > prev1.MansfieldRS && prev1.MansfieldRS > current.MansfieldRS && current.MansfieldRS < 0 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "Mansfield RS",
			Value:       current.MansfieldRS,
			Threshold:   prev2.MansfieldRS,
			Description: "3-day negative Mansfield RS pattern (strong downtrend)",
			Confidence:  0.85,
			CreatedAt:   time.Now(),
		})
	}

	return signals
}

// detectStageChanges detects Wyckoff stage transitions
func (sg *SignalGenerator) detectStageChanges(current, prev models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	if current.StockStage == prev.StockStage {
		return signals
	}

	// Stage 2 entry (bullish)
	if current.StockStage == "Stage 2: Advancing Phase" && prev.StockStage != "Stage 2: Advancing Phase" {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "Stock Stage",
			Value:       2.0,
			Threshold:   0.0,
			Description: fmt.Sprintf("Entered Stage 2 (Advancing Phase) from %s", prev.StockStage),
			Confidence:  0.80,
			CreatedAt:   time.Now(),
		})
	}

	// Stage 4 entry (bearish)
	if current.StockStage == "Stage 4: Declining Phase" && prev.StockStage != "Stage 4: Declining Phase" {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "Stock Stage",
			Value:       4.0,
			Threshold:   0.0,
			Description: fmt.Sprintf("Entered Stage 4 (Declining Phase) from %s", prev.StockStage),
			Confidence:  0.80,
			CreatedAt:   time.Now(),
		})
	}

	// Stage 3 warning (distribution)
	if current.StockStage == "Stage 3: Topping Phase" && prev.StockStage == "Stage 2: Advancing Phase" {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "Stock Stage",
			Value:       3.0,
			Threshold:   0.0,
			Description: "Entered Stage 3 (Topping/Distribution) - consider taking profits",
			Confidence:  0.70,
			CreatedAt:   time.Now(),
		})
	}

	return signals
}

// detectTrendChanges detects trend direction changes
func (sg *SignalGenerator) detectTrendChanges(current, prev models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	// Warrior Trend change
	if current.WarriorTrend != prev.WarriorTrend {
		if current.WarriorTrend == "LONG" {
			signals = append(signals, models.TradingSignal{
				Ticker:      current.Ticker,
				Exchange:    current.Exchange,
				TradeDate:   current.TradeDate,
				Timeframe:   current.Timeframe,
				SignalType:  models.SignalBuy,
				Indicator:   "Warrior Trend",
				Value:       float64(current.WarriorTrendLength),
				Threshold:   0.0,
				Description: "Warrior Trend switched to LONG",
				Confidence:  0.75,
				CreatedAt:   time.Now(),
			})
		} else if current.WarriorTrend == "SHORT" {
			signals = append(signals, models.TradingSignal{
				Ticker:      current.Ticker,
				Exchange:    current.Exchange,
				TradeDate:   current.TradeDate,
				Timeframe:   current.Timeframe,
				SignalType:  models.SignalSell,
				Indicator:   "Warrior Trend",
				Value:       float64(current.WarriorTrendLength),
				Threshold:   0.0,
				Description: "Warrior Trend switched to SHORT",
				Confidence:  0.75,
				CreatedAt:   time.Now(),
			})
		}
	}

	// TSD Trend change
	if current.TSDDirection != prev.TSDDirection {
		if current.TSDDirection == "UP" {
			signals = append(signals, models.TradingSignal{
				Ticker:      current.Ticker,
				Exchange:    current.Exchange,
				TradeDate:   current.TradeDate,
				Timeframe:   current.Timeframe,
				SignalType:  models.SignalBuy,
				Indicator:   "TSD",
				Value:       current.TSDLevel,
				Threshold:   0.0,
				Description: "Trend Strength Direction changed to UP",
				Confidence:  0.70,
				CreatedAt:   time.Now(),
			})
		} else if current.TSDDirection == "DOWN" {
			signals = append(signals, models.TradingSignal{
				Ticker:      current.Ticker,
				Exchange:    current.Exchange,
				TradeDate:   current.TradeDate,
				Timeframe:   current.Timeframe,
				SignalType:  models.SignalSell,
				Indicator:   "TSD",
				Value:       current.TSDLevel,
				Threshold:   0.0,
				Description: "Trend Strength Direction changed to DOWN",
				Confidence:  0.70,
				CreatedAt:   time.Now(),
			})
		}
	}

	return signals
}

// detectVolumeSpikes detects unusual volume activity
func (sg *SignalGenerator) detectVolumeSpikes(current models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	// High relative volume (>= 2.0x average)
	if current.RelativeVolume >= 2.0 {
		signalType := models.SignalHold
		description := "High volume detected"

		// If volume spike with positive trend, it's bullish
		if current.WarriorTrend == "LONG" || current.TSDDirection == "UP" {
			signalType = models.SignalBuy
			description = "High volume with uptrend (accumulation)"
		} else if current.WarriorTrend == "SHORT" || current.TSDDirection == "DOWN" {
			signalType = models.SignalSell
			description = "High volume with downtrend (distribution)"
		}

		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  signalType,
			Indicator:   "Relative Volume",
			Value:       current.RelativeVolume,
			Threshold:   2.0,
			Description: description,
			Confidence:  0.65,
			CreatedAt:   time.Now(),
		})
	}

	// VZO extremes
	if current.VZO14 > 60 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "VZO",
			Value:       current.VZO14,
			Threshold:   60.0,
			Description: "Volume Zone Oscillator in bullish zone",
			Confidence:  0.60,
			CreatedAt:   time.Now(),
		})
	} else if current.VZO14 < -60 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "VZO",
			Value:       current.VZO14,
			Threshold:   -60.0,
			Description: "Volume Zone Oscillator in bearish zone",
			Confidence:  0.60,
			CreatedAt:   time.Now(),
		})
	}

	return signals
}

// detectRSISignals detects RSI overbought/oversold conditions
func (sg *SignalGenerator) detectRSISignals(current models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	// RSI oversold (<30)
	if current.RSI14 < 30 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "RSI",
			Value:       current.RSI14,
			Threshold:   30.0,
			Description: "RSI oversold - potential bounce",
			Confidence:  0.60,
			CreatedAt:   time.Now(),
		})
	}

	// RSI overbought (>70)
	if current.RSI14 > 70 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "RSI",
			Value:       current.RSI14,
			Threshold:   70.0,
			Description: "RSI overbought - potential pullback",
			Confidence:  0.60,
			CreatedAt:   time.Now(),
		})
	}

	return signals
}

// detectMACrossover detects moving average crossovers
func (sg *SignalGenerator) detectMACrossover(current, prev models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	// Golden cross: EMA10 crosses above EMA50
	if prev.EMA10 <= prev.EMA50 && current.EMA10 > current.EMA50 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "EMA Crossover",
			Value:       current.EMA10,
			Threshold:   current.EMA50,
			Description: "Golden cross: EMA10 crossed above EMA50",
			Confidence:  0.70,
			CreatedAt:   time.Now(),
		})
	}

	// Death cross: EMA10 crosses below EMA50
	if prev.EMA10 >= prev.EMA50 && current.EMA10 < current.EMA50 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "EMA Crossover",
			Value:       current.EMA10,
			Threshold:   current.EMA50,
			Description: "Death cross: EMA10 crossed below EMA50",
			Confidence:  0.70,
			CreatedAt:   time.Now(),
		})
	}

	// Price above/below EMA200 (long-term trend)
	// Only generate if this is a recent change to avoid duplicate signals
	if prev.EMA200 > 0 && current.EMA200 > 0 {
		// Crossed above EMA200
		prevClose := prev.EMA10 // Using EMA10 as proxy for close
		currClose := current.EMA10
		if prevClose <= prev.EMA200 && currClose > current.EMA200 {
			signals = append(signals, models.TradingSignal{
				Ticker:      current.Ticker,
				Exchange:    current.Exchange,
				TradeDate:   current.TradeDate,
				Timeframe:   current.Timeframe,
				SignalType:  models.SignalBuy,
				Indicator:   "EMA200",
				Value:       currClose,
				Threshold:   current.EMA200,
				Description: "Price crossed above EMA200 (long-term bullish)",
				Confidence:  0.75,
				CreatedAt:   time.Now(),
			})
		}

		// Crossed below EMA200
		if prevClose >= prev.EMA200 && currClose < current.EMA200 {
			signals = append(signals, models.TradingSignal{
				Ticker:      current.Ticker,
				Exchange:    current.Exchange,
				TradeDate:   current.TradeDate,
				Timeframe:   current.Timeframe,
				SignalType:  models.SignalSell,
				Indicator:   "EMA200",
				Value:       currClose,
				Threshold:   current.EMA200,
				Description: "Price crossed below EMA200 (long-term bearish)",
				Confidence:  0.75,
				CreatedAt:   time.Now(),
			})
		}
	}

	return signals
}

// detectWilliamsRSignals detects Williams %R signals
func (sg *SignalGenerator) detectWilliamsRSignals(current models.TechnicalIndicators) []models.TradingSignal {
	signals := make([]models.TradingSignal, 0)

	// Williams %R oversold (< -80)
	if current.WilliamsR14 < -80 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalBuy,
			Indicator:   "Williams %R",
			Value:       current.WilliamsR14,
			Threshold:   -80.0,
			Description: "Williams %R oversold",
			Confidence:  0.55,
			CreatedAt:   time.Now(),
		})
	}

	// Williams %R overbought (> -20)
	if current.WilliamsR14 > -20 {
		signals = append(signals, models.TradingSignal{
			Ticker:      current.Ticker,
			Exchange:    current.Exchange,
			TradeDate:   current.TradeDate,
			Timeframe:   current.Timeframe,
			SignalType:  models.SignalSell,
			Indicator:   "Williams %R",
			Value:       current.WilliamsR14,
			Threshold:   -20.0,
			Description: "Williams %R overbought",
			Confidence:  0.55,
			CreatedAt:   time.Now(),
		})
	}

	return signals
}
