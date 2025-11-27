<template>
  <div class="tour-map-view">
    <div class="header">
      <div>
        <h1>{{ tour.name }} - Route Map</h1>
        <p class="subtitle">View the complete tour route and key points</p>
      </div>
      <button @click="goBack" class="btn-secondary">
        ← Back to Tour
      </button>
    </div>

    <div v-if="error" class="error-message">
      <span class="error-icon">⚠️</span>
      {{ error }}
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>Loading map...</p>
    </div>

    <div v-else class="map-container">
      <div id="map" ref="mapContainer"></div>
      
      <div class="map-overlay">
        <div class="map-info-card">
          <div class="info-header">
            <h3>📍 Tour Information</h3>
          </div>
          <div class="info-content">
            <div class="info-item">
              <span class="info-label">Total Points:</span>
              <span class="info-value">{{ keypoints.length }}</span>
            </div>
            <div v-if="keypoints.length > 1" class="info-item">
              <span class="info-label">Total Distance:</span>
              <span class="info-value">
                <span v-if="isCalculatingRoute" class="calculating">
                  <span class="pulse-dot"></span>
                  Calculating...
                </span>
                <span v-else>{{ totalDistance.toFixed(2) }} km</span>
              </span>
            </div>
            <div v-if="tour.durations" class="info-item travel-times">
              <span class="info-label">Travel Times:</span>
              <div class="travel-times-list">
                <div v-if="tour.durations.walking > 0" class="travel-time-item">
                  <span class="travel-icon">🚶</span>
                  <span>{{ formatDuration(tour.durations.walking) }}</span>
                </div>
                <div v-if="tour.durations.biking > 0" class="travel-time-item">
                  <span class="travel-icon">🚴</span>
                  <span>{{ formatDuration(tour.durations.biking) }}</span>
                </div>
                <div v-if="tour.durations.driving > 0" class="travel-time-item">
                  <span class="travel-icon">🚗</span>
                  <span>{{ formatDuration(tour.durations.driving) }}</span>
                </div>
              </div>
            </div>
            <div class="info-item">
              <span class="info-label">Difficulty:</span>
              <span class="info-value difficulty-badge" :class="tour.difficulty">
                {{ tour.difficulty }}
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">Price:</span>
              <span class="info-value">${{ tour.price }}</span>
            </div>
          </div>
        </div>

        <div class="keypoints-list-card">
          <div class="list-header">
            <h3>🗺️ Key Points</h3>
          </div>
          <div class="keypoints-list">
            <div 
              v-for="(kp, index) in keypoints" 
              :key="kp.id" 
              class="keypoint-item"
              @click="focusOnPoint(kp)"
            >
              <div class="keypoint-number">{{ index + 1 }}</div>
              <div class="keypoint-info">
                <h4>{{ kp.name }}</h4>
                <p v-if="kp.description" class="keypoint-desc">{{ kp.description }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../services/api'
import routingService from '../services/routingService'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

export default {
  name: 'TourMapView',
  setup() {
    const route = useRoute()
    const router = useRouter()
    const tourId = route.params.id

    const tour = ref({
      name: '',
      difficulty: '',
      price: 0,
      durations: {
        walking: 0,
        biking: 0,
        driving: 0
      }
    })
    const keypoints = ref([])
    const loading = ref(true)
    const error = ref('')
    const totalDistance = ref(0)
    const isCalculatingRoute = ref(false)

    const mapContainer = ref(null)
    let map = null
    let markers = []
    let routePolyline = null

    const initMap = () => {
      map = L.map(mapContainer.value).setView([45.2671, 19.8335], 13)

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors',
        maxZoom: 19
      }).addTo(map)
    }

    const renderMarkers = () => {
      // Clear existing markers
      markers.forEach(marker => map.removeLayer(marker))
      markers = []

      // Add numbered markers for each keypoint
      keypoints.value.forEach((kp, index) => {
        const customIcon = L.divIcon({
          className: 'custom-marker',
          html: `
            <div class="marker-pin">
              <span class="marker-number">${index + 1}</span>
            </div>
          `,
          iconSize: [40, 40],
          iconAnchor: [20, 40]
        })

        const marker = L.marker([kp.latitude, kp.longitude], { icon: customIcon })
          .bindPopup(`
            <div class="marker-popup">
              <h4>${kp.name}</h4>
              ${kp.description ? `<p>${kp.description}</p>` : ''}
            </div>
          `)
          .addTo(map)

        markers.push(marker)
      })

      // Fit map to show all markers
      if (keypoints.value.length > 0) {
        const bounds = L.latLngBounds(keypoints.value.map(kp => [kp.latitude, kp.longitude]))
        map.fitBounds(bounds, { padding: [100, 100] })
      }
    }

    const calculateAndDrawRoute = async () => {
      if (keypoints.value.length < 2) return

      // Clear existing route
      if (routePolyline) {
        map.removeLayer(routePolyline)
        routePolyline = null
      }

      isCalculatingRoute.value = true

      if (!routingService.isConfigured()) {
        // Fallback to straight line if no API key
        calculateStraightLineRoute()
        isCalculatingRoute.value = false
        return
      }

      try {
        // Use foot-walking as default for tour viewing
        const routes = await routingService.calculateTourRoute(
          keypoints.value,
          'foot-walking'
        )

        // Draw all route segments
        const allCoords = []
        routes.forEach((segment) => {
          if (segment.route && segment.route.geometry) {
            // Convert [lng, lat] to [lat, lng] for Leaflet
            const latLngs = segment.route.geometry.map(coord => [coord[1], coord[0]])
            allCoords.push(...latLngs)
          }
        })

        if (allCoords.length > 0) {
          routePolyline = L.polyline(allCoords, {
            color: '#42b983',
            weight: 4,
            opacity: 0.7,
            smoothFactor: 1
          }).addTo(map)

          // Calculate total distance from routes
          const stats = routingService.getTourStats(routes)
          totalDistance.value = stats.distance / 1000 // convert to km
        } else {
          // Fallback if no route geometry
          calculateStraightLineRoute()
        }
      } catch (err) {
        console.error('Error calculating route:', err)
        // Fallback to straight line
        calculateStraightLineRoute()
      } finally {
        isCalculatingRoute.value = false
      }
    }

    const calculateStraightLineRoute = () => {
      const routeCoordinates = keypoints.value.map(kp => [kp.latitude, kp.longitude])
      routePolyline = L.polyline(routeCoordinates, {
        color: '#ff9800',
        weight: 4,
        opacity: 0.7,
        dashArray: '10, 10',
        smoothFactor: 1
      }).addTo(map)

      let distance = 0
      for (let i = 0; i < keypoints.value.length - 1; i++) {
        const from = L.latLng(keypoints.value[i].latitude, keypoints.value[i].longitude)
        const to = L.latLng(keypoints.value[i + 1].latitude, keypoints.value[i + 1].longitude)
        distance += from.distanceTo(to)
      }
      totalDistance.value = distance / 1000
    }

    const focusOnPoint = (kp) => {
      map.setView([kp.latitude, kp.longitude], 16, {
        animate: true,
        duration: 0.5
      })

      const marker = markers.find(m => {
        const latlng = m.getLatLng()
        return latlng.lat === kp.latitude && latlng.lng === kp.longitude
      })

      if (marker) {
        marker.openPopup()
      }
    }

    const fetchTourData = async () => {
      loading.value = true
      error.value = ''

      try {
        const [tourData, keypointsData] = await Promise.all([
          api.getTourById(tourId),
          api.getKeyPoints(tourId)
        ])

        tour.value = {
          ...tourData,
          durations: tourData.durations || { walking: 0, biking: 0, driving: 0 }
        }
        keypoints.value = keypointsData

        if (keypoints.value.length === 0) {
          error.value = 'No key points available for this tour'
          loading.value = false
          return
        }

        // Set loading to false first to render the map container
        loading.value = false

        // Wait for DOM update, then initialize map
        await new Promise(resolve => setTimeout(resolve, 100))
        
        if (!mapContainer.value) {
          error.value = 'Map container not found'
          return
        }

        initMap()
        renderMarkers()
        
        if (keypoints.value.length > 1) {
          calculateAndDrawRoute()
        }
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to load tour data'
        loading.value = false
      }
    }

    const goBack = () => {
      router.push(`/tours/${tourId}`)
    }

    const formatDuration = (minutes) => {
      if (!minutes) return '0 min'
      if (minutes < 60) return `${minutes} min`
      const hours = Math.floor(minutes / 60)
      const mins = minutes % 60
      return mins > 0 ? `${hours}h ${mins}min` : `${hours}h`
    }

    onMounted(async () => {
      await fetchTourData()
    })

    onUnmounted(() => {
      if (map) {
        map.remove()
      }
    })

    return {
      mapContainer,
      tour,
      keypoints,
      loading,
      error,
      totalDistance,
      isCalculatingRoute,
      focusOnPoint,
      goBack,
      formatDuration
    }
  }
}
</script>

