"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Slider } from "@/components/ui/slider"
import { MapPin, Search, ArrowLeft, Plus, X, DollarSign, Clock } from "lucide-react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { PlacesAutocomplete } from "@/components/ui/places-autocomplete"

interface Stop {
  id: string
  address: string
  duration: number
  lat?: number
  lng?: number
}

export default function SearchPage() {
  const router = useRouter()
  const [stops, setStops] = useState<Stop[]>([{ id: "1", address: "", duration: 60 }])
  const [isLoading, setIsLoading] = useState(false)
  const [costWeight, setCostWeight] = useState(0.5)
  const [timeWeight, setTimeWeight] = useState(0.5)

  const addStop = () => {
    const newStop: Stop = {
      id: Date.now().toString(),
      address: "",
      duration: 60,
    }
    setStops([...stops, newStop])
  }

  const removeStop = (id: string) => {
    if (stops.length > 1) {
      setStops(stops.filter((stop) => stop.id !== id))
    }
  }

  const updateStop = (id: string, field: "address" | "duration", value: string | number) => {
    setStops(stops.map((stop) => 
      stop.id === id 
        ? { ...stop, [field]: value }
        : stop
    ))
  }

  const handlePlaceSelect = (id: string, place: { address: string; lat: number; lng: number }) => {
    setStops(stops.map((stop) => 
      stop.id === id 
        ? { ...stop, address: place.address, lat: place.lat, lng: place.lng }
        : stop
    ))
  }

  const handlePreferenceChange = (type: "cost" | "time", value: number) => {
    const weight = value / 100
    if (type === "cost") {
      setCostWeight(weight)
      setTimeWeight(1 - weight)
    } else {
      setTimeWeight(weight)
      setCostWeight(1 - weight)
    }
  }

  const handleSearch = async () => {
    const validStops = stops.filter((stop) => stop.address.trim() !== "")
    if (validStops.length < 1) {
      alert("Please enter at least 1 destination")
      return
    }

    setIsLoading(true)

    setTimeout(() => {
      const searchParams = new URLSearchParams()
      searchParams.set("stops", JSON.stringify(validStops))
      searchParams.set("costWeight", costWeight.toString())
      searchParams.set("timeWeight", timeWeight.toString())
      router.push(`/results?${searchParams.toString()}`)
    }, 1000)
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="flex items-center justify-between p-6 max-w-7xl mx-auto border-b border-border">
        <div className="flex items-center space-x-4">
          <Link href="/">
            <Button variant="ghost" size="sm">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back
            </Button>
          </Link>
          <div className="flex items-center space-x-2">
            <MapPin className="h-6 w-6 text-primary" />
            <h1 className="text-xl font-bold text-foreground font-[var(--font-playfair)]">WYHD</h1>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-12">
        <div className="text-center mb-12">
          <h2 className="text-4xl font-bold text-foreground font-[var(--font-playfair)] mb-4 text-balance">
            Plan Your Vancouver Trip
          </h2>
          <p className="text-lg text-muted-foreground text-pretty">
            Add your destinations and we'll find the best parking options for your journey
          </p>
        </div>

        <Card className="border-border/50">
          <CardContent className="p-8">
            <div className="space-y-8">
              <div className="space-y-4">
                <h3 className="text-lg font-semibold text-foreground mb-4">Your Destinations</h3>

                {stops.map((stop, index) => (
                  <div key={stop.id} className="flex items-center space-x-4 p-4 bg-muted/30 rounded-lg">
                    <div className="flex-shrink-0 w-8 h-8 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-sm font-medium">
                      {index + 1}
                    </div>

                    <div className="flex-1 space-y-3">
                      <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <PlacesAutocomplete
                          placeholder="Search for your destination"
                          value={stop.address}
                          onChange={(value) => updateStop(stop.id, "address", value)}
                          onPlaceSelect={(place) => handlePlaceSelect(stop.id, place)}
                          className="pl-10 text-base"
                        />
                      </div>

                      <div className="flex items-center space-x-2">
                        <label className="text-sm text-muted-foreground whitespace-nowrap">Duration:</label>
                        <Input
                          type="number"
                          min="15"
                          max="480"
                          step="15"
                          value={stop.duration}
                          onChange={(e) => updateStop(stop.id, "duration", Number.parseInt(e.target.value) || 60)}
                          className="w-20"
                        />
                        <span className="text-sm text-muted-foreground">minutes</span>
                      </div>
                    </div>

                    {stops.length > 1 && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => removeStop(stop.id)}
                        className="flex-shrink-0 text-muted-foreground hover:text-destructive"
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                ))}

                <Button
                  variant="outline"
                  onClick={addStop}
                  className="w-full border-dashed border-2 py-6 text-muted-foreground hover:text-foreground hover:border-primary/50 bg-transparent"
                >
                  <Plus className="h-4 w-4 mr-2" />
                  Add Another Destination
                </Button>
              </div>

              <div className="space-y-6 border-t border-border pt-6">
                <h3 className="text-lg font-semibold text-foreground">Optimization Preferences</h3>

                <div className="grid md:grid-cols-2 gap-6">
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <DollarSign className="h-4 w-4 text-primary" />
                        <label className="text-sm font-medium text-foreground">Cost Priority</label>
                      </div>
                      <span className="text-sm text-muted-foreground">{Math.round(costWeight * 100)}%</span>
                    </div>
                    <Slider
                      value={[costWeight * 100]}
                      onValueChange={(value) => handlePreferenceChange("cost", value[0])}
                      max={100}
                      step={5}
                      className="w-full"
                    />
                    <p className="text-xs text-muted-foreground">
                      Higher values prioritize finding cheaper parking options
                    </p>
                  </div>

                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Clock className="h-4 w-4 text-primary" />
                        <label className="text-sm font-medium text-foreground">Time Priority</label>
                      </div>
                      <span className="text-sm text-muted-foreground">{Math.round(timeWeight * 100)}%</span>
                    </div>
                    <Slider
                      value={[timeWeight * 100]}
                      onValueChange={(value) => handlePreferenceChange("time", value[0])}
                      max={100}
                      step={5}
                      className="w-full"
                    />
                    <p className="text-xs text-muted-foreground">
                      Higher values prioritize faster routes and shorter walking distances
                    </p>
                  </div>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setCostWeight(0.8)
                      setTimeWeight(0.2)
                    }}
                    className="text-xs"
                  >
                    Cheapest
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setCostWeight(0.5)
                      setTimeWeight(0.5)
                    }}
                    className="text-xs"
                  >
                    Balanced
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setCostWeight(0.2)
                      setTimeWeight(0.8)
                    }}
                    className="text-xs"
                  >
                    Fastest
                  </Button>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-foreground">Start Time</label>
                <Input
                  type="datetime-local"
                  defaultValue={new Date(Date.now() + 60 * 60 * 1000).toISOString().slice(0, 16)}
                  className="w-full md:w-auto"
                />
              </div>

              <Button
                onClick={handleSearch}
                disabled={isLoading}
                size="lg"
                className="w-full bg-primary hover:bg-primary/90 text-lg py-6"
              >
                {isLoading ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-foreground mr-2"></div>
                    Finding Best Routes...
                  </>
                ) : (
                  <>
                    <Search className="h-5 w-5 mr-2" />
                    Find Parking Options
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>

        <div className="mt-8 grid md:grid-cols-2 gap-4">
          <Card className="border-border/50">
            <CardContent className="p-4">
              <h4 className="font-medium text-foreground mb-2">💡 Pro Tip</h4>
              <p className="text-sm text-muted-foreground">
                Add multiple stops to optimize your entire trip route and save on parking costs.
              </p>
            </CardContent>
          </Card>

          <Card className="border-border/50">
            <CardContent className="p-4">
              <h4 className="font-medium text-foreground mb-2">⏰ Best Times</h4>
              <p className="text-sm text-muted-foreground">
                Parking is free after 10 PM and before 9 AM on weekdays in most Vancouver areas.
              </p>
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  )
}
