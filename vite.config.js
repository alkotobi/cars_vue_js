import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// Get environment
const isProduction = process.env.NODE_ENV === 'production'

// Set API URL based on environment
const apiUrl = isProduction ? 'https://cars.merhab.com' : 'http://localhost:8000'

export default defineConfig({
  plugins: [vue(), vueDevTools()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: apiUrl,
        changeOrigin: true,
        secure: isProduction,
      },
      '/uploads': {
        target: apiUrl,
        changeOrigin: true,
        secure: isProduction,
      },
    },
  },
  base: '/',
  build: {
    rollupOptions: {
      output: {
        entryFileNames: `[name].[hash].js`,
        chunkFileNames: `[name].[hash].js`,
        assetFileNames: `[name].[hash].[ext]`,
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['element-plus'],
          utils: ['lodash'],
        },
      },
    },
    manifest: true,
    sourcemap: true,
    chunkSizeWarningLimit: 1000, // Increase warning limit to 1MB
  },
})