<style scoped>
.tour-map-view {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  background: white;
  border-bottom: 2px solid #e0e0e0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.header h1 {
  color: #2c3e50;
  margin: 0 0 0.5rem 0;
  font-size: 1.75rem;
}

.subtitle {
  color: #666;
  margin: 0;
  font-size: 0.95rem;
}

.btn-secondary {
  background: #f0f0f0;
  color: #333;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background: #e0e0e0;
  transform: translateY(-1px);
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  margin: 1rem 2rem;
  border-radius: 8px;
  border: 1px solid #fcc;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #666;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f0f0f0;
  border-top: 4px solid #42b983;
  border-radius: 50%;
  margin-bottom: 1rem;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.map-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}

#map {
  width: 100%;
  height: 100%;
}

.map-overlay {
  position: absolute;
  top: 1rem;
  right: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-width: 350px;
  max-height: calc(100% - 2rem);
  pointer-events: none;
  z-index: 1000;
}

.map-overlay > * {
  pointer-events: auto;
}

.map-info-card,
.keypoints-list-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.info-header,
.list-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 1rem;
}

.info-header h3,
.list-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.info-content {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item.travel-times {
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
}

.travel-times-list {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 0.25rem;
}

.travel-time-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  background: #f8f9fa;
  border-radius: 6px;
  font-size: 0.9rem;
}

