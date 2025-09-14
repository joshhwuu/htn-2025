"use client"

import { useEffect, useRef, useState } from "react"
import { Wrapper, Status } from "@googlemaps/react-wrapper"

// Temporary type declaration for Google Maps
declare const google: any

interface GoogleMapProps {
  stops: Array<{
    id: string
    address: string
    lat?: number
    lng?: number
    duration: number
  }>
  route?: Array<{
    from_stop: {
      lat: number
      lng: number
      address: string
    }
    to_stop: {
      lat: number
      lng: number
      address: string
    }
    parking_meter: {
      meter_id: string
      lat: number
      lng: number
      meter_type: string
      local_area: string
      credit_card: boolean
      rate_mf_9a_6p: number
      rate_mf_6p_10: number
    }
  }>
  className?: string
}

const MapComponent = ({ stops, route, className }: GoogleMapProps) => {
  const mapRef = useRef<HTMLDivElement>(null)
  const [map, setMap] = useState<google.maps.Map | null>(null)
  const [markers, setMarkers] = useState<google.maps.Marker[]>([])
  const [directionsService, setDirectionsService] = useState<google.maps.DirectionsService | null>(null)
  const [directionsRenderer, setDirectionsRenderer] = useState<google.maps.DirectionsRenderer | null>(null)

  useEffect(() => {
    if (mapRef.current && !map) {
      const newMap = new google.maps.Map(mapRef.current, {
        center: { lat: 49.2827, lng: -123.1207 }, // Vancouver center
        zoom: 12,
        mapTypeControl: true,
        streetViewControl: true,
        fullscreenControl: true,
        zoomControl: true,
      })
      setMap(newMap)
      setDirectionsService(new google.maps.DirectionsService())
      setDirectionsRenderer(new google.maps.DirectionsRenderer())
    }
  }, [map])

  useEffect(() => {
    if (map && directionsRenderer) {
      directionsRenderer.setMap(map)
    }
  }, [map, directionsRenderer])

  // Clear existing markers
  useEffect(() => {
    markers.forEach(marker => marker.setMap(null))
    setMarkers([])
  }, [stops, route])

  // Add stop markers
  useEffect(() => {
    if (!map) return

    const newMarkers: google.maps.Marker[] = []

    stops.forEach((stop, index) => {
      if (stop.lat && stop.lng) {
        const marker = new google.maps.Marker({
          position: { lat: stop.lat, lng: stop.lng },
          map: map,
          title: stop.address,
          label: {
            text: `${index + 1}`,
            color: "white",
            fontWeight: "bold",
          },
          icon: {
            path: google.maps.SymbolPath.CIRCLE,
            scale: 25,
            fillColor: "#2563eb",
            fillOpacity: 0.9,
            strokeColor: "#ffffff",
            strokeWeight: 3,
          },
          animation: google.maps.Animation.DROP,
        })

        const infoWindow = new google.maps.InfoWindow({
          content: `
            <div class="p-3 min-w-[200px]">
              <h3 class="font-bold text-base text-blue-600 mb-2">🎯 Stop ${index + 1}</h3>
              <p class="text-sm text-gray-700 mb-1"><strong>📍 Address:</strong> ${stop.address}</p>
              <p class="text-sm text-gray-600 mb-1"><strong>⏱️ Duration:</strong> ${stop.duration} minutes</p>
              <p class="text-xs text-gray-500 mt-2">Click on parking meters nearby to see rates!</p>
            </div>
          `,
        })

        // Show info window on hover
        marker.addListener("mouseover", () => {
          infoWindow.open(map, marker)
        })

        // Hide info window when mouse leaves (with delay)
        marker.addListener("mouseout", () => {
          setTimeout(() => {
            infoWindow.close()
          }, 2000)
        })

        // Keep open on click
        marker.addListener("click", () => {
          infoWindow.open(map, marker)
        })

        newMarkers.push(marker)
      }
    })

    setMarkers(newMarkers)
  }, [map, stops])

  // Add parking meter markers
  useEffect(() => {
    if (!map || !route) return

    route.forEach((segment, index) => {
      const parkingMarker = new google.maps.Marker({
        position: { lat: segment.parking_meter.lat, lng: segment.parking_meter.lng },
        map: map,
        title: `Parking Meter ${segment.parking_meter.meter_id}`,
        icon: {
          url: "data:image/svg+xml;charset=UTF-8," + encodeURIComponent(`
            <svg width="30" height="30" viewBox="0 0 30 30" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="15" cy="15" r="12" fill="#10b981" stroke="#ffffff" stroke-width="3"/>
              <text x="15" y="19" text-anchor="middle" fill="white" font-size="12" font-weight="bold">P</text>
            </svg>
          `),
          scaledSize: new google.maps.Size(30, 30),
          anchor: new google.maps.Point(15, 15),
        },
        animation: google.maps.Animation.DROP,
      })

      const infoWindow = new google.maps.InfoWindow({
        content: `
          <div class="p-4 min-w-[250px]">
            <h3 class="font-bold text-base text-green-600 mb-3">🅿️ Parking Meter #${segment.parking_meter.meter_id}</h3>
            <div class="space-y-2">
              <p class="text-sm"><strong>📍 Area:</strong> ${segment.parking_meter.local_area}</p>
              <p class="text-sm"><strong>🏷️ Type:</strong> ${segment.parking_meter.meter_type}</p>
              <div class="bg-green-50 p-2 rounded">
                <p class="text-sm font-semibold text-green-800"><strong>💰 Rates:</strong></p>
                <p class="text-xs text-green-700">• 9AM-6PM: $${segment.parking_meter.rate_mf_9a_6p}/hour</p>
                <p class="text-xs text-green-700">• 6PM-10PM: $${segment.parking_meter.rate_mf_6p_10}/hour</p>
              </div>
              <p class="text-sm"><strong>💳 Credit Card:</strong> ${segment.parking_meter.credit_card ? "✅ Accepted" : "❌ Cash Only"}</p>
              <div class="bg-blue-50 p-2 rounded mt-2">
                <p class="text-sm font-bold text-blue-800">💵 Your Cost: $${(segment as any).parking_cost?.toFixed(2) || '0.00'}</p>
                <p class="text-xs text-blue-600">Based on your planned visit duration</p>
              </div>
            </div>
          </div>
        `,
      })

      // Show info window on hover
      parkingMarker.addListener("mouseover", () => {
        infoWindow.open(map, parkingMarker)
      })

      // Hide info window when mouse leaves (with delay)
      parkingMarker.addListener("mouseout", () => {
        setTimeout(() => {
          infoWindow.close()
        }, 3000)
      })

      // Keep open on click
      parkingMarker.addListener("click", () => {
        infoWindow.open(map, parkingMarker)
      })

      setMarkers(prev => [...prev, parkingMarker])
    })
  }, [map, route])

  // Draw route
  useEffect(() => {
    if (!directionsService || !directionsRenderer || !route || route.length === 0) return

    const waypoints = route.map(segment => ({
      location: { lat: segment.from_stop.lat, lng: segment.from_stop.lng },
      stopover: true,
    }))

    const request: google.maps.DirectionsRequest = {
      origin: { lat: route[0].from_stop.lat, lng: route[0].from_stop.lng },
      destination: { lat: route[route.length - 1].to_stop.lat, lng: route[route.length - 1].to_stop.lng },
      waypoints: waypoints.slice(1, -1),
      travelMode: google.maps.TravelMode.DRIVING,
      optimizeWaypoints: true,
    }

    directionsService.route(request, (result, status) => {
      if (status === google.maps.DirectionsStatus.OK && result) {
        directionsRenderer.setDirections(result)
      }
    })
  }, [directionsService, directionsRenderer, route])

  return <div ref={mapRef} className={className || "w-full h-full"} />
}

