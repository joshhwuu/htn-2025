# Vancouver Trip Planner API Documentation

Base URL: `http://localhost:8080`

## Authentication
Currently no authentication required. API key for Google Maps is configured server-side.

## Endpoints

### 1. Health Check

Check if the service is running.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T20:30:45Z",
  "service": "vancouver-trip-planner"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

---

### 2. Plan Trip

Plan an optimized multi-stop trip in Vancouver with parking cost and time optimization.

**Endpoint:** `POST /api/v1/trips/plan`

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "stops": [
    {
      "id": "optional-stop-id",
      "address": "800 Robson St, Vancouver, BC",
      "lat": 49.2827,
      "lng": -123.1207,
      "duration_minutes": 60
    },
    {
      "address": "1055 Canada Pl, Vancouver, BC",
      "duration_minutes": 90
    }
  ],
  "start_time": "2024-01-15T14:30:00-08:00",
  "timezone": "America/Vancouver",
  "preferences": {
    "cost_weight": 0.6,
    "time_weight": 0.4
  }
}
```

**Request Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `stops` | Array | Yes | Array of stops (minimum 2) |
| `stops[].id` | String | No | Optional unique identifier for the stop |
| `stops[].address` | String | Yes | Full address of the destination |
| `stops[].lat` | Number | No | Latitude (will geocode address if not provided) |
| `stops[].lng` | Number | No | Longitude (will geocode address if not provided) |
| `stops[].duration_minutes` | Integer | Yes | How long to stay at this stop (minimum 1) |
| `start_time` | String | Yes | ISO 8601 timestamp when trip starts |
| `timezone` | String | No | IANA timezone (defaults to "America/Vancouver") |
| `preferences` | Object | No | Optimization preferences |
| `preferences.cost_weight` | Number | No | Weight for cost optimization (0-1, default 0.5) |
| `preferences.time_weight` | Number | No | Weight for time optimization (0-1, default 0.5) |

**Response:**
```json
{
  "plans": [
    {
      "type": "cheapest",
      "total_cost": 12.50,
      "total_time_minutes": 180,
      "route": [
        {
          "from_stop": {
            "id": "stop_1",
            "address": "800 Robson St, Vancouver, BC",
            "lat": 49.2827,
            "lng": -123.1207,
            "duration_minutes": 60,
            "arrival_time": "2024-01-15T14:30:00-08:00",
            "departure_time": "2024-01-15T15:30:00-08:00"
          },
          "to_stop": {
            "id": "stop_2",
            "address": "1055 Canada Pl, Vancouver, BC",
            "lat": 49.2889,
            "lng": -123.1111,
            "duration_minutes": 90,
            "arrival_time": "2024-01-15T15:45:00-08:00",
            "departure_time": "2024-01-15T17:15:00-08:00"
          },
          "parking_meter": {
            "meter_id": "170127",
            "lat": 49.2888,
            "lng": -123.1110,
            "meter_type": "Twin",
            "local_area": "Downtown",
            "credit_card": false,
            "rate_mf_9a_6p": 3.50,
            "rate_mf_6p_10": 2.00
          },
          "travel_time_minutes": 12,
          "parking_cost": 5.25,
          "walking_time_minutes": 3
        }
      ],
      "metadata": {
        "optimization": "cost",
        "savings": "$3.25 vs fastest"
      }
    },
    {
      "type": "fastest",
      "total_cost": 15.75,
      "total_time_minutes": 150,
      "route": [...],
      "metadata": {
        "optimization": "time",
        "time_saved": "30 minutes vs cheapest"
      }
    },
    {
      "type": "hybrid",
      "total_cost": 14.10,
      "total_time_minutes": 165,
      "route": [...],
      "metadata": {
        "optimization": "balanced",
        "hybrid_score": 8.46
      }
    }
  ],
  "metadata": {
    "request_id": "req_1642272600000",
    "generated_at": "2024-01-15T22:30:00Z",
    "stops_count": 2,
    "timezone": "America/Vancouver",
    "optimization_weights": {
      "cost": 0.6,
      "time": 0.4
    }
  }
}
```

**Status Codes:**
- `200 OK` - Trip planned successfully
- `400 Bad Request` - Invalid request format or validation error
- `404 Not Found` - No valid routes found
- `500 Internal Server Error` - Server error (API failures, etc.)

**Error Response Format:**
```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "code": 400
}
```

**Common Error Codes:**
- `invalid_request` - Missing required fields or invalid format
- `invalid_start_time` - start_time not in RFC3339 format
- `invalid_preferences` - cost_weight and time_weight must sum to ~1.0
- `planning_failed` - Internal error during route planning
- `no_routes_found` - No valid routes for given stops

---

### 3. Get Parking Info

Get parking meter information for a specific location.

**Endpoint:** `GET /api/v1/parking/info`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `lat` | Number | Yes | Latitude of the location |
| `lng` | Number | Yes | Longitude of the location |

**Example Request:**
```
GET /api/v1/parking/info?lat=49.2827&lng=-123.1207
```

**Response:**
```json
{
  "message": "Parking info endpoint - to be implemented",
  "lat": "49.2827",
  "lng": "-123.1207"
}
```

**Status Codes:**
- `200 OK` - Information retrieved
- `400 Bad Request` - Missing lat/lng parameters

---

## Rate Limits

Currently no rate limits implemented. In production, consider:
- Google Maps API has usage limits
- Recommend caching geocoding results
- Consider implementing per-IP rate limiting

## Examples

### Basic Trip Planning
```bash
curl -X POST http://localhost:8080/api/v1/trips/plan \
  -H "Content-Type: application/json" \
  -d '{
    "stops": [
      {
        "address": "Vancouver Art Gallery, Vancouver, BC",
        "duration_minutes": 90
      },
      {
        "address": "Granville Island, Vancouver, BC",
        "duration_minutes": 120
      }
    ],
    "start_time": "2024-01-20T10:00:00-08:00"
  }'