.travel-icon {
  font-size: 1.2rem;
  width: 24px;
  text-align: center;
}

.info-label {
  color: #666;
  font-weight: 600;
  font-size: 0.9rem;
}

.info-value {
  color: #2c3e50;
  font-weight: 700;
  font-size: 1rem;
}

.difficulty-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  text-transform: capitalize;
}

.difficulty-badge.Easy {
  background: #d4edda;
  color: #155724;
}

.difficulty-badge.Medium {
  background: #fff3cd;
  color: #856404;
}

.difficulty-badge.Hard {
  background: #f8d7da;
  color: #721c24;
}

.calculating {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #ff9800;
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background: #ff9800;
  border-radius: 50%;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.2); }
}

.keypoints-list-card {
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.keypoints-list {
  overflow-y: auto;
  padding: 0.5rem;
}

.keypoint-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 8px;
  margin-bottom: 0.5rem;
  background: #f8f9fa;
  cursor: pointer;
  transition: all 0.2s;
}

.keypoint-item:hover {
  background: #e9ecef;
  transform: translateX(4px);
}

.keypoint-number {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.keypoint-info {
  flex: 1;
  min-width: 0;
}

.keypoint-info h4 {
  margin: 0 0 0.25rem 0;
  color: #2c3e50;
  font-size: 0.95rem;
}

.keypoint-desc {
  margin: 0;
  color: #666;
  font-size: 0.85rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Custom marker styles */
:deep(.custom-marker) {
  background: transparent;
  border: none;
}

:deep(.marker-pin) {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.3);
  border: 3px solid white;
}

:deep(.marker-number) {
  color: white;
  font-weight: 700;
  font-size: 1rem;
  transform: rotate(45deg);
}

:deep(.marker-popup) {
  min-width: 150px;
}

:deep(.marker-popup h4) {
  margin: 0 0 0.5rem 0;
  color: #2c3e50;
}

:deep(.marker-popup p) {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

/* Responsive */
@media (max-width: 768px) {
  .map-overlay {
    max-width: calc(100% - 2rem);
    left: 1rem;
    right: 1rem;
  }
  
  .keypoints-list-card {
    max-height: 200px;
  }
}

/* Ensure Leaflet controls don't overlap with overlay */
:deep(.leaflet-top),
:deep(.leaflet-bottom) {
  z-index: 999;
}

:deep(.leaflet-control) {
  margin-right: 370px !important;
}

@media (max-width: 768px) {
  :deep(.leaflet-control) {
    margin-right: 10px !important;
  }
}
</style>
