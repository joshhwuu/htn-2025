package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vancouver-trip-planner/internal/domain"
	"vancouver-trip-planner/internal/handler"
	"vancouver-trip-planner/internal/repository"
	"vancouver-trip-planner/internal/service"
	"vancouver-trip-planner/pkg/maps"
)

func TestTripPlanningIntegration(t *testing.T) {
	// Skip integration tests if no Google Maps API key is provided
	googleMapsAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleMapsAPIKey == "" {
		t.Skip("Skipping integration tests - GOOGLE_MAPS_API_KEY not set")
	}

	// Setup services
	parkingRepo := repository.NewVancouverParkingRepository()
	pricingService := service.NewPricingService()

	mapsService, err := maps.NewGoogleMapsService(googleMapsAPIKey)
	require.NoError(t, err)

	routingService := service.NewRoutingService(parkingRepo, mapsService, pricingService)
	tripHandler := handler.NewTripHandler(routingService)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/trips/plan", tripHandler.PlanTrip)
	router.GET("/health", tripHandler.HealthCheck)

	t.Run("Health check should return OK", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	t.Run("Should plan trip with valid Vancouver addresses", func(t *testing.T) {
		requestBody := handler.TripPlanRequest{
			Stops: []handler.StopRequest{
				{
					Address:         "800 Robson St, Vancouver, BC",
					DurationMinutes: 60,
				},
				{
					Address:         "1055 Canada Pl, Vancouver, BC", // Canada Place
					DurationMinutes: 90,
				},
			},
			StartTime: time.Now().Add(time.Hour).Format(time.RFC3339),
			Preferences: &handler.PreferencesRequest{
				CostWeight: 0.6,
				TimeWeight: 0.4,
			},
		}

		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/trips/plan", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response handler.TripPlanResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Should return 3 plans: cheapest, fastest, hybrid
		assert.Len(t, response.Plans, 3)

		planTypes := make(map[string]bool)
		for _, plan := range response.Plans {
			planTypes[plan.Type] = true
			assert.Greater(t, plan.TotalTime, 0)
			assert.GreaterOrEqual(t, plan.TotalCost, 0.0)
			assert.NotEmpty(t, plan.Route)
		}

		assert.True(t, planTypes["cheapest"])
		assert.True(t, planTypes["fastest"])
		assert.True(t, planTypes["hybrid"])
	})

	t.Run("Should return error for invalid request", func(t *testing.T) {
		requestBody := handler.TripPlanRequest{
			Stops: []handler.StopRequest{
				{
					Address:         "800 Robson St, Vancouver, BC",
					DurationMinutes: 60,
				},
				// Missing second stop - should fail validation
			},
			StartTime: "invalid-time-format",
		}

		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/trips/plan", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should handle preferences validation", func(t *testing.T) {
		requestBody := handler.TripPlanRequest{
			Stops: []handler.StopRequest{
				{
					Address:         "800 Robson St, Vancouver, BC",
					DurationMinutes: 60,
				},
				{
					Address:         "1055 Canada Pl, Vancouver, BC",
					DurationMinutes: 90,
				},
			},
			StartTime: time.Now().Add(time.Hour).Format(time.RFC3339),
			Preferences: &handler.PreferencesRequest{
				CostWeight: 0.8,
				TimeWeight: 0.8, // Total > 1.0, should fail
			},
		}

		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/trips/plan", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response handler.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_preferences", response.Error)
	})
}

func TestParkingRepositoryIntegration(t *testing.T) {
	t.Run("Should fetch parking meters from Vancouver API", func(t *testing.T) {
		repo := repository.NewVancouverParkingRepository()

		// Test fetching meters near downtown Vancouver
		meters, err := repo.GetParkingMetersNear(49.2827, -123.1207, 0.5) // 500m radius

		assert.NoError(t, err)
		assert.NotNil(t, meters)

		// Skip verification if no meters found (Vancouver API might not have data for this location)
		if len(meters) == 0 {
			t.Skip("No parking meters found in test location")
			return
		}

		// Verify meter structure for first few meters
		checkCount := len(meters)
		if checkCount > 5 {
			checkCount = 5
		}
		for _, meter := range meters[:checkCount] {
			assert.NotEmpty(t, meter.MeterID)
			assert.NotZero(t, meter.Lat)
			assert.NotZero(t, meter.Lng)
			assert.NotEmpty(t, meter.LocalArea)

			// At least one rate should be non-zero
			hasRate := meter.RateMF9A6P > 0 || meter.RateMF6P10 > 0 ||
				meter.RateSA9A6P > 0 || meter.RateSA6P10 > 0 ||
				meter.RateSU9A6P > 0 || meter.RateSU6P10 > 0

			assert.True(t, hasRate, "Meter should have at least one non-zero rate")
		}
	})

	t.Run("Should handle invalid coordinates gracefully", func(t *testing.T) {
		repo := repository.NewVancouverParkingRepository()

		// Test with coordinates outside Vancouver
		meters, err := repo.GetParkingMetersNear(0, 0, 0.5)

		// Should not error, but may return empty results
		assert.NoError(t, err)
		// meters could be empty, which is fine for invalid coordinates
		assert.NotNil(t, meters)
	})
}

func TestPricingServiceIntegration(t *testing.T) {
	t.Run("Should calculate realistic parking costs", func(t *testing.T) {
		service := service.NewPricingService()

		// Create a realistic Vancouver parking meter
		meter := &domain.ParkingMeter{
			MeterID:         "INTEGRATION_TEST",
			RateMF9A6P:      3.50, // Typical Vancouver rates
			RateMF6P10:      2.00,
			TimeLimitMF9A6P: 3,
			TimeLimitMF6P10: 4,
		}

		// Test different scenarios
		// Load Vancouver timezone
		vancouverTz, _ := time.LoadLocation("America/Vancouver")

		scenarios := []struct {
			name            string
			arrivalTime     time.Time
			durationMinutes int
			minCost         float64
			maxCost         float64
		}{
			{
				name:            "Weekday morning 2 hours",
				arrivalTime:     time.Date(2024, 1, 15, 10, 0, 0, 0, vancouverTz), // Monday 10 AM Vancouver time
				durationMinutes: 120,
				minCost:         6.0, // Should be around 2 * $3.50
				maxCost:         8.0,
			},
			{
				name:            "Evening 1 hour",
				arrivalTime:     time.Date(2024, 1, 15, 19, 0, 0, 0, vancouverTz), // Monday 7 PM Vancouver time
				durationMinutes: 60,
				minCost:         1.5, // Should be around 1 * $2.00
				maxCost:         2.5,
			},
			{
				name:            "Late night - should be free",
				arrivalTime:     time.Date(2024, 1, 15, 23, 0, 0, 0, vancouverTz), // Monday 11 PM Vancouver time
				durationMinutes: 60,
				minCost:         0.0,
				maxCost:         0.0,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				cost, err := service.CalculateParkingCost(meter, scenario.arrivalTime, scenario.durationMinutes)

				assert.NoError(t, err)
				assert.GreaterOrEqual(t, cost, scenario.minCost)
				assert.LessOrEqual(t, cost, scenario.maxCost)
			})
		}
	})
}
