import { themeStore } from './theme.store'

describe('themeStore', () => {
  beforeEach(() => {
    themeStore.reset()
  })

  it('should have initial theme set to light', () => {
    expect(themeStore.state.theme).toEqual({ isDark: false })
  })

  it('should update theme to dark', () => {
    themeStore.update((state) => ({
      ...state,
      theme: { isDark: true },
    }))

    expect(themeStore.state.theme).toEqual({ isDark: true })
  })

  it('should reset theme to initial state', () => {
    themeStore.update((state) => ({
      ...state,
      theme: { isDark: true },
    }))

    themeStore.reset()

    expect(themeStore.state.theme).toEqual({ isDark: false })
  })
})
