package service

import (
	"fmt"
	"sort"
	"time"

	"vancouver-trip-planner/internal/domain"
	"vancouver-trip-planner/internal/repository"
	"vancouver-trip-planner/pkg/maps"
)

// RoutingService handles multi-objective trip planning
type RoutingService interface {
	PlanTrip(request *domain.TripRequest) ([]*domain.TripPlan, error)
}

// DefaultRoutingService implements RoutingService
type DefaultRoutingService struct {
	parkingRepo    repository.ParkingRepository
	mapsService    maps.MapsService
	pricingService PricingService
}

// NewRoutingService creates a new routing service
func NewRoutingService(parkingRepo repository.ParkingRepository, mapsService maps.MapsService, pricingService PricingService) *DefaultRoutingService {
	return &DefaultRoutingService{
		parkingRepo:    parkingRepo,
		mapsService:    mapsService,
		pricingService: pricingService,
	}
}

// PlanTrip creates three optimized trip plans: cheapest, fastest, and hybrid
func (s *DefaultRoutingService) PlanTrip(request *domain.TripRequest) ([]*domain.TripPlan, error) {
	fmt.Printf("[DEBUG] PlanTrip started with %d stops\n", len(request.Stops))

	if len(request.Stops) < 2 {
		return nil, fmt.Errorf("at least 2 stops are required")
	}

	// Step 1: Geocode all stops if needed
	stops := make([]*domain.Stop, len(request.Stops))
	for i, stop := range request.Stops {
		fmt.Printf("[DEBUG] Processing stop %d: %s\n", i, stop.Address)
		stops[i] = &domain.Stop{
			ID:       stop.ID,
			Address:  stop.Address,
			Duration: stop.Duration,
			Lat:      stop.Lat,
			Lng:      stop.Lng,
		}

		// Geocode if coordinates are missing
		if stops[i].Lat == 0 && stops[i].Lng == 0 {
			fmt.Printf("[DEBUG] Geocoding address: %s\n", stop.Address)
			location, err := s.mapsService.GeocodeAddress(stop.Address)
			if err != nil {
				fmt.Printf("[DEBUG] Geocoding failed: %v\n", err)
				return nil, fmt.Errorf("failed to geocode address %s: %w", stop.Address, err)
			}
			stops[i].Lat = location.Lat
			stops[i].Lng = location.Lng
			fmt.Printf("[DEBUG] Geocoded to: %.6f, %.6f\n", location.Lat, location.Lng)
		}
	}

	// Step 2: Find parking options for each stop
	stopParkingOptions := make(map[string][]*domain.ParkingMeter)
	for _, stop := range stops {
		fmt.Printf("[DEBUG] Finding parking meters for stop: %s (%.6f, %.6f)\n", stop.Address, stop.Lat, stop.Lng)
		meters, err := s.parkingRepo.GetParkingMetersNear(stop.Lat, stop.Lng, 2.0) // 2km radius
		if err != nil {
			fmt.Printf("[DEBUG] Error getting parking meters: %v\n", err)
			return nil, fmt.Errorf("failed to get parking meters for stop %s: %w", stop.Address, err)
		}
		fmt.Printf("[DEBUG] Found %d parking meters for stop: %s\n", len(meters), stop.Address)

		// Limit to top 10 closest meters to avoid excessive combinations
		if len(meters) > 10 {
			// Sort by distance and take closest 10
			sort.Slice(meters, func(i, j int) bool {
				distI := maps.CalculateWalkingTime(&domain.Location{Lat: stop.Lat, Lng: stop.Lng},
					&domain.Location{Lat: meters[i].Lat, Lng: meters[i].Lng})
				distJ := maps.CalculateWalkingTime(&domain.Location{Lat: stop.Lat, Lng: stop.Lng},
					&domain.Location{Lat: meters[j].Lat, Lng: meters[j].Lng})
				return distI < distJ
			})
			meters = meters[:10]
			fmt.Printf("[DEBUG] Limited to top 10 meters for stop: %s\n", stop.Address)
		}

		stopParkingOptions[stop.ID] = meters
	}

	// Step 3: Generate and evaluate route combinations
	fmt.Printf("[DEBUG] Generating routes...\n")
	routes := s.generateRoutes(stops, stopParkingOptions, request)
	fmt.Printf("[DEBUG] Generated %d route candidates\n", len(routes))

	// Step 4: Select the best routes for each objective
	plans := s.selectOptimalPlans(routes)
	fmt.Printf("[DEBUG] Selected %d optimal plans\n", len(plans))

	return plans, nil
}

