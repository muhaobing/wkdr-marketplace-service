import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const isLocal = env.VITE_ENV === 'local'

  const proxy = isLocal
    ? {
        '/marketplace': { target: 'http://localhost:10302', changeOrigin: true },
        '/ops': { target: 'http://localhost:10302', changeOrigin: true }
      }
    : {
        '/market/api': {
          target: 'http://localhost:10302',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/market\/api/, '') || '/'
        }
      }

  return {
    base: '/market/',
    plugins: [vue()],
    server: {
      port: 10301,
      allowedHosts: ['www.lawmind.top', 'lawmind.top'],
      proxy
    },
    preview: {
      port: 10301,
      proxy
    }
  }
})
