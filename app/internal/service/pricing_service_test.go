package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"vancouver-trip-planner/internal/domain"
)

func TestPricingService_CalculateParkingCost(t *testing.T) {
	service := NewPricingService()

	// Create a test parking meter
	meter := &domain.ParkingMeter{
		MeterID:         "TEST001",
		RateMF9A6P:      3.50, // Mon-Fri 9AM-6PM: $3.50/hr
		RateMF6P10:      2.00, // Mon-Fri 6PM-10PM: $2.00/hr
		RateSA9A6P:      3.00, // Saturday 9AM-6PM: $3.00/hr
		RateSA6P10:      2.00, // Saturday 6PM-10PM: $2.00/hr
		RateSU9A6P:      3.00, // Sunday 9AM-6PM: $3.00/hr
		RateSU6P10:      2.00, // Sunday 6PM-10PM: $2.00/hr
		TimeLimitMF9A6P: 3,    // 3 hour limit
		TimeLimitMF6P10: 4,    // 4 hour limit
	}

	tests := []struct {
		name            string
		arrivalTime     string
		durationMinutes int
		expectedCost    float64
		expectError     bool
	}{
		{
			name:            "Weekday daytime parking - 2 hours",
			arrivalTime:     "2024-01-15T10:00:00-08:00", // Monday 10 AM
			durationMinutes: 120,
			expectedCost:    7.00, // 2 hours * $3.50
			expectError:     false,
		},
		{
			name:            "Weekday evening parking - 2 hours",
			arrivalTime:     "2024-01-15T19:00:00-08:00", // Monday 7 PM
			durationMinutes: 120,
			expectedCost:    4.00, // 2 hours * $2.00
			expectError:     false,
		},
		{
			name:            "Free parking after 10 PM",
			arrivalTime:     "2024-01-15T22:30:00-08:00", // Monday 10:30 PM
			durationMinutes: 120,
			expectedCost:    0.00, // Free after 10 PM
			expectError:     false,
		},
		{
			name:            "Cross-period parking (5:30 PM - 7:30 PM)",
			arrivalTime:     "2024-01-15T17:30:00-08:00", // Monday 5:30 PM
			durationMinutes: 120,
			expectedCost:    4.75, // 30 min @ $3.50 + 90 min @ $2.00 = 1.75 + 3.00 = 4.75
			expectError:     false,
		},
		{
			name:            "Saturday daytime parking",
			arrivalTime:     "2024-01-13T11:00:00-08:00", // Saturday 11 AM
			durationMinutes: 120,
			expectedCost:    6.00, // 2 hours * $3.00
			expectError:     false,
		},
		{
			name:            "Zero duration",
			arrivalTime:     "2024-01-15T10:00:00-08:00",
			durationMinutes: 0,
			expectedCost:    0.00,
			expectError:     false,
		},
		{
			name:            "Early morning - before 9 AM",
			arrivalTime:     "2024-01-15T08:00:00-08:00", // Monday 8 AM
			durationMinutes: 60,
			expectedCost:    0.00, // Free before 9 AM
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arrivalTime, err := time.Parse(time.RFC3339, tt.arrivalTime)
			assert.NoError(t, err)

			cost, err := service.CalculateParkingCost(meter, arrivalTime, tt.durationMinutes)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedCost, cost, 0.01, "Cost should match expected value")
			}
		})
	}
}

func TestPricingService_CrossPeriodParkingCost(t *testing.T) {
	service := NewPricingService()

	meter := &domain.ParkingMeter{
		MeterID:         "CROSS001",
		RateMF9A6P:      4.00, // Mon-Fri 9AM-6PM: $4.00/hr
		RateMF6P10:      2.50, // Mon-Fri 6PM-10PM: $2.50/hr
		TimeLimitMF9A6P: 4,
		TimeLimitMF6P10: 4,
	}

	// Park from 5:30 PM to 7:30 PM (crosses 6 PM boundary)
	arrivalTime, _ := time.Parse(time.RFC3339, "2024-01-15T17:30:00-08:00") // Monday 5:30 PM

	cost, err := service.CalculateParkingCost(meter, arrivalTime, 120) // 2 hours

	assert.NoError(t, err)

	// Expected: 30 minutes at $4.00/hr + 90 minutes at $2.50/hr
	// = 0.5 * $4.00 + 1.5 * $2.50 = $2.00 + $3.75 = $5.75
	assert.InDelta(t, 5.75, cost, 0.01)
}

