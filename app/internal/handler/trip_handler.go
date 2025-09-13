package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"vancouver-trip-planner/internal/domain"
	"vancouver-trip-planner/internal/service"
)

// TripHandler handles trip planning HTTP requests
type TripHandler struct {
	routingService service.RoutingService
}

// NewTripHandler creates a new trip handler
func NewTripHandler(routingService service.RoutingService) *TripHandler {
	return &TripHandler{
		routingService: routingService,
	}
}

// TripPlanRequest represents the HTTP request body for trip planning
type TripPlanRequest struct {
	Stops       []StopRequest       `json:"stops" binding:"required,min=2"`
	StartTime   string              `json:"start_time" binding:"required"` // RFC3339 format
	Timezone    string              `json:"timezone"`
	Preferences *PreferencesRequest `json:"preferences"`
}

// StopRequest represents a stop in the request
type StopRequest struct {
	ID              string  `json:"id"`
	Address         string  `json:"address" binding:"required"`
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	DurationMinutes int     `json:"duration_minutes" binding:"required,min=1"`
}

// PreferencesRequest represents optimization preferences
type PreferencesRequest struct {
	CostWeight float64 `json:"cost_weight" binding:"min=0,max=1"`
	TimeWeight float64 `json:"time_weight" binding:"min=0,max=1"`
}

// TripPlanResponse represents the HTTP response
type TripPlanResponse struct {
	Plans    []*domain.TripPlan     `json:"plans"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// PlanTrip handles POST /api/v1/trips/plan
func (h *TripHandler) PlanTrip(c *gin.Context) {
	var req TripPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Validate preferences weights sum to approximately 1
	if req.Preferences != nil {
		totalWeight := req.Preferences.CostWeight + req.Preferences.TimeWeight
		if totalWeight < 0.9 || totalWeight > 1.1 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_preferences",
				Message: "cost_weight and time_weight must sum to approximately 1.0",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_start_time",
			Message: "start_time must be in RFC3339 format (e.g., '2024-01-15T14:30:00-08:00')",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Set default timezone if not provided
	timezone := req.Timezone
	if timezone == "" {
		timezone = "America/Vancouver"
	}

	// Convert to domain request
	domainReq := &domain.TripRequest{
		StartTime: startTime,
		Timezone:  timezone,
		Stops:     make([]domain.Stop, len(req.Stops)),
		Preferences: domain.Preferences{
			CostWeight: 0.5, // Default equal weight
			TimeWeight: 0.5,
		},
	}

	// Set preferences if provided
	if req.Preferences != nil {
		domainReq.Preferences.CostWeight = req.Preferences.CostWeight
		domainReq.Preferences.TimeWeight = req.Preferences.TimeWeight
	}

	// Convert stops
	for i, stop := range req.Stops {
		domainReq.Stops[i] = domain.Stop{
			ID:       stop.ID,
			Address:  stop.Address,
			Lat:      stop.Lat,
			Lng:      stop.Lng,
			Duration: stop.DurationMinutes,
		}

		// Generate ID if not provided
		if domainReq.Stops[i].ID == "" {
			domainReq.Stops[i].ID = generateStopID(i)
		}
	}

	// Plan the trip
	plans, err := h.routingService.PlanTrip(domainReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "planning_failed",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if len(plans) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "no_routes_found",
			Message: "No valid routes could be found for the given stops",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Build response
	response := TripPlanResponse{
		Plans: plans,
		Metadata: map[string]interface{}{
			"request_id":   c.GetHeader("X-Request-ID"),
			"generated_at": time.Now().UTC(),
			"stops_count":  len(req.Stops),
			"timezone":     timezone,
			"optimization_weights": map[string]float64{
				"cost": domainReq.Preferences.CostWeight,
				"time": domainReq.Preferences.TimeWeight,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles GET /health
func (h *TripHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "vancouver-trip-planner",
	})
}

// GetParkingInfo handles GET /api/v1/parking/info
func (h *TripHandler) GetParkingInfo(c *gin.Context) {
	lat := c.Query("lat")
	lng := c.Query("lng")

	if lat == "" || lng == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_coordinates",
			Message: "lat and lng query parameters are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// This would be implemented to return parking information
	c.JSON(http.StatusOK, gin.H{
		"message": "Parking info endpoint - to be implemented",
		"lat":     lat,
		"lng":     lng,
	})
}

// generateStopID creates a unique ID for a stop
func generateStopID(index int) string {
	return fmt.Sprintf("stop_%d", index+1)
}
