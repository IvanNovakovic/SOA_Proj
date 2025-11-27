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
          <router-link to="/blogs" class="nav-link" @click="closeMenu">Blogs</router-link>
          <router-link to="/blogs/my" class="nav-link" @click="closeMenu">My Blogs</router-link>
          <router-link to="/recommendations" class="nav-link" @click="closeMenu">Discover</router-link>
          <router-link to="/profile" class="nav-link" @click="closeMenu">Profile</router-link>
          <router-link v-if="isAdmin" to="/admin/users" class="nav-link" @click="closeMenu">Admin</router-link>
        </div>
      </div>

      <div class="navbar-auth" :class="{ 'active': menuOpen }">
        <template v-if="!isLoggedIn">
          <router-link to="/login" class="auth-btn login-btn" @click="closeMenu">Login</router-link>
          <router-link to="/register" class="auth-btn register-btn" @click="closeMenu">Register</router-link>
        </template>
        <template v-else>
          <span class="user-name">{{ username }}</span>
          <button @click="logout" class="auth-btn logout-btn">Logout</button>
        </template>
      </div>
    </div>
  </nav>
</template>

<script>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authStore } from '../stores/authStore'
import { authService } from '../services/auth'

export default {
  name: 'Navbar',
  setup() {
    const router = useRouter()
    const menuOpen = ref(false)

    const isLoggedIn = authStore.isAuthenticated
    const username = authStore.username
    
    const isAdmin = computed(() => {
      const user = authService.getUserFromToken()
      return user && user.roles && user.roles.includes('admin')
    })

    const toggleMenu = () => {
      menuOpen.value = !menuOpen.value
    }

    const closeMenu = () => {
      menuOpen.value = false
    }

    const logout = () => {
      authStore.logout()
      closeMenu()
      // Force a full navigation to home
      router.replace('/')
    }

    return {
      menuOpen,
      isLoggedIn,
      username,
      isAdmin,
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
  gap: 2rem;
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
}

.nav-link:hover,
.nav-link.router-link-active {
  background: rgba(255, 255, 255, 0.2);
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
}
</style>