// RouteCandidate represents a possible route through all stops
type RouteCandidate struct {
	Stops       []*domain.Stop
	Segments    []domain.RouteSegment
	TotalCost   float64
	TotalTime   int
	HybridScore float64
}

// generateRoutes creates route candidates using different parking options
func (s *DefaultRoutingService) generateRoutes(stops []*domain.Stop, parkingOptions map[string][]*domain.ParkingMeter, request *domain.TripRequest) []*RouteCandidate {
	var routes []*RouteCandidate

	// For simplicity, we'll use a greedy approach to generate candidate routes
	// In a production system, you might want to use more sophisticated algorithms like genetic algorithms

	// Generate permutations of stops (for small numbers of stops)
	stopPermutations := s.generateStopPermutations(stops[1:]) // Exclude first stop as starting point

	for _, perm := range stopPermutations {
		// Add starting stop
		route := []*domain.Stop{stops[0]}
		route = append(route, perm...)

		// Try different parking combinations for this route
		routeCandidates := s.evaluateRouteWithParkingCombinations(route, parkingOptions, request)
		routes = append(routes, routeCandidates...)
	}

	return routes
}

// evaluateRouteWithParkingCombinations evaluates a route with different parking options
func (s *DefaultRoutingService) evaluateRouteWithParkingCombinations(stops []*domain.Stop, parkingOptions map[string][]*domain.ParkingMeter, request *domain.TripRequest) []*RouteCandidate {
	var candidates []*RouteCandidate

	// For each stop (except the first), try different parking options
	// Using a simplified approach: select best parking for each stop independently
	for i, stop := range stops {
		if i == 0 {
			continue // No parking needed for starting point
		}

		meters := parkingOptions[stop.ID]
		if len(meters) == 0 {
			continue // No parking available
		}

		// Calculate arrival time at this stop
		arrivalTime := s.calculateArrivalTime(stops[:i+1], request.StartTime)

		// Find best parking option for this stop
		bestMeter, cost, err := s.pricingService.GetOptimalParkingMeter(meters, arrivalTime, stop.Duration)
		if err != nil || bestMeter == nil {
			continue
		}

		// Build route candidate
		candidate := s.buildRouteCandidate(stops, i, bestMeter, cost, arrivalTime, request)
		if candidate != nil {
			candidates = append(candidates, candidate)
		}
	}

	return candidates
}

// buildRouteCandidate constructs a complete route candidate
func (s *DefaultRoutingService) buildRouteCandidate(stops []*domain.Stop, currentStopIndex int, meter *domain.ParkingMeter, parkingCost float64, arrivalTime time.Time, request *domain.TripRequest) *RouteCandidate {
	var segments []domain.RouteSegment
	totalCost := 0.0
	totalTime := 0
	currentTime := request.StartTime

	for i := 0; i < len(stops); i++ {
		if i == 0 {
			continue // Starting point
		}

		fromStop := stops[i-1]
		toStop := stops[i]

		// Calculate travel time
		travelTime, err := s.mapsService.GetTravelTime(
			&domain.Location{Lat: fromStop.Lat, Lng: fromStop.Lng},
			&domain.Location{Lat: toStop.Lat, Lng: toStop.Lng},
			currentTime,
		)
		if err != nil {
			return nil // Skip this route if we can't calculate travel time
		}

		// Use the best parking meter for this stop
		var segmentMeter *domain.ParkingMeter
		var segmentCost float64

		if i == currentStopIndex {
			segmentMeter = meter
			segmentCost = parkingCost
		} else {
			// Calculate optimal parking for other stops
			meters, _ := s.parkingRepo.GetParkingMetersNear(toStop.Lat, toStop.Lng, 0.5)
			segmentArrival := currentTime.Add(time.Duration(travelTime) * time.Minute)
			segmentMeter, segmentCost, _ = s.pricingService.GetOptimalParkingMeter(meters, segmentArrival, toStop.Duration)
			if segmentMeter == nil {
				return nil // No parking available
			}
		}

		// Calculate walking time from parking to destination
		walkingTime := maps.CalculateWalkingTime(
			&domain.Location{Lat: segmentMeter.Lat, Lng: segmentMeter.Lng},
			&domain.Location{Lat: toStop.Lat, Lng: toStop.Lng},
		)

		segment := domain.RouteSegment{
			FromStop:     fromStop,
			ToStop:       toStop,
			ParkingMeter: segmentMeter,
			TravelTime:   travelTime,
			ParkingCost:  segmentCost,
			WalkingTime:  walkingTime,
		}

		segments = append(segments, segment)
		totalCost += segmentCost
		totalTime += travelTime + walkingTime + toStop.Duration

		// Update current time
		currentTime = currentTime.Add(time.Duration(travelTime+walkingTime+toStop.Duration) * time.Minute)
	}

	// Calculate hybrid score
	hybridScore := request.Preferences.CostWeight*totalCost + request.Preferences.TimeWeight*float64(totalTime)/60.0

	return &RouteCandidate{
		Stops:       stops,
		Segments:    segments,
		TotalCost:   totalCost,
		TotalTime:   totalTime,
		HybridScore: hybridScore,
	}
}

