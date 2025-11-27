<template>
  <div class="blog-edit">
    <h1>Edit Blog</h1>

    <div v-if="loading" class="loading">Loading blog...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <form v-else @submit.prevent="handleSubmit" class="blog-form">
      <div class="form-group">
        <label for="title">Title</label>
        <input 
          id="title"
          v-model="form.title"
          type="text"
          required
          placeholder="Enter blog title"
        />
      </div>

      <div class="form-group">
        <label for="description">Description (Markdown supported)</label>
        <textarea 
          id="description"
          v-model="form.description"
          required
          rows="15"
          placeholder="Write your blog content here. You can use markdown formatting."
        ></textarea>
      </div>

      <div class="form-group">
        <label for="images">Image URLs (one per line)</label>
        <textarea 
          id="images"
          v-model="imageUrls"
          rows="4"
          placeholder="https://example.com/image1.jpg&#10;https://example.com/image2.jpg"
        ></textarea>
        <small>Optional: Add URLs to images you want to include in your blog</small>
      </div>

      <div class="form-actions">
        <button type="button" @click="$router.back()" class="cancel-btn">Cancel</button>
        <button type="submit" :disabled="submitting" class="submit-btn">
          {{ submitting ? 'Updating...' : 'Update Blog' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const route = useRoute()

const blogId = route.params.id
const form = ref({
  title: '',
  description: ''
})
const imageUrls = ref('')
const loading = ref(true)
const submitting = ref(false)
const error = ref(null)

onMounted(async () => {
  await loadBlog()
})

async function loadBlog() {
  loading.value = true
  error.value = null
  try {
    const blog = await api.getBlogById(blogId)
    form.value.title = blog.title
    form.value.description = blog.description
    if (blog.images && blog.images.length > 0) {
      imageUrls.value = blog.images.join('\n')
    }
  } catch (err) {
    error.value = 'Failed to load blog: ' + (err.message || '')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  submitting.value = true
  error.value = null
  
  try {
    const images = imageUrls.value
      .split('\n')
      .map(url => url.trim())
      .filter(url => url.length > 0)
    
    const blogData = {
      title: form.value.title,
      description: form.value.description,
      images: images
    }
    
    await api.updateBlog(blogId, blogData)
    router.push(`/blogs/${blogId}`)
  } catch (err) {
    console.error('Blog update error:', err)
    error.value = err.response?.data?.error || err.response?.data || err.message || 'Failed to update blog'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.blog-edit {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

h1 {
  margin-bottom: 30px;
  color: #2c3e50;
}

.loading, .error {
  text-align: center;
  padding: 40px;
}

.error {
  color: #d32f2f;
  background: #ffebee;
  border-radius: 5px;
}

.blog-form {
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 25px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
}

input, textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-size: 16px;
  font-family: inherit;
}

input:focus, textarea:focus {
  outline: none;
  border-color: #42b983;
}

textarea {
  resize: vertical;
}

small {
  display: block;
  margin-top: 5px;
  color: #666;
  font-size: 14px;
}

.form-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 30px;
}

.cancel-btn, .submit-btn {
  padding: 12px 24px;
  border: none;
  border-radius: 5px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f5f5f5;
  color: #333;
}

.cancel-btn:hover {
  background-color: #e0e0e0;
}

.submit-btn {
  background-color: #42b983;
  color: white;
}

.submit-btn:hover:not(:disabled) {
  background-color: #35a372;
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
