import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './en.json'
import ptBR from './pt-BR.json'

const STORAGE_KEY = 'los-lang'
const FALLBACK_LANG = 'en'
const SUPPORTED_LANGS = ['en', 'pt-BR'] as const

export type SupportedLanguage = (typeof SUPPORTED_LANGS)[number]

const getInitialLanguage = (): SupportedLanguage => {
  if (typeof window === 'undefined') {
    return FALLBACK_LANG
  }
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored === 'pt-BR') {
    return 'pt-BR'
  }
  return FALLBACK_LANG
}

const setDocumentLang = (lang: SupportedLanguage) => {
  if (typeof document === 'undefined') return
  document.documentElement.lang = lang === 'pt-BR' ? 'pt-BR' : 'en'
}

const initialLanguage = getInitialLanguage()

i18n
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: en },
      'pt-BR': { translation: ptBR }
    },
    lng: initialLanguage,
    fallbackLng: FALLBACK_LANG,
    interpolation: {
      escapeValue: false
    }
  })

setDocumentLang(initialLanguage)

i18n.on('languageChanged', (lang) => {
  if (lang === 'en' || lang === 'pt-BR') {
    setDocumentLang(lang)
  }
})

export const setLanguage = (lang: SupportedLanguage) => {
  if (!SUPPORTED_LANGS.includes(lang)) return
  i18n.changeLanguage(lang)
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, lang)
  }
}

export const getLocale = (lang: string = i18n.language) => (lang === 'pt-BR' ? 'pt-BR' : 'en-US')

export default i18n
