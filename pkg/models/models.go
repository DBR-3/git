package models

import "time"

// Exchange represents stock exchange
type Exchange string

const (
	ExchangeMOEX   Exchange = "MOEX"
	ExchangeNASDAQ Exchange = "NASDAQ"
	ExchangeNYSE   Exchange = "NYSE"
	ExchangeOTHER  Exchange = "OTHER"
)

// SecurityType represents type of security
type SecurityType string

const (
	SecurityTypeStock SecurityType = "STOCK"
	SecurityTypeETF   SecurityType = "ETF"
)

// Timeframe represents data timeframe
type Timeframe string

const (
	TimeframeDaily  Timeframe = "DAILY"
	TimeframeWeekly Timeframe = "WEEKLY"
)

// Ticker represents a stock ticker
type Ticker struct {
	Ticker       string       `json:"ticker" ch:"ticker"`
	Name         string       `json:"name" ch:"name"`
	Exchange     Exchange     `json:"exchange" ch:"exchange"`
	SecurityType SecurityType `json:"security_type" ch:"security_type"`
	Sector       string       `json:"sector" ch:"sector"`
	Industry     string       `json:"industry" ch:"industry"`
	IsActive     bool         `json:"is_active" ch:"is_active"`
	FirstSeen    time.Time    `json:"first_seen" ch:"first_seen"`
	LastUpdated  time.Time    `json:"last_updated" ch:"last_updated"`
}

// StockQuote represents OHLCV data
type StockQuote struct {
	Ticker    string    `json:"ticker" ch:"ticker"`
	Exchange  Exchange  `json:"exchange" ch:"exchange"`
	TradeDate time.Time `json:"trade_date" ch:"trade_date"`
	Open      float64   `json:"open" ch:"open"`
	High      float64   `json:"high" ch:"high"`
	Low       float64   `json:"low" ch:"low"`
	Close     float64   `json:"close" ch:"close"`
	Volume    uint64    `json:"volume" ch:"volume"`
	Value     float64   `json:"value" ch:"value"`         // For MOEX
	NumTrades uint32    `json:"num_trades" ch:"num_trades"` // For MOEX
	CreatedAt time.Time `json:"created_at" ch:"created_at"`
}

// BenchmarkIndex represents benchmark index data
type BenchmarkIndex struct {
	IndexName string    `json:"index_name" ch:"index_name"`
	TradeDate time.Time `json:"trade_date" ch:"trade_date"`
	Open      float64   `json:"open" ch:"open"`
	High      float64   `json:"high" ch:"high"`
	Low       float64   `json:"low" ch:"low"`
	Close     float64   `json:"close" ch:"close"`
	Volume    uint64    `json:"volume" ch:"volume"`
	CreatedAt time.Time `json:"created_at" ch:"created_at"`
}

// TechnicalIndicators represents all technical indicators for a stock
type TechnicalIndicators struct {
	Ticker    string    `json:"ticker" ch:"ticker"`
	Exchange  Exchange  `json:"exchange" ch:"exchange"`
	TradeDate time.Time `json:"trade_date" ch:"trade_date"`
	Timeframe Timeframe `json:"timeframe" ch:"timeframe"`

	// Trend Indicators
	EMA10  float64 `json:"ema_10" ch:"ema_10"`
	EMA30  float64 `json:"ema_30" ch:"ema_30"`
	EMA50  float64 `json:"ema_50" ch:"ema_50"`
	EMA100 float64 `json:"ema_100" ch:"ema_100"`
	EMA200 float64 `json:"ema_200" ch:"ema_200"`
	SMA20  float64 `json:"sma_20" ch:"sma_20"`
	SMA50  float64 `json:"sma_50" ch:"sma_50"`

	// Momentum Indicators
	RSI14       float64 `json:"rsi_14" ch:"rsi_14"`
	WilliamsR14 float64 `json:"williams_r_14" ch:"williams_r_14"`

	// Volume Indicators
	VolumeSMA20    uint64  `json:"volume_sma_20" ch:"volume_sma_20"`
	RelativeVolume float64 `json:"relative_volume" ch:"relative_volume"`
	VZO14          float64 `json:"vzo_14" ch:"vzo_14"`
	UpDownRatio50  float64 `json:"up_down_ratio_50" ch:"up_down_ratio_50"`

	// Volatility Indicators
	ATR14        float64 `json:"atr_14" ch:"atr_14"`
	Volatility15 float64 `json:"volatility_15" ch:"volatility_15"`
	RMV          float64 `json:"rmv" ch:"rmv"`

	// Relative Strength
	MansfieldRS   float64 `json:"mansfield_rs" ch:"mansfield_rs"`
	MansfieldRSMA float64 `json:"mansfield_rs_ma" ch:"mansfield_rs_ma"`

	// Money Flow
	MFI58 float64 `json:"mfi_58" ch:"mfi_58"`

	// Trend Strength
	WarriorTrend       string  `json:"warrior_trend" ch:"warrior_trend"`
	WarriorEMARSI      float64 `json:"warrior_ema_rsi" ch:"warrior_ema_rsi"`
	WarriorTrendLength uint32  `json:"warrior_trend_length" ch:"warrior_trend_length"`

	TSDDirection string  `json:"tsd_direction" ch:"tsd_direction"`
	TSDLevel     float64 `json:"tsd_level" ch:"tsd_level"`
	TSDLength    uint32  `json:"tsd_length" ch:"tsd_length"`

	// Stock Stage (Wyckoff)
	StockStage string `json:"stock_stage" ch:"stock_stage"`

	// AVWAP Bands
	AVWAP1Y    float64 `json:"avwap_1y" ch:"avwap_1y"`
	AVWAP1YStd float64 `json:"avwap_1y_std" ch:"avwap_1y_std"`
	AVWAP3Y    float64 `json:"avwap_3y" ch:"avwap_3y"`
	AVWAP3YStd float64 `json:"avwap_3y_std" ch:"avwap_3y_std"`
	AVWAP5Y    float64 `json:"avwap_5y" ch:"avwap_5y"`
	AVWAP5YStd float64 `json:"avwap_5y_std" ch:"avwap_5y_std"`

	CreatedAt time.Time `json:"created_at" ch:"created_at"`
}

