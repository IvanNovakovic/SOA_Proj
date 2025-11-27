<template>
  <div class="tour-edit">
    <div v-if="loading" class="loading">Loading tour...</div>
    <div v-else-if="loadError" class="error-message">{{ loadError }}</div>
    
    <template v-else>
      <h1>Edit Tour</h1>
      
      <div v-if="error" class="error-message">{{ error }}</div>
      <div v-if="success" class="success-message">Tour updated successfully!</div>

      <form @submit.prevent="handleSubmit" class="tour-form">
        <div class="form-group">
          <label for="name">Tour Name *</label>
          <input 
            id="name"
            v-model="form.name" 
            type="text" 
            placeholder="Enter tour name"
            required
          />
        </div>

        <div class="form-group">
          <label for="description">Description</label>
          <textarea 
            id="description"
            v-model="form.description" 
            placeholder="Describe your tour"
            rows="5"
          ></textarea>
        </div>

        <div class="form-group">
          <label for="difficulty">Difficulty *</label>
          <select id="difficulty" v-model="form.difficulty" required>
            <option value="">Select difficulty</option>
            <option value="Easy">Easy</option>
            <option value="Medium">Medium</option>
            <option value="Hard">Hard</option>
          </select>
        </div>

        <div class="form-group">
          <label for="status">Status *</label>
          <select id="status" v-model="form.status" required>
            <option value="draft">Draft</option>
            <option value="published">Published</option>
            <option value="archived">Archived</option>
          </select>
        </div>

        <div class="form-group">
          <label for="price">Price ($) *</label>
          <input 
            id="price"
            v-model.number="form.price" 
            type="number" 
            min="0"
            step="0.01"
            placeholder="Enter price"
            required
          />
        </div>

        <div class="form-group">
          <label for="tags">Tags</label>
          <div class="tags-input">
            <input 
              id="tags"
              v-model="tagInput" 
              type="text" 
              placeholder="Add a tag and press Enter"
              @keydown.enter.prevent="addTag"
            />
            <div class="tags-list">
              <span v-for="(tag, index) in form.tags" :key="index" class="tag">
                {{ tag }}
                <button type="button" @click="removeTag(index)" class="remove-tag">&times;</button>
              </span>
            </div>
          </div>
        </div>

        <div class="form-group info-section">
          <label>Transport Durations</label>
          <p class="field-note">⏱️ Durations are automatically calculated from your tour's key points</p>
          <div class="durations-display">
            <div class="duration-badge">
              <span class="duration-icon">🚶</span>
              <span class="duration-label">Walking</span>
              <span class="duration-value">{{ form.durations.walking > 0 ? `${form.durations.walking} min` : 'Not set' }}</span>
            </div>
            <div class="duration-badge">
              <span class="duration-icon">🚴</span>
              <span class="duration-label">Cycling</span>
              <span class="duration-value">{{ form.durations.biking > 0 ? `${form.durations.biking} min` : 'Not set' }}</span>
            </div>
            <div class="duration-badge">
              <span class="duration-icon">🚗</span>
              <span class="duration-label">Driving</span>
              <span class="duration-value">{{ form.durations.driving > 0 ? `${form.durations.driving} min` : 'Not set' }}</span>
            </div>
          </div>
        </div>

        <div class="form-actions">
          <button type="submit" :disabled="saving" class="btn-primary">
            {{ saving ? 'Saving...' : 'Save Changes' }}
          </button>
          <button type="button" @click="goBack" class="btn-secondary">Cancel</button>
        </div>
      </form>
    </template>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { api } from '../services/api'

export default {
  name: 'TourEdit',
  setup() {
    const router = useRouter()
    const route = useRoute()
    const tourId = route.params.id

    const form = ref({
      name: '',
      description: '',
      difficulty: '',
      tags: [],
      status: 'draft',
      price: 0,
      durations: {
        walking: 0,
        biking: 0,
        driving: 0
      }
    })
    const tagInput = ref('')
    const loading = ref(true)
    const saving = ref(false)
    const error = ref('')
    const loadError = ref('')
    const success = ref(false)

    const fetchTour = async () => {
      loading.value = true
      loadError.value = ''

      try {
        const tour = await api.getTourById(tourId)
        form.value = {
          name: tour.name,
          description: tour.description || '',
          difficulty: tour.difficulty,
          tags: tour.tags || [],
          status: tour.status,
          price: tour.price,
          durations: tour.durations || { walking: 0, biking: 0, driving: 0 }
        }
      } catch (err) {
        loadError.value = err.response?.data?.error || err.message || 'Failed to load tour'
      } finally {
        loading.value = false
      }
    }

    const addTag = () => {
      const tag = tagInput.value.trim()
      if (tag && !form.value.tags.includes(tag)) {
        form.value.tags.push(tag)
        tagInput.value = ''
      }
    }

    const removeTag = (index) => {
      form.value.tags.splice(index, 1)
    }

    const handleSubmit = async () => {
      saving.value = true
      error.value = ''
      success.value = false

      try {
        const tourData = {
          name: form.value.name,
          description: form.value.description,
          difficulty: form.value.difficulty,
          tags: form.value.tags,
          status: form.value.status,
          price: form.value.price
          // durations are auto-updated when keypoints change
        }

        await api.updateTour(tourId, tourData)
        success.value = true
        
        setTimeout(() => {
          router.push('/tours/my-tours')
        }, 1500)
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to update tour'
      } finally {
        saving.value = false
      }
    }

    const goBack = () => {
      router.back()
    }

    onMounted(() => {
      fetchTour()
    })

    return {
      form,
      tagInput,
      loading,
      saving,
      error,
      loadError,
      success,
      addTag,
      removeTag,
      handleSubmit,
      goBack
    }
  }
}
</script>

<style scoped>
.tour-edit {
  max-width: 800px;
  margin: 2rem auto;
  padding: 2rem;
}

h1 {
  color: #2c3e50;
  margin-bottom: 2rem;
}

.loading {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.error-message {
  padding: 1rem;
  background: #fee;
  color: #c33;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.success-message {
  padding: 1rem;
  background: #efe;
  color: #3c3;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.tour-form {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.form-group {
  margin-bottom: 1.5rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
  color: #333;
  font-weight: 600;
}

input[type="text"],
input[type="number"],
textarea,
select {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input:focus,
textarea:focus,
select:focus {
  outline: none;
  border-color: #42b983;
}

textarea {
  resize: vertical;
  font-family: inherit;
}

.tags-input {
  margin-top: 0.5rem;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.tag {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.4rem 0.8rem;
  background: #e8f5e9;
  color: #2e7d32;
  border-radius: 20px;
  font-size: 0.9rem;
}

.remove-tag {
  background: transparent;
  border: none;
  color: #2e7d32;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-tag:hover {
  color: #c33;
}

.info-section {
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
  padding: 1.5rem;
  border-radius: 12px;
  border: 2px solid #d4dce6;
}

.info-section label {
  color: #2c3e50;
  font-size: 1.1rem;
  margin-bottom: 0.25rem;
}

.durations-display {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.duration-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 2px solid #e0e6ed;
  transition: transform 0.2s, box-shadow 0.2s;
}

.duration-badge:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.duration-icon {
  font-size: 2rem;
  margin-bottom: 0.25rem;
}

.duration-label {
  font-weight: 600;
  color: #555;
  font-size: 0.95rem;
}

.duration-value {
  color: #42b983;
  font-size: 0.9rem;
  font-weight: 500;
}

.field-note {
  font-size: 0.85rem;
  color: #666;
  margin: 0.25rem 0 0 0;
  font-style: italic;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.form-actions button {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
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
</style>
