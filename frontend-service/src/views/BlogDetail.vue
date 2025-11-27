<template>
  <div class="blog-detail">
    <div v-if="loading" class="loading">Loading blog...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="blog" class="blog-content">
      <!-- Blog Header -->
      <div class="blog-header">
        <h1>{{ blog.title }}</h1>
        <div class="meta">
          <span class="author">by {{ blog.author_name }}</span>
          <span class="date">{{ formatDate(blog.created_at) }}</span>
        </div>
      </div>

      <!-- Blog Images -->
      <div v-if="blog.images && blog.images.length > 0" class="images">
        <img v-for="(img, idx) in blog.images" :key="idx" :src="img" :alt="`Blog image ${idx + 1}`" />
      </div>

      <!-- Blog Description (Markdown) -->
      <div class="description" v-html="renderedDescription"></div>

      <!-- Like Section -->
      <div class="like-section">
        <button @click="toggleLike" :disabled="likeLoading" class="like-btn" :class="{ liked: isLiked }">
          {{ isLiked ? '❤️' : '🤍' }} {{ blog.likes_count || 0 }}
        </button>
      </div>

      <!-- Comments Section -->
      <div class="comments-section">
        <h2>Comments</h2>
        
        <!-- Comment Form -->
        <div v-if="canComment" class="comment-form">
          <textarea
            v-model="newComment"
            placeholder="Write a comment..."
            rows="3"
          ></textarea>
          <button @click="submitComment" :disabled="!newComment.trim() || commentLoading">
            {{ commentLoading ? 'Posting...' : 'Post Comment' }}
          </button>
        </div>
        <div v-else class="comment-disabled">
          You must follow {{ blog.author_name }} to comment on this blog.
        </div>

        <!-- Comments List -->
        <div v-if="commentsLoading" class="loading">Loading comments...</div>
        <div v-else-if="!comments || comments.length === 0" class="no-comments">No comments yet. Be the first!</div>
        <div v-else class="comments-list">
          <div v-for="comment in comments" :key="comment.id" class="comment">
            <div class="comment-header">
              <router-link :to="`/user/${comment.author_name}`" class="comment-author">
                {{ comment.author_name }}
              </router-link>
              <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
            </div>
            <p class="comment-text">{{ comment.text }}</p>
            <span v-if="comment.last_edited_at" class="edited">(edited)</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { marked } from 'marked'
import api from '../services/api'
import { authStore } from '../stores/authStore'

const route = useRoute()
const router = useRouter()
const blog = ref(null)
const comments = ref([])
const loading = ref(true)
const commentsLoading = ref(true)
const error = ref(null)
const isLiked = ref(false)
const likeLoading = ref(false)
const newComment = ref('')
const commentLoading = ref(false)
const isFollowing = ref(false)

const blogId = route.params.id

const renderedDescription = computed(() => {
  if (!blog.value?.description) return ''
  return marked(blog.value.description)
})

const canComment = computed(() => {
  // User can comment if they are the author or if they follow the author
  return blog.value?.author_id === authStore.getUserId() || isFollowing.value
})

onMounted(async () => {
  await fetchBlog()
  await fetchComments()
  await checkIfFollowing()
  await checkLikeStatus()
})

async function fetchBlog() {
  loading.value = true
  error.value = null
  try {
    blog.value = await api.getBlogById(blogId)
    // Check if user has liked this blog (you'd need to add this to backend response or make separate call)
    // For now, we'll assume backend doesn't return this info
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load blog'
    if (err.response?.status === 403) {
      error.value = 'You must follow the author to view this blog.'
    }
  } finally {
    loading.value = false
  }
}

async function fetchComments() {
  commentsLoading.value = true
  try {
    const result = await api.getComments(blogId)
    comments.value = result || []
  } catch (err) {
    console.error('Failed to load comments:', err)
    comments.value = []
  } finally {
    commentsLoading.value = false
  }
}

async function checkIfFollowing() {
  if (!blog.value || blog.value.author_id === authStore.getUserId()) {
    isFollowing.value = true
    return
  }
  
  try {
    const currentUserId = authStore.getUserId()
    isFollowing.value = await api.isFollowing(currentUserId, blog.value.author_id)
  } catch (err) {
    console.error('Failed to check following status:', err)
  }
}

