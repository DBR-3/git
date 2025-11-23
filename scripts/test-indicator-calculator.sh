#!/bin/bash

# Test Indicator Calculator Service
# This script tests the indicator calculator API endpoints

set -e

BASE_URL="http://localhost:8083"

echo "=== Testing Indicator Calculator Service ==="
echo ""

# Test health check
echo "1. Testing health check..."
curl -s "${BASE_URL}/health" | jq '.'
echo ""

# Test manual indicator calculation for MOEX tickers
echo "2. Starting manual indicator calculation for MOEX tickers..."
START_DATE=$(date -d "1 year ago" +%Y-%m-%d)
END_DATE=$(date +%Y-%m-%d)

echo "Date range: ${START_DATE} to ${END_DATE}"
echo ""

# Test with small list of tickers
TICKERS='["SBER", "GAZP", "LKOH"]'

echo "Triggering calculation for tickers: ${TICKERS}"
RESULT=$(curl -s -X POST "${BASE_URL}/api/v1/calculate" \
  -H "Content-Type: application/json" \
  -d "{
    \"exchange\": \"MOEX\",
    \"tickers\": ${TICKERS},
    \"start_date\": \"${START_DATE}\",
    \"end_date\": \"${END_DATE}\"
  }")

echo "${RESULT}" | jq '.'
echo ""

# Extract job_id from result
JOB_ID=$(echo "${RESULT}" | jq -r '.job_id')

echo "3. Calculation completed."
echo "Job ID: ${JOB_ID}"
echo ""

# Wait a bit for data to be written
echo "4. Waiting 5 seconds for data to be written..."
sleep 5

echo "5. Verifying calculated indicators in ClickHouse..."
echo ""
echo "To check calculated indicators, run:"
echo "  make db-console"
echo ""
echo "Then execute:"
echo "  SELECT ticker, trade_date, ema10, ema50, rsi14, mansfield_rs, stock_stage"
echo "  FROM technical_indicators"
echo "  WHERE ticker IN ('SBER', 'GAZP', 'LKOH')"
echo "  ORDER BY ticker, trade_date DESC"
echo "  LIMIT 30;"
echo ""

echo "6. Checking for generated trading signals..."
echo ""
echo "To check generated signals, run:"
echo "  make db-console"
echo ""
echo "Then execute:"
echo "  SELECT ticker, signal_date, signal_type, indicator, description, confidence"
echo "  FROM trading_signals"
echo "  WHERE ticker IN ('SBER', 'GAZP', 'LKOH')"
echo "  ORDER BY signal_date DESC"
echo "  LIMIT 20;"
echo ""

echo "=== Test completed ===\"
echo ""
echo "To monitor Kafka messages:"
echo "  - Check indicators.calculated topic:"
echo "    docker exec -it kafka kafka-console-consumer \\"
echo "      --bootstrap-server localhost:9092 \\"
echo "      --topic stock.indicators.calculated \\"
echo "      --from-beginning"
echo ""
echo "  - Check signals.generated topic:"
echo "    docker exec -it kafka kafka-console-consumer \\"
echo "      --bootstrap-server localhost:9092 \\"
echo "      --topic stock.signals.generated \\"
echo "      --from-beginning"
echo ""

echo "To check service logs:"
echo "  docker logs -f indicator-calculator"
echo ""
