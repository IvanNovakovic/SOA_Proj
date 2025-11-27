<template>
  <div class="position-simulator">
    <div class="header">
      <div>
        <h1>📍 Position Simulator</h1>
        <p class="subtitle">Click on the map to set your current location</p>
      </div>
      <button @click="goBack" class="btn-secondary">
        ← Back
      </button>
    </div>

    <div class="content">
      <div class="map-section">
        <div id="map" ref="mapContainer"></div>
        
        <!-- Nearby Keypoint Notification -->
        <transition name="slide-up">
          <div v-if="nearbyKeypoint" class="nearby-notification">
            <div class="notification-header">
              <div class="notification-icon">🎯</div>
              <div class="notification-title">
                <h4>Key Point Reached!</h4>
                <span class="keypoint-number">Point #{{ nearbyKeypoint.index }}</span>
              </div>
              <button @click="closeNearbyNotification" class="close-btn">×</button>
            </div>
            <div class="notification-body">
              <h3>{{ nearbyKeypoint.name }}</h3>
              <p v-if="nearbyKeypoint.description">{{ nearbyKeypoint.description }}</p>
              <div class="notification-footer">
                <span class="distance-badge">📍 {{ nearbyKeypoint.distance }}m away</span>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <div class="info-panel">
        <div class="panel-card">
          <h3>Current Position</h3>
          <div v-if="currentPosition" class="position-info">
            <div class="coord-row">
              <span class="label">Latitude:</span>
              <span class="value">{{ currentPosition.lat.toFixed(6) }}</span>
            </div>
            <div class="coord-row">
              <span class="label">Longitude:</span>
              <span class="value">{{ currentPosition.lng.toFixed(6) }}</span>
            </div>
            <div class="timestamp">
              Set at: {{ formatTime(currentPosition.timestamp) }}
            </div>
          </div>
          <div v-else class="no-position">
            <p>No position set yet</p>
            <p class="hint">Click anywhere on the map to set your position</p>
          </div>
        </div>

        <div class="panel-card" v-if="!hideInstructions">
          <h3>Instructions</h3>
          <ul class="instructions">
            <li>Click on the map to simulate your current position</li>
            <li>The marker will show your selected location</li>
            <li>Your position is saved automatically</li>
            <li>This simulates GPS tracking for tour execution</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { api } from '../services/api'
import routingService from '../services/routingService'

