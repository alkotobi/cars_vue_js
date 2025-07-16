// Development helper for i18n missing keys
class I18nDevHelper {
  constructor() {
    this.missingKeys = new Set()
    this.isDevelopment = process.env.NODE_ENV === 'development'
  }

  // Track missing key
  trackMissingKey(key, locale) {
    if (this.isDevelopment) {
      this.missingKeys.add(`${locale}:${key}`)
      console.warn(`Missing translation key: ${key} for locale: ${locale}`)
    }
  }

  // Get all missing keys
  getMissingKeys() {
    return Array.from(this.missingKeys)
  }

  // Clear missing keys
  clearMissingKeys() {
    this.missingKeys.clear()
  }

  // Export missing keys for translation files
  exportMissingKeys() {
    if (this.isDevelopment) {
      const missingByLocale = {}

      this.missingKeys.forEach((key) => {
        const [locale, translationKey] = key.split(':', 2)
        if (!missingByLocale[locale]) {
          missingByLocale[locale] = []
        }
        missingByLocale[locale].push(translationKey)
      })

      console.log('Missing translation keys by locale:', missingByLocale)
      return missingByLocale
    }
    return {}
  }

  // Check if a key exists in all locales
  checkKeyExists(key, i18nInstance) {
    const locales = i18nInstance.global.availableLocales
    const missingLocales = []

    locales.forEach((locale) => {
      try {
        const translation = i18nInstance.global.t(key, {}, locale)
        if (translation === key) {
          missingLocales.push(locale)
        }
      } catch (error) {
        missingLocales.push(locale)
      }
    })

    if (missingLocales.length > 0 && this.isDevelopment) {
      console.warn(`Key "${key}" missing in locales:`, missingLocales)
    }

    return missingLocales.length === 0
  }
}

// Create singleton instance
const i18nDevHelper = new I18nDevHelper()

export default i18nDevHelper
