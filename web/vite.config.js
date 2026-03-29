import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  // 与 src/api/index.js 一致：仅开发模式用直连 /marketplace、/ops；勿依赖 VITE_ENV
  const isViteDev = mode === 'development'

  const proxy = isViteDev
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
