<template>
  <div class="my-blogs">
    <div class="header">
      <h1>My Blogs</h1>
      <button @click="$router.push('/blogs/create')" class="create-btn">Create New Blog</button>
    </div>

    <div v-if="loading" class="loading">Loading your blogs...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="blogs.length === 0" class="empty">
      <p>You haven't created any blogs yet.</p>
      <button @click="$router.push('/blogs/create')" class="create-btn">Create Your First Blog</button>
    </div>
    <div v-else class="blogs">
      <div v-for="blog in blogs" :key="blog.id" class="blog-card">
        <div class="blog-header">
          <h2 @click="goToBlog(blog.id)" class="blog-title">{{ blog.title }}</h2>
          <span class="date">{{ formatDate(blog.created_at) }}</span>
        </div>
        <p class="description">{{ getPreview(blog.description) }}</p>
        <div class="blog-footer">
          <span class="likes">❤️ {{ blog.likes_count || 0 }}</span>
          <div class="actions">
            <button @click="editBlog(blog.id)" class="edit-btn">Edit</button>
            <button @click="confirmDelete(blog)" class="delete-btn">Delete</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="closeDeleteModal">
      <div class="modal" @click.stop>
        <h3>Delete Blog?</h3>
        <p>Are you sure you want to delete "{{ blogToDelete?.title }}"? This action cannot be undone.</p>
        <div class="modal-actions">
          <button @click="closeDeleteModal" class="cancel-btn">Cancel</button>
          <button @click="deleteBlog" :disabled="deleting" class="confirm-delete-btn">
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const blogs = ref([])
const loading = ref(true)
const error = ref(null)
const showDeleteModal = ref(false)
const blogToDelete = ref(null)
const deleting = ref(false)

onMounted(async () => {
  await fetchBlogs()
})

async function fetchBlogs() {
  loading.value = true
  error.value = null
  try {
    const result = await api.getMyBlogs()
    blogs.value = result || []
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load your blogs'
    blogs.value = []
  } finally {
    loading.value = false
  }
}

function goToBlog(id) {
  router.push(`/blogs/${id}`)
}

function editBlog(id) {
  router.push(`/blogs/${id}/edit`)
}

function confirmDelete(blog) {
  blogToDelete.value = blog
  showDeleteModal.value = true
}

function closeDeleteModal() {
  showDeleteModal.value = false
  blogToDelete.value = null
}

async function deleteBlog() {
  if (!blogToDelete.value) return
  
  deleting.value = true
  try {
    await api.deleteBlog(blogToDelete.value.id)
    blogs.value = blogs.value.filter(b => b.id !== blogToDelete.value.id)
    closeDeleteModal()
  } catch (err) {
    error.value = 'Failed to delete blog: ' + (err.message || '')
  } finally {
    deleting.value = false
  }
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric' 
  })
}

function getPreview(description) {
  const plain = description.replace(/[#*_~`\[\]]/g, '')
  return plain.length > 200 ? plain.substring(0, 200) + '...' : plain
}
</script>

<style scoped>
.my-blogs {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.create-btn {
  background-color: #42b983;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
}

.create-btn:hover {
  background-color: #35a372;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px;
  color: #666;
}

.error {
  color: #d32f2f;
}

.empty p {
  margin-bottom: 20px;
}

.blogs {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.blog-card {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 20px;
  transition: all 0.2s;
}

.blog-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.blog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.blog-title {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
  cursor: pointer;
}

.blog-title:hover {
  color: #42b983;
}

.date {
  color: #666;
  font-size: 14px;
}

.description {
  color: #333;
  line-height: 1.6;
  margin: 15px 0;
}

.blog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

.likes {
  font-size: 16px;
  color: #666;
}

.actions {
  display: flex;
  gap: 10px;
}

.edit-btn, .delete-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
}

.edit-btn {
  background-color: #667eea;
  color: white;
}

.edit-btn:hover {
  background-color: #5568d3;
}

.delete-btn {
  background-color: #dc3545;
  color: white;
}

.delete-btn:hover {
  background-color: #c82333;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background: white;
  padding: 30px;
  border-radius: 8px;
  max-width: 500px;
  width: 90%;
}

.modal h3 {
  margin-top: 0;
  color: #333;
}

.modal p {
  color: #666;
  margin: 15px 0 25px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.cancel-btn {
  padding: 10px 20px;
  border: 1px solid #ddd;
  border-radius: 5px;
  background: white;
  cursor: pointer;
}

.cancel-btn:hover {
  background-color: #f5f5f5;
}

.confirm-delete-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 5px;
  background-color: #dc3545;
  color: white;
  cursor: pointer;
}

.confirm-delete-btn:hover:not(:disabled) {
  background-color: #c82333;
}

.confirm-delete-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