const render = (status: Status): React.ReactElement => {
  switch (status) {
    case Status.LOADING:
      return (
        <div className="flex items-center justify-center h-full">
          <div className="text-center space-y-4">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
            <p className="text-sm text-muted-foreground">Loading Google Maps...</p>
          </div>
        </div>
      )
    case Status.FAILURE:
      return (
        <div className="flex items-center justify-center h-full">
          <div className="text-center space-y-4">
            <div className="w-16 h-16 bg-destructive/10 rounded-full flex items-center justify-center mx-auto">
              <span className="text-destructive text-2xl">⚠️</span>
            </div>
            <h3 className="text-lg font-semibold text-foreground">Map Loading Failed</h3>
            <p className="text-sm text-muted-foreground max-w-md">
              Please check your Google Maps API key and internet connection.
            </p>
          </div>
        </div>
      )
    default:
      return <div></div>
  }
}

export default function GoogleMap({ stops, route, className }: GoogleMapProps) {
  const apiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || process.env.GOOGLE_MAPS_API_KEY

  if (!apiKey || apiKey === "your_google_maps_api_key_here") {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center space-y-4 p-8">
          <div className="w-16 h-16 bg-yellow-100 dark:bg-yellow-900/20 rounded-full flex items-center justify-center mx-auto">
            <span className="text-yellow-600 dark:text-yellow-400 text-2xl">🔑</span>
          </div>
          <h3 className="text-lg font-semibold text-foreground">Google Maps API Key Required</h3>
          <p className="text-sm text-muted-foreground max-w-md">
            Please add your Google Maps API key to the .env.local file to view the interactive map.
          </p>
          <div className="bg-muted/50 p-3 rounded text-xs text-left max-w-sm mx-auto">
            <p className="font-mono">NEXT_PUBLIC_GOOGLE_MAPS_API_KEY=your_actual_api_key</p>
            <p className="font-mono mt-1">or GOOGLE_MAPS_API_KEY=your_actual_api_key</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <Wrapper apiKey={apiKey} render={render}>
      <MapComponent stops={stops} route={route} className={className} />
    </Wrapper>
  )
}