export default {
  name: 'PositionSimulator',
  props: {
    hideInstructions: {
      type: Boolean,
      default: false
    },
    tourId: {
      type: String,
      default: null
    },
    execution: {
      type: Object,
      default: null
    }
  },
  setup(props, { emit }) {
    const router = useRouter()
    const mapContainer = ref(null)
    const currentPosition = ref(null)
    const keyPoints = ref([])
    const nearbyKeypoint = ref(null)
    const visitedKeypoints = ref(new Set())

    let map = null
    let marker = null
    let keypointMarkers = []
    let routeLine = null

    const initMap = () => {
      // Try to get stored position or default to Novi Sad
      const stored = localStorage.getItem('touristPosition')
      let center = [45.2671, 19.8335]
      
      if (stored) {
        try {
          const pos = JSON.parse(stored)
          currentPosition.value = pos
          center = [pos.lat, pos.lng]
        } catch (e) {
          console.error('Failed to parse stored position:', e)
        }
      }

      map = L.map(mapContainer.value).setView(center, 13)

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors',
        maxZoom: 19
      }).addTo(map)

      // Add marker if position exists
      if (currentPosition.value) {
        addMarker(currentPosition.value.lat, currentPosition.value.lng)
      }

      // Handle map clicks
      map.on('click', (e) => {
        setPosition(e.latlng.lat, e.latlng.lng)
      })

      // Fetch and display key points if tourId is provided
      if (props.tourId) {
        fetchKeyPoints()
      }
    }

    const fetchKeyPoints = async () => {
      try {
        const data = await api.getKeyPoints(props.tourId)
        keyPoints.value = data || []
        displayKeyPoints()
      } catch (err) {
        console.error('Failed to fetch key points:', err)
      }
    }

    const displayKeyPoints = () => {
      // Clear existing keypoint markers
      keypointMarkers.forEach(m => map.removeLayer(m))
      keypointMarkers = []

      if (keyPoints.value.length === 0) return

      // Create markers for each keypoint
      keyPoints.value.forEach((kp, index) => {
        const keypointIcon = L.divIcon({
          className: 'keypoint-marker',
          html: `
            <div class="keypoint-number-marker">
              <span>${index + 1}</span>
            </div>
          `,
          iconSize: [40, 40],
          iconAnchor: [20, 20]
        })

        const keypointMarker = L.marker([kp.latitude, kp.longitude], { icon: keypointIcon })
          .bindPopup(`
            <div class="keypoint-popup">
              <h4>${kp.name}</h4>
              ${kp.description ? `<p>${kp.description}</p>` : ''}
              <p class="coords"><strong>Coordinates:</strong><br>${kp.latitude.toFixed(4)}, ${kp.longitude.toFixed(4)}</p>
            </div>
          `)
          .addTo(map)

        keypointMarkers.push(keypointMarker)
      })

      // Fit map bounds to show all keypoints
      if (keyPoints.value.length > 0) {
        const bounds = L.latLngBounds(keyPoints.value.map(kp => [kp.latitude, kp.longitude]))
        map.fitBounds(bounds, { padding: [50, 50] })
      }
    }

    const getNextUncompletedKeyPoint = (executionData = null) => {
      if (!keyPoints.value || keyPoints.value.length === 0) return null
      
      const execToUse = executionData || props.execution
      if (!execToUse?.completedPoints) return keyPoints.value[0]

      const completedIds = new Set(execToUse.completedPoints.map(cp => cp.keyPointId))
      
      // Find first uncompleted key point
      for (const kp of keyPoints.value) {
        if (!completedIds.has(kp.id)) {
          return kp
        }
      }
      
      return null // All completed
    }

    const updateRouteToNextPoint = async (executionData = null) => {
      // Remove existing route line
      if (routeLine) {
        try {
          map.removeLayer(routeLine)
        } catch (e) {
          console.warn('Failed to remove route line:', e)
        }
        routeLine = null
      }

      if (!currentPosition.value) return

      const nextPoint = getNextUncompletedKeyPoint(executionData)
      if (!nextPoint) {
        // No more points, don't draw a route
        return
      }

      try {
        // Calculate route using OpenRouteService
        const coordinates = [
          [currentPosition.value.lng, currentPosition.value.lat], // ORS uses [lng, lat]
          [nextPoint.longitude, nextPoint.latitude]
        ]

        const routeData = await routingService.calculateRoute(coordinates, 'foot-walking')
        
        // Remove any existing route before adding new one
        if (routeLine) {
          try {
            map.removeLayer(routeLine)
          } catch (e) {
            console.warn('Failed to remove previous route:', e)
          }
        }
        
        // Convert ORS coordinates [lng, lat] to Leaflet format [lat, lng]
        const latlngs = routeData.geometry.map(coord => [coord[1], coord[0]])

        routeLine = L.polyline(latlngs, {
          color: '#667eea',
          weight: 4,
          opacity: 0.7,
          lineJoin: 'round'
        }).addTo(map)

      } catch (error) {
        console.error('Failed to calculate route, using straight line:', error)
        
        // Remove any existing route before adding fallback
        if (routeLine) {
          try {
            map.removeLayer(routeLine)
          } catch (e) {
            console.warn('Failed to remove previous route:', e)
          }
        }
        
        // Fallback to straight line if routing fails
        const latlngs = [
          [currentPosition.value.lat, currentPosition.value.lng],
          [nextPoint.latitude, nextPoint.longitude]
        ]

        routeLine = L.polyline(latlngs, {
          color: '#667eea',
          weight: 4,
          opacity: 0.7,
          dashArray: '10, 10',
          lineJoin: 'round'
        }).addTo(map)
      }
    }

    const addMarker = (lat, lng) => {
      if (marker) {
        map.removeLayer(marker)
      }

      const customIcon = L.divIcon({
        className: 'current-position-marker',
        html: `
          <div class="marker-pulse">
            <div class="marker-inner"></div>
          </div>
        `,
        iconSize: [30, 30],
        iconAnchor: [15, 15]
      })

      marker = L.marker([lat, lng], { icon: customIcon })
        .bindPopup(`
          <div class="position-popup">
            <h4>Your Position</h4>
            <p><strong>Lat:</strong> ${lat.toFixed(6)}</p>
            <p><strong>Lng:</strong> ${lng.toFixed(6)}</p>
          </div>
        `)
        .addTo(map)
    }

    const setPosition = async (lat, lng) => {
      const position = {
        lat,
        lng,
        timestamp: new Date().toISOString()
      }

      currentPosition.value = position
      localStorage.setItem('touristPosition', JSON.stringify(position))
      
      addMarker(lat, lng)
      map.setView([lat, lng], map.getZoom())

      // Update route to next key point
      updateRouteToNextPoint()

      // Send location to backend if execution is active
      if (props.execution && props.execution.id) {
        await sendLocationToBackend(lat, lng)
      }
    }

    const sendLocationToBackend = async (lat, lng) => {
      try {
        // Get completed points before sending location
        const beforeCompleted = new Set(props.execution.completedPoints?.map(cp => cp.keyPointId) || [])
        
        // Send location to backend
        await api.addExecutionLocation(props.execution.id, { latitude: lat, longitude: lng })
        
        // Get updated execution to check for newly completed points
        const updatedExec = await api.getActiveExecution(props.tourId)
        
        // Check for newly completed points
        const afterCompleted = new Set(updatedExec.completedPoints?.map(cp => cp.keyPointId) || [])
        
        for (const kpId of afterCompleted) {
          if (!beforeCompleted.has(kpId) && !visitedKeypoints.value.has(kpId)) {
            // Find the key point that was just completed
            const kpIndex = keyPoints.value.findIndex(kp => kp.id === kpId)
            if (kpIndex !== -1) {
              const kp = keyPoints.value[kpIndex]
              visitedKeypoints.value.add(kpId)
              
              // Show notification
              nearbyKeypoint.value = {
                ...kp,
                distance: 0,
                index: kpIndex + 1
              }

              // Auto-open the marker popup
              if (keypointMarkers[kpIndex]) {
                keypointMarkers[kpIndex].openPopup()
              }

              // Clear after 10 seconds
              setTimeout(() => {
                if (nearbyKeypoint.value?.id === kpId) {
                  nearbyKeypoint.value = null
                }
              }, 10000)

              break // Only show one at a time
            }
          }
        }
        
        // Emit update to parent
        emit('update-location', updatedExec)
        
        // Update route to next uncompleted point with fresh execution data
        await updateRouteToNextPoint(updatedExec)
      } catch (err) {
        console.error('Failed to send location:', err)
      }
    }

    const closeNearbyNotification = () => {
      nearbyKeypoint.value = null
    }

    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString()
    }

    const goBack = () => {
      router.back()
    }

    onMounted(() => {
      initMap()
    })

    onUnmounted(() => {
      if (map) {
        map.remove()
      }
    })

    // Watch for changes in execution to update route
    watch(() => props.execution, (newExec) => {
      if (newExec && currentPosition.value && map) {
        updateRouteToNextPoint()
      }
    }, { deep: true })

    return {
      mapContainer,
      currentPosition,
      nearbyKeypoint,
      formatTime,
      goBack,
      closeNearbyNotification,
      hideInstructions: props.hideInstructions
    }
  }
}
</script>

