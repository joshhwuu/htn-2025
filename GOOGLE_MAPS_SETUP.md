# Google Maps API Setup

## Overview
The search functionality now uses Google Maps Places API for real-time destination autocomplete. When users type in the search bar, they'll get actual location suggestions from Google Places, and selecting a destination will provide accurate coordinates.

## Setup Instructions

### 1. Get a Google Maps API Key
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the following APIs:
   - **Places API** (for autocomplete suggestions)
   - **Geocoding API** (for getting coordinates)
   - **Maps JavaScript API** (if using maps display)
4. Go to "Credentials" and create an API Key
5. Restrict the API key (recommended):
   - Set HTTP referrers for web usage
   - Limit to only the APIs you need

### 2. Configure Environment Variables
Create a `.env.local` file in the client directory:

```bash
# In /client/.env.local
GOOGLE_MAPS_API_KEY=your_actual_api_key_here
```

### 3. Restart Development Server
After adding the environment variable, restart your Next.js development server:

```bash
cd client
npm run dev
```

## How It Works

### Autocomplete Flow
1. User types 3+ characters in the destination search bar
2. Component calls `/api/places/autocomplete` with the input
3. API route forwards request to Google Places Autocomplete API
4. Results are displayed as a dropdown with main text and secondary text
5. User clicks on a suggestion

### Place Selection Flow
1. When a place is selected, the component calls `/api/places/details`
2. API route gets detailed information including coordinates from Google
3. The destination is added to the trip with accurate lat/lng coordinates
4. This enables precise parking calculations and routing

### Fallback Behavior
- If the API key is missing or invalid, mock suggestions are shown
- If the API call fails, fallback mock coordinates are used
- The app continues to function even without the API key (with reduced accuracy)

## API Endpoints Created

### `/api/places/autocomplete`
- **Method**: GET
- **Parameters**: `input` (search query)
- **Returns**: Array of place predictions with structured formatting
- **Features**: Restricted to Canada, focused on Vancouver area

### `/api/places/details`
- **Method**: GET  
- **Parameters**: `place_id` (from autocomplete results)
- **Returns**: Detailed place information including exact coordinates
- **Usage**: Called automatically when user selects a destination

## Testing
1. Start typing in any destination search bar (e.g., "Granville")
2. You should see real Vancouver locations appear
3. Select a location and verify it appears in the destination list
4. The coordinates should be accurate (not random mock values)

## Cost Considerations
- Autocomplete API: ~$2.83 per 1,000 requests
- Place Details API: ~$17 per 1,000 requests  
- Set up billing alerts and quotas in Google Cloud Console
- Consider implementing request caching for production

## Security Notes
- API key is stored server-side only (not exposed to client)
- All requests go through your Next.js API routes
- Consider adding rate limiting for production use
