import { NextResponse } from "next/server"

export async function GET() {
  try {
    // Check Vancouver API health
    const vancouverApiUrl = process.env.VANCOUVER_API_URL || "http://localhost:8080"

    const healthResponse = await fetch(`${vancouverApiUrl}/health`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    })

    const healthData = await healthResponse.json()

    return NextResponse.json({
      status: "healthy",
      timestamp: new Date().toISOString(),
      service: "wyhd-frontend",
      vancouver_api: {
        status: healthResponse.ok ? "healthy" : "unhealthy",
        response: healthData,
      },
    })
  } catch (error) {
    console.error("Health check error:", error)
    return NextResponse.json(
      {
        status: "unhealthy",
        timestamp: new Date().toISOString(),
        service: "wyhd-frontend",
        vancouver_api: {
          status: "unreachable",
          error: error instanceof Error ? error.message : "Unknown error",
        },
      },
      { status: 503 },
    )
  }
}
