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

        <!-- 用户积分（点击跳转到积分中心） -->
        <router-link to="/ecoin" class="nav-item ecoin-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 6v12M8 10h8M8 14h8"/>
          </svg>
          <span>积分: {{ balance.toFixed(2) }}</span>
        </router-link>

        <!-- 运营中心（仅管理员可见） -->
        <router-link v-if="isAdmin" to="/ops" class="nav-item ops-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="3" width="7" height="7"/>
            <rect x="14" y="3" width="7" height="7"/>
            <rect x="14" y="14" width="7" height="7"/>
            <rect x="3" y="14" width="7" height="7"/>
          </svg>
          <span>运营中心</span>
        </router-link>

        <!-- 用户信息下拉菜单 -->
        <div class="nav-item user-item" @click="toggleDropdown">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
          <span>{{ userName }} ({{ userId }})</span>
          <svg class="dropdown-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
          
          <!-- 下拉菜单 -->
          <div v-if="showDropdown" class="dropdown-menu" @click.stop>
            <router-link to="/ecoin" class="dropdown-item ecoin-center-btn" @click="showDropdown = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 6v12M8 10h8M8 14h8"/>
              </svg>
              <span>积分中心</span>
            </router-link>
            <button @click="handleLogout" class="dropdown-item logout-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                <polyline points="16 17 21 12 16 7"/>
                <line x1="21" y1="12" x2="9" y2="12"/>
              </svg>
              <span>退出登录</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCartStore } from '../stores/cart'
import { useUserStore } from '../stores/user'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

const showDropdown = ref(false)

const cartCount = computed(() => cartStore.count)
const userId = computed(() => userStore.userId)
const userName = computed(() => userStore.userName)
const balance = computed(() => userStore.balance)
const isAdmin = computed(() => userStore.isAdmin)

function toggleDropdown() {
  showDropdown.value = !showDropdown.value
}

function handleLogout() {
  userStore.logout()
  // 清空购物车本地数据
  localStorage.removeItem('cart_selected')
  showDropdown.value = false
  router.push('/login')
}

// 点击外部关闭下拉菜单
function handleClickOutside(event) {
  if (!event.target.closest('.user-item')) {
    showDropdown.value = false
  }
}

onMounted(() => {
  userStore.fetchEcoin()
  cartStore.init()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
  width: 100%;
  padding: 0 32px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
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

.ops-item {
  padding: 6px 12px;
  background-color: var(--primary-color);
  color: white !important;
  border-radius: 6px;
  font-weight: 500;
}

.ops-item:hover {
  background-color: #152a47;
  color: white !important;
}

.user-item {
  padding: 6px 12px;
  background-color: var(--gray-100);
  border-radius: 20px;
  cursor: pointer;
  position: relative;
}

.user-item:hover {
  background-color: var(--gray-200);
}

.dropdown-arrow {
  width: 16px;
  height: 16px;
  transition: transform 0.2s;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  min-width: 160px;
  padding: 8px 0;
  z-index: 1001;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: transparent;
  font-size: 14px;
  color: var(--gray-600);
  cursor: pointer;
  transition: all 0.2s;
}

.dropdown-item:hover {
  background-color: var(--gray-100);
}

.dropdown-item svg {
  width: 18px;
  height: 18px;
}

.ecoin-center-btn {
  color: var(--warning);
  text-decoration: none;
}

.ecoin-center-btn:hover {
  background-color: #fffaf0;
}

.logout-btn {
  color: var(--danger);
}

.logout-btn:hover {
  background-color: #fff5f5;
}

@media (max-width: 768px) {
  .navbar-container {
    padding: 0 16px;
  }

  .navbar-brand span {
    display: none;
  }

  .navbar-right {
    gap: 12px;
  }

  .nav-item span {
    display: none;
  }

  .badge {
    right: -8px;
  }

  .dropdown-arrow {
    display: none;
  }
}
</style>
