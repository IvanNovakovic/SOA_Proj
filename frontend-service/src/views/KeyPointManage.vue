<template>
  <div class="keypoint-manage">
    <div class="header">
      <div>
        <h1>Manage Tour Key Points</h1>
        <p class="subtitle">Add, reorder, and manage points of interest for your tour</p>
      </div>
      <button @click="goBack" class="btn-secondary">
        ← Back to Tour
      </button>
    </div>

    <div v-if="error" class="error-message">
      <span class="error-icon">⚠️</span>
      {{ error }}
    </div>
    <div v-if="success" class="success-message">{{ success }}</div>

    <div class="content-layout">
      <!-- Map Section -->
      <div class="map-section">
        <div class="section-header">
          <h2>📍 Map View</h2>
          <p>Click on the map to add a new point or view existing ones</p>
        </div>
        <div id="map" ref="mapContainer"></div>
        <div class="map-info">
          <div class="map-legend">
            <span class="legend-item"><span class="marker-dot new"></span> New Point</span>
            <span class="legend-item"><span class="marker-dot existing"></span> Existing Points</span>
            <span class="legend-item"><span class="route-line"></span> Tour Route</span>
          </div>
          <div v-if="keypoints.length > 1" class="distance-info">
            <span class="distance-icon">📏</span>
            <strong>Total Distance:</strong> {{ totalDistance.toFixed(2) }} km
          </div>
        </div>
      </div>

      <!-- KeyPoints List & Form Section -->
      <div class="keypoints-section">
        <!-- Add New KeyPoint Form -->
        <div v-if="showForm" class="keypoint-form-card">
          <div class="form-header">
            <h3>{{ editingKeyPoint ? '✏️ Edit' : '➕ Add New' }} Key Point</h3>
            <button @click="cancelForm" class="btn-close">✕</button>
          </div>
          
          <form @submit.prevent="handleSubmit">
            <div class="form-group">
              <label for="name">Name *</label>
              <input 
                id="name"
                v-model="form.name" 
                type="text" 
                placeholder="e.g., City Museum"
                required
              />
            </div>

            <div class="form-group">
              <label for="description">Description</label>
              <textarea 
                id="description"
                v-model="form.description" 
                placeholder="Describe this location"
                rows="3"
              ></textarea>
            </div>

            <div class="form-group">
              <label for="imageUrl">Image URL</label>
              <input 
                id="imageUrl"
                v-model="form.imageUrl" 
                type="url" 
                placeholder="https://example.com/image.jpg"
              />
            </div>

            <div class="form-group">
              <label>Location</label>
              <div class="location-display">
                <div v-if="form.latitude && form.longitude">
                  <p><strong>Lat:</strong> {{ form.latitude.toFixed(6) }}</p>
                  <p><strong>Lng:</strong> {{ form.longitude.toFixed(6) }}</p>
                </div>
                <p v-else class="no-location">Click on map to select location</p>
              </div>
            </div>

            <div class="form-actions">
              <button type="submit" :disabled="loading || !form.latitude" class="btn-primary">
                {{ loading ? 'Saving...' : editingKeyPoint ? 'Update' : 'Add' }} Point
              </button>
              <button type="button" @click="cancelForm" class="btn-secondary">Cancel</button>
            </div>
          </form>
        </div>

        <!-- KeyPoints List -->
        <div class="keypoints-list-card">
          <div class="list-header">
            <h3>🗺️ Tour Points ({{ keypoints.length }})</h3>
            <button v-if="!showForm" @click="startNewKeyPoint" class="btn-add">
              ➕ Add Point
            </button>
          </div>

          <div v-if="loading && keypoints.length === 0" class="loading">
            <div class="spinner"></div>
            <p>Loading points...</p>
          </div>

          <div v-else-if="keypoints.length === 0" class="empty-state">
            <div class="empty-icon">📍</div>
            <p>No key points yet</p>
            <p class="hint">Click "Add Point" to create your first point of interest</p>
          </div>

          <draggable 
            v-else
            v-model="keypoints" 
            @end="handleReorder"
            item-key="id"
            class="keypoints-list"
            handle=".drag-handle"
          >
            <template #item="{ element: kp, index }">
              <div class="keypoint-item">
                <div class="drag-handle" title="Drag to reorder">
                  <span>☰</span>
                </div>
                <div class="keypoint-number">{{ index + 1 }}</div>
                <div class="keypoint-image">
                  <img 
                    v-if="kp.imageUrl" 
                    :src="kp.imageUrl" 
                    :alt="kp.name"
                    @error="handleImageError"
                  />
                  <div v-else class="image-placeholder">
                    <span class="placeholder-icon">🏞️</span>
                  </div>
                </div>
                <div class="keypoint-content">
                  <h4>{{ kp.name }}</h4>
                  <p v-if="kp.description" class="keypoint-desc">{{ kp.description }}</p>
                  <div class="keypoint-coords">
                    📍 {{ kp.latitude.toFixed(4) }}, {{ kp.longitude.toFixed(4) }}
                  </div>
                </div>
                <div class="keypoint-actions">
                  <button @click="editKeyPoint(kp)" class="btn-icon" title="Edit">
                    ✏️
                  </button>
                  <button @click="deleteKeyPointConfirm(kp)" class="btn-icon danger" title="Delete">
                    🗑️
                  </button>
                </div>
              </div>
            </template>
          </draggable>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../services/api'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import draggable from 'vuedraggable'

