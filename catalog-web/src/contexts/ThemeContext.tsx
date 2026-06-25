import React, { createContext, useContext, useEffect, useState, useCallback } from 'react'

/**
 * ThemeContext — light/dark theme management for the Tailwind `darkMode: 'class'`
 * strategy. Applies/removes the `.dark` class on `document.documentElement`,
 * persists the explicit choice to localStorage, and defaults to the OS preference
 * (`matchMedia('(prefers-color-scheme: dark)')`) when no choice is stored.
 *
 * Wires the Catalogizer Blue token system (src/styles/tokens.ts → src/index.css
 * `:root`/`.dark`) into a real, user-toggleable theme per §11.4.162 (every
 * component ships light + dark variants). Before this provider the `.dark`
 * palette was dead CSS (no code ever applied the class).
 */

export type Theme = 'light' | 'dark'

interface ThemeContextValue {
  theme: Theme
  /** Explicitly set the theme (persisted). */
  setTheme: (theme: Theme) => void
  /** Toggle between light and dark (persisted). */
  toggle: () => void
}

const STORAGE_KEY = 'catalog-web-theme'

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined)

/** Read the initial theme: stored choice wins, else OS preference, else light. */
function getInitialTheme(): Theme {
  if (typeof window === 'undefined') return 'light'
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') return stored
  } catch {
    // localStorage may be unavailable (private mode / sandbox); fall through.
  }
  if (typeof window.matchMedia === 'function') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'light'
}

/** Apply the theme to the document root by toggling the `.dark` class. */
function applyTheme(theme: Theme): void {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  root.classList.toggle('dark', theme === 'dark')
}

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setThemeState] = useState<Theme>(getInitialTheme)

  // Apply the active theme to the document on mount + whenever it changes.
  useEffect(() => {
    applyTheme(theme)
  }, [theme])

  // Follow OS preference changes ONLY while the user has not made an explicit choice.
  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return
    const mql = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = (e: MediaQueryListEvent) => {
      let stored: string | null = null
      try {
        stored = window.localStorage.getItem(STORAGE_KEY)
      } catch {
        stored = null
      }
      if (stored !== 'light' && stored !== 'dark') {
        setThemeState(e.matches ? 'dark' : 'light')
      }
    }
    mql.addEventListener('change', handleChange)
    return () => mql.removeEventListener('change', handleChange)
  }, [])

  const setTheme = useCallback((next: Theme) => {
    setThemeState(next)
    try {
      window.localStorage.setItem(STORAGE_KEY, next)
    } catch {
      // Persisting is best-effort; the in-memory state still applies.
    }
  }, [])

  const toggle = useCallback(() => {
    setThemeState((current) => {
      const next: Theme = current === 'dark' ? 'light' : 'dark'
      try {
        window.localStorage.setItem(STORAGE_KEY, next)
      } catch {
        // best-effort
      }
      return next
    })
  }, [])

  return (
    <ThemeContext.Provider value={{ theme, setTheme, toggle }}>
      {children}
    </ThemeContext.Provider>
  )
}

/** Access the active theme + setters. Throws if used outside <ThemeProvider>. */
export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext)
  if (ctx === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return ctx
}
