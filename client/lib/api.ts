interface Stop {
  id?: string
  address: string
  lat?: number
  lng?: number
  duration_minutes: number
}

interface TripPlanRequest {
  stops: Stop[]
  start_time: string
  timezone?: string
  preferences?: {
    cost_weight: number
    time_weight: number
  }
}

interface ParkingMeter {
  meter_id: string
  lat: number
  lng: number
  meter_type: string
  local_area: string
  credit_card: boolean
  rate_mf_9a_6p: number
  rate_mf_6p_10: number
}

interface RouteSegment {
  from_stop: {
    id: string
    address: string
    lat: number
    lng: number
    duration_minutes: number
    arrival_time: string
    departure_time: string
  }
  to_stop: {
    id: string
    address: string
    lat: number
    lng: number
    duration_minutes: number
    arrival_time: string
    departure_time: string
  }
  parking_meter: ParkingMeter
  travel_time_minutes: number
  parking_cost: number
  walking_time_minutes: number
}

interface TripPlan {
  type: "cheapest" | "fastest" | "hybrid"
  total_cost: number
  total_time_minutes: number
  route: RouteSegment[]
  metadata: {
    optimization: string
    savings?: string
    time_saved?: string
    hybrid_score?: number
  }
}

interface TripPlanResponse {
  plans: TripPlan[]
  metadata: {
    request_id: string
    generated_at: string
    stops_count: number
    timezone: string
    optimization_weights: {
      cost: number
      time: number
    }
  }
}

export class VancouverAPI {
  private baseUrl: string

  constructor(baseUrl = "/api") {
    this.baseUrl = baseUrl
  }

  async planTrip(request: TripPlanRequest): Promise<TripPlanResponse> {
    const response = await fetch(`${this.baseUrl}/trips/plan`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
    }

    return response.json()
  }

  async getParkingInfo(lat: number, lng: number): Promise<any> {
    const response = await fetch(`${this.baseUrl}/parking/info?lat=${lat}&lng=${lng}`)

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
    }

    return response.json()
  }

  async checkHealth(): Promise<any> {
    const response = await fetch(`${this.baseUrl}/health`)

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
    }

    return response.json()
  }
}

// Export singleton instance
export const vancouverAPI = new VancouverAPI()
