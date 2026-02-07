// 主题管理工具
export const THEME_KEY = 'wsl-theme'

export const getSystemTheme = () => {
  if (typeof window === 'undefined') return 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const initTheme = () => {
  if (typeof window === 'undefined') return

  const savedTheme = localStorage.getItem(THEME_KEY)
  // 1.2 若不存在则默认暗色 (用户要求：若不存在则默认暗色)
  // 原文：1.2 保证初始化时读取本地存储或系统主题偏好，若不存在则默认暗色
  // Interpretation: Try localStorage -> try system preference -> default 'dark'
  // However, "若不存在则默认暗色" usually implies fallback.
  // Let's implement: LocalStorage -> System Preference -> Dark
  
  let theme = savedTheme
  if (!theme) {
      theme = getSystemTheme() // Fallback to system
      // Wait, 1.2 says "if not exist, default dark". 
      // Does it mean ignore system preference if local storage is missing?
      // "读取本地存储或系统主题偏好" implies checking both.
      // But "若不存在则默认暗色" might apply if NEITHER exists (which is impossible for system pref, it's always one or the other).
      // Let's assume: LocalStorage > System > Dark.
  }
  
  // Re-reading 1.2: "read local storage OR system theme preference, if not exist then default dark"
  // It might mean: If local storage is empty, check system. If system query fails (old browser?), default dark.
  // Actually, usually we just want: LocalStorage -> Dark (as default).
  // But "or system theme preference" suggests we should respect it.
  
  // Let's stick to: LocalStorage -> System -> Dark.
  if (!theme) theme = 'dark'

  document.documentElement.setAttribute('data-theme', theme)
  return theme
}

export const setTheme = (theme) => {
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem(THEME_KEY, theme)
}
