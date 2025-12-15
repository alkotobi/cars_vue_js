import './assets/main.css'
import '@fortawesome/fontawesome-free/css/all.css'

import { createApp } from 'vue'

// Debug: Log base path detection at app startup
console.log('=== APP STARTUP DEBUG ===')
console.log('window.location.href:', window.location.href)
console.log('window.location.pathname:', window.location.pathname)
console.log('window.location.origin:', window.location.origin)
console.log('import.meta.env.BASE_URL:', import.meta.env.BASE_URL)
console.log('document.baseURI:', document.baseURI)

// Check if there's a <base> tag
const baseTag = document.querySelector('base')
if (baseTag) {
  console.log('<base> tag found:', baseTag.href)
} else {
  console.log('No <base> tag found')
}

// Detect base path
const detectBasePath = () => {
  let baseUrl = import.meta.env.BASE_URL || './'
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    const match = pathname.match(/^(\/[^/]+\/)/)
    return match ? match[1] : '/'
  }
  return baseUrl
}
const detectedBase = detectBasePath()
console.log('Detected base path:', detectedBase)
console.log('========================')
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'

const app = createApp(App)

// Add global translation helper
app.config.globalProperties.$translate = (key, values = {}) => {
  try {
    const translation = i18n.global.t(key, values)
    if (translation === key) {
      // Try fallback to English
      try {
        const fallback = i18n.global.t(key, values, 'en')
        if (fallback && fallback !== key) {
          return fallback
        }
      } catch (fallbackError) {
        // Fallback failed, continue to next step
      }
      
      // Format key for display as last resort
      try {
        return key
          .split('.')
          .pop()
          .replace(/([A-Z])/g, ' $1')
          .replace(/_/g, ' ')
          .replace(/^\w/, (c) => c.toUpperCase())
          .trim()
      } catch (formatError) {
        return key
      }
    }
    return translation
  } catch (error) {
    console.warn(`Global translation error for key "${key}":`, error)
    try {
      return key
        .split('.')
        .pop()
        .replace(/([A-Z])/g, ' $1')
        .replace(/_/g, ' ')
        .replace(/^\w/, (c) => c.toUpperCase())
        .trim()
    } catch (formatError) {
      return key
    }
  }
}

app.use(createPinia())
app.use(router)
app.use(i18n)

app.mount('#app')
