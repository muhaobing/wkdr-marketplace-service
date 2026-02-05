<template>
  <nav class="navbar">
    <div class="navbar-container">
      <!-- 左侧：平台名称 -->
      <router-link to="/" class="navbar-brand">
        <svg class="brand-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
          <polyline points="9 22 9 12 15 12 15 22"/>
        </svg>
        <span>商城中心</span>
      </router-link>

      <!-- 右侧：用户信息 -->
      <div class="navbar-right">
        <!-- 购物车 -->
        <router-link to="/cart" class="nav-item cart-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="9" cy="21" r="1"/>
            <circle cx="20" cy="21" r="1"/>
            <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
          </svg>
          <span>购物车</span>
          <span v-if="cartCount > 0" class="badge">{{ cartCount }}</span>
        </router-link>

        <!-- 订单中心 -->
        <router-link to="/orders" class="nav-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <polyline points="10 9 9 9 8 9"/>
          </svg>
          <span>订单中心</span>
        </router-link>

        <!-- 用户积分 -->
        <div class="nav-item ecoin-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 6v12M8 10h8M8 14h8"/>
          </svg>
          <span>积分: {{ balance.toFixed(2) }}</span>
        </div>

        <!-- 用户ID -->
        <div class="nav-item user-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
          <span>{{ userName }} ({{ userId }})</span>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useCartStore } from '../stores/cart'
import { useUserStore } from '../stores/user'

const cartStore = useCartStore()
const userStore = useUserStore()

const cartCount = computed(() => cartStore.count)
const userId = computed(() => userStore.userId)
const userName = computed(() => userStore.userName)
const balance = computed(() => userStore.balance)

onMounted(() => {
  userStore.fetchEcoin()
})
</script>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background-color: var(--white);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.navbar-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.navbar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
  font-weight: 600;
  color: var(--primary-color);
}

.brand-icon {
  width: 28px;
  height: 28px;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--gray-600);
  transition: color 0.2s;
  position: relative;
}

.nav-item:hover {
  color: var(--primary-color);
}

.nav-item svg {
  width: 20px;
  height: 20px;
}

.cart-item {
  position: relative;
}

.badge {
  position: absolute;
  top: -8px;
  right: -12px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background-color: var(--danger);
  color: var(--white);
  font-size: 12px;
  font-weight: 500;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ecoin-item {
  color: var(--warning);
  font-weight: 500;
}

.user-item {
  padding: 6px 12px;
  background-color: var(--gray-100);
  border-radius: 20px;
}

@media (max-width: 768px) {
  .navbar-right {
    gap: 12px;
  }

  .nav-item span {
    display: none;
  }

  .badge {
    right: -8px;
  }

  .user-item span {
    display: block;
  }
}
</style>