async function checkLikeStatus() {
  try {
    const result = await api.checkLikeStatus(blogId)
    isLiked.value = result.liked || false
  } catch (err) {
    console.error('Failed to check like status:', err)
    isLiked.value = false
  }
}

async function toggleLike() {
  likeLoading.value = true
  try {
    if (isLiked.value) {
      await api.unlikeBlog(blogId)
      blog.value.likes_count = Math.max(0, (blog.value.likes_count || 0) - 1)
      isLiked.value = false
    } else {
      await api.likeBlog(blogId)
      blog.value.likes_count = (blog.value.likes_count || 0) + 1
      isLiked.value = true
    }
  } catch (err) {
    console.error('Failed to toggle like:', err)
  } finally {
    likeLoading.value = false
  }
}

async function submitComment() {
  if (!newComment.value.trim()) return
  
  commentLoading.value = true
  try {
    const comment = await api.createComment(blogId, { text: newComment.value })
    comments.value.push(comment)
    newComment.value = ''
  } catch (err) {
    alert(err.response?.data?.error || 'Failed to post comment')
  } finally {
    commentLoading.value = false
  }
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('en-US', { 
    year: 'numeric', 
    month: 'long', 
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.blog-detail {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.loading, .error {
  text-align: center;
  padding: 40px;
  color: #666;
}

.error {
  color: #d32f2f;
}

.blog-content {
  background: white;
  border-radius: 8px;
  padding: 30px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.blog-header h1 {
  margin: 0 0 15px 0;
  color: #2c3e50;
  font-size: 32px;
}

.meta {
  display: flex;
  gap: 20px;
  color: #666;
  font-size: 14px;
  margin-bottom: 30px;
  padding-bottom: 15px;
  border-bottom: 1px solid #eee;
}

.author {
  color: #42b983;
  font-weight: 500;
}

.images {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 15px;
  margin-bottom: 30px;
}

.images img {
  width: 100%;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.description {
  line-height: 1.8;
  color: #333;
  margin-bottom: 30px;
  font-size: 16px;
}

.description :deep(h1),
.description :deep(h2),
.description :deep(h3) {
  margin-top: 25px;
  margin-bottom: 15px;
  color: #2c3e50;
}

.description :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}

.description :deep(pre) {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 5px;
  overflow-x: auto;
}

.like-section {
  margin: 30px 0;
  padding: 20px 0;
  border-top: 1px solid #eee;
  border-bottom: 1px solid #eee;
}

.like-btn {
  background: white;
  border: 2px solid #ddd;
  padding: 10px 20px;
  border-radius: 25px;
  font-size: 18px;
  cursor: pointer;
  transition: all 0.2s;
}

.like-btn:hover:not(:disabled) {
  border-color: #ff6b6b;
  transform: scale(1.05);
}

.like-btn.liked {
  border-color: #ff6b6b;
  background: #fff5f5;
}

.like-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.comments-section {
  margin-top: 40px;
}

.comments-section h2 {
  margin-bottom: 20px;
  color: #2c3e50;
}

.comment-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 30px;
}

.comment-form textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 5px;
  font-family: inherit;
  resize: vertical;
}

.comment-form button {
  align-self: flex-end;
  background: #42b983;
  color: white;
  border: none;
  padding: 10px 25px;
  border-radius: 5px;
  cursor: pointer;
}

.comment-form button:hover:not(:disabled) {
  background: #35a372;
}

.comment-form button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.comment-disabled {
  padding: 15px;
  background: #fff3cd;
  border: 1px solid #ffc107;
  border-radius: 5px;
  color: #856404;
  margin-bottom: 30px;
}

.no-comments {
  text-align: center;
  color: #666;
  padding: 20px;
  font-style: italic;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.comment {
  background: #f9f9f9;
  padding: 15px;
  border-radius: 5px;
  border-left: 3px solid #42b983;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}

.comment-author {
  font-weight: 500;
  color: #2c3e50;
  text-decoration: none;
  transition: color 0.2s;
}

.comment-author:hover {
  color: #42b983;
  text-decoration: underline;
}

.comment-date {
  color: #999;
  font-size: 13px;
}

.comment-text {
  margin: 0;
  color: #333;
  line-height: 1.6;
}

.edited {
  font-size: 12px;
  color: #999;
  font-style: italic;
}
</style>