export default {
  name: 'KeyPointManage',
  components: {
    draggable
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const tourId = route.params.tourId

    const mapContainer = ref(null)
    let map = null
    let markers = []
    let newMarker = null
    let routePolyline = null

    const keypoints = ref([])
    const showForm = ref(false)
    const editingKeyPoint = ref(null)
    const totalDistance = ref(0)

    const form = ref({
      name: '',
      description: '',
      imageUrl: '',
      latitude: null,
      longitude: null
    })

    const loading = ref(false)
    const error = ref('')
    const success = ref('')

    const initMap = () => {
      map = L.map('map').setView([44.8176, 20.4633], 13)

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      }).addTo(map)

      map.on('click', (e) => {
        const { lat, lng } = e.latlng
        
        if (newMarker) {
          map.removeLayer(newMarker)
        }

        const newIcon = L.divIcon({
          className: 'custom-marker new-marker',
          html: '<div class="marker-pin new">📍</div>',
          iconSize: [30, 42],
          iconAnchor: [15, 42]
        })

        newMarker = L.marker([lat, lng], { icon: newIcon }).addTo(map)
        
        form.value.latitude = lat
        form.value.longitude = lng

        if (!showForm.value) {
          showForm.value = true
        }
      })
    }

    const renderMarkers = () => {
      markers.forEach(m => map.removeLayer(m))
      markers = []

      // Remove existing route
      if (routePolyline) {
        map.removeLayer(routePolyline)
        routePolyline = null
      }

      keypoints.value.forEach((kp, index) => {
        const icon = L.divIcon({
          className: 'custom-marker existing-marker',
          html: `<div class="marker-pin existing"><span class="marker-number">${index + 1}</span></div>`,
          iconSize: [30, 42],
          iconAnchor: [15, 42]
        })

        const marker = L.marker([kp.latitude, kp.longitude], { icon })
          .bindPopup(`<strong>${kp.name}</strong><br/>${kp.description || ''}`)
          .addTo(map)
        
        markers.push(marker)
      })

      // Draw route line between points
      if (keypoints.value.length > 1) {
        const routeCoordinates = keypoints.value.map(kp => [kp.latitude, kp.longitude])
        routePolyline = L.polyline(routeCoordinates, {
          color: '#42b983',
          weight: 4,
          opacity: 0.7,
          smoothFactor: 1
        }).addTo(map)

        // Calculate total distance
        calculateTotalDistance()
      } else {
        totalDistance.value = 0
      }

      if (keypoints.value.length > 0) {
        const bounds = L.latLngBounds(keypoints.value.map(kp => [kp.latitude, kp.longitude]))
        map.fitBounds(bounds, { padding: [50, 50] })
      }
    }

    const calculateTotalDistance = () => {
      let distance = 0
      for (let i = 0; i < keypoints.value.length - 1; i++) {
        const from = L.latLng(keypoints.value[i].latitude, keypoints.value[i].longitude)
        const to = L.latLng(keypoints.value[i + 1].latitude, keypoints.value[i + 1].longitude)
        distance += from.distanceTo(to) // distance in meters
      }
      totalDistance.value = distance / 1000 // convert to kilometers
    }

    const updateTourDistance = async () => {
      try {
        await api.updateTour(tourId, { distance: totalDistance.value })
      } catch (err) {
        console.error('Failed to update tour distance:', err)
      }
    }

    const fetchKeyPoints = async () => {
      try {
        const data = await api.getKeyPoints(tourId)
        keypoints.value = data || []
        await nextTick()
        if (map) {
          renderMarkers()
        }
      } catch (err) {
        error.value = 'Failed to load key points'
        console.error(err)
      }
    }

    const startNewKeyPoint = () => {
      editingKeyPoint.value = null
      form.value = {
        name: '',
        description: '',
        imageUrl: '',
        latitude: null,
        longitude: null
      }
      showForm.value = true
      
      if (newMarker) {
        map.removeLayer(newMarker)
        newMarker = null
      }
    }

    const editKeyPoint = (kp) => {
      editingKeyPoint.value = kp
      form.value = {
        name: kp.name,
        description: kp.description || '',
        imageUrl: kp.imageUrl || '',
        latitude: kp.latitude,
        longitude: kp.longitude
      }
      showForm.value = true

      if (newMarker) {
        map.removeLayer(newMarker)
      }

      const newIcon = L.divIcon({
        className: 'custom-marker new-marker',
        html: '<div class="marker-pin new">📍</div>',
        iconSize: [30, 42],
        iconAnchor: [15, 42]
      })

      newMarker = L.marker([kp.latitude, kp.longitude], { icon: newIcon }).addTo(map)
      map.setView([kp.latitude, kp.longitude], 15)
    }

    const handleSubmit = async () => {
      if (!form.value.latitude || !form.value.longitude) {
        error.value = 'Please select a location on the map'
        return
      }

      loading.value = true
      error.value = ''
      success.value = ''

      try {
        const data = {
          name: form.value.name,
          description: form.value.description,
          imageUrl: form.value.imageUrl,
          latitude: form.value.latitude,
          longitude: form.value.longitude
        }

        if (editingKeyPoint.value) {
          await api.updateKeyPoint(editingKeyPoint.value.id, data)
          success.value = 'Key point updated successfully!'
        } else {
          await api.createKeyPoint(tourId, data)
          success.value = 'Key point added successfully!'
        }

        showForm.value = false
        editingKeyPoint.value = null
        
        if (newMarker) {
          map.removeLayer(newMarker)
          newMarker = null
        }

        await fetchKeyPoints()
        await updateTourDistance()

        setTimeout(() => {
          success.value = ''
        }, 3000)
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to save key point'
      } finally {
        loading.value = false
      }
    }

    const cancelForm = () => {
      showForm.value = false
      editingKeyPoint.value = null
      form.value = {
        name: '',
        description: '',
        imageUrl: '',
        latitude: null,
        longitude: null
      }
      
      if (newMarker) {
        map.removeLayer(newMarker)
        newMarker = null
      }
    }

    const deleteKeyPointConfirm = async (kp) => {
      if (!confirm(`Delete "${kp.name}"? This cannot be undone.`)) {
        return
      }

      loading.value = true
      error.value = ''

      try {
        await api.deleteKeyPoint(kp.id)
        success.value = 'Key point deleted successfully!'
        await fetchKeyPoints()
        await updateTourDistance()
        
        setTimeout(() => {
          success.value = ''
        }, 3000)
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to delete key point'
      } finally {
        loading.value = false
      }
    }

    const handleReorder = async () => {
      const orderedIds = keypoints.value.map(kp => kp.id)
      
      try {
        await api.reorderKeyPoints(tourId, orderedIds)
        renderMarkers()
        await updateTourDistance()
        success.value = 'Order updated!'
        setTimeout(() => {
          success.value = ''
        }, 2000)
      } catch (err) {
        error.value = 'Failed to update order'
        console.error(err)
      }
    }

    const goBack = () => {
      router.push(`/tours/${tourId}`)
    }

    const handleImageError = (event) => {
      event.target.style.display = 'none'
      event.target.nextElementSibling?.classList.add('show')
    }

    onMounted(async () => {
      initMap()
      await fetchKeyPoints()
    })

    onUnmounted(() => {
      if (map) {
        map.remove()
      }
    })

    return {
      mapContainer,
      keypoints,
      showForm,
      editingKeyPoint,
      form,
      loading,
      error,
      success,
      totalDistance,
      handleSubmit,
      cancelForm,
      startNewKeyPoint,
      editKeyPoint,
      deleteKeyPointConfirm,
      handleReorder,
      handleImageError,
      goBack
    }
  }
}
</script>

