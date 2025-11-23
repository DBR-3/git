#!/bin/bash

# Test Data Collector Service
# This script tests the data collector API endpoints

set -e

BASE_URL="http://localhost:8082"

echo "=== Testing Data Collector Service ==="
echo ""

# Test health check
echo "1. Testing health check..."
curl -s "${BASE_URL}/health" | jq '.'
echo ""

# Test data collection for MOEX tickers
echo "2. Starting data collection for MOEX tickers..."
START_DATE=$(date -d "6 months ago" +%Y-%m-%d)
END_DATE=$(date +%Y-%m-%d)

echo "Date range: ${START_DATE} to ${END_DATE}"
echo ""

# Test with small list of tickers
TICKERS='["SBER", "GAZP", "LKOH", "YNDX", "GMKN"]'

curl -X POST "${BASE_URL}/api/v1/collect" \
  -H "Content-Type: application/json" \
  -d "{
    \"exchange\": \"MOEX\",
    \"tickers\": ${TICKERS},
    \"start_date\": \"${START_DATE}\",
    \"end_date\": \"${END_DATE}\"
  }" | jq '.'

echo ""
echo "3. Data collection started in background."
echo "Check logs to see progress."
echo ""

# Wait a bit and check last collection date
echo "4. Waiting 10 seconds..."
sleep 10

echo "5. Checking last collection date for SBER..."
curl -s "${BASE_URL}/api/v1/last-date/MOEX/SBER" | jq '.'

echo ""
echo "=== Test completed ==="
echo ""
echo "To monitor progress:"
echo "  - Check service logs"
echo "  - Check Kafka UI: http://localhost:8080"
echo "  - Check ClickHouse: make db-console"
echo "    SELECT count(*) FROM stock_quotes WHERE ticker IN ('SBER', 'GAZP', 'LKOH', 'YNDX', 'GMKN');"
