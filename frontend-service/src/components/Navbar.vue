<template>
  <nav class="navbar">
    <div class="navbar-container">
      <div class="navbar-brand">
        <router-link to="/" class="brand-link">
          <h2>MyApp</h2>
        </router-link>
      </div>

      <div class="navbar-menu" v-if="isLoggedIn">
        <button class="menu-toggle" @click="toggleMenu">
          <span></span>
          <span></span>
          <span></span>
        </button>

        <div class="navbar-links" :class="{ 'active': menuOpen }">
          <router-link to="/" class="nav-link" @click="closeMenu">Home</router-link>
          <router-link to="/tours" class="nav-link" @click="closeMenu">Tours</router-link>
          
          <!-- Blogs Dropdown -->
          <div class="nav-dropdown" @mouseenter="showBlogsDropdown = true" @mouseleave="showBlogsDropdown = false">
            <button class="nav-link dropdown-toggle">
              Blogs <span class="arrow">▼</span>
            </button>
            <div class="dropdown-menu" :class="{ 'show': showBlogsDropdown }">
              <router-link to="/blogs" class="dropdown-item" @click="closeMenu">All Blogs</router-link>
              <router-link to="/blogs/my" class="dropdown-item" @click="closeMenu">My Blogs</router-link>
            </div>
          </div>

          <router-link to="/recommendations" class="nav-link" @click="closeMenu">Discover</router-link>
          
          <!-- More Dropdown -->
          <div class="nav-dropdown" @mouseenter="showMoreDropdown = true" @mouseleave="showMoreDropdown = false">
            <button class="nav-link dropdown-toggle">
              More <span class="arrow">▼</span>
            </button>
            <div class="dropdown-menu" :class="{ 'show': showMoreDropdown }">
              <router-link v-if="isTourist" to="/purchased-tours" class="dropdown-item" @click="closeMenu">Purchased Tours</router-link>
              <router-link to="/profile" class="dropdown-item" @click="closeMenu">Profile</router-link>
              <router-link v-if="isAdmin" to="/admin/users" class="dropdown-item" @click="closeMenu">Admin Panel</router-link>
            </div>
          </div>
        </div>
      </div>

      <div class="navbar-auth" :class="{ 'active': menuOpen }">
        <template v-if="!isLoggedIn">
          <router-link to="/login" class="auth-btn login-btn" @click="closeMenu">Login</router-link>
          <router-link to="/register" class="auth-btn register-btn" @click="closeMenu">Register</router-link>
        </template>
        <template v-else>
          <span class="user-name">{{ username }}</span>

          <!-- CART ICON -->
          <router-link v-if="isTourist" to="/cart" class="cart-icon" title="Shopping Cart">
            🛒
            <span v-if="cartCount > 0" class="cart-count">{{ cartCount }}</span>
          </router-link>

          <button @click="logout" class="auth-btn logout-btn">Logout</button>
        </template>
      </div>
    </div>
  </nav>
</template>

<script>
import { ref, computed, onMounted  } from 'vue'
import { useRouter } from 'vue-router'
import { authStore } from '../stores/authStore'
import { authService } from '../services/auth'

export default {
  name: 'Navbar',
  setup() {
    const router = useRouter()
    const menuOpen = ref(false)
    const showBlogsDropdown = ref(false)
    const showMoreDropdown = ref(false)

    const isLoggedIn = authStore.isAuthenticated
    const username = authStore.username
    
    const isAdmin = computed(() => {
      const user = authService.getUserFromToken()
      return user && user.roles && user.roles.includes('admin')
    })

    const isTourist = computed(() => {
      const user = authService.getUserFromToken()
      return user && user.roles && user.roles.includes('tourist')
    })

     const cartCount = ref(0)

    const fetchCartCount = async () => {
      try {
        const cart = await api.getCart() 
        cartCount.value = cart.items.length
      } catch (err) {
        cartCount.value = 0
      }
    }

    onMounted(() => {
      if (isLoggedIn) fetchCartCount()
    })

    const toggleMenu = () => {
      menuOpen.value = !menuOpen.value
    }

    const closeMenu = () => {
      menuOpen.value = false
      showBlogsDropdown.value = false
      showMoreDropdown.value = false
    }

    const logout = () => {
      authStore.logout()
      closeMenu()
      // Force a full navigation to home
      router.replace('/')
    }
    return {
      menuOpen,
      showBlogsDropdown,
      showMoreDropdown,
      isLoggedIn,
      username,
      isAdmin,
      isTourist,
      cartCount,
      fetchCartCount,
      toggleMenu,
      closeMenu,
      logout
    }
  }
}
</script>

