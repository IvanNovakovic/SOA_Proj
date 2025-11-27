import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import { authService } from '../services/auth'
import SeeCart from '../views/SeeCart.vue'
import PurchasedTours from '../views/PurchasedTours.vue'
import PurchasedTourDetail from '../views/PurchasedTourDetail.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { requiresGuest: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('../views/Profile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/profile/edit',
    name: 'EditProfile',
    component: () => import('../views/EditProfile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: () => import('../views/AdminUsers.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/tours',
    name: 'Tours',
    component: () => import('../views/Tours.vue')
  },
  {
    path: '/tours/my-tours',
    name: 'MyTours',
    component: () => import('../views/MyTours.vue'),
    meta: { requiresAuth: true, requiresGuide: true }
  },
  {
    path: '/tours/create',
    name: 'TourCreate',
    component: () => import('../views/TourCreate.vue'),
    meta: { requiresAuth: true, requiresGuide: true }
  },
  {
    path: '/tours/:id/edit',
    name: 'TourEdit',
    component: () => import('../views/TourEdit.vue'),
    meta: { requiresAuth: true, requiresGuide: true }
  },
  {
    path: '/tours/:id',
    name: 'TourDetail',
    component: () => import('../views/TourDetail.vue')
  },
  {
    path: '/tours/:id/map',
    name: 'TourMapView',
    component: () => import('../views/TourMapView.vue')
  },
  {
    path: '/position-simulator',
    name: 'PositionSimulator',
    component: () => import('../views/PositionSimulator.vue'),
    meta: { requiresAuth: true, requiresTourist: true }
  },
  {
    path: '/tours/:tourId/keypoints/create',
    name: 'KeyPointCreate',
    component: () => import('../views/KeyPointCreate.vue'),
    meta: { requiresAuth: true, requiresGuide: true }
  },
  {
    path: '/tours/:tourId/keypoints/manage',
    name: 'KeyPointManage',
    component: () => import('../views/KeyPointManage.vue'),
    meta: { requiresAuth: true, requiresGuide: true }
  },
  {
    path: '/blogs',
    name: 'Blogs',
    component: () => import('../views/Blogs.vue'),
    children: [
      {
        path: '',
        name: 'BlogList',
        component: () => import('../views/BlogList.vue')
      },
      {
        path: 'my',
        name: 'MyBlogs',
        component: () => import('../views/MyBlogs.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: 'create',
        name: 'BlogCreate',
        component: () => import('../views/BlogCreate.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: ':id/edit',
        name: 'BlogEdit',
        component: () => import('../views/BlogEdit.vue'),
        meta: { requiresAuth: true }
      },
      {
        path: ':id',
        name: 'BlogDetail',
        component: () => import('../views/BlogDetail.vue'),
        meta: { requiresAuth: true }
      }
    ]
  },
  {
    path: '/user/:username',
    name: 'UserProfile',
    component: () => import('../views/UserProfile.vue')
  },
  {
    path: '/recommendations',
    name: 'FollowRecommendations',
    component: () => import('../views/FollowRecommendations.vue'),
    meta: { requiresAuth: true }
  },
  { path: '/cart', 
    name: 'Cart', 
    component: SeeCart 
  },
  {
    path: '/purchased-tours',
    name: 'PurchasedTours',
    component: PurchasedTours
  },
  {
    path: '/purchased-tour-detail/:id',
    name: 'PurchasedTourDetail',
    component: PurchasedTourDetail,
    props: true 
  },

]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guards
router.beforeEach((to, from, next) => {
  const isAuthenticated = authService.isAuthenticated()
  
  // Check if user is admin
  const user = authService.getUserFromToken()
  const isAdmin = user && user.roles && user.roles.includes('admin')
  const isGuide = user && user.roles && (user.roles.includes('guide') || user.roles.includes('Guide'))

  // Redirect authenticated users away from login/register
  if (to.meta.requiresGuest && isAuthenticated) {
    next('/')
  }
  // Redirect unauthenticated users to login for protected routes
  else if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  }
  // Redirect non-admin users from admin routes
  else if (to.meta.requiresAdmin && !isAdmin) {
    next('/')
  }
  // Redirect non-guide users from guide routes
  else if (to.meta.requiresGuide && !isGuide) {
    next('/tours')
  }
  else {
    next()
  }
})

export default router
