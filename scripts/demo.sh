#!/bin/bash

# Demo script for Vancouver Trip Planner
set -e

echo "🚗 Vancouver Trip Planner - API Demo"
echo

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ Server is not running on localhost:8080"
    echo "   Start it with: ./scripts/run.sh"
    exit 1
fi

echo "✅ Server is running!"
echo

echo "🏥 Testing health endpoint..."
curl -s http://localhost:8080/health | jq .
echo
echo

echo "🗺️  Testing trip planning with Vancouver destinations..."
echo "Request: Science World → Granville Island"

curl -s -X POST http://localhost:8080/api/v1/trips/plan \
  -H "Content-Type: application/json" \
  -d '{
    "stops": [
      {
        "address": "Science World, Vancouver, BC",
        "duration_minutes": 120
      },
      {
        "address": "Granville Island, Vancouver, BC",
        "duration_minutes": 90
      }
    ],
    "start_time": "2024-01-20T14:00:00-08:00",
    "preferences": {
      "cost_weight": 0.6,
      "time_weight": 0.4
    }
  }' | jq .

echo
echo "✅ Demo completed!"
echo
echo "Try your own destinations:"
echo "curl -X POST http://localhost:8080/api/v1/trips/plan \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"stops\": [{\"address\": \"Your Address\", \"duration_minutes\": 60}], \"start_time\": \"2024-01-20T10:00:00-08:00\"}'"