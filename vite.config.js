import { existsSync, rmSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Build ID so every build produces different filenames (avoids stale cache after deploy)
const buildId = Date.now().toString(36)

// Plugin to remove vendor chunk preload links
const removeVendorPreload = () => {
  return {
    name: 'remove-vendor-preload',
    transformIndexHtml(html) {
      // Remove preload/modulepreload links for vendor chunks
      // This prevents browser warnings about unused preloaded resources
      return html.replace(
        /<link[^>]*rel=["'](modulepreload|preload)["'][^>]*vendor[^>]*>/gi,
        ''
      )
    },
  }
}

// db_code.json is PER SERVER, not per build: it names the database a deployment
// talks to. public/db_code.json exists so `npm run dev` works locally, but it
// must never reach dist/ — deploying it would point every client at this
// machine's database. deploy/deploy.sh writes the real one on the server and
// fails the deploy if this file leaks into a build.
const removeDbCode = () => {
  return {
    name: 'remove-db-code-json',
    apply: 'build',
    closeBundle() {
      const target = fileURLToPath(new URL('./dist/db_code.json', import.meta.url))
      if (existsSync(target)) {
        rmSync(target)
        this.warn('removed dist/db_code.json (per-server file, written by deploy/deploy.sh)')
      }
    },
  }
}

// Get environment
const isProduction = process.env.NODE_ENV === 'production'

// Set API URL based on environment
// Note: This is only used for the dev server proxy during development
// In production, the app uses relative URLs (window.location.origin)
// so it works from any client domain without code changes
const apiUrl = 'http://localhost:8000'

export default defineConfig({
  // vue-plugin-vue-devtools is intentionally not loaded: importing it evaluates
  // @vue/devtools-kit at config load, which touches localStorage and throws
  // "localStorage.getItem is not a function" under the Node build, breaking
  // `vite build` and `vitest`. Re-add it once the plugin is fixed or pinned.
  plugins: [vue(), removeVendorPreload(), removeDbCode()],
  test: {
    globals: true,
    environment: 'node',
    include: ['src/**/*.spec.js', 'src/**/*.test.js'],
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    allowedHosts: ['python-platinum-rainbow-printed.trycloudflare.com'],
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
  base: './',
  build: {
    rollupOptions: {
      output: {
        entryFileNames: `[name].[hash].${buildId}.js`,
        chunkFileNames: `[name].[hash].${buildId}.js`,
        assetFileNames: (assetInfo) => {
          // Preserve exact filenames for logo.png, letter_head.png, and gml2.png
          const preservedAssets = ['logo.png', 'letter_head.png', 'gml2.png']
          const assetName = assetInfo.name || ''

          // Check if the asset name ends with any of the preserved asset names
          // This handles cases where the path might be included
          const matchesPreserved = preservedAssets.some((name) => {
            return (
              assetName.endsWith(name) ||
              assetName.includes(`/${name}`) ||
              assetName.includes(`\\${name}`)
            )
          })

          if (matchesPreserved) {
            // Extract just the filename from the path
            const fileName = assetName.split('/').pop() || assetName.split('\\').pop() || assetName
            return fileName
          }

          // For all other assets: hash + buildId so names change every build
          return `[name].[hash].${buildId}.[ext]`
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
