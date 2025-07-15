import { createI18n } from 'vue-i18n'
import en from '@/locales/en.json'
import ar from '@/locales/ar.json'
import fr from '@/locales/fr.json'
import zh from '@/locales/zh.json'

// Get stored language or detect from browser
const getStoredLanguage = () => {
  const stored = localStorage.getItem('app_language')
  if (stored) return stored
  
  // Detect browser language
  const browserLang = navigator.language || navigator.userLanguage
  if (browserLang.startsWith('ar')) return 'ar'
  if (browserLang.startsWith('fr')) return 'fr'
  if (browserLang.startsWith('zh')) return 'zh'
  return 'en'
}

// Create i18n instance
const i18n = createI18n({
  legacy: false, // Use Composition API
  locale: getStoredLanguage(),
  fallbackLocale: 'en',
  messages: {
    en,
    ar,
    fr,
    zh
  },
  // Global properties
  globalInjection: true,
  // RTL support
  numberFormats: {
    en: {
      currency: {
        style: 'currency',
        currency: 'USD'
      }
    },
    ar: {
      currency: {
        style: 'currency',
        currency: 'DZD'
      }
    },
    fr: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      }
    },
    zh: {
      currency: {
        style: 'currency',
        currency: 'CNY'
      }
    }
  },
  // Date formats
  datetimeFormats: {
    en: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      }
    },
    ar: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      }
    },
    fr: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      }
    },
    zh: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      }
    }
  }
})

// Function to change language
export const changeLanguage = (locale) => {
  i18n.global.locale.value = locale
  localStorage.setItem('app_language', locale)
  
  // Update document direction for RTL languages
  if (locale === 'ar') {
    document.documentElement.dir = 'rtl'
    document.documentElement.lang = 'ar'
  } else {
    document.documentElement.dir = 'ltr'
    document.documentElement.lang = locale
  }
}

// Initialize document direction
if (i18n.global.locale.value === 'ar') {
  document.documentElement.dir = 'rtl'
  document.documentElement.lang = 'ar'
} else {
  document.documentElement.dir = 'ltr'
  document.documentElement.lang = i18n.global.locale.value
}

export default i18n
