import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import en from '../locales/en.json'
import es from '../locales/es.json'
import fr from '../locales/fr.json'

// Get saved language from localStorage or use browser language
const savedLanguage = localStorage.getItem('language')
const browserLanguage = navigator.language.split('-')[0]
const defaultLanguage = savedLanguage || browserLanguage || 'en'

i18n
  .use(initReactI18next) // passes i18n down to react-i18next
  .init({
    resources: {
      en: { translation: en },
      es: { translation: es },
      fr: { translation: fr },
    },
    lng: defaultLanguage, // default language
    fallbackLng: 'en', // use en if detected lng is not available
    interpolation: {
      escapeValue: false, // react already safes from xss
    },
    debug: import.meta.env.DEV, // enable debug in development mode
  })

// Save language preference when it changes
i18n.on('languageChanged', (lng) => {
  localStorage.setItem('language', lng)
})

export default i18n
