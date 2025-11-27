import axios from 'axios';

class RoutingService {
  constructor() {
    const apiKey = import.meta.env.VITE_OPENROUTE_API_KEY;
    
    if (!apiKey || apiKey === 'your_api_key_here') {
      console.warn('OpenRouteService API key not configured. Please add VITE_OPENROUTE_API_KEY to .env file');
      console.warn('Get your free API key at: https://openrouteservice.org/dev/#/signup');
      this.apiKey = null;
    } else {
      this.apiKey = apiKey;
    }
    
    this.baseURL = 'https://api.openrouteservice.org/v2';
  }

  /**
   * Calculate route between multiple points
   * @param {Array} coordinates - Array of [lng, lat] pairs
   * @param {String} profile - Routing profile: 'driving-car', 'cycling-regular', 'foot-walking'
   * @returns {Promise<Object>} Route data with geometry, distance, duration
   */
  async calculateRoute(coordinates, profile = 'foot-walking') {
    if (!this.apiKey) {
      throw new Error('OpenRouteService API key not configured');
    }

    try {
      const response = await axios.post(
        `${this.baseURL}/directions/${profile}/geojson`,
        {
          coordinates: coordinates,
          instructions: true,
          elevation: false
        },
        {
          headers: {
            'Authorization': this.apiKey,
            'Content-Type': 'application/json'
          }
        }
      );

      const route = response.data.features[0];
      return {
        geometry: route.geometry.coordinates, // Array of [lng, lat] points along the route
        distance: route.properties.segments[0].distance, // Distance in meters
        duration: route.properties.segments[0].duration, // Duration in seconds
        segments: route.properties.segments, // Detailed segment information
        instructions: route.properties.segments[0].steps || [] // Turn-by-turn instructions
      };
    } catch (error) {
      console.error('Routing error:', error);
      throw new Error(`Failed to calculate route: ${error.message}`);
    }
  }

  /**
   * Calculate routes between consecutive keypoints
   * @param {Array} keypoints - Array of {latitude, longitude} objects
   * @param {String} profile - Routing profile
   * @returns {Promise<Array>} Array of route segments
   */
  async calculateTourRoute(keypoints, profile = 'foot-walking') {
    if (!keypoints || keypoints.length < 2) {
      return [];
    }

    const routes = [];
    
    // Calculate route for each consecutive pair of keypoints
    for (let i = 0; i < keypoints.length - 1; i++) {
      const start = keypoints[i];
      const end = keypoints[i + 1];
      
      const coordinates = [
        [start.longitude, start.latitude],
        [end.longitude, end.latitude]
      ];

      try {
        const route = await this.calculateRoute(coordinates, profile);
        routes.push({
          from: i,
          to: i + 1,
          route: route
        });
      } catch (error) {
        console.error(`Failed to calculate route from ${i} to ${i + 1}:`, error);
        // Fallback to straight line if routing fails
        routes.push({
          from: i,
          to: i + 1,
          route: null,
          error: error.message
        });
      }
    }

    return routes;
  }

  /**
   * Get total distance and duration for a tour
   * @param {Array} routes - Array of route segments from calculateTourRoute
   * @returns {Object} Total distance (m) and duration (s)
   */
  getTourStats(routes) {
    return routes.reduce((acc, segment) => {
      if (segment.route) {
        acc.distance += segment.route.distance;
        acc.duration += segment.route.duration;
      }
      return acc;
    }, { distance: 0, duration: 0 });
  }

  /**
   * Format distance for display
   * @param {Number} meters - Distance in meters
   * @returns {String} Formatted distance
   */
  formatDistance(meters) {
    if (meters < 1000) {
      return `${Math.round(meters)} m`;
    }
    return `${(meters / 1000).toFixed(2)} km`;
  }

  /**
   * Format duration for display
   * @param {Number} seconds - Duration in seconds
   * @returns {String} Formatted duration
   */
  formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  }

  /**
   * Get available routing profiles
   * @returns {Array} List of available profiles with labels
   */
  getProfiles() {
    return [
      { value: 'foot-walking', label: '🚶 Walking', icon: '🚶' },
      { value: 'cycling-regular', label: '🚴 Cycling', icon: '🚴' },
      { value: 'driving-car', label: '🚗 Driving', icon: '🚗' },
      { value: 'foot-hiking', label: '🥾 Hiking', icon: '🥾' }
    ];
  }

  /**
   * Check if the service is configured
   * @returns {Boolean}
   */
  isConfigured() {
    return this.apiKey !== null;
  }
}

export default new RoutingService();
