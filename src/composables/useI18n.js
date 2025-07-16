import { useI18n } from 'vue-i18n'
import i18nDevHelper from '../utils/i18nDevHelper'

export function useEnhancedI18n() {
  const { t, locale, availableLocales } = useI18n()

  // Enhanced translation function with better fallback handling
  const translate = (key, values = {}) => {
    try {
      // First try to get the translation
      const translation = t(key, values)

      // If the translation is the same as the key, it means the key is missing
      if (translation === key) {
        // Track missing key in development
        i18nDevHelper.trackMissingKey(key, locale.value)

        // Try to find a fallback in English
        try {
          const fallback = t(key, values, 'en')
          if (fallback && fallback !== key) {
            return fallback
          }
        } catch (fallbackError) {
          // Fallback failed, continue to next step
        }

        // If still not found, try to find a similar key
        const similarKey = findSimilarKeyInMessages(key)
        if (similarKey) {
          try {
            return t(similarKey, values)
          } catch (similarError) {
            // Similar key translation failed, continue to next step
          }
        }

        // Last resort: return a user-friendly version of the key
        return formatKeyForDisplay(key)
      }

      return translation
    } catch (error) {
      console.warn(`Translation error for key "${key}":`, error)
      i18nDevHelper.trackMissingKey(key, locale.value)
      return formatKeyForDisplay(key)
    }
  }

  // Helper function to find similar keys in messages
  const findSimilarKeyInMessages = (key) => {
    try {
      // Get English messages directly from the import
      const englishMessages = require('@/locales/en.json')
      const keys = getAllKeysFromObject(englishMessages)

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
      console.warn('Error finding similar keys:', error)
      return null
    }
  }

  // Helper function to get all keys from nested object
  const getAllKeysFromObject = (obj, prefix = '') => {
    const keys = []
    try {
      for (const key in obj) {
        const fullKey = prefix ? `${prefix}.${key}` : key
        if (typeof obj[key] === 'object' && obj[key] !== null) {
          keys.push(...getAllKeysFromObject(obj[key], fullKey))
        } else {
          keys.push(fullKey)
        }
      }
    } catch (error) {
      console.warn('Error getting keys from object:', error)
    }
    return keys
  }

  // Format key for display when translation is missing
  const formatKeyForDisplay = (key) => {
    try {
      // Convert camelCase or snake_case to Title Case
      return key
        .split('.')
        .pop()
        .replace(/([A-Z])/g, ' $1')
        .replace(/_/g, ' ')
        .replace(/^\w/, (c) => c.toUpperCase())
        .trim()
    } catch (error) {
      // If formatting fails, return the key as is
      return key
    }
  }

  // Check if a translation key exists
  const hasTranslation = (key) => {
    try {
      const translation = t(key)
      return translation !== key
    } catch {
      return false
    }
  }

  // Get all missing keys for the current locale
  const getMissingKeys = () => {
    try {
      const currentMessages = t('', {}, locale.value)
      const englishMessages = t('', {}, 'en')
      const missingKeys = []

      const checkKeys = (obj, prefix = '') => {
        for (const key in obj) {
          const fullKey = prefix ? `${prefix}.${key}` : key
          if (typeof obj[key] === 'object') {
            checkKeys(obj[key], fullKey)
          } else {
            if (!currentMessages[fullKey] && englishMessages[fullKey]) {
              missingKeys.push(fullKey)
            }
          }
        }
      }

      checkKeys(englishMessages)
      return missingKeys
    } catch (error) {
      console.warn('Error getting missing keys:', error)
      return []
    }
  }

  // Get development helper methods
  const getDevHelper = () => {
    return {
      trackMissingKey: i18nDevHelper.trackMissingKey.bind(i18nDevHelper),
      getMissingKeys: i18nDevHelper.getMissingKeys.bind(i18nDevHelper),
      clearMissingKeys: i18nDevHelper.clearMissingKeys.bind(i18nDevHelper),
      exportMissingKeys: i18nDevHelper.exportMissingKeys.bind(i18nDevHelper),
      checkKeyExists: i18nDevHelper.checkKeyExists.bind(i18nDevHelper),
    }
  }

  return {
    t: translate,
    locale,
    availableLocales,
    hasTranslation,
    getMissingKeys,
    formatKeyForDisplay,
    getDevHelper,
  }
}
