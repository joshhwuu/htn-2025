import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams
  const placeId = searchParams.get('place_id')
  
  if (!placeId) {
    return NextResponse.json({ error: 'place_id is required' }, { status: 400 })
  }

  const apiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || process.env.GOOGLE_MAPS_API_KEY

  if (!apiKey || apiKey === "your_google_maps_api_key_here") {
    // Return mock coordinates for Vancouver area
    const mockResult = {
      result: {
        place_id: placeId,
        formatted_address: "Mock Address, Vancouver, BC, Canada",
        geometry: {
          location: {
            lat: 49.2827 + (Math.random() - 0.5) * 0.1, // Random location around Vancouver
            lng: -123.1207 + (Math.random() - 0.5) * 0.1
          }
        },
        name: "Mock Location"
      }
    }
    
    return NextResponse.json(mockResult)
  }

  try {
    const url = new URL('https://maps.googleapis.com/maps/api/place/details/json')
    url.searchParams.set('place_id', placeId)
    url.searchParams.set('key', apiKey)
    url.searchParams.set('fields', 'place_id,formatted_address,geometry,name')

    const response = await fetch(url.toString())
    
    if (!response.ok) {
      throw new Error(`Google Maps API error: ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching place details:', error)
    
    // Fallback to mock data if API fails
    const mockResult = {
      result: {
        place_id: placeId,
        formatted_address: "Fallback Address, Vancouver, BC, Canada",
        geometry: {
          location: {
            lat: 49.2827 + (Math.random() - 0.5) * 0.1,
            lng: -123.1207 + (Math.random() - 0.5) * 0.1
          }
        },
        name: "Fallback Location"
      }
    }
    
    return NextResponse.json(mockResult)
  }
}