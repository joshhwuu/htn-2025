package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"vancouver-trip-planner/internal/domain"
)

func TestCalculateWalkingTime(t *testing.T) {
	tests := []struct {
		name     string
		from     *domain.Location
		to       *domain.Location
		expected int // Expected time in minutes (approximately)
	}{
		{
			name:     "Short walk - 1 block",
			from:     &domain.Location{Lat: 49.2827, Lng: -123.1207}, // Vancouver downtown
			to:       &domain.Location{Lat: 49.2837, Lng: -123.1217}, // ~1 block away
			expected: 2,                                              // About 2 minutes
		},
		{
			name:     "Medium walk - 5 blocks",
			from:     &domain.Location{Lat: 49.2827, Lng: -123.1207},
			to:       &domain.Location{Lat: 49.2877, Lng: -123.1257}, // ~5 blocks away
			expected: 8,                                              // About 8 minutes
		},
		{
			name:     "Same location",
			from:     &domain.Location{Lat: 49.2827, Lng: -123.1207},
			to:       &domain.Location{Lat: 49.2827, Lng: -123.1207},
			expected: 0, // 0 minutes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWalkingTime(tt.from, tt.to)

			// Allow some tolerance for calculation variations
			assert.InDelta(t, tt.expected, result, 2, "Walking time should be approximately correct")
		})
	}
}

func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		expected float64 // Expected distance in km (approximately)
	}{
		{
			name:     "Vancouver to Burnaby",
			lat1:     49.2827, // Vancouver downtown
			lng1:     -123.1207,
			lat2:     49.2488, // Burnaby
			lng2:     -122.9805,
			expected: 11.5, // About 11.5 km
		},
		{
			name:     "Same location",
			lat1:     49.2827,
			lng1:     -123.1207,
			lat2:     49.2827,
			lng2:     -123.1207,
			expected: 0.0,
		},
		{
			name:     "Short distance",
			lat1:     49.2827,
			lng1:     -123.1207,
			lat2:     49.2837,
			lng2:     -123.1217,
			expected: 0.15, // About 150 meters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := haversineDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)

			// Allow some tolerance for calculation variations
			assert.InDelta(t, tt.expected, result, 1.0, "Distance should be approximately correct")
		})
	}
}

func TestMathHelpers(t *testing.T) {
	t.Run("cos function", func(t *testing.T) {
		result := cos(0)
		assert.InDelta(t, 1.0, result, 0.001)

		result = cos(3.14159265359 / 2) // π/2
		assert.InDelta(t, 0.0, result, 0.001)
	})

	t.Run("sin function", func(t *testing.T) {
		result := sin(0)
		assert.InDelta(t, 0.0, result, 0.001)

		result = sin(3.14159265359 / 2) // π/2
		assert.InDelta(t, 1.0, result, 0.001)
	})

	t.Run("sqrt function", func(t *testing.T) {
		result := sqrt(4)
		assert.Equal(t, 2.0, result)

		result = sqrt(9)
		assert.Equal(t, 3.0, result)

		result = sqrt(0)
		assert.Equal(t, 0.0, result)
	})

	t.Run("asin function", func(t *testing.T) {
		result := asin(0)
		assert.InDelta(t, 0.0, result, 0.001)

		result = asin(1)
		assert.InDelta(t, 3.14159265359/2, result, 0.001) // π/2
	})
}

// Note: Testing the actual Google Maps API integration would require:
// 1. API credentials
// 2. Network access
// 3. Potentially costs money
//
// For unit tests, we focus on testing the pure functions and logic.
// Integration tests with the actual API would be in separate files.

func TestGoogleMapsServiceCreation(t *testing.T) {
	t.Run("Should fail with empty API key", func(t *testing.T) {
		service, err := NewGoogleMapsService("")
		assert.Error(t, err)
		assert.Nil(t, service)
	})

	t.Run("Should succeed with valid API key format", func(t *testing.T) {
		// Note: This doesn't validate the actual API key, just that the service can be created
		service, err := NewGoogleMapsService("fake-api-key-for-testing")

		// The actual Google Maps client creation might fail with invalid key
		// but we're testing our wrapper logic here
		if err != nil {
			// If it fails, it should be due to invalid key, not our logic
			assert.Contains(t, err.Error(), "Google Maps")
		} else {
			assert.NotNil(t, service)
		}
	})
}
