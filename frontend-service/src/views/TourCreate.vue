<template>
  <div class="tour-create">
    <div v-if="!isGuide" class="access-denied">
      <h2>⛔ Access Denied</h2>
      <p>Only guides can create tours. Please contact an administrator if you believe this is an error.</p>
      <router-link to="/tours" class="btn-back">← Back to Tours</router-link>
    </div>

    <template v-else>
      <h1>Create New Tour</h1>
      
      <div v-if="error" class="error-message">{{ error }}</div>
      <div v-if="success" class="success-message">Tour created successfully!</div>

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
        <p class="field-note">⏱️ Durations are automatically calculated when you add key points to your tour</p>
        <div class="durations-display">
          <div class="duration-badge">
            <span class="duration-icon">🚶</span>
            <span class="duration-label">Walking</span>
            <span class="duration-value">Auto-calculated</span>
          </div>
          <div class="duration-badge">
            <span class="duration-icon">🚴</span>
            <span class="duration-label">Cycling</span>
            <span class="duration-value">Auto-calculated</span>
          </div>
          <div class="duration-badge">
            <span class="duration-icon">🚗</span>
            <span class="duration-label">Driving</span>
            <span class="duration-value">Auto-calculated</span>
          </div>
        </div>
      </div>

      <div class="form-actions">
        <button type="submit" :disabled="loading" class="btn-primary">
          {{ loading ? 'Creating...' : 'Create Tour (Draft)' }}
        </button>
        <button type="button" @click="goBack" class="btn-secondary">Cancel</button>
      </div>

      <p class="form-note">
        * Tours are created with "draft" status and price set to 0 by default
      </p>
      </form>
    </template>
  </div>
</template>

<script>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../services/api'
import { authStore } from '../stores/authStore'

export default {
  name: 'TourCreate',
  setup() {
    const router = useRouter()
    const isGuide = computed(() => authStore.isGuide())
    const form = ref({
      name: '',
      description: '',
      difficulty: '',
      tags: []
    })
    const tagInput = ref('')
    const loading = ref(false)
    const error = ref('')
    const success = ref(false)

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
      loading.value = true
      error.value = ''
      success.value = false

      try {
        const tourData = {
          name: form.value.name,
          description: form.value.description,
          difficulty: form.value.difficulty,
          tags: form.value.tags,
          status: 'draft'
          // durations will be auto-calculated when keypoints are added
        }

        const createdTour = await api.createTour(tourData)
        success.value = true
        
        setTimeout(() => {
          router.push('/tours/my-tours')
        }, 1500)
      } catch (err) {
        error.value = err.response?.data?.error || err.message || 'Failed to create tour'
      } finally {
        loading.value = false
      }
    }

    const goBack = () => {
      router.back()
    }

    return {
      isGuide,
      form,
      tagInput,
      loading,
      error,
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
.tour-create {
  max-width: 800px;
  margin: 2rem auto;
  padding: 2rem;
}

h1 {
  color: #2c3e50;
  margin-bottom: 2rem;
}

.error-message {
  background: #fee;
  color: #c33;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  border: 1px solid #fcc;
}

.success-message {
  background: #efe;
  color: #3c3;
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  border: 1px solid #cfc;
}

.tour-form {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.form-group input[type="text"],
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
  outline: none;
  border-color: #42b983;
}

.tags-input {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.tag {
  background: #42b983;
  color: white;
  padding: 0.4rem 0.8rem;
  border-radius: 20px;
  font-size: 0.9rem;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.remove-tag {
  background: none;
  border: none;
  color: white;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
}

.remove-tag:hover {
  background: rgba(255, 255, 255, 0.2);
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
  font-size: 0.85rem;
  font-weight: 500;
  font-style: italic;
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

.btn-primary,
.btn-secondary {
  padding: 0.75rem 2rem;
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

.form-note {
  margin-top: 1rem;
  color: #666;
  font-size: 0.9rem;
  font-style: italic;
}

.access-denied {
  text-align: center;
  padding: 3rem 2rem;
  background: #fff3cd;
  border: 2px solid #ffc107;
  border-radius: 12px;
  max-width: 600px;
  margin: 2rem auto;
}

.access-denied h2 {
  color: #856404;
  margin-bottom: 1rem;
}

.access-denied p {
  color: #856404;
  margin-bottom: 1.5rem;
  font-size: 1.1rem;
}

.btn-back {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  background: #42b983;
  color: white;
  text-decoration: none;
  border-radius: 8px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-back:hover {
  background: #35a372;
}
</style>