<style scoped>
.keypoint-manage {
  max-width: 1600px;
  margin: 0 auto;
  padding: 2rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
}

h1 {
  color: #2c3e50;
  margin-bottom: 0.5rem;
  font-size: 2rem;
}

.subtitle {
  color: #666;
  font-size: 1.1rem;
}

.error-message,
.success-message {
  padding: 1rem 1.5rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  animation: slideIn 0.3s ease-out;
}

.error-message {
  background: linear-gradient(135deg, #fee 0%, #fdd 100%);
  color: #c33;
  border: 2px solid #fcc;
}

.success-message {
  background: linear-gradient(135deg, #efe 0%, #dfd 100%);
  color: #3c3;
  border: 2px solid #cfc;
}

.content-layout {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 2rem;
  align-items: start;
}

.map-section {
  background: white;
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  position: sticky;
  top: 2rem;
}

.section-header {
  margin-bottom: 1rem;
}

.section-header h2 {
  color: #2c3e50;
  font-size: 1.4rem;
  margin-bottom: 0.5rem;
}

.section-header p {
  color: #666;
  font-size: 0.9rem;
}

#map {
  width: 100%;
  height: 600px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 1rem;
}

.map-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.map-legend {
  display: flex;
  gap: 1.5rem;
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 8px;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  color: #555;
}

.marker-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  display: inline-block;
}