<style scoped>
.position-simulator {
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

.content {
  flex: 1;
  display: flex;
  gap: 1.5rem;
  padding: 1.5rem;
  overflow: hidden;
}

.map-section {
  flex: 1;
  min-height: 500px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  position: relative;
}

#map {
  width: 100%;
  height: 100%;
  min-height: 500px;
}

/* Nearby Keypoint Notification */
.nearby-notification {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  min-width: 400px;
  max-width: 500px;
  overflow: hidden;
  animation: bounce-in 0.5s ease-out;
}

.notification-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.notification-icon {
  font-size: 2rem;
  animation: pulse-icon 2s ease-in-out infinite;
}

@keyframes pulse-icon {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.notification-title {
  flex: 1;
}

.notification-title h4 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
}

.keypoint-number {
  font-size: 0.85rem;
  opacity: 0.9;
}

.close-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  line-height: 1;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.notification-body {
  padding: 1.5rem;
}

.notification-body h3 {
  margin: 0 0 0.75rem 0;
  color: #2c3e50;
  font-size: 1.25rem;
}

.notification-body p {
  margin: 0 0 1rem 0;
  color: #666;
  line-height: 1.5;
}

.notification-footer {
  display: flex;
  justify-content: flex-end;
}

.distance-badge {
  background: #e8f5e9;
  color: #2e7d32;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
}

