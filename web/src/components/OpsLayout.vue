<template>
  <div class="ops-layout">
    <!-- 左侧导航 -->
    <aside class="ops-sidebar">
      <div class="sidebar-header">
        <router-link to="/" class="back-link">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="19" y1="12" x2="5" y2="12"/>
            <polyline points="12 19 5 12 12 5"/>
          </svg>
          <span>返回商城</span>
        </router-link>
      </div>
      
      <div class="sidebar-brand">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="7" height="7"/>
          <rect x="14" y="3" width="7" height="7"/>
          <rect x="14" y="14" width="7" height="7"/>
          <rect x="3" y="14" width="7" height="7"/>
        </svg>
        <span>运营中心</span>
      </div>

      <nav class="sidebar-nav">
        <!-- 商城管理 -->
        <div class="nav-group">
          <div class="nav-group-title" @click="toggleGroup('mall')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="9" cy="21" r="1"/>
              <circle cx="20" cy="21" r="1"/>
              <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/>
            </svg>
            <span>商城管理</span>
            <svg class="arrow" :class="{ expanded: expandedGroups.mall }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </div>
          <div class="nav-group-items" v-show="expandedGroups.mall">
            <router-link to="/ops/skus" class="nav-item" active-class="active">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
              </svg>
              <span>商品管理</span>
            </router-link>
          </div>
        </div>

        <!-- 数据中心 -->
        <div class="nav-group">
          <div class="nav-group-title" @click="toggleGroup('data')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="20" x2="18" y2="10"/>
              <line x1="12" y1="20" x2="12" y2="4"/>
              <line x1="6" y1="20" x2="6" y2="14"/>
            </svg>
            <span>数据中心</span>
            <svg class="arrow" :class="{ expanded: expandedGroups.data }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </div>
          <div class="nav-group-items" v-show="expandedGroups.data">
            <div class="nav-item empty">
              <span>暂无内容</span>
            </div>
          </div>
        </div>
      </nav>

      <!-- 用户信息 -->
      <div class="sidebar-footer">
        <div class="user-info">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
          <span>{{ userName }}</span>
        </div>
        <button class="logout-btn" @click="handleLogout">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
            <polyline points="16 17 21 12 16 7"/>
            <line x1="21" y1="12" x2="9" y2="12"/>
          </svg>
        </button>
      </div>
    </aside>

    <!-- 主内容区 -->
    <main class="ops-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const userName = computed(() => userStore.userName)

const expandedGroups = reactive({
  mall: true,
  data: false
})

function toggleGroup(group) {
  expandedGroups[group] = !expandedGroups[group]
}

function handleLogout() {
  userStore.logout()
  localStorage.removeItem('cart_selected')
  router.push('/login')
}
</script>

<style scoped>
.ops-layout {
  display: flex;
  min-height: 100vh;
}

.ops-sidebar {
  width: 260px;
  background: linear-gradient(180deg, #0f172a 0%, #1e293b 100%);
  color: white;
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
}

.sidebar-header {
  padding: 18px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.back-link {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 13px;
  font-weight: 500;
  transition: color 0.2s;
}

.back-link:hover {
  color: rgba(255, 255, 255, 0.9);
}

.back-link svg {
  width: 16px;
  height: 16px;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 28px 24px 24px;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.sidebar-brand svg {
  width: 24px;
  height: 24px;
  opacity: 0.9;
}

.sidebar-nav {
  flex: 1;
  padding: 8px 12px;
  overflow-y: auto;
}

.nav-group {
  margin-bottom: 4px;
}

.nav-group-title {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.7);
  font-size: 13px;
  font-weight: 600;
  transition: all 0.15s;
  border-radius: 8px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.nav-group-title:hover {
  background-color: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.9);
}

.nav-group-title svg {
  width: 18px;
  height: 18px;
}

.nav-group-title .arrow {
  width: 16px;
  height: 16px;
  margin-left: auto;
  transition: transform 0.2s;
  opacity: 0.5;
}

.nav-group-title .arrow.expanded {
  transform: rotate(180deg);
}

.nav-group-items {
  padding-left: 8px;
  margin-top: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.15s;
  border-radius: 8px;
  margin-bottom: 2px;
}

.nav-item:hover {
  color: white;
  background-color: rgba(255, 255, 255, 0.08);
}

.nav-item.active {
  color: white;
  background-color: rgba(99, 102, 241, 0.2);
  font-weight: 600;
}

.nav-item svg {
  width: 18px;
  height: 18px;
}

.nav-item.empty {
  color: rgba(255, 255, 255, 0.25);
  font-size: 13px;
  cursor: default;
}

.nav-item.empty:hover {
  background-color: transparent;
}

.sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  font-weight: 500;
}

.user-info svg {
  width: 20px;
  height: 20px;
}

.logout-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.5);
  border-radius: 8px;
  transition: all 0.2s;
}

.logout-btn:hover {
  background-color: rgba(239, 68, 68, 0.15);
  color: #fca5a5;
}

.logout-btn svg {
  width: 16px;
  height: 16px;
}

.ops-main {
  flex: 1;
  margin-left: 260px;
  background-color: #f0f2f5;
  min-height: 100vh;
}
</style>
