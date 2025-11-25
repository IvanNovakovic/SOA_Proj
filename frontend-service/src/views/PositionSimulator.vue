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

        <div class="panel-card">
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

export default {
  name: 'PositionSimulator',
  setup() {
    const router = useRouter()
    const mapContainer = ref(null)
    const currentPosition = ref(null)

    let map = null
    let marker = null

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

    const setPosition = (lat, lng) => {
      const position = {
        lat,
        lng,
        timestamp: new Date().toISOString()
      }

      currentPosition.value = position
      localStorage.setItem('touristPosition', JSON.stringify(position))
      
      addMarker(lat, lng)
      map.setView([lat, lng], map.getZoom())
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

    return {
      mapContainer,
      currentPosition,
      formatTime,
      goBack
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
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

#map {
  width: 100%;
  height: 100%;
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
