import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 10301,
    proxy: {
      '/marketplace': {
        target: 'http://localhost:10302',
        changeOrigin: true
      },
      '/ops': {
        target: 'http://localhost:10302',
        changeOrigin: true,
        bypass(req) {
          if (req.headers.accept?.includes('text/html')) {
            return '/index.html'
          }
        }
      }
    }
  }
})
