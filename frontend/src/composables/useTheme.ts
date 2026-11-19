import { ref } from 'vue'
import {
  loadPreference,
  savePreference,
  migratePreferencesIntoUser,
} from './preferenceStorage'

export type ThemeMode = 'light' | 'dark' | 'system'

const THEME_KEY = 'theme'

/** 浅蓝玻璃拟态为浅色 only；历史 dark/system 偏好统一落到 light */
function loadTheme(): ThemeMode {
  const v = loadPreference(THEME_KEY)
  if (v === 'light' || v === 'dark' || v === 'system') return 'light'
  return 'light'
}

// Shared reactive state across all consumers
const currentTheme = ref<ThemeMode>(loadTheme())

let lastEffective: 'light' | 'dark' | null = null

/** Wails：与原生窗口底色对齐，减轻 Ctrl+R 整窗白闪 */
function syncWailsNativeChrome() {
  const bg = '#e4f0ff'
  document.documentElement.style.background = bg
  document.documentElement.style.minHeight = '100%'
  document.documentElement.style.colorScheme = 'light'
  if (document.body) {
    document.body.style.background = bg
    document.body.style.minHeight = '100%'
  }
  const appEl = document.getElementById('app')
  if (appEl) {
    appEl.style.background = 'transparent'
    appEl.style.minHeight = '100%'
  }
  const w = (window as unknown as {
    runtime?: {
      WindowSetDarkTheme?: () => void
      WindowSetLightTheme?: () => void
      WindowSetBackgroundColour?: (r: number, g: number, b: number, a: number) => void
    }
  }).runtime
  if (!w?.WindowSetBackgroundColour) return
  try {
    w.WindowSetLightTheme?.()
    // #e4f0ff
    w.WindowSetBackgroundColour(228, 240, 255, 255)
  } catch {
    /* 非桌面壳或未注入 runtime */
  }
}

function applyTheme(_mode: ThemeMode) {
  const effective: 'light' = 'light'
  if (lastEffective === effective) return
  lastEffective = effective
  document.documentElement.setAttribute('theme-mode', effective)
  syncWailsNativeChrome()
}

export function useTheme() {
  function setTheme(mode: ThemeMode): boolean {
    if (mode !== 'light' && mode !== 'dark' && mode !== 'system') return false
    // 浅色 only：持久化与生效都固定为 light
    currentTheme.value = 'light'
    savePreference(THEME_KEY, 'light')
    applyTheme('light')
    return true
  }

  return { currentTheme, setTheme }
}

/** Call once in main.ts to initialise theme and listen for OS changes. */
export function initTheme() {
  currentTheme.value = 'light'
  savePreference(THEME_KEY, 'light')
  applyTheme('light')
}

/** Re-read preferences from storage (call after login / logout). */
export function reloadThemeFromStorage() {
  migratePreferencesIntoUser()
  currentTheme.value = 'light'
  savePreference(THEME_KEY, 'light')
  applyTheme('light')
}
