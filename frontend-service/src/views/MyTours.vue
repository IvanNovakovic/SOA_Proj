<template>
  <div class="my-tours">
    <div class="header">
      <h1>My Tours</h1>
      <router-link to="/tours/create" class="btn-create">
        + Create New Tour
      </router-link>
    </div>

    <div v-if="loading" class="loading">Loading tours...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <div v-else-if="tours.length === 0" class="empty-state">
      <p>You haven't created any tours yet.</p>
      <router-link to="/tours/create" class="btn-primary">Create Your First Tour</router-link>
    </div>

    <div v-else class="tours-grid">
      <div v-for="tour in tours" :key="tour.id" class="tour-card">
        <div class="tour-header">
          <h3>{{ tour.name }}</h3>
          <span class="status-badge" :class="tour.status">{{ tour.status }}</span>
        </div>
        
        <p v-if="tour.description" class="tour-description">{{ tour.description }}</p>
        
        <div class="tour-meta">
          <span class="meta-item">
            <strong>Difficulty:</strong> {{ tour.difficulty || 'Not set' }}
          </span>
          <span class="meta-item">
            <strong>Price:</strong> ${{ tour.price }}
          </span>
        </div>

        <div v-if="tour.tags && tour.tags.length > 0" class="tour-tags">
          <span v-for="tag in tour.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <div class="tour-actions">
          <router-link :to="`/tours/${tour.id}`" class="btn-view">View Details</router-link>
          <router-link :to="`/tours/${tour.id}/edit`" class="btn-edit">Edit Tour</router-link>
        </div>

        <div class="tour-date">
          Created: {{ formatDate(tour.createdAt) }}
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
  name: 'MyTours',
  setup() {
    const tours = ref([])
    const loading = ref(true)
    const error = ref('')

    const fetchTours = async () => {
      loading.value = true
      error.value = ''
      
      try {
        const userId = authStore.getUserId()
        const data = await api.getToursByAuthor(userId)
        tours.value = data || []
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to load tours'
        tours.value = []
      } finally {
        loading.value = false
      }
    }

    const formatDate = (dateString) => {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      })
    }

    onMounted(() => {
      fetchTours()
    })

    return {
      tours,
      loading,
      error,
      formatDate
    }
  }
}
</script>

<style scoped>
.my-tours {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 2rem;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

h1 {
  color: #2c3e50;
  margin: 0;
}

.btn-create {
  background: #42b983;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-create:hover {
  background: #35a372;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(66, 185, 131, 0.3);
}

.loading {
  text-align: center;
  padding: 3rem;
  color: #666;
  font-size: 1.2rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #fcc;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-state p {
  color: #666;
  font-size: 1.2rem;
  margin-bottom: 2rem;
}

.btn-primary {
  background: #42b983;
  color: white;
  padding: 1rem 2rem;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  display: inline-block;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: #35a372;
}

.tours-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 2rem;
}

.tour-card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 12px;
  padding: 1.5rem;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
}

.tour-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #42b983;
}

.tour-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
  gap: 1rem;
}

.tour-header h3 {
  margin: 0;
  color: #2c3e50;
  font-size: 1.3rem;
  flex: 1;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.85rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.draft {
  background: #fff3e0;
  color: #f57c00;
}

.status-badge.published {
  background: #e8f5e9;
  color: #4caf50;
}

.tour-description {
  color: #666;
  margin: 0 0 1rem 0;
  line-height: 1.5;
}

.tour-meta {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.meta-item {
  color: #666;
}

.meta-item strong {
  color: #333;
}

.tour-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.tag {
  background: #f0f0f0;
  color: #666;
  padding: 0.3rem 0.8rem;
  border-radius: 12px;
  font-size: 0.85rem;
}

.tour-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: auto;
  padding-top: 1rem;
  border-top: 1px solid #eee;
  flex-wrap: wrap;
}

.btn-view,
.btn-edit,
.btn-keypoints {
  flex: 1;
  min-width: 100px;
  text-align: center;
  padding: 0.6rem 1rem;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 500;
  font-size: 0.9rem;
  transition: all 0.2s;
}

.btn-view {
  background: #42b983;
  color: white;
}

.btn-view:hover {
  background: #35a372;
}

.btn-edit {
  background: #3498db;
  color: white;
}

.btn-edit:hover {
  background: #2980b9;
}

.btn-keypoints {
  background: #f5f5f5;
  color: #666;
}

.btn-keypoints:hover {
  background: #e0e0e0;
}

.tour-date {
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: #999;
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .tours-grid {
    grid-template-columns: 1fr;
  }
}
</style>
