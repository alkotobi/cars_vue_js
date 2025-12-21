import { createI18n } from 'vue-i18n'
import en from '@/locales/en.json'
import ar from '@/locales/ar.json'
import fr from '@/locales/fr.json'
import zh from '@/locales/zh.json'

// Get stored language or detect from browser
const getStoredLanguage = () => {
  const stored = localStorage.getItem('app_language')

  if (stored && ['en', 'ar', 'fr', 'zh'].includes(stored)) {
    return stored
  }

  // Only detect browser language if no stored preference
  const browserLang = navigator.language || navigator.userLanguage

  if (browserLang.startsWith('ar')) return 'ar'
  if (browserLang.startsWith('fr')) return 'fr'
  if (browserLang.startsWith('zh')) return 'zh'

  // Default to English for all other cases
  return 'en'
}

// Custom missing handler to provide fallbacks
const missingHandler = (locale, key, instance, values) => {
  try {
    // Try to find the key in the fallback locale (English)
    const englishMessages = en
    const fallbackValue = getNestedValue(englishMessages, key)
    if (fallbackValue) {
      return fallbackValue
    }

    // If not found in fallback, try to find a similar key
    const similarKey = findSimilarKey(key, englishMessages)
    if (similarKey) {
      return getNestedValue(englishMessages, similarKey)
    }

    // If still not found, return the key itself as a last resort
    return key
  } catch (error) {
    return key
  }
}

// Helper function to get nested object values by dot notation
const getNestedValue = (obj, path) => {
  try {
    return path.split('.').reduce((current, key) => current && current[key], obj)
  } catch (error) {
    return undefined
  }
}

// Helper function to find similar keys
const findSimilarKey = (key, messages) => {
  try {
    const keys = getAllKeys(messages)

    // Try to find keys that contain the same words
    const keyWords = key.split('.')
    for (const msgKey of keys) {
      const msgKeyWords = msgKey.split('.')
      if (keyWords.some((word) => msgKeyWords.includes(word))) {
        return msgKey
      }
    }

    // Try to find keys that end with the same suffix
    for (const msgKey of keys) {
      if (msgKey.endsWith(key.split('.').pop())) {
        return msgKey
      }
    }

    return null
  } catch (error) {
    return null
  }
}

// Helper function to get all keys from nested object
const getAllKeys = (obj, prefix = '') => {
  const keys = []
  for (const key in obj) {
    const fullKey = prefix ? `${prefix}.${key}` : key
    if (typeof obj[key] === 'object' && obj[key] !== null) {
      keys.push(...getAllKeys(obj[key], fullKey))
    } else {
      keys.push(fullKey)
    }
  }
  return keys
}

// Create i18n instance
const i18n = createI18n({
  legacy: false, // Use Composition API
  locale: getStoredLanguage() || 'en', // Ensure we always have a valid locale
  fallbackLocale: 'en',
  messages: {
    en,
    ar,
    fr,
    zh,
  },
  // Global properties
  globalInjection: true,
  // Silent mode to prevent console warnings for missing keys
  silentTranslationWarn: true,
  silentFallbackWarn: true,
  // Custom missing handler - temporarily disabled
  // missing: missingHandler,
  // RTL support
  numberFormats: {
    en: {
      currency: {
        style: 'currency',
        currency: 'USD',
      },
    },
    ar: {
      currency: {
        style: 'currency',
        currency: 'DZD',
      },
    },
    fr: {
      currency: {
        style: 'currency',
        currency: 'EUR',
      },
    },
    zh: {
      currency: {
        style: 'currency',
        currency: 'CNY',
      },
    },
  },
  // Date formats
  datetimeFormats: {
    en: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      },
    },
    ar: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      },
    },
    fr: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      },
    },
    zh: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      },
    },
  },
})

// Function to change language
export const changeLanguage = (locale) => {
  // Validate locale
  if (!['en', 'ar', 'fr', 'zh'].includes(locale)) {
    locale = 'en'
  }

  // Update i18n locale
  i18n.global.locale.value = locale

  // Store in localStorage
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

// Function to force reset language to English
export const resetToEnglish = () => {
  localStorage.removeItem('app_language')
  changeLanguage('en')
}

// Initialize document direction
if (i18n.global.locale.value === 'ar') {
  document.documentElement.dir = 'rtl'
  document.documentElement.lang = 'ar'
} else {
  document.documentElement.dir = 'ltr'
  document.documentElement.lang = i18n.global.locale.value
}

// Temporary fix: Manually add the missing key if it doesn't exist
if (
  i18n.global.messages.value.en?.loading &&
  !i18n.global.messages.value.en.loading.loading_management
) {
  i18n.global.messages.value.en.loading.loading_management = 'Loading Management'
}

export default i18n
