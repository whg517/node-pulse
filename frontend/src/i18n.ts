/**
 * i18n Configuration
 *
 * Internationalization setup using react-i18next.
 * Supports English (en) and Simplified Chinese (zh-CN).
 */

import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from './locales/en.json'
import zhCN from './locales/zh-CN.json'

// Language resources
const resources = {
  en: { translation: en },
  'zh-CN': { translation: zhCN },
}

// Get saved language from localStorage or use browser preference
const getSavedLanguage = (): string => {
  const saved = localStorage.getItem('settings:language')
  if (saved && (saved === 'en' || saved === 'zh-CN')) {
    return saved
  }

  // Fall back to browser language
  const browserLang = navigator.language
  if (browserLang.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en'
}

// Initialize i18n and export the promise so app can wait for it
const i18nInitPromise = i18n.use(initReactI18next).init({
  resources,
  lng: getSavedLanguage(),
  fallbackLng: 'en',
  interpolation: {
    escapeValue: false, // React already escapes values
  },
  react: {
    useSuspense: false, // Avoid suspense for faster initial load
  },
})

export default i18n
export { i18nInitPromise }

// Supported languages - re-exported from i18n-config to avoid bundling i18next
// in components that only need the language list
export { supportedLanguages, type LanguageCode } from './i18n-config'
