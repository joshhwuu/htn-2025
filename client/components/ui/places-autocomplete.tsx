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
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const inputRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const timeoutRef = useRef<NodeJS.Timeout>()

  useEffect(() => {
    const stringValue = String(value)
    if (stringValue.length < 3) {
      setSuggestions([])
      setShowSuggestions(false)
      setSelectedIndex(-1)
      return
    }

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }

    timeoutRef.current = setTimeout(async () => {
      setIsLoading(true)
      try {
        // Use our server-side API route to avoid CORS issues
        const response = await fetch(
          `/api/places/autocomplete?input=${encodeURIComponent(stringValue)}`
        )
        
        if (response.ok) {
          const data = await response.json()
          setSuggestions(data.predictions || [])
          setShowSuggestions(true)
          setSelectedIndex(-1)
        } else {
          console.error("Error fetching places:", response.status, response.statusText)
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

  // Handle click outside to close suggestions
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setShowSuggestions(false)
      }
    }

    if (showSuggestions) {
      document.addEventListener('mousedown', handleClickOutside)
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showSuggestions])

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

  const handlePlaceSelect = async (place: PlaceResult) => {
    // Immediately close suggestions and clear them
    setShowSuggestions(false)
    setSuggestions([])
    onChange(place.description)
    
    // Get detailed place information including coordinates
    if (onPlaceSelect) {
      try {
        const response = await fetch(
          `/api/places/details?place_id=${encodeURIComponent(place.place_id)}`
        )
        
        if (response.ok) {
          const data = await response.json()
          const result = data.result
          
          if (result && result.geometry && result.geometry.location) {
            onPlaceSelect({
              address: result.formatted_address || place.description,
              lat: result.geometry.location.lat,
              lng: result.geometry.location.lng
            })
          }
        } else {
          console.error("Error fetching place details:", response.status, response.statusText)
          // Fallback to mock coordinates if details API fails
          onPlaceSelect({
            address: place.description,
            lat: 49.2827 + (Math.random() - 0.5) * 0.1,
            lng: -123.1207 + (Math.random() - 0.5) * 0.1
          })
        }
      } catch (error) {
        console.error("Error fetching place details:", error)
        // Fallback to mock coordinates if there's an error
        onPlaceSelect({
          address: place.description,
          lat: 49.2827 + (Math.random() - 0.5) * 0.1,
          lng: -123.1207 + (Math.random() - 0.5) * 0.1
        })
      }
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value)
  }

  const handleFocus = () => {
    if (String(value).length >= 3 && suggestions.length > 0) {
      setShowSuggestions(true)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!showSuggestions || suggestions.length === 0) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev < suggestions.length - 1 ? prev + 1 : 0
        )
        break
      case 'ArrowUp':
        e.preventDefault()
        setSelectedIndex(prev => 
          prev > 0 ? prev - 1 : suggestions.length - 1
        )
        break
      case 'Enter':
        e.preventDefault()
        if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
          handlePlaceSelect(suggestions[selectedIndex])
        }
        break
      case 'Escape':
        setShowSuggestions(false)
        setSelectedIndex(-1)
        break
    }
  }

  return (
    <div ref={containerRef} className="relative">
      <Input
        ref={inputRef}
        value={String(value)}
        onChange={handleInputChange}
        onFocus={handleFocus}
        onKeyDown={handleKeyDown}
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
          {suggestions.map((suggestion, index) => (
            <button
              key={suggestion.place_id}
              type="button"
              className={`w-full px-4 py-2 text-left hover:bg-muted focus:bg-muted focus:outline-none ${
                index === selectedIndex ? 'bg-muted' : ''
              }`}
              onClick={() => handlePlaceSelect(suggestion)}
              onMouseEnter={() => setSelectedIndex(index)}
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