<template>
  <div class="tours-list">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>Loading tours...</p>
    </div>
    <div v-else-if="error" class="error-message">
      <span class="error-icon">⚠️</span>
      {{ error }}
    </div>
    <div v-else-if="tours.length === 0" class="empty-state">
      <div class="empty-icon">🗺️</div>
      <h2>No Tours Available</h2>
      <p>Tours from guides you follow will appear here.</p>
      <p class="hint">Start by following some guides to discover their amazing tours!</p>
      <router-link to="/recommendations" class="btn-primary">
        <span>🔍</span> Discover Guides
      </router-link>
    </div>

    <div v-else>
      <div class="tours-header">
        <h2>Discover Tours</h2>
        <p class="tours-count">{{ tours.length }} tour{{ tours.length !== 1 ? 's' : '' }} available</p>
      </div>
      
      <div class="tours-grid">
        <div v-for="tour in tours" :key="tour.id" class="tour-card">
          <div class="tour-image-placeholder">
            <span class="tour-icon">🎯</span>
          </div>
          
          <div class="tour-content">
            <div class="tour-header">
              <h3>{{ tour.name }}</h3>
              <span class="status-badge published">Published</span>
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
                  <span v-if="tour.durations && tour.durations.walking > 0" class="travel-time-compact">
                    🚶 {{ formatDuration(tour.durations.walking) }}
                  </span>
                  <span v-if="tour.durations && tour.durations.biking > 0" class="travel-time-compact">
                    🚴 {{ formatDuration(tour.durations.biking) }}
                  </span>
                  <span v-if="tour.durations && tour.durations.driving > 0" class="travel-time-compact">
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
              <router-link :to="`/tours/${tour.id}`" class="btn-view">
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
  name: 'ToursList',
  setup() {
    const tours = ref([])
    const loading = ref(true)
    const error = ref('')

    const fetchToursFromFollowing = async () => {
      loading.value = true
      error.value = ''
      
      try {
        const userId = authStore.getUserId()
        
        // Get list of users the current user is following
        const followingData = await api.getFollowing(userId)
        console.log('Following data:', followingData)
        
        // Handle different response formats - it's an array of user IDs
        let followingIds = []
        if (Array.isArray(followingData)) {
          // Check if it's an array of objects or array of strings
          if (followingData.length > 0 && typeof followingData[0] === 'string') {
            followingIds = followingData
          } else if (followingData.length > 0 && typeof followingData[0] === 'object') {
            followingIds = followingData.map(u => u.user_id || u.userId || u.id)
          }
        }
        
        if (!followingIds || followingIds.length === 0) {
          tours.value = []
          loading.value = false
          return
        }

        console.log('Following user IDs:', followingIds)

        // First, fetch usernames for all followed users
        const usernameMap = new Map()
        await Promise.all(
          followingIds.map(async (userId) => {
            try {
              const user = await api.getUserById(userId)
              usernameMap.set(userId, user.username || 'Guide')
            } catch (err) {
              console.error(`Failed to fetch username for user ${userId}:`, err)
              usernameMap.set(userId, 'Guide')
            }
          })
        )

        // Fetch tours from each followed guide
        const tourPromises = followingIds.map(async (followedUserId) => {
          try {
            console.log(`Fetching tours for user ${followedUserId}`)
            const userTours = await api.getToursByAuthor(followedUserId)
            console.log(`Tours for ${followedUserId}:`, userTours)
            
            // Filter only published tours and add the username
            return (userTours || [])
              .filter(tour => tour.status === 'published')
              .map(tour => ({
                ...tour,
                authorName: usernameMap.get(tour.authorId) || 'Guide'
              }))
          } catch (err) {
            console.error(`Failed to fetch tours for user ${followedUserId}:`, err)
            return []
          }
        })

        const allTours = await Promise.all(tourPromises)
        console.log('All tours fetched:', allTours)
        
        // Flatten the array and sort by creation date
        tours.value = allTours
          .flat()
          .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
        
        console.log('Final tours:', tours.value)
        
      } catch (err) {
        console.error('Error fetching tours:', err)
        error.value = err.response?.data?.error || err.message || 'Failed to load tours'
        tours.value = []
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      fetchToursFromFollowing()
    })

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

.status-badge.published {
  background: linear-gradient(135deg, #d4edda 0%, #c3e6cb 100%);
  color: #155724;
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
