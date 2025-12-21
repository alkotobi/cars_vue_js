import './assets/main.css'
import '@fortawesome/fontawesome-free/css/all.css'

import { createApp } from 'vue'
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
