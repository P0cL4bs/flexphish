import { createStore, withProps } from '@ngneat/elf'

interface ThemeProps {
  theme: { isDark: boolean } | null
}

export const THEMES = {
  DARK: 'dark',
  LIGHT: 'light',
}

export const themeStore = createStore({ name: 'theme' }, withProps<ThemeProps>({ theme: { isDark: false } }))
