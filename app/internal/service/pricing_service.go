package service

import (
	"math"
	"time"

	"vancouver-trip-planner/internal/domain"
)

// PricingService handles time-dependent parking cost calculations
type PricingService interface {
	CalculateParkingCost(meter *domain.ParkingMeter, arrivalTime time.Time, durationMinutes int) (float64, error)
	GetParkingRateAtTime(meter *domain.ParkingMeter, t time.Time) (float64, int)
	IsMeterActive(t time.Time) bool
	GetOptimalParkingMeter(meters []*domain.ParkingMeter, arrivalTime time.Time, durationMinutes int) (*domain.ParkingMeter, float64, error)
}

type DefaultPricingService struct{}

func NewPricingService() PricingService {
	return &DefaultPricingService{}
}

// CalculateParkingCost calculates the total cost for parking at a specific time and duration
func (s *DefaultPricingService) CalculateParkingCost(meter *domain.ParkingMeter, arrivalTime time.Time, durationMinutes int) (float64, error) {
	if durationMinutes <= 0 {
		return 0.0, nil
	}

	// Convert to Vancouver timezone if needed
	loc, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		return 0.0, err
	}
	localArrival := arrivalTime.In(loc)

	totalCost := 0.0
	currentTime := localArrival
	remainingMinutes := durationMinutes

	for remainingMinutes > 0 {
		if !s.IsMeterActive(currentTime) {
			// Parking is free outside of 9 AM - 10 PM
			break
		}

		rate, timeLimit := s.GetParkingRateAtTime(meter, currentTime)

		// Find the next time boundary (either rate change or meter inactive)
		nextBoundary := s.getNextTimeBoundary(currentTime)
		minutesToBoundary := int(nextBoundary.Sub(currentTime).Minutes())

		// Calculate how many minutes to charge at this rate
		minutesAtThisRate := int(math.Min(float64(remainingMinutes), float64(minutesToBoundary)))

		// Apply time limit if it exists and is lower
		if timeLimit > 0 {
			timeLimitMinutes := timeLimit * 60
			minutesAtThisRate = int(math.Min(float64(minutesAtThisRate), float64(timeLimitMinutes)))
		}

		if minutesAtThisRate > 0 {
			cost := rate * (float64(minutesAtThisRate) / 60.0) // Convert minutes to hours
			totalCost += cost
		}

		currentTime = currentTime.Add(time.Duration(minutesAtThisRate) * time.Minute)
		remainingMinutes -= minutesAtThisRate

		// If we hit a time limit, we can't park longer at this meter
		if timeLimit > 0 && minutesAtThisRate >= timeLimit*60 {
			break
		}
	}

	return totalCost, nil
}

// GetParkingRateAtTime returns the parking rate and time limit for a specific time
func (s *DefaultPricingService) GetParkingRateAtTime(meter *domain.ParkingMeter, t time.Time) (float64, int) {
	if !s.IsMeterActive(t) {
		return 0.0, 0
	}

	weekday := t.Weekday()
	hour := t.Hour()

	switch weekday {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		if hour >= 9 && hour < 18 { // 9 AM - 6 PM
			return meter.RateMF9A6P, meter.TimeLimitMF9A6P
		} else if hour >= 18 && hour < 22 { // 6 PM - 10 PM
			return meter.RateMF6P10, meter.TimeLimitMF6P10
		}
	case time.Saturday:
		if hour >= 9 && hour < 18 {
			return meter.RateSA9A6P, meter.TimeLimitSA9A6P
		} else if hour >= 18 && hour < 22 {
			return meter.RateSA6P10, meter.TimeLimitSA6P10
		}
	case time.Sunday:
		if hour >= 9 && hour < 18 {
			return meter.RateSU9A6P, meter.TimeLimitSU9A6P
		} else if hour >= 18 && hour < 22 {
			return meter.RateSU6P10, meter.TimeLimitSU6P10
		}
	}

	return 0.0, 0 // Free parking
}

// IsMeterActive checks if parking meters are active at a given time
func (s *DefaultPricingService) IsMeterActive(t time.Time) bool {
	hour := t.Hour()
	return hour >= 9 && hour < 22 // 9 AM to 10 PM
}

// getNextTimeBoundary finds the next time when pricing might change
func (s *DefaultPricingService) getNextTimeBoundary(t time.Time) time.Time {
	year, month, day := t.Date()
	loc := t.Location()

	// Check boundaries: 9 AM, 6 PM, 10 PM, and next day 9 AM
	boundaries := []time.Time{
		time.Date(year, month, day, 9, 0, 0, 0, loc),   // 9 AM
		time.Date(year, month, day, 18, 0, 0, 0, loc),  // 6 PM
		time.Date(year, month, day, 22, 0, 0, 0, loc),  // 10 PM
		time.Date(year, month, day+1, 9, 0, 0, 0, loc), // Next day 9 AM
	}

	for _, boundary := range boundaries {
		if boundary.After(t) {
			return boundary
		}
	}

	// Default to next day 9 AM
	return time.Date(year, month, day+1, 9, 0, 0, 0, loc)
}

// GetOptimalParkingMeter finds the best parking meter for a given arrival time and duration
func (s *DefaultPricingService) GetOptimalParkingMeter(meters []*domain.ParkingMeter, arrivalTime time.Time, durationMinutes int) (*domain.ParkingMeter, float64, error) {
	if len(meters) == 0 {
		return nil, 0.0, nil
	}

	var bestMeter *domain.ParkingMeter
	bestCost := math.Inf(1)

	for _, meter := range meters {
		cost, err := s.CalculateParkingCost(meter, arrivalTime, durationMinutes)
		if err != nil {
			continue
		}

		// Check if meter can accommodate the duration
		_, timeLimit := s.GetParkingRateAtTime(meter, arrivalTime)
		if timeLimit > 0 && durationMinutes > timeLimit*60 {
			continue // Skip meters that can't accommodate the full duration
		}

		if cost < bestCost {
			bestCost = cost
			bestMeter = meter
		}
	}

	if bestMeter == nil {
		return nil, 0.0, nil
	}

	return bestMeter, bestCost, nil
}
