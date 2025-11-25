import axios from 'axios'

const API_BASE_URL = '/api'

// Create axios instance
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor to add JWT token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('jwt_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor to handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid
      localStorage.removeItem('jwt_token')
      localStorage.removeItem('username')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const api = {
  // Auth endpoints
  async login(credentials) {
    const response = await apiClient.post('/auth/login', credentials)
    return response.data
  },

  async register(userData) {
    const response = await apiClient.post('/auth/register', userData)
    return response.data
  },

  // User endpoints
  async getUsers() {
    const response = await apiClient.get('/users')
    return response.data
  },

  async getUserByUsername(username) {
    const response = await apiClient.get(`/users?username=${username}`)
    return response.data
  },

  async getUserById(userId) {
    const response = await apiClient.get(`/users/${userId}`)
    return response.data
  },

  async getCurrentUser() {
    const response = await apiClient.get('/users/me')
    return response.data
  },

  async updateCurrentUser(userData) {
    const response = await apiClient.patch('/users/me', userData)
    return response.data
  },

  async blockUser(userId) {
    const response = await apiClient.patch(`/users/${userId}/block`)
    return response.data
  },

  async unblockUser(userId) {
    const response = await apiClient.patch(`/users/${userId}/unblock`)
    return response.data
  },

  // Tours endpoints
  async getTours() {
    const response = await apiClient.get('/tours')
    return response.data
  },

  async getTourById(id) {
    const response = await apiClient.get(`/tours/${id}`)
    return response.data
  },

  async getToursByAuthor(authorId) {
    const response = await apiClient.get(`/tours/author/${authorId}`)
    return response.data
  },

  async createTour(tourData) {
    const response = await apiClient.post('/tours', tourData)
    return response.data
  },

  async updateTour(id, tourData) {
    const response = await apiClient.put(`/tours/${id}`, tourData)
    return response.data
  },

  async publishTour(id) {
    const response = await apiClient.post(`/tours/${id}/publish`)
    return response.data
  },

  async archiveTour(id) {
    const response = await apiClient.post(`/tours/${id}/archive`)
    return response.data
  },

  async activateTour(id) {
    const response = await apiClient.post(`/tours/${id}/activate`)
    return response.data
  },

  // KeyPoints endpoints
  async getKeyPoints(tourId) {
    const response = await apiClient.get(`/tours/${tourId}/keypoints`)
    return response.data
  },

  async createKeyPoint(tourId, keyPointData) {
    const response = await apiClient.post(`/tours/${tourId}/keypoints`, keyPointData)
    return response.data
  },

  async updateKeyPoint(keypointId, keyPointData) {
    const response = await apiClient.put(`/keypoints/${keypointId}`, keyPointData)
    return response.data
  },

  async deleteKeyPoint(keypointId) {
    const response = await apiClient.delete(`/keypoints/${keypointId}`)
    return response.data
  },

  async reorderKeyPoints(tourId, keypointIds) {
    const response = await apiClient.put(`/tours/${tourId}/keypoints/reorder`, { keypointIds })
    return response.data
  },

  // Reviews endpoints
  async getReviews(tourId) {
    const response = await apiClient.get(`/tours/${tourId}/reviews`)
    return response.data
  },

  async createReview(tourId, reviewData) {
    const response = await apiClient.post(`/tours/${tourId}/reviews`, reviewData)
    return response.data
  },

  // Blogs endpoints
  async getBlogs() {
    const response = await apiClient.get('/blogs')
    return response.data
  },

  async getBlogById(id) {
    const response = await apiClient.get(`/blogs/${id}`)
    return response.data
  },

  async createBlog(blogData) {
    const response = await apiClient.post('/blogs', blogData)
    return response.data
  },

  // Comments endpoints
  async getComments(blogId) {
    const response = await apiClient.get(`/blogs/${blogId}/comments`)
    return response.data
  },

  async createComment(blogId, commentData) {
    const response = await apiClient.post(`/blogs/${blogId}/comments`, commentData)
    return response.data
  },

  async updateComment(blogId, commentId, commentData) {
    const response = await apiClient.patch(`/blogs/${blogId}/comments/${commentId}`, commentData)
    return response.data
  },

  // Likes endpoints
  async likeBlog(blogId) {
    const response = await apiClient.post(`/blogs/${blogId}/likes`, {})
    return response.data
  },

  async unlikeBlog(blogId) {
    const response = await apiClient.delete(`/blogs/${blogId}/likes`, { data: {} })
    return response.data
  },

  async checkLikeStatus(blogId) {
    const response = await apiClient.get(`/blogs/${blogId}/likes/check`)
    return response.data
  },

  // Followers endpoints
  async follow(userId) {
    const response = await apiClient.post(`/follow`, { followee: userId })
    return response.data
  },

  async unfollow(userId) {
    const response = await apiClient.delete(`/follow`, { data: { followee: userId } })
    return response.data
  },

  async getFollowers(userId) {
    const response = await apiClient.get(`/followers/${userId}`)
    return response.data
  },

  async getFollowing(userId) {
    const response = await apiClient.get(`/following/${userId}`)
    return response.data
  },

  async getRecommendations(userId, limit = 20) {
    const response = await apiClient.get(`/recommendations/${userId}?limit=${limit}`)
    return response.data
  },

  async isFollowing(currentUserId, targetUserId) {
    try {
      const following = await this.getFollowing(currentUserId)
      return following.includes(targetUserId)
    } catch {
      return false
    }
  }
}

export default api
