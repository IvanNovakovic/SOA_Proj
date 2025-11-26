<template>
  <div class="tours-list">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>Loading purchased tours...</p>
    </div>
    <div v-else-if="error" class="error-message">
      <span class="error-icon">⚠️</span>
      {{ error }}
    </div>
    <div v-else-if="tours.length === 0" class="empty-state">
      <div class="empty-icon">🗺️</div>
      <h2>No Purchased Tours</h2>
      <p>You haven't purchased any tours yet.</p>
      <p class="hint">Discover amazing tours and add them to your cart!</p>
      <router-link to="/tours" class="btn-primary">
        <span>🔍</span> Discover Tours
      </router-link>
    </div>

    <div v-else>
      <div class="tours-header">
        <h2>My Purchased Tours</h2>
        <p class="tours-count">{{ tours.length }} tour{{ tours.length !== 1 ? 's' : '' }} purchased</p>
      </div>
      
      <div class="tours-grid">
        <div v-for="tour in tours" :key="tour.id" class="tour-card">
          <div class="tour-image-placeholder">
            <span class="tour-icon">🎯</span>
          </div>
          
          <div class="tour-content">
            <div class="tour-header">
              <h3>{{ tour.name }}</h3>
              <span class="status-badge purchased">Purchased</span>
            </div>
            
            <p v-if="tour.description" class="tour-description">{{ tour.description }}</p>
            
            <div class="tour-meta">
              <div class="meta-row">
                <span class="meta-item">
                  <span class="icon">👤</span>
                  <strong>Guide:</strong> {{ tour.authorName }}
                </span>
              </div>
              <div class="meta-row">
                <span class="meta-item">
                  <span class="icon">📊</span>
                  <strong>Difficulty:</strong> 
                  <span :class="['difficulty-badge', tour.difficulty?.toLowerCase()]">
                    {{ tour.difficulty || 'Not set' }}
                  </span>
                </span>
                <span class="meta-item">
                  <span class="icon">💰</span>
                  <strong>Price:</strong> ${{ tour.price?.toFixed(2) || '0.00' }}
                </span>
              </div>
              <div v-if="(tour.distance > 0) || (tour.durations && hasDurations(tour.durations))" class="meta-row travel-times-row">
                <div class="travel-times-compact">
                  <span v-if="tour.distance > 0" class="travel-time-compact distance-badge">
                    📏 {{ tour.distance.toFixed(2) }} km
                  </span>
                  <span v-if="tour.durations?.walking > 0" class="travel-time-compact">
                    🚶 {{ formatDuration(tour.durations.walking) }}
                  </span>
                  <span v-if="tour.durations?.biking > 0" class="travel-time-compact">
                    🚴 {{ formatDuration(tour.durations.biking) }}
                  </span>
                  <span v-if="tour.durations?.driving > 0" class="travel-time-compact">
                    🚗 {{ formatDuration(tour.durations.driving) }}
                  </span>
                </div>
              </div>
            </div>

            <div v-if="tour.tags && tour.tags.length > 0" class="tour-tags">
              <span v-for="tag in tour.tags.slice(0, 4)" :key="tag" class="tag">
                {{ tag }}
              </span>
              <span v-if="tour.tags.length > 4" class="tag more">
                +{{ tour.tags.length - 4 }} more
              </span>
            </div>

            <div class="tour-footer">
              <router-link :to="`/purchased-tour-detail/${tour.id}`" class="btn-view">
                View Details →
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { api } from '../services/api'
import { authStore } from '../stores/authStore'

export default {
  name: 'PurchasedTours',
  setup() {
    const tours = ref([])
    const loading = ref(true)
    const error = ref('')

    const fetchPurchasedTours = async () => {
      loading.value = true
      error.value = ''

      try {
        const userId = authStore.getUserId()
        const purchasedTokens = await api.getPurchasedTours(userId)

        if (!purchasedTokens || purchasedTokens.length === 0) {
          tours.value = []
          return
        }

        const detailedTours = await Promise.all(
          purchasedTokens.map(async (token) => {
            try {
              const tour = await api.getTourById(token.tour_id)

              return {
                ...tour,
                authorName: tour.authorName || tour.author?.username || 'Guide',
                difficulty: tour.difficulty || 'Not set',
                price: tour.price || 0,
                distance: tour.distance || 0,
                durations: {
                  walking: tour.durations?.walking || tour.tour_duration || 0,
                  biking: tour.durations?.biking || 0,
                  driving: tour.durations?.driving || 0
                },
                tags: tour.tags || [],
                start_point: tour.start_point || tour.startLocation || 'Unknown',
                token: token.token
              }
            } catch (err) {
              console.error(`Failed to fetch tour ${token.tour_id}:`, err)
              return null
            }
          })
        )

        tours.value = detailedTours.filter(t => t !== null)
          .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))

      } catch (err) {
        console.error('Error fetching purchased tours:', err)
        error.value = err.response?.data?.error || err.message || 'Failed to load purchased tours'
        tours.value = []
      } finally {
        loading.value = false
      }
    }

    const formatDuration = (minutes) => {
      if (!minutes || minutes === 0) return ''
      const hours = Math.floor(minutes / 60)
      const mins = minutes % 60
      if (hours === 0) return `${mins} min`
      if (mins === 0) return `${hours}h`
      return `${hours}h ${mins}min`
    }

    const hasDurations = (durations) => {
      return durations && (durations.walking > 0 || durations.biking > 0 || durations.driving > 0)
    }

    onMounted(() => {
      fetchPurchasedTours()
    })

    return {
      tours,
      loading,
      error,
      formatDuration,
      hasDurations
    }
  }
}
</script>

