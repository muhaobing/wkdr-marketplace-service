import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 10301,
    allowedHosts: ['www.lawmind.top'],
    // 与线上一致：/market/api 由 nginx 去掉前缀后转发；本地开发由 Vite 去掉前缀再转发到后端
    proxy: {
      '/market/api': {
        target: 'http://localhost:10302',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/market\/api/, '') || '/'
      }
    }
  },
  preview: {
    port: 10301,
    proxy: {
      '/market/api': {
        target: 'http://localhost:10302',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/market\/api/, '') || '/'
      }
    }
  }
})
