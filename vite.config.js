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
        assetFileNames: (assetInfo) => {
          // Preserve exact filenames for logo.png, letter_head.png, and gml2.png
          const preservedAssets = ['logo.png', 'letter_head.png', 'gml2.png']
          const assetName = assetInfo.name || ''
          
          // Check if the asset name ends with any of the preserved asset names
          // This handles cases where the path might be included
          const matchesPreserved = preservedAssets.some(name => {
            return assetName.endsWith(name) || assetName.includes(`/${name}`) || assetName.includes(`\\${name}`)
          })
          
          if (matchesPreserved) {
            // Extract just the filename from the path
            const fileName = assetName.split('/').pop() || assetName.split('\\').pop() || assetName
            return fileName
          }
          
          // For all other assets, use the default hashed naming
          return `[name].[hash].[ext]`
        },
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['element-plus'],
          utils: [],
        },
      },
    },
    manifest: true,
    sourcemap: true,
    chunkSizeWarningLimit: 1000, // Increase warning limit to 1MB
  },
})