@keyframes bounce-in {
  0% {
    transform: translateX(-50%) translateY(100px);
    opacity: 0;
  }
  50% {
    transform: translateX(-50%) translateY(-10px);
  }
  100% {
    transform: translateX(-50%) translateY(0);
    opacity: 1;
  }
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from {
  transform: translateX(-50%) translateY(100px);
  opacity: 0;
}

.slide-up-leave-to {
  transform: translateX(-50%) translateY(100px);
  opacity: 0;
}

.info-panel {
  width: 350px;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.panel-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.panel-card h3 {
  margin: 0 0 1rem 0;
  color: #2c3e50;
  font-size: 1.2rem;
  padding-bottom: 0.75rem;
  border-bottom: 2px solid #f0f0f0;
}

.position-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.coord-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem;
  background: #f8f9fa;
  border-radius: 6px;
}

.label {
  color: #666;
  font-weight: 600;
}

.value {
  color: #2c3e50;
  font-family: 'Courier New', monospace;
  font-weight: 700;
}

.timestamp {
  margin-top: 0.5rem;
  padding: 0.5rem;
  background: #e8f5e9;
  border-radius: 6px;
  color: #2e7d32;
  font-size: 0.9rem;
  text-align: center;
}

.no-position {
  text-align: center;
  padding: 2rem 1rem;
  color: #999;
}

.no-position p {
  margin: 0.5rem 0;
}

.hint {
  font-size: 0.9rem;
  font-style: italic;
}

.instructions {
  list-style: none;
  padding: 0;
  margin: 0;
}

.instructions li {
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  background: #f8f9fa;
  border-radius: 6px;
  border-left: 3px solid #42b983;
  color: #555;
  font-size: 0.95rem;
}

:deep(.current-position-marker) {
  background: transparent;
  border: none;
}

:deep(.route-arrow) {
  background: transparent;
  border: none;
}

:deep(.marker-pulse) {
  width: 30px;
  height: 30px;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.marker-pulse::before) {
  content: '';
  position: absolute;
  width: 100%;
  height: 100%;
  background: #42b983;
  border-radius: 50%;
  opacity: 0.3;
  animation: pulse 2s ease-out infinite;
}

:deep(.marker-inner) {
  width: 16px;
  height: 16px;
  background: #42b983;
  border-radius: 50%;
  border: 3px solid white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  z-index: 1;
}

@keyframes pulse {
  0% {
    transform: scale(0.5);
    opacity: 0.5;
  }
  100% {
    transform: scale(2);
    opacity: 0;
  }
}

:deep(.position-popup h4) {
  margin: 0 0 0.5rem 0;
  color: #2c3e50;
}

:deep(.position-popup p) {
  margin: 0.25rem 0;
  color: #666;
  font-size: 0.9rem;
}

:deep(.keypoint-marker) {
  background: transparent;
  border: none;
}

:deep(.keypoint-number-marker) {
  width: 40px;
  height: 40px;
  background: #ff6b6b;
  border: 3px solid white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  font-weight: bold;
  color: white;
  font-size: 16px;
}

:deep(.keypoint-popup) {
  min-width: 200px;
}

:deep(.keypoint-popup h4) {
  margin: 0 0 0.5rem 0;
  color: #2c3e50;
  font-size: 1.1rem;
}

:deep(.keypoint-popup p) {
  margin: 0.5rem 0;
  color: #666;
  font-size: 0.9rem;
}

:deep(.keypoint-popup .coords) {
  font-family: 'Courier New', monospace;
  font-size: 0.85rem;
  color: #888;
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid #e0e0e0;
}

@media (max-width: 968px) {
  .content {
    flex-direction: column;
  }

  .map-section {
    height: 400px;
  }

  .info-panel {
    width: 100%;
  }
}
</style>