```

### Trip with Coordinates and Preferences
```bash
curl -X POST http://localhost:8080/api/v1/trips/plan \
  -H "Content-Type: application/json" \
  -d '{
    "stops": [
      {
        "address": "Science World, Vancouver, BC",
        "lat": 49.2732,
        "lng": -123.1037,
        "duration_minutes": 180
      },
      {
        "address": "Queen Elizabeth Park, Vancouver, BC",
        "lat": 49.2404,
        "lng": -123.1144,
        "duration_minutes": 120
      }
    ],
    "start_time": "2024-01-20T14:00:00-08:00",
    "preferences": {
      "cost_weight": 0.8,
      "time_weight": 0.2
    }
  }'
```

### Health Check
```bash
curl http://localhost:8080/health
```

### Parking Info
```bash
curl "http://localhost:8080/api/v1/parking/info?lat=49.2827&lng=-123.1207"
```

## Vancouver Parking Pricing

The system uses Vancouver's time-dependent parking meter pricing:

| Time Period | Weekdays | Saturdays | Sundays |
|-------------|----------|-----------|---------|
| 9:00 AM - 6:00 PM | $3.50/hr (typical) | $3.00/hr (typical) | $3.00/hr (typical) |
| 6:00 PM - 10:00 PM | $2.00/hr (typical) | $2.00/hr (typical) | $2.00/hr (typical) |
| 10:00 PM - 9:00 AM | FREE | FREE | FREE |

**Notes:**
- Rates vary by location and meter type
- Time limits typically 2-4 hours depending on area
- Some meters accept credit cards, others require coins/app
- Pricing automatically calculated based on arrival time and duration

## Data Sources

- **Parking Data**: [Vancouver Open Data Portal](https://opendata.vancouver.ca/explore/dataset/parking-meters)
- **Travel Times**: Google Maps Distance Matrix API
- **Geocoding**: Google Maps Geocoding API