<template>
  <nav class="navbar">
    <div class="navbar-container">
      <!-- 左侧：平台名称 -->
      <router-link to="/" class="navbar-brand">
        <div class="brand-logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="9" cy="21" r="1"/>
            <circle cx="20" cy="21" r="1"/>
            <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
          </svg>
        </div>
        <span class="brand-text">商城中心</span>
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
        <router-link to="/ecoin" class="nav-item ecoin-item">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 6v12M8 10h8M8 14h8"/>
          </svg>
          <span>{{ balance.toFixed(2) }}</span>
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
          <div class="user-avatar">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          </div>
          <span>{{ userName }}</span>
          <svg class="dropdown-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
          
          <!-- 下拉菜单 -->
          <div v-if="showDropdown" class="dropdown-menu" @click.stop>
            <div class="dropdown-user-info">
              <div class="dropdown-avatar">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  <circle cx="12" cy="7" r="4"/>
                </svg>
              </div>
              <div>
                <div class="dropdown-name">{{ userName }}</div>
                <div class="dropdown-id">ID: {{ userId }}</div>
              </div>
            </div>
            <div class="dropdown-divider"></div>
            <router-link to="/ecoin" class="dropdown-item" @click="showDropdown = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 6v12M8 10h8M8 14h8"/>
              </svg>
              <span>积分中心</span>
            </router-link>
            <div class="dropdown-divider"></div>
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
  localStorage.removeItem('cart_selected')
  showDropdown.value = false
  router.push('/login')
}

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
  height: 64px;
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  z-index: 1000;
}

.navbar-container {
  width: 100%;
  padding: 0 40px;
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
  font-size: 18px;
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.02em;
}

.brand-logo {
  width: 34px;
  height: 34px;
  background: linear-gradient(135deg, var(--primary-color) 0%, #475569 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-logo svg {
  width: 18px;
  height: 18px;
  color: white;
}

.brand-text {
  font-weight: 700;
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--gray-500);
  padding: 7px 14px;
  border-radius: 8px;
  transition: all 0.2s;
  position: relative;
}

.nav-item:hover {
  color: var(--gray-800);
  background-color: var(--gray-100);
}

.nav-item svg {
  width: 18px;
  height: 18px;
}

.cart-item {
  position: relative;
}

.badge {
  position: absolute;
  top: 0;
  right: 4px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: var(--white);
  font-size: 11px;
  font-weight: 600;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ecoin-item {
  color: var(--warning);
  font-weight: 600;
}

.ecoin-item:hover {
  background-color: #fffbeb;
  color: #b45309;
}

.ops-item {
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-light) 100%);
  color: white !important;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.ops-item:hover {
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
  transform: translateY(-1px);
  background: linear-gradient(135deg, #4f46e5 0%, #6366f1 100%);
  color: white !important;
}

.user-item {
  cursor: pointer;
  position: relative;
  gap: 8px;
}

.user-avatar {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, var(--gray-200) 0%, var(--gray-300) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-avatar svg {
  width: 16px;
  height: 16px;
  color: var(--gray-500);
}

.dropdown-arrow {
  width: 14px !important;
  height: 14px !important;
  transition: transform 0.2s;
  color: var(--gray-400);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: white;
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.12), 0 0 0 1px rgba(0, 0, 0, 0.04);
  min-width: 220px;
  padding: 8px;
  z-index: 1001;
  animation: dropdownIn 0.15s ease-out;
}

@keyframes dropdownIn {
  from { opacity: 0; transform: translateY(-4px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.dropdown-user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
}

.dropdown-avatar {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--accent) 0%, var(--accent-light) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.dropdown-avatar svg {
  width: 18px;
  height: 18px;
  color: white;
}

.dropdown-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--gray-800);
}

.dropdown-id {
  font-size: 12px;
  color: var(--gray-400);
  margin-top: 1px;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--gray-100);
  margin: 4px 8px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  background: transparent;
  font-size: 14px;
  color: var(--gray-600);
  cursor: pointer;
  transition: all 0.15s;
  border-radius: 8px;
  text-decoration: none;
}

.dropdown-item:hover {
  background-color: var(--gray-50);
  color: var(--gray-800);
}

.dropdown-item svg {
  width: 18px;
  height: 18px;
}

.logout-btn:hover {
  background-color: #fef2f2;
  color: var(--danger);
}

@media (max-width: 768px) {
  .navbar-container {
    padding: 0 16px;
  }

  .brand-text {
    display: none;
  }

  .navbar-right {
    gap: 2px;
  }

  .nav-item {
    padding: 7px 8px;
  }

  .nav-item span {
    display: none;
  }

  .badge {
    right: 0;
  }

  .dropdown-arrow {
    display: none;
  }
}
</style>