func TestPricingService_GetParkingRateAtTime(t *testing.T) {
	service := NewPricingService()

	meter := &domain.ParkingMeter{
		RateMF9A6P:      3.50,
		RateMF6P10:      2.00,
		TimeLimitMF9A6P: 3,
		TimeLimitMF6P10: 4,
	}

	tests := []struct {
		name          string
		timeStr       string
		expectedRate  float64
		expectedLimit int
	}{
		{
			name:          "Monday morning",
			timeStr:       "2024-01-15T10:00:00-08:00",
			expectedRate:  3.50,
			expectedLimit: 3,
		},
		{
			name:          "Monday evening",
			timeStr:       "2024-01-15T19:00:00-08:00",
			expectedRate:  2.00,
			expectedLimit: 4,
		},
		{
			name:          "Monday late night",
			timeStr:       "2024-01-15T23:00:00-08:00",
			expectedRate:  0.00,
			expectedLimit: 0,
		},
		{
			name:          "Early morning",
			timeStr:       "2024-01-15T08:00:00-08:00",
			expectedRate:  0.00,
			expectedLimit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime, err := time.Parse(time.RFC3339, tt.timeStr)
			assert.NoError(t, err)

			rate, limit := service.GetParkingRateAtTime(meter, testTime)

			assert.Equal(t, tt.expectedRate, rate)
			assert.Equal(t, tt.expectedLimit, limit)
		})
	}
}

func TestPricingService_IsMeterActive(t *testing.T) {
	service := NewPricingService()

	tests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"9 AM - Active", "2024-01-15T09:00:00-08:00", true},
		{"12 PM - Active", "2024-01-15T12:00:00-08:00", true},
		{"9:59 PM - Active", "2024-01-15T21:59:00-08:00", true},
		{"10 PM - Inactive", "2024-01-15T22:00:00-08:00", false},
		{"11 PM - Inactive", "2024-01-15T23:00:00-08:00", false},
		{"8 AM - Inactive", "2024-01-15T08:00:00-08:00", false},
		{"6 AM - Inactive", "2024-01-15T06:00:00-08:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime, err := time.Parse(time.RFC3339, tt.timeStr)
			assert.NoError(t, err)

			result := service.IsMeterActive(testTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPricingService_GetOptimalParkingMeter(t *testing.T) {
	service := NewPricingService()

	meters := []*domain.ParkingMeter{
		{
			MeterID:         "CHEAP001",
			RateMF9A6P:      2.00,
			TimeLimitMF9A6P: 4,
		},
		{
			MeterID:         "EXPENSIVE001",
			RateMF9A6P:      5.00,
			TimeLimitMF9A6P: 4,
		},
		{
			MeterID:         "SHORT_LIMIT001",
			RateMF9A6P:      1.00,
			TimeLimitMF9A6P: 1, // Only 1 hour limit
		},
	}

	arrivalTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00-08:00") // Monday 10 AM

	t.Run("Should choose cheapest viable option", func(t *testing.T) {
		bestMeter, cost, err := service.GetOptimalParkingMeter(meters, arrivalTime, 120) // 2 hours

		assert.NoError(t, err)
		assert.NotNil(t, bestMeter)
		assert.Equal(t, "CHEAP001", bestMeter.MeterID)
		assert.Equal(t, 4.00, cost) // 2 hours * $2.00
	})

	t.Run("Should skip meters with insufficient time limits", func(t *testing.T) {
		// Request 3 hours parking - should skip SHORT_LIMIT001 (1 hour limit)
		bestMeter, cost, err := service.GetOptimalParkingMeter(meters, arrivalTime, 180)

		assert.NoError(t, err)
		assert.NotNil(t, bestMeter)
		assert.Equal(t, "CHEAP001", bestMeter.MeterID)
		assert.Equal(t, 6.00, cost) // 3 hours * $2.00
	})

	t.Run("Should handle empty meter list", func(t *testing.T) {
		bestMeter, cost, err := service.GetOptimalParkingMeter([]*domain.ParkingMeter{}, arrivalTime, 120)

		assert.NoError(t, err)
		assert.Nil(t, bestMeter)
		assert.Equal(t, 0.00, cost)
	})
}
