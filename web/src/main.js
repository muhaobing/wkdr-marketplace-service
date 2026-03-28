import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useUserStore } from './stores/user.js'
import './style.css'

async function bootstrap() {
  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)
  await useUserStore().hydrateBindingsIfNeeded()
  app.use(router)
  app.mount('#app')
}

bootstrap()
