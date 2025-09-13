package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"vancouver-trip-planner/internal/domain"
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

// GetParkingMetersNear fetches parking meters within a radius of the given location
func (r *VancouverParkingRepository) GetParkingMetersNear(lat, lng, radiusKm float64) ([]*domain.ParkingMeter, error) {
	// Create a bounding box query (approximation for simplicity)
	// More precise would be to use PostGIS distance queries if supported
	latDelta := radiusKm / 111.32            // Rough conversion: 1 degree lat ≈ 111.32 km
	lngDelta := radiusKm / (111.32 * 0.7071) // Adjust for Vancouver's latitude (~49°N)

	minLat := lat - latDelta
	maxLat := lat + latDelta
	minLng := lng - lngDelta
	maxLng := lng + lngDelta

	query := fmt.Sprintf(`geo_point_2d.lat >= %f AND geo_point_2d.lat <= %f AND geo_point_2d.lon >= %f AND geo_point_2d.lon <= %f`,
		minLat, maxLat, minLng, maxLng)

	params := url.Values{}
	params.Add("where", query)
	params.Add("limit", "100")
	params.Add("select", "*")

	url := fmt.Sprintf("%s?%s", r.baseURL, params.Encode())

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parking meters: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp VancouverParkingResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	meters := make([]*domain.ParkingMeter, 0, len(apiResp.Results))
	for _, data := range apiResp.Results {
		meter := r.convertToDomainModel(data)
		meters = append(meters, meter)
	}

	return meters, nil
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
