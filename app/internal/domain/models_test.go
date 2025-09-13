package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "Valid rate with dollar sign",
			input:    "$3.50",
			expected: 3.50,
		},
		{
			name:     "Valid rate without dollar sign",
			input:    "3.50",
			expected: 3.50,
		},
		{
			name:     "Integer rate",
			input:    "$5",
			expected: 5.0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0.0,
		},
		{
			name:     "Null value",
			input:    "null",
			expected: 0.0,
		},
		{
			name:     "Zero rate",
			input:    "$0.00",
			expected: 0.0,
		},
		{
			name:     "High precision rate",
			input:    "$4.25",
			expected: 4.25,
		},
		{
			name:     "Invalid format",
			input:    "invalid",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTimeLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Valid hour format",
			input:    "3 Hr",
			expected: 3,
		},
		{
			name:     "Single hour",
			input:    "1 Hr",
			expected: 1,
		},
		{
			name:     "Multiple hours",
			input:    "4 Hr",
			expected: 4,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Null value",
			input:    "null",
			expected: 0,
		},
		{
			name:     "Invalid format",
			input:    "invalid",
			expected: 0,
		},
		{
			name:     "No units",
			input:    "3",
			expected: 3,
		},
		{
			name:     "Different case",
			input:    "2 hr",
			expected: 2,
		},
		{
			name:     "With extra spaces",
			input:    " 5 Hr ",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTimeLimit(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParkingMeterModel(t *testing.T) {
	t.Run("Should create valid parking meter", func(t *testing.T) {
		meter := &ParkingMeter{
			MeterID:         "TEST001",
			Lat:             49.2827,
			Lng:             -123.1207,
			MeterType:       "Twin",
			LocalArea:       "Downtown",
			CreditCard:      true,
			RateMF9A6P:      3.50,
			RateMF6P10:      2.00,
			TimeLimitMF9A6P: 3,
			TimeLimitMF6P10: 4,
		}

		assert.Equal(t, "TEST001", meter.MeterID)
		assert.Equal(t, 49.2827, meter.Lat)
		assert.Equal(t, -123.1207, meter.Lng)
		assert.Equal(t, "Twin", meter.MeterType)
		assert.Equal(t, "Downtown", meter.LocalArea)
		assert.True(t, meter.CreditCard)
		assert.Equal(t, 3.50, meter.RateMF9A6P)
		assert.Equal(t, 2.00, meter.RateMF6P10)
		assert.Equal(t, 3, meter.TimeLimitMF9A6P)
		assert.Equal(t, 4, meter.TimeLimitMF6P10)
	})
}

func TestStopModel(t *testing.T) {
	t.Run("Should create valid stop", func(t *testing.T) {
		stop := &Stop{
			ID:       "stop_1",
			Address:  "123 Main St, Vancouver, BC",
			Lat:      49.2827,
			Lng:      -123.1207,
			Duration: 60,
		}

		assert.Equal(t, "stop_1", stop.ID)
		assert.Equal(t, "123 Main St, Vancouver, BC", stop.Address)
		assert.Equal(t, 49.2827, stop.Lat)
		assert.Equal(t, -123.1207, stop.Lng)
		assert.Equal(t, 60, stop.Duration)
	})
}

func TestLocationModel(t *testing.T) {
	t.Run("Should create valid location", func(t *testing.T) {
		location := &Location{
			Lat: 49.2827,
			Lng: -123.1207,
		}

		assert.Equal(t, 49.2827, location.Lat)
		assert.Equal(t, -123.1207, location.Lng)
	})
}

func TestPreferencesModel(t *testing.T) {
	t.Run("Should create valid preferences", func(t *testing.T) {
		preferences := &Preferences{
			CostWeight: 0.6,
			TimeWeight: 0.4,
		}

		assert.Equal(t, 0.6, preferences.CostWeight)
		assert.Equal(t, 0.4, preferences.TimeWeight)
	})

	t.Run("Should handle equal weights", func(t *testing.T) {
		preferences := &Preferences{
			CostWeight: 0.5,
			TimeWeight: 0.5,
		}

		totalWeight := preferences.CostWeight + preferences.TimeWeight
		assert.Equal(t, 1.0, totalWeight)
	})
}
