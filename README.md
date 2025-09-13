# Vancouver Trip Planner 🚗

A smart trip planning API that helps you visit multiple destinations in Vancouver while minimizing both parking costs and travel time. Perfect for tourists, locals, or anyone trying to optimize their day out in the city!

## The Problem

Planning a multi-stop day in Vancouver is frustrating:
- Parking meters have different rates throughout the day
- Some spots are free after 10 PM, others cost $3.50/hour during peak times
- You never know how far you'll have to walk from parking to your destination
- Traffic changes throughout the day affecting travel time

## The Solution

This API takes your list of destinations and gives you 3 optimized route options:
- 💰 **Cheapest Route** - Saves you money on parking
- ⚡ **Fastest Route** - Gets you there in minimum time
- 🎯 **Balanced Route** - Best compromise between cost and time

## How to Use

### 1. Get API Keys
You need a Google Maps API key with these services enabled:
- Distance Matrix API (for travel times)
- Geocoding API (for address lookup)

Get yours at: https://console.cloud.google.com/apis/credentials

### 2. Run the Server
```bash
git clone https://github.com/joshhwuu/htn-2025.git
cd htn-2025

# Add your API key
export GOOGLE_MAPS_API_KEY="your-key-here"

# Easy way - use the script
./scripts/run.sh

# Or use make
make run

# Or traditional way
go mod tidy && go run cmd/main.go
```

### 3. Plan Your Trip
```bash
curl -X POST http://localhost:8080/api/v1/trips/plan \
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
    "start_time": "2024-01-20T14:00:00-08:00"
  }'
```

You'll get back 3 route options with total costs, travel times, and parking details for each stop.

## API Documentation

See [docs/API.md](docs/API.md) for complete API reference.

**Main Endpoints:**
- `POST /api/v1/trips/plan` - Plan your multi-stop trip
- `GET /health` - Check if server is running

## Example Response

```json
{
  "plans": [
    {
      "type": "cheapest",
      "total_cost": 8.50,
      "total_time_minutes": 195,
      "route": [...]
    },
    {
      "type": "fastest",
      "total_cost": 12.00,
      "total_time_minutes": 150,
      "route": [...]
    },
    {
      "type": "hybrid",
      "total_cost": 10.25,
      "total_time_minutes": 170,
      "route": [...]
    }
  ]
}
```

## Vancouver Parking Info

The system uses real Vancouver parking meter data:

| Time Period | Weekdays | Weekends |
|-------------|----------|----------|
| 9 AM - 6 PM | ~$3.50/hr | ~$3.00/hr |
| 6 PM - 10 PM | ~$2.00/hr | ~$2.00/hr |
| 10 PM - 9 AM | **FREE** | **FREE** |

*Rates vary by location and are pulled live from Vancouver's Open Data API*

## Tech Stack

- **Go** - Backend API server
- **Gin** - HTTP web framework
- **Google Maps APIs** - Travel times and geocoding
- **Vancouver Open Data** - Live parking meter rates and locations

## Testing

```bash
# Easy way - use the script
./scripts/test.sh

# Or use make
make test

# Or traditional way
go test ./...
```

## Available Scripts

- `make help` - Show all available commands
- `make test` - Run all tests
- `make run` - Start the server
- `make dev` - Setup development environment
- `make demo` - Run API demo
- `make build` - Build binary

## Project Structure

```
├── cmd/main.go              # Server entry point
├── internal/
│   ├── domain/              # Core business models
│   ├── service/             # Business logic (pricing, routing)
│   ├── repository/          # Data access (Vancouver API)
│   └── handler/             # HTTP request handlers
├── pkg/maps/                # Google Maps integration
├── test/                    # Integration tests
└── docs/                    # API documentation
```

## Future Ideas

- Add transit/bike route options
- Consider weather in route planning
- Real-time parking availability
- Mobile app interface
- User accounts and trip history

## Contributing

1. Fork the repo
2. Create a feature branch
3. Add tests for new features
4. Make sure all tests pass
5. Submit a pull request

Built for Hack the North 2025 🍁