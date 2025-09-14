# Google Maps API Setup

This application now includes Google Maps Places Autocomplete functionality for the search bar.

## Setup Instructions

1. **Get a Google Maps API Key**
   - Go to [Google Cloud Console](https://console.developers.google.com/)
   - Create a new project or select an existing one
   - Enable the following APIs:
     - Maps JavaScript API
     - Places API
     - Geocoding API

2. **Configure Environment Variables**
   - Create a `.env.local` file in the client directory
   - Add your API key:
   ```
   GOOGLE_MAPS_API_KEY=your_actual_api_key_here
   ```

3. **API Key Restrictions (Recommended)**
   - In the Google Cloud Console, restrict your API key:
     - For HTTP referrers: Add your domain (e.g., `localhost:3000/*` for development)
     - For APIs: Limit to only the APIs you need

## Features

- **Real-time Autocomplete**: As users type in the search bar, they get suggestions from Google Places API
- **Coordinate Resolution**: When a place is selected, the system fetches precise coordinates
- **Fallback Support**: Works with mock data when no API key is configured (for development)
- **CORS-Safe**: Uses server-side API routes to avoid browser CORS restrictions

## API Endpoints

- `GET /api/places/autocomplete?input=query` - Get place suggestions
- `GET /api/places/details?place_id=id` - Get detailed place information with coordinates

## Development Mode

Without an API key, the system will use mock data that simulates Vancouver-area locations. This allows development to continue without requiring immediate API setup.