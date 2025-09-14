package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"vancouver-trip-planner/internal/domain"
	"vancouver-trip-planner/pkg/maps"
)

// VancouverParkingResponse represents the API response structure
type VancouverParkingResponse struct {
	TotalCount int                    `json:"total_count"`
	Results    []VancouverParkingData `json:"results"`
}

// VancouverParkingData represents a single parking meter from Vancouver API
type VancouverParkingData struct {
	MeterHead  string `json:"meterhead"`
	RateMF9A6P string `json:"r_mf_9a_6p"`
	RateMF6P10 string `json:"r_mf_6p_10"`
	RateSA9A6P string `json:"r_sa_9a_6p"`
	RateSA6P10 string `json:"r_sa_6p_10"`
	RateSU9A6P string `json:"r_su_9a_6p"`
	RateSU6P10 string `json:"r_su_6p_10"`
	TimeMF9A6P string `json:"t_mf_9a_6p"`
	TimeMF6P10 string `json:"t_mf_6p_10"`
	TimeSA9A6P string `json:"t_sa_9a_6p"`
	TimeSA6P10 string `json:"t_sa_6p_10"`
	TimeSU9A6P string `json:"t_su_9a_6p"`
	TimeSU6P10 string `json:"t_su_6p_10"`
	CreditCard string `json:"creditcard"`
	MeterID    string `json:"meterid"`
	LocalArea  string `json:"geo_local_area"`
	GeoPoint2D struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lon"`
	} `json:"geo_point_2d"`
}

// MeterWithDistance holds a parking meter and its distance from the target location
type MeterWithDistance struct {
	Meter   *domain.ParkingMeter
	Distance float64 // in kilometers
}

// ParkingRepository handles parking meter data operations
type ParkingRepository interface {
	GetParkingMetersNear(lat, lng, radiusKm float64) ([]*domain.ParkingMeter, error)
	GetAllParkingMeters() ([]*domain.ParkingMeter, error)
}

// VancouverParkingRepository implements ParkingRepository using Vancouver Open Data API
type VancouverParkingRepository struct {
	baseURL    string
	httpClient *http.Client
}