// SignalType represents type of trading signal
type SignalType string

const (
	SignalRSCrossedZeroUp         SignalType = "RS_CROSSED_ZERO_UP"
	SignalRSCrossedZeroDown       SignalType = "RS_CROSSED_ZERO_DOWN"
	SignalRSGrowing3DaysPositive  SignalType = "RS_GROWING_3DAYS_POSITIVE"
	SignalRSGrowing3DaysNegative  SignalType = "RS_GROWING_3DAYS_NEGATIVE"
	SignalRSDeclining3DaysPositive SignalType = "RS_DECLINING_3DAYS_POSITIVE"
	SignalRSDeclining3DaysNegative SignalType = "RS_DECLINING_3DAYS_NEGATIVE"
	SignalStageChange             SignalType = "STAGE_CHANGE"
	SignalTrendChange             SignalType = "TREND_CHANGE"
	SignalVolumeSpike             SignalType = "VOLUME_SPIKE"
	SignalBreakout                SignalType = "BREAKOUT"
)

// TradingSignal represents a trading signal
type TradingSignal struct {
	SignalID       string     `json:"signal_id" ch:"signal_id"`
	Ticker         string     `json:"ticker" ch:"ticker"`
	Exchange       Exchange   `json:"exchange" ch:"exchange"`
	SignalDate     time.Time  `json:"signal_date" ch:"signal_date"`
	SignalType     SignalType `json:"signal_type" ch:"signal_type"`
	SignalStrength float32    `json:"signal_strength" ch:"signal_strength"` // 0-100
	Description    string     `json:"description" ch:"description"`
	Metadata       string     `json:"metadata" ch:"metadata"` // JSON
	CreatedAt      time.Time  `json:"created_at" ch:"created_at"`
}

// JobType represents type of job
type JobType string

const (
	JobTypeTickerFetch          JobType = "TICKER_FETCH"
	JobTypeDataCollection       JobType = "DATA_COLLECTION"
	JobTypeIndicatorCalculation JobType = "INDICATOR_CALCULATION"
	JobTypeSignalGeneration     JobType = "SIGNAL_GENERATION"
)

// JobStatus represents job execution status
type JobStatus string

const (
	JobStatusStarted JobStatus = "STARTED"
	JobStatusSuccess JobStatus = "SUCCESS"
	JobStatusFailed  JobStatus = "FAILED"
	JobStatusPartial JobStatus = "PARTIAL"
)

// JobExecutionLog represents job execution log
type JobExecutionLog struct {
	JobID            string    `json:"job_id" ch:"job_id"`
	JobType          JobType   `json:"job_type" ch:"job_type"`
	Exchange         Exchange  `json:"exchange" ch:"exchange"`
	Status           JobStatus `json:"status" ch:"status"`
	StartTime        time.Time `json:"start_time" ch:"start_time"`
	EndTime          time.Time `json:"end_time" ch:"end_time"`
	DurationSeconds  uint32    `json:"duration_seconds" ch:"duration_seconds"`
	TickersProcessed uint32    `json:"tickers_processed" ch:"tickers_processed"`
	TickersFailed    uint32    `json:"tickers_failed" ch:"tickers_failed"`
	ErrorMessage     string    `json:"error_message" ch:"error_message"`
	CreatedAt        time.Time `json:"created_at" ch:"created_at"`
}

// KafkaMessage represents message structure for Kafka
type KafkaMessage struct {
	MessageID string      `json:"message_id"`
	Timestamp time.Time   `json:"timestamp"`
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
}

// DataCollectedEvent represents event when data is collected
type DataCollectedEvent struct {
	Ticker    string       `json:"ticker"`
	Exchange  Exchange     `json:"exchange"`
	TradeDate time.Time    `json:"trade_date"`
	Quotes    []StockQuote `json:"quotes"`
}

// IndicatorCalculatedEvent represents event when indicators are calculated
type IndicatorCalculatedEvent struct {
	Ticker     string              `json:"ticker"`
	Exchange   Exchange            `json:"exchange"`
	TradeDate  time.Time           `json:"trade_date"`
	Timeframe  Timeframe           `json:"timeframe"`
	Indicators TechnicalIndicators `json:"indicators"`
}
