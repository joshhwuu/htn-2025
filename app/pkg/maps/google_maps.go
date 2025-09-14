package maps

import (
	"context"
	"fmt"
	"math"
	"time"

	"googlemaps.github.io/maps"
	"vancouver-trip-planner/internal/domain"
)

// MapsService provides travel time and routing functionality
type MapsService interface {
	GetTravelTime(from, to *domain.Location, departureTime time.Time) (int, error)
	GetTravelTimeMatrix(locations []*domain.Location, departureTime time.Time) ([][]int, error)
	GeocodeAddress(address string) (*domain.Location, error)
}

// GoogleMapsService implements MapsService using Google Maps API
type GoogleMapsService struct {
	client *maps.Client
}

// NewGoogleMapsService creates a new Google Maps service
func NewGoogleMapsService(apiKey string) (*GoogleMapsService, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps client: %w", err)
	}

	return &GoogleMapsService{
		client: client,
	}, nil
}

// GetTravelTime calculates travel time between two locations
func (s *GoogleMapsService) GetTravelTime(from, to *domain.Location, departureTime time.Time) (int, error) {
	ctx := context.Background()

	req := &maps.DistanceMatrixRequest{
		Origins:      []string{fmt.Sprintf("%f,%f", from.Lat, from.Lng)},
		Destinations: []string{fmt.Sprintf("%f,%f", to.Lat, to.Lng)},
		Mode:         maps.TravelModeDriving,
		Units:        maps.UnitsMetric,
		// Remove traffic parameters that require premium APIs
	}

	resp, err := s.client.DistanceMatrix(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get distance matrix: %w", err)
	}

	if len(resp.Rows) == 0 || len(resp.Rows[0].Elements) == 0 {
		return 0, fmt.Errorf("no route found")
	}

	element := resp.Rows[0].Elements[0]
	if element.Status != "OK" {
		return 0, fmt.Errorf("route calculation failed: %s", element.Status)
	}

	// Return duration in minutes (use regular duration since we're not using traffic)
	return int(element.Duration.Minutes()), nil
}

// GetTravelTimeMatrix calculates travel times between all pairs of locations
func (s *GoogleMapsService) GetTravelTimeMatrix(locations []*domain.Location, departureTime time.Time) ([][]int, error) {
	ctx := context.Background()
	n := len(locations)

	// Convert locations to string format
	coords := make([]string, n)
	for i, loc := range locations {
		coords[i] = fmt.Sprintf("%f,%f", loc.Lat, loc.Lng)
	}

	req := &maps.DistanceMatrixRequest{
		Origins:      coords,
		Destinations: coords,
		Mode:         maps.TravelModeDriving,
		Units:        maps.UnitsMetric,
		// Remove traffic parameters that require premium APIs
	}

	resp, err := s.client.DistanceMatrix(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get distance matrix: %w", err)
	}

	// Build the travel time matrix
	matrix := make([][]int, n)
	for i := 0; i < n; i++ {
		matrix[i] = make([]int, n)
		for j := 0; j < n; j++ {
			if i == j {
				matrix[i][j] = 0
				continue
			}

			if len(resp.Rows) <= i || len(resp.Rows[i].Elements) <= j {
				matrix[i][j] = -1 // No route found
				continue
			}

			element := resp.Rows[i].Elements[j]
			if element.Status != "OK" {
				matrix[i][j] = -1 // Route calculation failed
				continue
			}

			// Use duration in traffic if available, otherwise use regular duration
			duration := element.DurationInTraffic
			if duration == 0 {
				duration = element.Duration
			}

			matrix[i][j] = int(duration.Minutes())
		}
	}

	return matrix, nil
}

// GeocodeAddress converts an address to coordinates
func (s *GoogleMapsService) GeocodeAddress(address string) (*domain.Location, error) {
	ctx := context.Background()

	req := &maps.GeocodingRequest{
		Address: address,
	}

	resp, err := s.client.Geocode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to geocode address: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", address)
	}

	// Take the first result
	result := resp[0]
	location := &domain.Location{
		Lat: result.Geometry.Location.Lat,
		Lng: result.Geometry.Location.Lng,
	}

	return location, nil
}

// CalculateWalkingTime calculates walking time between two points using Haversine distance
func CalculateWalkingTime(from, to *domain.Location) int {
	distance := haversineDistance(from.Lat, from.Lng, to.Lat, to.Lng)

	// Assume walking speed of 5 km/h
	walkingSpeedKmH := 5.0
	timeHours := distance / walkingSpeedKmH
	timeMinutes := timeHours * 60

	return int(timeMinutes)
}

// CalculateDistance calculates the distance between two points on Earth using Haversine formula
func CalculateDistance(from, to *domain.Location) float64 {
	return haversineDistance(from.Lat, from.Lng, to.Lat, to.Lng)
}

// haversineDistance calculates the distance between two points on Earth using Haversine formula
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * (3.14159265359 / 180)
	lng1Rad := lng1 * (3.14159265359 / 180)
	lat2Rad := lat2 * (3.14159265359 / 180)
	lng2Rad := lng2 * (3.14159265359 / 180)

	// Calculate differences
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	// Haversine formula
	a := (1-cos(dlat))/2 + cos(lat1Rad)*cos(lat2Rad)*(1-cos(dlng))/2
	c := 2 * asin(sqrt(a))

	return R * c
}

// Helper functions for math operations
func cos(x float64) float64 {
	return math.Cos(x)
}

func sin(x float64) float64 {
	return math.Sin(x)
}

func asin(x float64) float64 {
	return math.Asin(x)
}

func sqrt(x float64) float64 {
	return math.Sqrt(x)
}
