import { type NextRequest, NextResponse } from "next/server"

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = request.nextUrl
    const lat = searchParams.get("lat")
    const lng = searchParams.get("lng")

    if (!lat || !lng) {
      return NextResponse.json(
        { error: "invalid_request", message: "Missing lat/lng parameters", code: 400 },
        { status: 400 },
      )
    }

    // Validate coordinates
    const latitude = Number.parseFloat(lat)
    const longitude = Number.parseFloat(lng)

    if (isNaN(latitude) || isNaN(longitude)) {
      return NextResponse.json(
        { error: "invalid_request", message: "Invalid lat/lng format", code: 400 },
        { status: 400 },
      )
    }

    // Call Vancouver Trip Planner API
    const vancouverApiUrl = process.env.VANCOUVER_API_URL || "http://localhost:8080"

    const apiResponse = await fetch(`${vancouverApiUrl}/api/v1/parking/info?lat=${latitude}&lng=${longitude}`)

    if (!apiResponse.ok) {
      const errorData = await apiResponse.json().catch(() => ({}))
      return NextResponse.json(
        {
          error: "parking_info_failed",
          message: errorData.message || "Failed to get parking info",
          code: apiResponse.status,
        },
        { status: apiResponse.status },
      )
    }

    const parkingData = await apiResponse.json()
    return NextResponse.json(parkingData)
  } catch (error) {
    console.error("Parking info error:", error)
    return NextResponse.json(
      { error: "parking_info_failed", message: "Internal server error", code: 500 },
      { status: 500 },
    )
  }
}
