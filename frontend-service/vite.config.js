import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 8087, // 8085 promeniti za doker
     proxy: {
      '/api': {
        target: 'http://localhost:8080',  // gateway lokalno
        changeOrigin: true,
        rewrite: path => path.replace(/^\/api/, '') // ako gateway ne očekuje /api prefiks
      }
    } // obrisati za docker deploy
  }
})