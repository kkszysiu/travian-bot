import { writable, derived, get } from 'svelte/store'
import en from './en'
import pl from './pl'

export type Locale = 'en' | 'pl'

const translations: Record<Locale, typeof en> = { en, pl }

const STORAGE_KEY = 'travian-bot-locale'

function getInitialLocale(): Locale {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'en' || stored === 'pl') return stored
  } catch {}
  return 'en'
}

export const locale = writable<Locale>(getInitialLocale())

locale.subscribe(v => {
  try { localStorage.setItem(STORAGE_KEY, v) } catch {}
})

export const t = derived(locale, ($locale) => {
  const msgs = translations[$locale]
  return (key: string, params?: Record<string, string | number>): string => {
    const parts = key.split('.')
    let val: any = msgs
    for (const p of parts) {
      if (val == null) return key
      val = val[p]
    }
    if (typeof val !== 'string') return key
    if (params) {
      return val.replace(/\{(\w+)\}/g, (_, k) => String(params[k] ?? `{${k}}`))
    }
    return val
  }
})

export const localeNames: Record<Locale, string> = {
  en: 'English',
  pl: 'Polski',
}
