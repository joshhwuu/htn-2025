"use client"

import { useState, useEffect, useRef } from "react"
import { Input } from "./input"

interface PlaceResult {
  place_id: string
  description: string
  structured_formatting: {
    main_text: string
    secondary_text: string
  }
}

interface PlacesAutocompleteProps {
  value: string | number
  onChange: (value: string) => void
  onPlaceSelect?: (place: { address: string; lat: number; lng: number }) => void
  placeholder?: string
  className?: string
  type?: string
  min?: string
  max?: string
  step?: string
  defaultValue?: string
}

export function PlacesAutocomplete({
  value,
  onChange,
  onPlaceSelect,
  placeholder = "Search for a location",
  className,
  type,
  min,
  max,
  step,
  defaultValue,
}: PlacesAutocompleteProps) {
  const [suggestions, setSuggestions] = useState<PlaceResult[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const timeoutRef = useRef<NodeJS.Timeout>()

  // If this is not a places autocomplete (has type prop), render as regular input
  if (type) {
    return (
      <Input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className={className}
        min={min}
        max={max}
        step={step}
        defaultValue={defaultValue}
      />
    )
  }

  useEffect(() => {
    const stringValue = String(value)
    if (stringValue.length < 3) {
      setSuggestions([])
      setShowSuggestions(false)
      return
    }

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(async () => {
      setIsLoading(true)
      try {
        const apiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || process.env.GOOGLE_MAPS_API_KEY
        if (!apiKey || apiKey === "your_google_maps_api_key_here") {
          // Mock suggestions when API key is not available
          const mockSuggestions: PlaceResult[] = [
            {
              place_id: "1",
              description: `${stringValue}, Vancouver, BC, Canada`,
              structured_formatting: {
                main_text: stringValue,
                secondary_text: "Vancouver, BC, Canada"
              }
            },
            {
              place_id: "2", 
              description: `${stringValue} Street, Vancouver, BC, Canada`,
              structured_formatting: {
                main_text: `${stringValue} Street`,
                secondary_text: "Vancouver, BC, Canada"
              }
            }
          ]
          
          setSuggestions(mockSuggestions)
          setShowSuggestions(true)
          return
        }

        // Use actual Google Places API
        const response = await fetch(
          `https://maps.googleapis.com/maps/api/place/autocomplete/json?input=${encodeURIComponent(
            stringValue
          )}&key=${apiKey}&types=establishment|geocode&components=country:ca`
        )
        
        if (response.ok) {
          const data = await response.json()
          setSuggestions(data.predictions || [])
          setShowSuggestions(true)
        }
      } catch (error) {
        console.error("Error fetching places:", error)
      } finally {
        setIsLoading(false)
      }
    }, 300)

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [value])

  const handlePlaceSelect = (place: PlaceResult) => {
    onChange(place.description)
    
    // Mock coordinates for Vancouver area
    const mockCoordinates = {
      address: place.description,
      lat: 49.2827 + (Math.random() - 0.5) * 0.1,
      lng: -123.1207 + (Math.random() - 0.5) * 0.1
    }
    
    onPlaceSelect?.(mockCoordinates)
    setShowSuggestions(false)
    setSuggestions([])
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value)
  }

  const handleBlur = () => {
    setTimeout(() => setShowSuggestions(false), 200)
  }

  return (
    <div className="relative">
      <Input
        ref={inputRef}
        value={String(value)}
        onChange={handleInputChange}
        onBlur={handleBlur}
        onFocus={() => String(value).length >= 3 && setShowSuggestions(true)}
        placeholder={placeholder}
        className={className}
        defaultValue={defaultValue}
      />
      
      {isLoading && (
        <div className="absolute right-3 top-1/2 transform -translate-y-1/2">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
        </div>
      )}

      {showSuggestions && suggestions.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-background border border-border rounded-md shadow-lg max-h-60 overflow-y-auto">
          {suggestions.map((suggestion) => (
            <button
              key={suggestion.place_id}
              type="button"
              className="w-full px-4 py-2 text-left hover:bg-muted focus:bg-muted focus:outline-none"
              onClick={() => handlePlaceSelect(suggestion)}
            >
              <div className="font-medium text-sm">
                {suggestion.structured_formatting.main_text}
              </div>
              <div className="text-xs text-muted-foreground">
                {suggestion.structured_formatting.secondary_text}
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}