<template>
  <div class="app">
    <Navbar v-if="showNavbar" />
    <main :class="['main-content', { 'no-navbar': !showNavbar }]">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'

const route = useRoute()

// 登录页面和运营页面不显示商城导航栏
const showNavbar = computed(() => {
  return route.name !== 'Login' && !route.path.startsWith('/ops')
})
</script>

<style scoped>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  padding-top: 70px;
  padding-bottom: 40px;
}

.main-content.no-navbar {
  padding-top: 0;
}
</style>
