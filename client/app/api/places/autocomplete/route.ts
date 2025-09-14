import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams
  const input = searchParams.get('input')
  
  if (!input || input.length < 3) {
    return NextResponse.json({ predictions: [] })
  }

  const apiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || process.env.GOOGLE_MAPS_API_KEY

  if (!apiKey || apiKey === "your_google_maps_api_key_here") {
    // Return mock data for development
    const mockPredictions = [
      {
        place_id: `mock_${Date.now()}_1`,
        description: `${input}, Vancouver, BC, Canada`,
        structured_formatting: {
          main_text: input,
          secondary_text: "Vancouver, BC, Canada"
        } 
      },
      {
        place_id: `mock_${Date.now()}_2`,
        description: `${input} Street, Vancouver, BC, Canada`,
        structured_formatting: {
          main_text: `${input} Street`,
          secondary_text: "Vancouver, BC, Canada"
        }
      },
      {
        place_id: `mock_${Date.now()}_3`,
        description: `${input} Avenue, Vancouver, BC, Canada`,
        structured_formatting: {
          main_text: `${input} Avenue`,
          secondary_text: "Vancouver, BC, Canada"
        }
      }
    ]
    
    return NextResponse.json({ predictions: mockPredictions })
  }

  try {
    const url = new URL('https://maps.googleapis.com/maps/api/place/autocomplete/json')
    url.searchParams.set('input', input)
    url.searchParams.set('key', apiKey)
    url.searchParams.set('types', 'establishment|geocode')
    url.searchParams.set('components', 'country:ca')
    url.searchParams.set('location', '49.2827,-123.1207') // Vancouver center
    url.searchParams.set('radius', '50000') // 50km radius

    const response = await fetch(url.toString())
    
    if (!response.ok) {
      throw new Error(`Google Maps API error: ${response.status}`)
    }

    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching places:', error)
    
    // Fallback to mock data if API fails
    const mockPredictions = [
      {
        place_id: `fallback_${Date.now()}_1`,
        description: `${input}, Vancouver, BC, Canada`,
        structured_formatting: {
          main_text: input,
          secondary_text: "Vancouver, BC, Canada"
        }
      },
      {
        place_id: `fallback_${Date.now()}_2`,
        description: `${input} Street, Vancouver, BC, Canada`,
        structured_formatting: {
          main_text: `${input} Street`,
          secondary_text: "Vancouver, BC, Canada"
        }
      }
    ]
    
    return NextResponse.json({ predictions: mockPredictions })
  }
}
