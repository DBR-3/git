-- ClickHouse Schema for Stock Market Data Collection System
-- Database: stock_market

CREATE DATABASE IF NOT EXISTS stock_market;

USE stock_market;

-- ========================================
-- 1. TICKERS TABLE
-- ========================================
-- Stores list of all available tickers
CREATE TABLE IF NOT EXISTS tickers (
    ticker String,
    name String,
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'OTHER' = 4),
    security_type Enum8('STOCK' = 1, 'ETF' = 2),
    sector String DEFAULT '',
    industry String DEFAULT '',
    is_active UInt8 DEFAULT 1,
    first_seen DateTime DEFAULT now(),
    last_updated DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(last_updated)
ORDER BY (exchange, ticker)
SETTINGS index_granularity = 8192;

-- ========================================
-- 2. STOCK QUOTES (OHLCV) TABLE
-- ========================================
-- Main table for storing daily stock quotes
CREATE TABLE IF NOT EXISTS stock_quotes (
    ticker String,
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'OTHER' = 4),
    trade_date Date,
    open Decimal64(4),
    high Decimal64(4),
    low Decimal64(4),
    close Decimal64(4),
    volume UInt64,
    value Decimal64(2) DEFAULT 0, -- For MOEX (ruble value)
    num_trades UInt32 DEFAULT 0,  -- For MOEX
    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(trade_date)
ORDER BY (ticker, exchange, trade_date)
SETTINGS index_granularity = 8192;

-- Index for faster queries by date range
ALTER TABLE stock_quotes ADD INDEX idx_trade_date trade_date TYPE minmax GRANULARITY 3;

-- ========================================
-- 3. INTRADAY QUOTES (for future real-time support)
-- ========================================
CREATE TABLE IF NOT EXISTS stock_quotes_intraday (
    ticker String,
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'OTHER' = 4),
    timestamp DateTime,
    price Decimal64(4),
    volume UInt64,
    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (ticker, exchange, timestamp)
TTL timestamp + INTERVAL 30 DAY  -- Keep intraday data for 30 days
SETTINGS index_granularity = 8192;

-- ========================================
-- 4. WEEKLY AGGREGATED QUOTES
-- ========================================
-- Pre-aggregated weekly data for faster queries
CREATE TABLE IF NOT EXISTS stock_quotes_weekly (
    ticker String,
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'OTHER' = 4),
    week_start Date,
    open Decimal64(4),
    high Decimal64(4),
    low Decimal64(4),
    close Decimal64(4),
    volume UInt64,
    value Decimal64(2) DEFAULT 0,
    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(week_start)
ORDER BY (ticker, exchange, week_start)
SETTINGS index_granularity = 8192;

-- Materialized view to auto-populate weekly aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_stock_quotes_weekly
TO stock_quotes_weekly
AS SELECT
    ticker,
    exchange,
    toMonday(trade_date) as week_start,
    argMin(open, trade_date) as open,
    max(high) as high,
    min(low) as low,
    argMax(close, trade_date) as close,
    sum(volume) as volume,
    sum(value) as value,
    now() as created_at
FROM stock_quotes
GROUP BY ticker, exchange, week_start;