<style scoped>
.navbar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 1000;
}

.navbar-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.navbar-brand .brand-link {
  text-decoration: none;
  color: white;
}

.navbar-brand h2 {
  margin: 0;
  font-size: 1.8rem;
  font-weight: 600;
  color: white;
}

.navbar-menu {
  display: flex;
  align-items: center;
}

.menu-toggle {
  display: none;
  flex-direction: column;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
}

.menu-toggle span {
  width: 25px;
  height: 3px;
  background: white;
  margin: 3px 0;
  transition: 0.3s;
  border-radius: 2px;
}

.navbar-links {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}

.nav-link {
  color: white;
  text-decoration: none;
  font-size: 1rem;
  font-weight: 500;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  transition: background 0.3s ease;
  background: none;
  border: none;
  cursor: pointer;
}

.nav-link:hover,
.nav-link.router-link-active {
  background: rgba(255, 255, 255, 0.2);
}

/* Dropdown Styles */
.nav-dropdown {
  position: relative;
}

.dropdown-toggle {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.arrow {
  font-size: 0.7rem;
  transition: transform 0.3s ease;
}

.nav-dropdown:hover .arrow {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 180px;
  opacity: 0;
  visibility: hidden;
  transform: translateY(-10px);
  transition: all 0.3s ease;
  margin-top: 0.5rem;
  overflow: hidden;
}

.dropdown-menu.show {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.dropdown-item {
  display: block;
  padding: 0.75rem 1.25rem;
  color: #333;
  text-decoration: none;
  transition: background 0.2s ease;
  font-weight: 500;
}

.dropdown-item:hover {
  background: #f0f0f0;
  color: #667eea;
}

.dropdown-item:first-child {
  padding-top: 1rem;
}

.dropdown-item:last-child {
  padding-bottom: 1rem;
}

.navbar-auth {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.user-name {
  color: white;
  font-weight: 500;
  margin-right: 0.5rem;
}

.auth-btn {
  padding: 0.6rem 1.5rem;
  border-radius: 6px;
  font-size: 0.95rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.3s ease;
  border: none;
  cursor: pointer;
  display: inline-block;
  text-align: center;
}

.login-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 2px solid white;
}

.login-btn:hover {
  background: white;
  color: #667eea;
}

.register-btn {
  background: white;
  color: #667eea;
}

.register-btn:hover {
  background: rgba(255, 255, 255, 0.9);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.logout-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 2px solid white;
}

.logout-btn:hover {
  background: #ff4757;
  border-color: #ff4757;
}

@media (max-width: 768px) {
  .navbar-container {
    flex-wrap: wrap;
    padding: 1rem;
  }

  .menu-toggle {
    display: flex;
    order: 3;
  }

  .navbar-brand {
    order: 1;
  }

  .navbar-auth {
    order: 2;
  }

  .navbar-links {
    display: none;
    width: 100%;
    flex-direction: column;
    order: 4;
    margin-top: 1rem;
    gap: 0;
  }

  .navbar-links.active {
    display: flex;
  }

  .navbar-auth.active {
    display: flex;
  }

  .nav-link {
    width: 100%;
    text-align: center;
    padding: 1rem;
  }

  /* Mobile Dropdown Styles */
  .nav-dropdown {
    width: 100%;
  }

  .dropdown-toggle {
    width: 100%;
    justify-content: center;
  }

  .dropdown-menu {
    position: static;
    opacity: 1;
    visibility: visible;
    transform: none;
    box-shadow: none;
    background: rgba(255, 255, 255, 0.1);
    margin: 0;
    display: none;
  }

  .nav-dropdown:hover .dropdown-menu,
  .dropdown-menu.show {
    display: block;
  }

  .dropdown-item {
    color: white;
    text-align: center;
  }

  .dropdown-item:hover {
    background: rgba(255, 255, 255, 0.15);
    color: white;
  }
}

.cart-icon {
  position: relative;
  font-size: 1.5rem;
  color: white;
  text-decoration: none;
  margin-right: 0.5rem;
}

.cart-count {
  position: absolute;
  top: -6px;
  right: -10px;
  background: red;
  color: white;
  border-radius: 50%;
  padding: 0 6px;
  font-size: 0.75rem;
  font-weight: bold;
}

</style>
