<template>
  <div class="app">
    <Navbar v-if="showNavbar" />
    <main :class="['main-content', { 'no-navbar': !showNavbar }]">
      <router-view />
    </main>
    <ToastContainer />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import ToastContainer from './components/ToastContainer.vue'

const route = useRoute()

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
  padding-top: 80px;
  padding-bottom: 60px;
}

.main-content.no-navbar {
  padding-top: 0;
}
</style>
