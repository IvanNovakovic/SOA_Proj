<template>
  <div class="tour-detail">
    <div v-if="loading" class="loading">Loading tour...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>

    <div v-else class="tour-content">
      <!-- Tour Header -->
      <div class="tour-header">
        <div>
          <h1>{{ tour.name }}</h1>
          <div class="tour-meta">
            <span class="status-badge" :class="tour.status">{{ tour.status }}</span>
            <span class="difficulty">Difficulty: {{ tour.difficulty }}</span>
            <span class="price">Price: ${{ tour.price }}</span>
          </div>
        </div>
      </div>

      <!-- Tour Description -->
      <div v-if="tour.description" class="tour-description">
        <h2>Description</h2>
        <p>{{ tour.description }}</p>
      </div>

      <!-- Tags -->
      <div v-if="tour.tags && tour.tags.length > 0" class="tour-tags">
        <h3>Tags</h3>
        <div class="tags-list">
          <span v-for="tag in tour.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>
      </div>

      <!-- Travel Times & Distance -->
      <div v-if="(tour.distance > 0) || (tour.durations && (tour.durations.walking > 0 || tour.durations.biking > 0 || tour.durations.driving > 0))" class="tour-travel-times">
        <h3>Tour Information</h3>
        <div class="travel-times-grid">
          <div v-if="tour.distance > 0" class="travel-time-card distance-card">
            <span class="travel-icon">📏</span>
            <span class="travel-label">Distance</span>
            <span class="travel-value">{{ tour.distance.toFixed(2) }} km</span>
          </div>
          <div v-if="tour.durations.walking > 0" class="travel-time-card">
            <span class="travel-icon">🚶</span>
            <span class="travel-label">Walking</span>
            <span class="travel-value">{{ formatDuration(tour.durations.walking) }}</span>
          </div>
          <div v-if="tour.durations.biking > 0" class="travel-time-card">
            <span class="travel-icon">🚴</span>
            <span class="travel-label">Biking</span>
            <span class="travel-value">{{ formatDuration(tour.durations.biking) }}</span>
          </div>
          <div v-if="tour.durations.driving > 0" class="travel-time-card">
            <span class="travel-icon">🚗</span>
            <span class="travel-label">Driving</span>
            <span class="travel-value">{{ formatDuration(tour.durations.driving) }}</span>
          </div>
        </div>
      </div>

      <!-- Key Points Section -->
      <div class="section keypoints-section">
        <div class="section-header">
          <h2>Key Points</h2>
          <div class="header-actions">
            <router-link 
              v-if="keyPoints.length > 0"
              :to="`/tours/${tour.id}/map`" 
              class="btn-view-map"
            >
              🗺️ View Route Map
            </router-link>
          </div>
        </div>

        <div v-if="loadingKeyPoints" class="loading-small">Loading key points...</div>
        <div v-else-if="keyPoints.length === 0" class="empty-message">
          <p>No key points available.</p>
        </div>
        <div v-else class="keypoints-grid">
          <div v-for="(kp, index) in keyPoints" :key="kp.id" class="keypoint-card">
            <div class="keypoint-number">{{ index + 1 }}</div>
            <div class="keypoint-image-container">
              <img 
                v-if="kp.imageUrl" 
                :src="kp.imageUrl" 
                :alt="kp.name" 
                class="keypoint-image"
                @error="handleImageError"
              />
              <div v-else class="keypoint-image-placeholder">
                <span class="placeholder-icon">🏞️</span>
              </div>
            </div>
            <div class="keypoint-content">
              <h3>{{ kp.name }}</h3>
              <p v-if="kp.description">{{ kp.description }}</p>
              <div class="keypoint-location">
                📍 {{ kp.latitude.toFixed(4) }}, {{ kp.longitude.toFixed(4) }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Reviews Section -->
      <div class="section reviews-section">
        <div class="section-header">
          <h2>Reviews</h2>
          <button 
            v-if="canAddReview" 
            @click="showReviewForm = true" 
            class="btn-add"
          >
            + Add Review
          </button>
        </div>

        <!-- Review Form -->
        <div v-if="showReviewForm" class="review-form">
          <h3>Write a Review</h3>
          <form @submit.prevent="submitReview">
            <div class="form-group">
              <label>Rating *</label>
              <div class="rating-input">
                <button
                  v-for="n in 5"
                  :key="n"
                  type="button"
                  @click="reviewForm.rating = n"
                  class="star-btn"
                  :class="{ active: n <= reviewForm.rating }"
                >
                  ★
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>Comment</label>
              <textarea 
                v-model="reviewForm.comment" 
                placeholder="Share your experience..."
                rows="4"
              ></textarea>
            </div>

            <div class="form-group">
              <label>When did you visit?</label>
              <input 
                v-model="reviewForm.visitedAt" 
                type="date" 
                :max="today"
              />
            </div>

            <div class="form-group">
              <label>Image URLs (one per line)</label>
              <textarea 
                v-model="imagesText" 
                placeholder="https://example.com/image1.jpg&#10;https://example.com/image2.jpg"
                rows="3"
              ></textarea>
            </div>

            <div class="form-actions">
              <button type="submit" :disabled="!reviewForm.rating || reviewLoading" class="btn-primary">
                {{ reviewLoading ? 'Submitting...' : 'Submit Review' }}
              </button>
              <button type="button" @click="cancelReview" class="btn-secondary">Cancel</button>
            </div>
          </form>
        </div>

        <!-- Reviews List -->
        <div v-if="loadingReviews" class="loading-small">Loading reviews...</div>
        <div v-else-if="reviews.length === 0" class="empty-message">
          No reviews yet. Be the first to review!
        </div>
        <div v-else class="reviews-list">
          <div v-for="review in reviews" :key="review.id" class="review-card">
            <div class="review-header">
              <div>
                <h4>{{ review.authorName || 'Anonymous' }}</h4>
                <div class="review-meta">
                  <span class="rating">{{ '★'.repeat(review.rating) }}{{ '☆'.repeat(5 - review.rating) }}</span>
                  <span v-if="review.visitedAt" class="visited-date">
                    Visited: {{ formatDate(review.visitedAt) }}
                  </span>
                </div>
              </div>
              <span class="review-date">{{ formatDate(review.createdAt) }}</span>
            </div>

            <p v-if="review.comment" class="review-comment">{{ review.comment }}</p>

            <div v-if="review.images && review.images.length > 0" class="review-images">
              <img 
                v-for="(img, index) in review.images" 
                :key="index" 
                :src="img" 
                :alt="`Review image ${index + 1}`"
                class="review-image"
              />
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../services/api'
import { authStore } from '../stores/authStore'

export default {
  name: 'PurchaseTourDetail',
  setup() {
    const route = useRoute()
    const tourId = route.params.id

    const tour = ref({
      tags: [],
      name: '',
      description: '',
      status: '',
      difficulty: '',
      price: 0,
      durations: { walking: 0, biking: 0, driving: 0 }
    })
    const keyPoints = ref([])
    const reviews = ref([])

    const loading = ref(true)
    const loadingKeyPoints = ref(false)
    const loadingReviews = ref(false)
    const error = ref('')

    const showReviewForm = ref(false)
    const reviewForm = ref({ rating: 0, comment: '', visitedAt: null })
    const imagesText = ref('')
    const reviewLoading = ref(false)

    const today = new Date().toISOString().split('T')[0]

    const canAddReview = computed(() => authStore.isAuthenticated.value)

    const fetchTour = async () => {
      loading.value = true
      error.value = ''
      try {
        const data = await api.getTourById(tourId)
        // Ensure status is 'purchased' if needed
        tour.value = { ...data, status: 'purchased', tags: data.tags || [] }
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to load tour'
      } finally {
        loading.value = false
      }
    }

    const fetchKeyPoints = async () => {
      loadingKeyPoints.value = true
      try {
        const data = await api.getKeyPoints(tourId)
        keyPoints.value = data || []
      } catch (err) {
        keyPoints.value = []
      } finally {
        loadingKeyPoints.value = false
      }
    }

    const fetchReviews = async () => {
      loadingReviews.value = true
      try {
        const data = await api.getReviews(tourId)
        reviews.value = data || []
      } catch (err) {
        reviews.value = []
      } finally {
        loadingReviews.value = false
      }
    }

    const submitReview = async () => {
      if (!reviewForm.value.rating) return alert('Please select a rating')
      reviewLoading.value = true

      try {
        const images = imagesText.value
          .split('\n')
          .map(url => url.trim())
          .filter(url => url.length > 0)

        const reviewData = {
          rating: reviewForm.value.rating,
          comment: reviewForm.value.comment,
          visitedAt: reviewForm.value.visitedAt ? new Date(reviewForm.value.visitedAt).toISOString() : null,
          images
        }

        await api.createReview(tourId, reviewData)
        reviewForm.value = { rating: 0, comment: '', visitedAt: null }
        imagesText.value = ''
        showReviewForm.value = false
        await fetchReviews()
      } catch (err) {
        alert(err.response?.data?.error || err.message || 'Failed to submit review')
      } finally {
        reviewLoading.value = false
      }
    }

    const cancelReview = () => {
      reviewForm.value = { rating: 0, comment: '', visitedAt: null }
      imagesText.value = ''
      showReviewForm.value = false
    }

    const formatDate = (dateString) => {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      })
    }

    const handleImageError = (event) => {
      event.target.style.display = 'none'
      const placeholder = event.target.nextElementSibling
      if (placeholder && placeholder.classList.contains('keypoint-image-placeholder')) {
        placeholder.style.display = 'flex'
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

    onMounted(async () => {
      await fetchTour()
      fetchKeyPoints()
      fetchReviews()
    })

    return {
      tour,
      keyPoints,
      reviews,
      loading,
      loadingKeyPoints,
      loadingReviews,
      error,
      showReviewForm,
      reviewForm,
      reviewLoading,
      imagesText,
      today,
      canAddReview,
      submitReview,
      cancelReview,
      formatDate,
      formatDuration,
      handleImageError
    }
  }
}
</script>

<style scoped>
.tour-detail {
  max-width: 1200px;
  margin: 2rem auto;
  padding: 2rem;
}

.loading,
.loading-small {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.loading-small {
  padding: 1rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid #fcc;
}

.tour-header {
  margin-bottom: 2rem;
}

.tour-header h1 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.tour-meta {
  display: flex;
  gap: 1.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.status-badge {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
  text-transform: uppercase;
}

.status-badge.draft {
  background: #fff3e0;
  color: #f57c00;
}

.status-badge.purchased {
  background: #f1d21f;
  color: #826909;
}

.status-badge.archived {
  background: #eeeeee;
  color: #757575;
}

.tour-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
}

.author-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.btn-action {
  padding: 0.6rem 1.2rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-publish {
  background: #af914c;
  color: white;
}

.btn-publish:hover:not(:disabled) {
  background: #dca931;
}

.btn-archive {
  background: #ff9800;
  color: white;
}

.btn-archive:hover:not(:disabled) {
  background: #fb8c00;
}

.btn-activate {
  background: #2196f3;
  color: white;
}

.btn-activate:hover:not(:disabled) {
  background: #1976d2;
}

.difficulty,
.price {
  color: #666;
  font-weight: 500;
}

.tour-description {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.tour-description h2 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.tour-description p {
  color: #666;
  line-height: 1.6;
}

.tour-tags {
  margin-bottom: 2rem;
}

.tour-tags h3 {
  color: #2c3e50;
  margin-bottom: 0.5rem;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tag {
  background: #42b983;
  color: white;
  padding: 0.4rem 1rem;
  border-radius: 20px;
  font-size: 0.9rem;
}

.tour-travel-times {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.tour-travel-times h3 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.travel-times-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.travel-time-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.5rem;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 12px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.travel-time-card .travel-icon {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.travel-time-card .travel-label {
  font-size: 0.9rem;
  color: #666;
  margin-bottom: 0.25rem;
}

.travel-time-card .travel-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: #2c3e50;
}

.travel-time-card.distance-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.travel-time-card.distance-card .travel-label,
.travel-time-card.distance-card .travel-value {
  color: white;
}

.section {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.section-header h2 {
  color: #2c3e50;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-add,
.btn-manage,
.btn-view-map {
  background: #42b983;
  color: white;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  border: none;
  text-decoration: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-view-map {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.btn-add:hover,
.btn-manage:hover {
  background: #35a372;
  transform: translateY(-1px);
}

.btn-view-map:hover {
  background: linear-gradient(135deg, #5568d3 0%, #65408e 100%);
  transform: translateY(-1px);
}

.btn-add-inline {
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  display: inline-block;
  margin-top: 1rem;
  transition: all 0.2s;
}

.btn-add-inline:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(66, 185, 131, 0.3);
}

.empty-message {
  text-align: center;
  color: #999;
  padding: 2rem;
}

.empty-message p {
  font-style: italic;
  margin-bottom: 0.5rem;
}

.keypoints-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.keypoint-card {
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s;
  position: relative;
}

.keypoint-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #42b983;
  transform: translateY(-2px);
}

.keypoint-number {
  position: absolute;
  top: 1rem;
  left: 1rem;
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
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  z-index: 1;
}

.keypoint-image-container {
  width: 100%;
  height: 200px;
  position: relative;
  overflow: hidden;
}

.keypoint-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.keypoint-image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #e0e0e0 0%, #f5f5f5 100%);
}

.placeholder-icon {
  font-size: 4rem;
  opacity: 0.5;
}

.keypoint-content {
  padding: 1rem;
}

.keypoint-content h3 {
  color: #2c3e50;
  margin: 0 0 0.5rem 0;
}

.keypoint-content p {
  color: #666;
  margin: 0 0 0.5rem 0;
}

.keypoint-location {
  color: #42b983;
  font-size: 0.9rem;
  font-weight: 500;
}

.review-form {
  background: #f8f9fa;
  padding: 1.5rem;
  border-radius: 8px;
  margin-bottom: 2rem;
}

.review-form h3 {
  color: #2c3e50;
  margin-bottom: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.rating-input {
  display: flex;
  gap: 0.5rem;
}

.star-btn {
  background: none;
  border: none;
  font-size: 2rem;
  color: #ddd;
  cursor: pointer;
  transition: color 0.2s;
}

.star-btn.active {
  color: #ffc107;
}

.star-btn:hover {
  color: #ffb300;
}

.form-group input[type="date"],
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
}

.form-group textarea {
  resize: vertical;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.btn-primary,
.btn-secondary {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #42b983;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #35a372;
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

.reviews-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.review-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 1.5rem;
}

.review-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.review-header h4 {
  color: #2c3e50;
  margin: 0 0 0.5rem 0;
}

.review-meta {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.rating {
  color: #ffc107;
  font-size: 1.1rem;
}

.visited-date,
.review-date {
  color: #999;
  font-size: 0.9rem;
}

.review-comment {
  color: #666;
  line-height: 1.6;
  margin: 0 0 1rem 0;
}

.review-images {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.review-image {
  width: 150px;
  height: 150px;
  object-fit: cover;
  border-radius: 8px;
}

@media (max-width: 768px) {
  .keypoints-grid {
    grid-template-columns: 1fr;
  }
}

</style>



