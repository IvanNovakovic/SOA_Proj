<template>
  <div class="see-cart">
    <h2>Your Cart</h2>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>Loading cart...</p>
    </div>

    <div v-else-if="error" class="error-message">
      <span class="error-icon">⚠️</span>
      {{ error }}
    </div>

    <div v-else-if="cart.items.length === 0" class="empty-state">
      <p>No items in your cart.</p>
    </div>

    <ul v-else>
      <li v-for="item in cart.items" :key="item.id" class="cart-item">
        <span>Tour: {{ item.name }}</span>
        <span> Price: ${{ item.price.toFixed(2) }}</span>
        <button @click="removeItem(item.id)">Remove</button>
      </li>
    </ul>

    <div v-if="cart.items.length > 0" class="cart-total">
      <strong>Total: ${{ cart.total.toFixed(2) }}</strong>
      <button @click="checkoutCart">Checkout</button>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from "vue"
import { api } from "../services/api"
import { authStore } from "../stores/authStore"

export default {
  name: "SeeCart",
  setup() {
    const cart = ref({ items: [], total: 0 })
    const loading = ref(true)
    const error = ref("")

    const fetchCart = async () => {
      loading.value = true
      error.value = ""
      try {
        const userId = authStore.getUserId()
        const data = await api.getCart(userId)
        cart.value = data
      } catch (err) {
        console.error(err)
        error.value = err.response?.data?.detail || "Failed to load cart."
      } finally {
        loading.value = false
      }
    }

    const removeItem = async (itemId) => {
      try {
        const data = await api.removeCartItem(itemId) // samo itemId
        
        cart.value.items = data.items
        cart.value.total = data.total
      } catch (err) {
        console.error(err)
        error.value = err.response?.data?.detail || "Failed to remove item."
      }
    }

    const checkoutCart = async () => {
      try {
        const userId = authStore.getUserId()
        await api.checkoutCart(userId)
        alert("Checkout successful!")
        cart.value = { items: [], total: 0 }
        router.push("/purchased-tours")
      } catch (err) {
        console.error(err)
        error.value = err.response?.data?.detail 
      }
    }

    onMounted(fetchCart)

    return { cart, 
             loading, 
             error, 
             removeItem,
             checkoutCart }
  }
}
</script>

<style>
.see-cart {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 5rem 2rem;
  gap: 1rem;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #42b983;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading p {
  color: #666;
  font-size: 1.1rem;
}

.error-message {
  background: linear-gradient(135deg, #fee 0%, #fdd 100%);
  color: #c33;
  padding: 1.5rem;
  border-radius: 12px;
  text-align: center;
  border: 2px solid #fcc;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  font-size: 1.1rem;
}

.error-icon {
  font-size: 1.5rem;
}

.empty-state {
  text-align: center;
  padding: 5rem 2rem;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border-radius: 20px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.08);
  max-width: 600px;
  margin: 2rem auto;
}

.empty-state p {
  color: #666;
  font-size: 1.1rem;
  margin-bottom: 1rem;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  padding: 1rem 2.5rem;
  border-radius: 12px;
  text-decoration: none;
  font-weight: 600;
  font-size: 1.1rem;
  transition: all 0.3s;
  box-shadow: 0 4px 15px rgba(66, 185, 131, 0.3);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 25px rgba(66, 185, 131, 0.4);
}

.cart-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding: 1rem;
  border-radius: 12px;
  background: #f8f9fa;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  transition: all 0.3s;
}

.cart-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
}

.cart-item button {
  background: linear-gradient(135deg, #ff7e5f 0%, #feb47b 100%);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 0.5rem 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.cart-item button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(255, 126, 95, 0.4);
}

.cart-total {
  margin-top: 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.cart-total button {
  background: linear-gradient(135deg, #42b983 0%, #35a372 100%);
  color: white;
  border: none;
  border-radius: 10px;
  padding: 0.75rem 1.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.cart-total button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(66, 185, 131, 0.4);
}

@media (max-width: 768px) {
  .see-cart {
    padding: 1rem;
  }

  .cart-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .cart-total {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }
}
</style>

