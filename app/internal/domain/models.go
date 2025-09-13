package domain

import (
	"strconv"
	"strings"
	"time"
)

// ParkingMeter represents a Vancouver parking meter with time-dependent pricing
type ParkingMeter struct {
	MeterID    string  `json:"meter_id"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	MeterType  string  `json:"meter_type"`
	LocalArea  string  `json:"local_area"`
	CreditCard bool    `json:"credit_card"`

	// Time-dependent rates (hourly)
	RateMF9A6P float64 `json:"rate_mf_9a_6p"` // Mon-Fri 9AM-6PM
	RateMF6P10 float64 `json:"rate_mf_6p_10"` // Mon-Fri 6PM-10PM
	RateSA9A6P float64 `json:"rate_sa_9a_6p"` // Saturday 9AM-6PM
	RateSA6P10 float64 `json:"rate_sa_6p_10"` // Saturday 6PM-10PM
	RateSU9A6P float64 `json:"rate_su_9a_6p"` // Sunday 9AM-6PM
	RateSU6P10 float64 `json:"rate_su_6p_10"` // Sunday 6PM-10PM

	// Time limits (in hours)
	TimeLimitMF9A6P int `json:"time_limit_mf_9a_6p"`
	TimeLimitMF6P10 int `json:"time_limit_mf_6p_10"`
	TimeLimitSA9A6P int `json:"time_limit_sa_9a_6p"`
	TimeLimitSA6P10 int `json:"time_limit_sa_6p_10"`
	TimeLimitSU9A6P int `json:"time_limit_su_9a_6p"`
	TimeLimitSU6P10 int `json:"time_limit_su_6p_10"`
}

// Stop represents a destination in the trip
type Stop struct {
	ID            string    `json:"id"`
	Address       string    `json:"address"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	Duration      int       `json:"duration_minutes"`
	ArrivalTime   time.Time `json:"arrival_time"`
	DepartureTime time.Time `json:"departure_time"`
}

// RouteSegment represents a segment of the trip route
type RouteSegment struct {
	FromStop     *Stop         `json:"from_stop"`
	ToStop       *Stop         `json:"to_stop"`
	ParkingMeter *ParkingMeter `json:"parking_meter"`
	TravelTime   int           `json:"travel_time_minutes"`
	ParkingCost  float64       `json:"parking_cost"`
	WalkingTime  int           `json:"walking_time_minutes"`
}

// TripPlan represents a complete trip plan
type TripPlan struct {
	Type      string                 `json:"type"` // "cheapest", "fastest", "hybrid"
	TotalCost float64                `json:"total_cost"`
	TotalTime int                    `json:"total_time_minutes"`
	Route     []RouteSegment         `json:"route"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// TripRequest represents the input for trip planning
type TripRequest struct {
	Stops       []Stop      `json:"stops"`
	StartTime   time.Time   `json:"start_time"`
	Timezone    string      `json:"timezone"`
	Preferences Preferences `json:"preferences"`
}

// Preferences for trip optimization
type Preferences struct {
	CostWeight float64 `json:"cost_weight"`
	TimeWeight float64 `json:"time_weight"`
}

// Location represents a geographical point
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// ParseRate converts rate string (e.g., "$3.50") to float64
func ParseRate(rateStr string) float64 {
	if rateStr == "" || rateStr == "null" {
		return 0.0
	}

	// Remove $ sign and parse
	cleanRate := strings.TrimPrefix(rateStr, "$")
	rate, err := strconv.ParseFloat(cleanRate, 64)
	if err != nil {
		return 0.0
	}

	return rate
}

// ParseTimeLimit converts time limit string (e.g., "3 Hr") to hours
func ParseTimeLimit(timeLimitStr string) int {
	if timeLimitStr == "" || timeLimitStr == "null" {
		return 0
	}

	// Extract number from "X Hr" format
	parts := strings.Fields(timeLimitStr)
	if len(parts) == 0 {
		return 0
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	return hours
}