<style scoped>
.tours-list {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5rem 2rem;
  gap: 1rem;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #42b983;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading p {
  color: #666;
  font-size: 1.1rem;
}

.error-message {
  background: linear-gradient(135deg, #fee 0%, #fdd 100%);
  color: #c33;
  padding: 1.5rem;
  border-radius: 12px;
  text-align: center;
  border: 2px solid #fcc;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 1.1rem;
}

.error-icon {
  font-size: 1.5rem;
}

.empty-state {
  text-align: center;
  padding: 5rem 2rem;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border-radius: 20px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.08);
  max-width: 600px;
  margin: 2rem auto;
}

.empty-icon {
  font-size: 5rem;
  margin-bottom: 1.5rem;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.empty-state h2 {
  color: #2c3e50;
  margin-bottom: 1rem;
  font-size: 1.8rem;
}

.empty-state p {
  color: #666;
  font-size: 1.1rem;
  margin-bottom: 0.5rem;
  line-height: 1.6;
}

.empty-state .hint {
  color: #999;
  font-size: 1rem;
  margin-bottom: 2rem;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  padding: 1rem 2.5rem;
  border-radius: 12px;
  text-decoration: none;
  font-weight: 600;
  font-size: 1.1rem;
  transition: all 0.3s;
  box-shadow: 0 4px 15px rgba(66, 185, 131, 0.3);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 25px rgba(66, 185, 131, 0.4);
}

.tours-header {
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid #e0e0e0;
}

.tours-header h2 {
  color: #2c3e50;
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.tours-count {
  color: #666;
  font-size: 1rem;
}

.tours-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 2rem;
}

.tour-card {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  transition: all 0.3s;
  border: 2px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.tour-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
  border-color: #42b983;
}

.tour-image-placeholder {
  height: 180px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.tour-icon {
  font-size: 4rem;
  filter: drop-shadow(0 4px 8px rgba(0,0,0,0.2));
}

.tour-content {
  padding: 1.5rem;
}

.tour-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  gap: 1rem;
}

.tour-header h3 {
  color: #2c3e50;
  margin: 0;
  flex: 1;
  font-size: 1.4rem;
  line-height: 1.3;
}

.status-badge {
  padding: 0.35rem 0.8rem;
  border-radius: 20px;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
  flex-shrink: 0;
  align-self: flex-start;
}

.status-badge.purchased {
  background: linear-gradient(135deg, #D4AF37 0%, #B8860B 100%); /* mat zlatni gradijent */
  color: #5C4000; 
  border: 1px solid #A67C00; 
  font-weight: 600;
  text-transform: uppercase;
  padding: 0.35rem 0.8rem;
  border-radius: 20px;
  font-size: 0.7rem;
  letter-spacing: 0.5px;
}

.tour-description {
  color: #555;
  margin-bottom: 1.25rem;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 0.95rem;
}

.tour-meta {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 10px;
}

.meta-row {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: #666;
  font-size: 0.9rem;
}

.meta-item .icon {
  font-size: 1.1rem;
}

.meta-item strong {
  color: #333;
  margin-right: 0.25rem;
}

.travel-times-row {
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid #e9ecef;
}

.travel-times-compact {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  font-size: 0.9rem;
}

.travel-time-compact {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.75rem;
  background: #f8f9fa;
  border-radius: 12px;
  color: #495057;
  font-weight: 500;
}

.travel-time-compact.distance-badge {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
}

.difficulty-badge {
  padding: 0.2rem 0.6rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.85rem;
}

.difficulty-badge.easy {
  background: #d1f2eb;
  color: #0f6848;
}

.difficulty-badge.medium {
  background: #fff3cd;
  color: #856404;
}

.difficulty-badge.hard {
  background: #f8d7da;
  color: #721c24;
}

.tour-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}

.tag {
  background: linear-gradient(135deg, #e3f2fd 0%, #bbdefb 100%);
  color: #1565c0;
  padding: 0.4rem 0.9rem;
  border-radius: 16px;
  font-size: 0.8rem;
  font-weight: 500;
  transition: all 0.2s;
}

.tag:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(21, 101, 192, 0.2);
}

.tag.more {
  background: linear-gradient(135deg, #e0e0e0 0%, #bdbdbd 100%);
  color: #555;
}

.tour-footer {
  padding-top: 1rem;
  border-top: 1px solid #e0e0e0;
}

.btn-view {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.9rem 1.5rem;
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  border-radius: 10px;
  text-decoration: none;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.3s;
  box-shadow: 0 2px 8px rgba(66, 185, 131, 0.3);
}

.btn-view:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(66, 185, 131, 0.4);
}

@media (max-width: 768px) {
  .tours-grid {
    grid-template-columns: 1fr;
  }
  
  .tours-list {
    padding: 1rem;
  }
  
  .tour-header h3 {
    font-size: 1.2rem;
  }
}
</style>