// NewVancouverParkingRepository creates a new Vancouver parking repository
func NewVancouverParkingRepository() *VancouverParkingRepository {
	return &VancouverParkingRepository{
		baseURL:    "https://opendata.vancouver.ca/api/explore/v2.1/catalog/datasets/parking-meters/records",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetParkingMetersNear fetches parking meters within a radius of the given location using spatial query
func (r *VancouverParkingRepository) GetParkingMetersNear(lat, lng, radiusKm float64) ([]*domain.ParkingMeter, error) {
	fmt.Printf("[DEBUG] Finding parking meters for stop: (%.6f, %.6f) within %.1fkm radius\n", lat, lng, radiusKm)
	
	// Use bounding box approach - this works reliably with the Vancouver API
	// Create a bounding box around the target location (±0.01 degrees ≈ 1km)
	latMin := lat - 0.01
	latMax := lat + 0.01
	lngMin := lng - 0.01
	lngMax := lng + 0.01
	
	whereClause := fmt.Sprintf("in_bbox(geo_point_2d, %f, %f, %f, %f)", latMin, lngMin, latMax, lngMax)
	
	params := url.Values{}
	params.Add("where", whereClause)
	params.Add("limit", "50") // Get up to 50 meters within the bounding box
	params.Add("select", "*")
	
	url := fmt.Sprintf("%s?%s", r.baseURL, params.Encode())
	fmt.Printf("[DEBUG] Calling Vancouver API: %s\n", url)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		fmt.Printf("[DEBUG] HTTP request failed: %v\n", err)
		return nil, fmt.Errorf("failed to fetch parking meters: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[DEBUG] Vancouver API response status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to read response body: %v\n", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("[DEBUG] Vancouver API response length: %d bytes\n", len(body))

	// Always print response body for debugging
	maxLen := len(body)
	if maxLen > 500 {
		maxLen = 500
	}
	fmt.Printf("[DEBUG] Response body: %s\n", string(body)[:maxLen])

	var apiResp VancouverParkingResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("[DEBUG] JSON unmarshal failed: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	fmt.Printf("[DEBUG] Vancouver API returned %d results within bounding box\n", len(apiResp.Results))

	// Convert API results to domain models and calculate exact distances for sorting
	var metersWithDistance []MeterWithDistance
	for _, data := range apiResp.Results {
		meter := r.convertToDomainModel(data)
		
		// Calculate exact distance using haversine formula for precise sorting
		distance := maps.CalculateDistance(
			&domain.Location{Lat: lat, Lng: lng},
			&domain.Location{Lat: meter.Lat, Lng: meter.Lng},
		)
		
		// Convert distance from meters to kilometers
		distanceKm := distance / 1000.0
		
		// Filter by actual distance (bounding box might include some meters slightly outside radius)
		if distanceKm <= radiusKm {
			metersWithDistance = append(metersWithDistance, MeterWithDistance{
				Meter:   meter,
				Distance: distanceKm,
			})
		}
	}

	fmt.Printf("[DEBUG] Found %d meters within %.1fkm radius after distance filtering\n", len(metersWithDistance), radiusKm)

	// Sort by distance (closest first)
	sort.Slice(metersWithDistance, func(i, j int) bool {
		return metersWithDistance[i].Distance < metersWithDistance[j].Distance
	})
	
	// Convert back to domain models and limit to top 10
	var nearbyMeters []*domain.ParkingMeter
	maxMeters := 10
	if len(metersWithDistance) < maxMeters {
		maxMeters = len(metersWithDistance)
	}
	
	for i := 0; i < maxMeters; i++ {
		nearbyMeters = append(nearbyMeters, metersWithDistance[i].Meter)
		fmt.Printf("[DEBUG] Meter %s at distance %.3fkm\n", 
			metersWithDistance[i].Meter.MeterID, 
			metersWithDistance[i].Distance)
	}

	return nearbyMeters, nil
}

// GetAllParkingMeters fetches all parking meters (paginated)
func (r *VancouverParkingRepository) GetAllParkingMeters() ([]*domain.ParkingMeter, error) {
	var allMeters []*domain.ParkingMeter
	limit := 1000
	offset := 0

	for {
		params := url.Values{}
		params.Add("limit", strconv.Itoa(limit))
		params.Add("offset", strconv.Itoa(offset))
		params.Add("select", "*")

		url := fmt.Sprintf("%s?%s", r.baseURL, params.Encode())

		resp, err := r.httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parking meters: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var apiResp VancouverParkingResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(apiResp.Results) == 0 {
			break
		}

		for _, data := range apiResp.Results {
			meter := r.convertToDomainModel(data)
			allMeters = append(allMeters, meter)
		}

		offset += limit
	}

	return allMeters, nil
}

// convertToDomainModel converts Vancouver API data to domain model
func (r *VancouverParkingRepository) convertToDomainModel(data VancouverParkingData) *domain.ParkingMeter {
	return &domain.ParkingMeter{
		MeterID:         data.MeterID,
		Lat:             data.GeoPoint2D.Lat,
		Lng:             data.GeoPoint2D.Lng,
		MeterType:       data.MeterHead,
		LocalArea:       data.LocalArea,
		CreditCard:      data.CreditCard == "Yes",
		RateMF9A6P:      domain.ParseRate(data.RateMF9A6P),
		RateMF6P10:      domain.ParseRate(data.RateMF6P10),
		RateSA9A6P:      domain.ParseRate(data.RateSA9A6P),
		RateSA6P10:      domain.ParseRate(data.RateSA6P10),
		RateSU9A6P:      domain.ParseRate(data.RateSU9A6P),
		RateSU6P10:      domain.ParseRate(data.RateSU6P10),
		TimeLimitMF9A6P: domain.ParseTimeLimit(data.TimeMF9A6P),
		TimeLimitMF6P10: domain.ParseTimeLimit(data.TimeMF6P10),
		TimeLimitSA9A6P: domain.ParseTimeLimit(data.TimeSA9A6P),
		TimeLimitSA6P10: domain.ParseTimeLimit(data.TimeSA6P10),
		TimeLimitSU9A6P: domain.ParseTimeLimit(data.TimeSU9A6P),
		TimeLimitSU6P10: domain.ParseTimeLimit(data.TimeSU6P10),
	}
}