.marker-dot.new {
  background: #42b983;
}

.marker-dot.existing {
  background: #667eea;
}

.route-line {
  width: 30px;
  height: 4px;
  background: #42b983;
  display: inline-block;
  border-radius: 2px;
}

.distance-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 100%);
  border-radius: 8px;
  font-size: 1rem;
  color: #2e7d32;
  border: 2px solid #a5d6a7;
  font-weight: 500;
}

.distance-icon {
  font-size: 1.3rem;
}

.keypoints-section {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.keypoint-form-card,
.keypoints-list-card {
  background: white;
  border-radius: 16px;
  padding: 1.5rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.form-header,
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid #f0f0f0;
}

.form-header h3,
.list-header h3 {
  color: #2c3e50;
  font-size: 1.3rem;
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #999;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.btn-close:hover {
  background: #f0f0f0;
  color: #666;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
  font-size: 0.95rem;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
  font-family: inherit;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #42b983;
}

.location-display {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 8px;
  border: 2px solid #e0e0e0;
}

.location-display p {
  margin: 0.25rem 0;
  color: #333;
}

.no-location {
  color: #999;
  font-style: italic;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.btn-primary,
.btn-secondary,
.btn-add {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  flex: 1;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(66, 185, 131, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f5f5f5;
  color: #666;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

.btn-add {
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  font-size: 0.9rem;
  padding: 0.6rem 1.25rem;
}

.btn-add:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(66, 185, 131, 0.3);
}

.loading {
  text-align: center;
  padding: 3rem 2rem;
  color: #666;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f0f0f0;
  border-top: 4px solid #42b983;
  border-radius: 50%;
  margin: 0 auto 1rem;
  animation: spin 1s linear infinite;
}

.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  color: #999;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.empty-state .hint {
  font-size: 0.9rem;
  margin-top: 0.5rem;
}

.keypoints-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.keypoint-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #f8f9fa;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  transition: all 0.2s;
  cursor: move;
}

.keypoint-item:hover {
  border-color: #42b983;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.drag-handle {
  color: #999;
  cursor: grab;
  font-size: 1.2rem;
  padding: 0.5rem;
  display: flex;
  align-items: center;
}

.drag-handle:active {
  cursor: grabbing;
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

.keypoint-image {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  background: #f0f0f0;
  position: relative;
}

.keypoint-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #e0e0e0 0%, #f5f5f5 100%);
}

.placeholder-icon {
  font-size: 2.5rem;
  opacity: 0.5;
}

.keypoint-content {
  flex: 1;
  min-width: 0;
}

.keypoint-content h4 {
  color: #2c3e50;
  margin: 0 0 0.25rem 0;
  font-size: 1.05rem;
}

.keypoint-desc {
  color: #666;
  font-size: 0.9rem;
  margin: 0.25rem 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.keypoint-coords {
  color: #999;
  font-size: 0.85rem;
  margin-top: 0.25rem;
}

.keypoint-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 8px;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: rgba(0, 0, 0, 0.05);
}

.btn-icon.danger:hover {
  background: rgba(255, 0, 0, 0.1);
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1200px) {
  .content-layout {
    grid-template-columns: 1fr;
  }

  .map-section {
    position: relative;
    top: 0;
  }

  #map {
    height: 400px;
  }
}

/* Custom marker styles */
:deep(.custom-marker) {
  background: none;
  border: none;
}

:deep(.marker-pin) {
  font-size: 2rem;
  text-shadow: 0 2px 4px rgba(0,0,0,0.3);
  position: relative;
}

:deep(.marker-pin.new) {
  filter: drop-shadow(0 0 8px rgba(66, 185, 131, 0.6));
}

:deep(.marker-pin.existing) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  width: 30px;
  height: 30px;
  border-radius: 50% 50% 50% 0;
  transform: rotate(-45deg);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.3);
}

:deep(.marker-number) {
  transform: rotate(45deg);
  font-size: 0.85rem;
  font-weight: 700;
}
</style>