-- ========================================
-- 5. BENCHMARK INDICES
-- ========================================
-- Store benchmark indices (IMOEX, S&P500, NASDAQ Composite)
CREATE TABLE IF NOT EXISTS benchmark_indices (
    index_name String,  -- 'IMOEX', 'SPX', 'IXIC'
    trade_date Date,
    open Decimal64(4),
    high Decimal64(4),
    low Decimal64(4),
    close Decimal64(4),
    volume UInt64 DEFAULT 0,
    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(trade_date)
ORDER BY (index_name, trade_date)
SETTINGS index_granularity = 8192;

-- ========================================
-- 6. TECHNICAL INDICATORS TABLE
-- ========================================
-- Stores all calculated technical indicators
CREATE TABLE IF NOT EXISTS technical_indicators (
    ticker String,
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'OTHER' = 4),
    trade_date Date,
    timeframe Enum8('DAILY' = 1, 'WEEKLY' = 2),

    -- Trend Indicators
    ema_10 Decimal64(4) DEFAULT 0,
    ema_30 Decimal64(4) DEFAULT 0,
    ema_50 Decimal64(4) DEFAULT 0,
    ema_100 Decimal64(4) DEFAULT 0,
    ema_200 Decimal64(4) DEFAULT 0,
    sma_20 Decimal64(4) DEFAULT 0,
    sma_50 Decimal64(4) DEFAULT 0,

    -- Momentum Indicators
    rsi_14 Decimal64(4) DEFAULT 0,
    williams_r_14 Decimal64(4) DEFAULT 0,

    -- Volume Indicators
    volume_sma_20 UInt64 DEFAULT 0,
    relative_volume Decimal64(4) DEFAULT 0,
    vzo_14 Decimal64(4) DEFAULT 0,
    up_down_ratio_50 Decimal64(4) DEFAULT 0,

    -- Volatility Indicators
    atr_14 Decimal64(4) DEFAULT 0,
    volatility_15 Decimal64(4) DEFAULT 0,
    rmv Decimal64(4) DEFAULT 0,

    -- Relative Strength
    mansfield_rs Decimal64(4) DEFAULT 0,
    mansfield_rs_ma Decimal64(4) DEFAULT 0,

    -- Money Flow
    mfi_58 Decimal64(4) DEFAULT 0,

    -- Trend Strength
    warrior_trend String DEFAULT '',
    warrior_ema_rsi Decimal64(4) DEFAULT 0,
    warrior_trend_length UInt32 DEFAULT 0,

    tsd_direction String DEFAULT '',
    tsd_level Decimal64(4) DEFAULT 0,
    tsd_length UInt32 DEFAULT 0,

    -- Stock Stage (Wyckoff)
    stock_stage String DEFAULT '',

    -- AVWAP Bands
    avwap_1y Decimal64(4) DEFAULT 0,
    avwap_1y_std Decimal64(4) DEFAULT 0,
    avwap_3y Decimal64(4) DEFAULT 0,
    avwap_3y_std Decimal64(4) DEFAULT 0,
    avwap_5y Decimal64(4) DEFAULT 0,
    avwap_5y_std Decimal64(4) DEFAULT 0,

    created_at DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(trade_date)
ORDER BY (ticker, exchange, timeframe, trade_date)
SETTINGS index_granularity = 8192;

-- ========================================
-- 7. SIGNALS TABLE
-- ========================================
-- Stores generated trading signals
CREATE TABLE IF NOT EXISTS trading_signals (
    signal_id UUID DEFAULT generateUUIDv4(),
    ticker String,
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'OTHER' = 4),
    signal_date Date,
    signal_type Enum8(
        'RS_CROSSED_ZERO_UP' = 1,
        'RS_CROSSED_ZERO_DOWN' = 2,
        'RS_GROWING_3DAYS_POSITIVE' = 3,
        'RS_GROWING_3DAYS_NEGATIVE' = 4,
        'RS_DECLINING_3DAYS_POSITIVE' = 5,
        'RS_DECLINING_3DAYS_NEGATIVE' = 6,
        'STAGE_CHANGE' = 7,
        'TREND_CHANGE' = 8,
        'VOLUME_SPIKE' = 9,
        'BREAKOUT' = 10
    ),
    signal_strength Decimal32(2) DEFAULT 0, -- 0-100
    description String,
    metadata String DEFAULT '', -- JSON with additional data
    created_at DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(signal_date)
ORDER BY (signal_date, ticker, signal_type)
SETTINGS index_granularity = 8192;

-- ========================================
-- 8. JOB EXECUTION LOG
-- ========================================
-- Tracks execution of data collection and calculation jobs
CREATE TABLE IF NOT EXISTS job_execution_log (
    job_id UUID DEFAULT generateUUIDv4(),
    job_type Enum8(
        'TICKER_FETCH' = 1,
        'DATA_COLLECTION' = 2,
        'INDICATOR_CALCULATION' = 3,
        'SIGNAL_GENERATION' = 4
    ),
    exchange Enum8('MOEX' = 1, 'NASDAQ' = 2, 'NYSE' = 3, 'ALL' = 4),
    status Enum8('STARTED' = 1, 'SUCCESS' = 2, 'FAILED' = 3, 'PARTIAL' = 4),
    start_time DateTime,
    end_time DateTime DEFAULT now(),
    duration_seconds UInt32,
    tickers_processed UInt32 DEFAULT 0,
    tickers_failed UInt32 DEFAULT 0,
    error_message String DEFAULT '',
    created_at DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (start_time, job_type)
TTL start_time + INTERVAL 90 DAY  -- Keep logs for 90 days
SETTINGS index_granularity = 8192;

-- ========================================
-- HELPER VIEWS
-- ========================================

-- View for latest quotes
CREATE VIEW IF NOT EXISTS v_latest_quotes AS
SELECT
    ticker,
    exchange,
    trade_date,
    open,
    high,
    low,
    close,
    volume,
    value
FROM stock_quotes
WHERE trade_date = (
    SELECT max(trade_date)
    FROM stock_quotes AS sq2
    WHERE sq2.ticker = stock_quotes.ticker
      AND sq2.exchange = stock_quotes.exchange
);

-- View for latest indicators with signals
CREATE VIEW IF NOT EXISTS v_latest_indicators_with_signals AS
SELECT
    ti.ticker,
    ti.exchange,
    ti.trade_date,
    ti.timeframe,
    ti.mansfield_rs,
    ti.rsi_14,
    ti.williams_r_14,
    ti.warrior_trend,
    ti.warrior_ema_rsi,
    ti.stock_stage,
    ti.relative_volume,
    ti.vzo_14,
    sq.close as price,
    sq.volume,
    COUNT(ts.signal_id) as signals_count
FROM technical_indicators AS ti
LEFT JOIN stock_quotes AS sq
    ON ti.ticker = sq.ticker
    AND ti.exchange = sq.exchange
    AND ti.trade_date = sq.trade_date
LEFT JOIN trading_signals AS ts
    ON ti.ticker = ts.ticker
    AND ti.exchange = ts.exchange
    AND ti.trade_date = ts.signal_date
WHERE ti.trade_date = (
    SELECT max(trade_date)
    FROM technical_indicators AS ti2
    WHERE ti2.ticker = ti.ticker
      AND ti2.exchange = ti.exchange
      AND ti2.timeframe = ti.timeframe
)
GROUP BY
    ti.ticker, ti.exchange, ti.trade_date, ti.timeframe,
    ti.mansfield_rs, ti.rsi_14, ti.williams_r_14,
    ti.warrior_trend, ti.warrior_ema_rsi, ti.stock_stage,
    ti.relative_volume, ti.vzo_14, sq.close, sq.volume;

-- ========================================
-- GRANTS (adjust user as needed)
-- ========================================
-- CREATE USER IF NOT EXISTS stock_app IDENTIFIED BY 'secure_password';
-- GRANT SELECT, INSERT, ALTER UPDATE, DELETE ON stock_market.* TO stock_app;
