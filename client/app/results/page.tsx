"use client"

import { useState, useEffect, Suspense } from "react"
import { useSearchParams, useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { MapPin, ArrowLeft, Clock, DollarSign, Navigation, Car, AlertCircle } from "lucide-react"
import Link from "next/link"
import { vancouverAPI } from "@/lib/api"
import GoogleMap from "@/components/maps/GoogleMap"

interface Stop {
  id: string
  address: string
  duration: number
  lat?: number
  lng?: number
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

function ResultsContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const [stops, setStops] = useState<Stop[]>([])
  const [plans, setPlans] = useState<TripPlan[]>([])
  const [selectedPlan, setSelectedPlan] = useState<"cheapest" | "fastest" | "hybrid">("hybrid")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const stopsParam = searchParams.get("stops")
    const costWeight = Number.parseFloat(searchParams.get("costWeight") || "0.5")
    const timeWeight = Number.parseFloat(searchParams.get("timeWeight") || "0.5")
    const startTime = searchParams.get("startTime") || new Date(Date.now() + 60 * 60 * 1000).toISOString()

    if (stopsParam) {
      try {
        const parsedStops = JSON.parse(stopsParam)
        setStops(parsedStops)
        
        // Determine the preferred plan type based on weights
        let preferredPlan: "cheapest" | "fastest" | "hybrid" = "hybrid"
        if (costWeight >= 0.7) {
          preferredPlan = "cheapest"
        } else if (timeWeight >= 0.7) {
          preferredPlan = "fastest"
        }
        
        setSelectedPlan(preferredPlan)
        fetchTripPlans(parsedStops, costWeight, timeWeight, startTime)
      } catch (error) {
        console.error("Error parsing stops:", error)
        setError("Invalid search parameters")
        setIsLoading(false)
      }
    } else {
      router.push("/search")
    }
  }, [searchParams, router])

  const fetchTripPlans = async (stops: Stop[], costWeight: number, timeWeight: number, startTime: string) => {
    setIsLoading(true)
    setError(null)

    // Validate that we have at least 2 stops
    if (stops.length < 2) {
      setError("At least 2 destinations are required for trip planning")
      setIsLoading(false)
      return
    }

    try {
      const response = await vancouverAPI.planTrip({
        stops: stops.map((stop, index) => ({
          id: stop.id || `stop_${index + 1}`,
          address: stop.address,
          duration_minutes: stop.duration,
        })),
        start_time: startTime,
        timezone: "America/Vancouver",
        preferences: {
          cost_weight: costWeight,
          time_weight: timeWeight,
        },
      })

      setPlans(response.plans)
    } catch (error) {
      console.error("Error fetching trip plans:", error)
      setError(error instanceof Error ? error.message : "Failed to fetch trip plans")

      // Fallback to mock data with coordinates
      const mockPlans: TripPlan[] = [
        {
          type: "cheapest",
          total_cost: 12.5,
          total_time_minutes: 180,
          route: [
            {
              from_stop: {
                id: "stop_1",
                address: stops[0]?.address || "Vancouver Art Gallery",
                lat: 49.2827,
                lng: -123.1207,
                duration_minutes: stops[0]?.duration || 60,
                arrival_time: "2024-01-15T14:30:00-08:00",
                departure_time: "2024-01-15T15:30:00-08:00",
              },
              to_stop: {
                id: "stop_2",
                address: stops[1]?.address || "Granville Island",
                lat: 49.2889,
                lng: -123.1111,
                duration_minutes: stops[1]?.duration || 90,
                arrival_time: "2024-01-15T15:45:00-08:00",
                departure_time: "2024-01-15T17:15:00-08:00",
              },
              parking_meter: {
                meter_id: "170127",
                lat: 49.2888,
                lng: -123.111,
                meter_type: "Twin",
                local_area: "Downtown",
                credit_card: false,
                rate_mf_9a_6p: 3.5,
                rate_mf_6p_10: 2.0,
              },
              travel_time_minutes: 12,
              parking_cost: 5.25,
              walking_time_minutes: 3,
            },
          ],
          metadata: {
            optimization: "cost",
            savings: "$3.25 vs fastest",
          },
        },
        {
          type: "fastest",
          total_cost: 15.75,
          total_time_minutes: 150,
          route: [
            {
              from_stop: {
                id: "stop_1",
                address: stops[0]?.address || "Vancouver Art Gallery",
                lat: 49.2827,
                lng: -123.1207,
                duration_minutes: stops[0]?.duration || 60,
                arrival_time: "2024-01-15T14:30:00-08:00",
                departure_time: "2024-01-15T15:30:00-08:00",
              },
              to_stop: {
                id: "stop_2",
                address: stops[1]?.address || "Granville Island",
                lat: 49.2889,
                lng: -123.1111,
                duration_minutes: stops[1]?.duration || 90,
                arrival_time: "2024-01-15T15:35:00-08:00",
                departure_time: "2024-01-15T17:05:00-08:00",
              },
              parking_meter: {
                meter_id: "170128",
                lat: 49.289,
                lng: -123.1112,
                meter_type: "Single",
                local_area: "Downtown",
                credit_card: true,
                rate_mf_9a_6p: 4.0,
                rate_mf_6p_10: 2.5,
              },
              travel_time_minutes: 8,
              parking_cost: 7.0,
              walking_time_minutes: 2,
            },
          ],
          metadata: {
            optimization: "time",
            time_saved: "30 minutes vs cheapest",
          },
        },
        {
          type: "hybrid",
          total_cost: 14.1,
          total_time_minutes: 165,
          route: [
            {
              from_stop: {
                id: "stop_1",
                address: stops[0]?.address || "Vancouver Art Gallery",
                lat: 49.2827,
                lng: -123.1207,
                duration_minutes: stops[0]?.duration || 60,
                arrival_time: "2024-01-15T14:30:00-08:00",
                departure_time: "2024-01-15T15:30:00-08:00",
              },
              to_stop: {
                id: "stop_2",
                address: stops[1]?.address || "Granville Island",
                lat: 49.2889,
                lng: -123.1111,
                duration_minutes: stops[1]?.duration || 90,
                arrival_time: "2024-01-15T15:40:00-08:00",
                departure_time: "2024-01-15T17:10:00-08:00",
              },
              parking_meter: {
                meter_id: "170129",
                lat: 49.2889,
                lng: -123.1111,
                meter_type: "Twin",
                local_area: "Downtown",
                credit_card: true,
                rate_mf_9a_6p: 3.75,
                rate_mf_6p_10: 2.25,
              },
              travel_time_minutes: 10,
              parking_cost: 6.15,
              walking_time_minutes: 2,
            },
          ],
          metadata: {
            optimization: "balanced",
            hybrid_score: 8.46,
          },
        },
      ]
      setPlans(mockPlans)
    } finally {
      setIsLoading(false)
    }
  }

  const currentPlan = plans.find((plan) => plan.type === selectedPlan)

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-center space-y-4">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <h2 className="text-xl font-semibold text-foreground">Finding optimal parking routes...</h2>
          <p className="text-muted-foreground">Analyzing Vancouver parking data and traffic patterns</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="flex items-center justify-between p-4 max-w-7xl mx-auto border-b border-border">
        <div className="flex items-center space-x-4">
          <Link href="/search">
            <Button variant="ghost" size="sm">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Search
            </Button>
          </Link>
          <div className="flex items-center space-x-2">
            <MapPin className="h-6 w-6 text-primary" />
            <h1 className="text-xl font-bold text-foreground font-[var(--font-playfair)]">WYHD</h1>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Button
            variant={selectedPlan === "hybrid" ? "default" : "outline"}
            size="sm"
            onClick={() => setSelectedPlan("hybrid")}
            className={selectedPlan === "hybrid" ? "bg-primary hover:bg-primary/90" : ""}
            title="Balanced approach considering both cost and time based on your preferences"
          >
            <div className="flex items-center space-x-1">
              <span>⚖️</span>
              <span>Balanced</span>
            </div>
          </Button>
          <Button
            variant={selectedPlan === "cheapest" ? "default" : "outline"}
            size="sm"
            onClick={() => setSelectedPlan("cheapest")}
            className={selectedPlan === "cheapest" ? "bg-green-600 hover:bg-green-700 text-white" : ""}
            title="Minimize parking costs - may take longer but saves money"
          >
            <div className="flex items-center space-x-1">
              <DollarSign className="h-4 w-4" />
              <span>Cheapest</span>
            </div>
          </Button>
          <Button
            variant={selectedPlan === "fastest" ? "default" : "outline"}
            size="sm"
            onClick={() => setSelectedPlan("fastest")}
            className={selectedPlan === "fastest" ? "bg-blue-600 hover:bg-blue-700 text-white" : ""}
            title="Minimize total travel time - may cost more but saves time"
          >
            <div className="flex items-center space-x-1">
              <Clock className="h-4 w-4" />
              <span>Fastest</span>
            </div>
          </Button>
        </div>
      </header>

      <div className="flex flex-col lg:flex-row max-w-7xl mx-auto">
        <div className="lg:w-2/3 h-96 lg:h-screen bg-muted/20 relative">
          <GoogleMap 
            stops={stops} 
            route={currentPlan?.route} 
            className="w-full h-full"
          />
        </div>

        <div className="lg:w-1/3 p-6 space-y-6 lg:h-screen lg:overflow-y-auto">
          {error && (
            <Card className="border-destructive/50 bg-destructive/5">
              <CardContent className="p-4">
                <div className="flex items-center space-x-2 text-destructive">
                  <AlertCircle className="h-4 w-4" />
                  <p className="text-sm font-medium">API Connection Issue</p>
                </div>
                <p className="text-sm text-muted-foreground mt-1">Using demo data. {error}</p>
              </CardContent>
            </Card>
          )}

          {/* Algorithm Comparison */}
          {plans.length > 0 && (
            <Card className="border-border/50">
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <span>📊</span>
                  <span>Route Comparison</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {plans.map((plan) => (
                    <div 
                      key={plan.type}
                      className={`p-3 rounded-lg border cursor-pointer transition-all ${
                        selectedPlan === plan.type 
                          ? 'border-primary bg-primary/5' 
                          : 'border-border hover:border-primary/50'
                      }`}
                      onClick={() => setSelectedPlan(plan.type as "cheapest" | "fastest" | "hybrid")}
                    >
                      <div className="flex justify-between items-center">
                        <div className="flex items-center space-x-2">
                          <span className="text-sm font-medium capitalize">
                            {plan.type === "hybrid" ? "⚖️ Balanced" : 
                             plan.type === "cheapest" ? "💰 Cheapest" : "⚡ Fastest"}
                          </span>
                        </div>
                        <div className="flex space-x-4 text-sm">
                          <span className="text-green-600 font-medium">${plan.total_cost.toFixed(2)}</span>
                          <span className="text-blue-600 font-medium">
                            {Math.floor(plan.total_time_minutes / 60)}h {plan.total_time_minutes % 60}m
                          </span>
                        </div>
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">
                        {plan.type === "cheapest" && "Finds the lowest cost parking options, may involve longer walks"}
                        {plan.type === "fastest" && "Minimizes total travel time including walking and driving"}
                        {plan.type === "hybrid" && "Balances cost and time based on your preferences"}
                      </p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {currentPlan && (
            <>
              <Card className="border-border/50">
                <CardHeader>
                  <CardTitle className="flex items-center justify-between">
                    <span className="capitalize">{currentPlan.type} Route</span>
                    <Badge variant="secondary" className="bg-primary/10 text-primary">
                      {currentPlan.metadata.optimization}
                    </Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="flex items-center space-x-2">
                      <DollarSign className="h-4 w-4 text-primary" />
                      <div>
                        <p className="text-sm text-muted-foreground">Total Cost</p>
                        <p className="font-semibold">${currentPlan.total_cost.toFixed(2)}</p>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Clock className="h-4 w-4 text-primary" />
                      <div>
                        <p className="text-sm text-muted-foreground">Total Time</p>
                        <p className="font-semibold">
                          {Math.floor(currentPlan.total_time_minutes / 60)}h {currentPlan.total_time_minutes % 60}m
                        </p>
                      </div>
                    </div>
                  </div>

                  {currentPlan.metadata.savings && (
                    <div className="bg-green-50 dark:bg-green-950/20 p-3 rounded-lg">
                      <p className="text-sm text-green-700 dark:text-green-300">
                        💰 Saves {currentPlan.metadata.savings}
                      </p>
                    </div>
                  )}

                  {currentPlan.metadata.time_saved && (
                    <div className="bg-blue-50 dark:bg-blue-950/20 p-3 rounded-lg">
                      <p className="text-sm text-blue-700 dark:text-blue-300">
                        ⚡ Saves {currentPlan.metadata.time_saved}
                      </p>
                    </div>
                  )}
                </CardContent>
              </Card>

              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-foreground">Route Details</h3>

                {currentPlan.route.map((segment, index) => (
                  <Card key={index} className="border-border/50">
                    <CardContent className="p-4 space-y-3">
                      <div className="flex items-start space-x-3">
                        <div className="flex-shrink-0 w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-medium mt-1">
                          {index + 1}
                        </div>
                        <div className="flex-1 space-y-2">
                          <h4 className="font-medium text-foreground text-sm">{segment.to_stop.address}</h4>

                          <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                            <div className="flex items-center space-x-1">
                              <Car className="h-3 w-3" />
                              <span>{segment.travel_time_minutes}min drive</span>
                            </div>
                            <div className="flex items-center space-x-1">
                              <Navigation className="h-3 w-3" />
                              <span>{segment.walking_time_minutes}min walk</span>
                            </div>
                          </div>

                          <div className="bg-muted/50 p-2 rounded text-xs space-y-1">
                            <div className="flex justify-between">
                              <span className="text-muted-foreground">Parking:</span>
                              <span className="font-medium">${segment.parking_cost.toFixed(2)}</span>
                            </div>
                            <div className="flex justify-between">
                              <span className="text-muted-foreground">Meter #{segment.parking_meter.meter_id}</span>
                              <span
                                className={segment.parking_meter.credit_card ? "text-green-600" : "text-orange-600"}
                              >
                                {segment.parking_meter.credit_card ? "Card OK" : "Coins only"}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>

              <Button className="w-full bg-primary hover:bg-primary/90" size="lg">
                Start Navigation
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export default function ResultsPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-background flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      }
    >
      <ResultsContent />
    </Suspense>
  )
}