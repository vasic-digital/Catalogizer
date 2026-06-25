import React from 'react'
import { Sun, Moon } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useTheme } from '@/contexts/ThemeContext'

/**
 * ThemeToggle — a single icon button that flips the app between light and dark
 * themes via {@link useTheme}. Shows a Moon when light (click → go dark) and a
 * Sun when dark (click → go light). Reuses the local ui Button `ghost` variant
 * + the global focus-visible ring, so it inherits the Catalogizer Blue token
 * system rather than introducing any ad-hoc one-off CSS (§11.4.162).
 */
export const ThemeToggle: React.FC<{ className?: string }> = ({ className }) => {
  const { theme, toggle } = useTheme()
  const isDark = theme === 'dark'

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggle}
      className={className ?? 'h-8 w-8'}
      aria-label={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
      aria-pressed={isDark}
      title={isDark ? 'Switch to light theme' : 'Switch to dark theme'}
      data-testid="theme-toggle"
    >
      {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </Button>
  )
}

export default ThemeToggle
