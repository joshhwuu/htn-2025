import { type NextRequest, NextResponse } from "next/server"

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

export async function POST(request: NextRequest) {
  try {
    const body: TripPlanRequest = await request.json()

    // Validate request
    if (!body.stops || body.stops.length < 1) {
      return NextResponse.json(
        { error: "invalid_request", message: "At least one stop is required", code: 400 },
        { status: 400 },
      )
    }

    if (!body.start_time) {
      return NextResponse.json(
        { error: "invalid_start_time", message: "start_time is required in RFC3339 format", code: 400 },
        { status: 400 },
      )
    }

    // Set defaults
    const preferences = {
      cost_weight: body.preferences?.cost_weight || 0.5,
      time_weight: body.preferences?.time_weight || 0.5,
    }

    // Validate preferences sum to ~1.0
    const sum = preferences.cost_weight + preferences.time_weight
    if (Math.abs(sum - 1.0) > 0.1) {
      return NextResponse.json(
        { error: "invalid_preferences", message: "cost_weight and time_weight must sum to ~1.0", code: 400 },
        { status: 400 },
      )
    }

    // Call Vancouver Trip Planner API
    const vancouverApiUrl = process.env.VANCOUVER_API_URL || "http://localhost:8080"

    const apiResponse = await fetch(`${vancouverApiUrl}/api/v1/trips/plan`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        stops: body.stops.map((stop, index) => ({
          id: stop.id || `stop_${index + 1}`,
          address: stop.address,
          lat: stop.lat,
          lng: stop.lng,
          duration_minutes: stop.duration_minutes,
        })),
        start_time: body.start_time,
        timezone: body.timezone || "America/Vancouver",
        preferences,
      }),
    })

    if (!apiResponse.ok) {
      const errorData = await apiResponse.json().catch(() => ({}))
      return NextResponse.json(
        {
          error: "planning_failed",
          message: errorData.message || "Failed to plan trip",
          code: apiResponse.status,
        },
        { status: apiResponse.status },
      )
    }

    const planData = await apiResponse.json()
    return NextResponse.json(planData)
  } catch (error) {
    console.error("Trip planning error:", error)
    return NextResponse.json(
      { error: "planning_failed", message: "Internal server error during trip planning", code: 500 },
      { status: 500 },
    )
  }
}