// selectOptimalPlans selects the best routes for each objective
func (s *DefaultRoutingService) selectOptimalPlans(routes []*RouteCandidate) []*domain.TripPlan {
	if len(routes) == 0 {
		return nil
	}

	// Find cheapest route
	cheapestRoute := routes[0]
	for _, route := range routes {
		if route.TotalCost < cheapestRoute.TotalCost {
			cheapestRoute = route
		}
	}

	// Find fastest route
	fastestRoute := routes[0]
	for _, route := range routes {
		if route.TotalTime < fastestRoute.TotalTime {
			fastestRoute = route
		}
	}

	// Find hybrid route (best balance)
	hybridRoute := routes[0]
	for _, route := range routes {
		if route.HybridScore < hybridRoute.HybridScore {
			hybridRoute = route
		}
	}

	plans := []*domain.TripPlan{
		{
			Type:      "cheapest",
			TotalCost: cheapestRoute.TotalCost,
			TotalTime: cheapestRoute.TotalTime,
			Route:     cheapestRoute.Segments,
			Metadata: map[string]interface{}{
				"optimization": "cost",
				"savings":      fmt.Sprintf("$%.2f vs fastest", fastestRoute.TotalCost-cheapestRoute.TotalCost),
			},
		},
		{
			Type:      "fastest",
			TotalCost: fastestRoute.TotalCost,
			TotalTime: fastestRoute.TotalTime,
			Route:     fastestRoute.Segments,
			Metadata: map[string]interface{}{
				"optimization": "time",
				"time_saved":   fmt.Sprintf("%d minutes vs cheapest", cheapestRoute.TotalTime-fastestRoute.TotalTime),
			},
		},
		{
			Type:      "hybrid",
			TotalCost: hybridRoute.TotalCost,
			TotalTime: hybridRoute.TotalTime,
			Route:     hybridRoute.Segments,
			Metadata: map[string]interface{}{
				"optimization": "balanced",
				"hybrid_score": hybridRoute.HybridScore,
			},
		},
	}

	return plans
}

// Helper functions

func (s *DefaultRoutingService) generateStopPermutations(stops []*domain.Stop) [][]*domain.Stop {
	if len(stops) <= 1 {
		return [][]*domain.Stop{stops}
	}

	var permutations [][]*domain.Stop
	for i, stop := range stops {
		remaining := make([]*domain.Stop, 0, len(stops)-1)
		remaining = append(remaining, stops[:i]...)
		remaining = append(remaining, stops[i+1:]...)

		subPerms := s.generateStopPermutations(remaining)
		for _, subPerm := range subPerms {
			perm := []*domain.Stop{stop}
			perm = append(perm, subPerm...)
			permutations = append(permutations, perm)
		}
	}

	return permutations
}

func (s *DefaultRoutingService) calculateArrivalTime(stopsToHere []*domain.Stop, startTime time.Time) time.Time {
	currentTime := startTime

	for i := 1; i < len(stopsToHere); i++ {
		fromStop := stopsToHere[i-1]
		toStop := stopsToHere[i]

		// Estimate travel time (use cached or approximate)
		travelTime, _ := s.mapsService.GetTravelTime(
			&domain.Location{Lat: fromStop.Lat, Lng: fromStop.Lng},
			&domain.Location{Lat: toStop.Lat, Lng: toStop.Lng},
			currentTime,
		)

		currentTime = currentTime.Add(time.Duration(travelTime+toStop.Duration) * time.Minute)
	}

	return currentTime
}
