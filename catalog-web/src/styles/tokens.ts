/**
 * Catalogizer Blue — unified design-token module (OpenDesign DESIGN.md taxonomy).
 *
 * Single source of truth for the web client's brand palette (light + dark),
 * typography, spacing and radius. Per §11.4.162 (OpenDesign mandate) the
 * project consumes OpenDesign's token *structure* (color / typography / space /
 * radius sections); per docs/design/OPENDESIGN_CONSUMABILITY_VERDICT.md
 * `nexu-io/open-design` is NOT npm-consumable (it ships no token package), so
 * this hand-authored module mirrors its taxonomy and encodes the unified
 * "Catalogizer Blue" palette from docs/design/OPENDESIGN_INTEGRATION_AUDIT.md §3.
 *
 * Every value below is copied verbatim from the audit §3.1–§3.3 (themselves
 * traced to real per-app source files). The CSS-var layer (src/index.css) and
 * the Tailwind config consume FROM this module so a brand change is one edit.
 *
 * Two representations per color: `hex` (human-authored, audit value) and `hsl`
 * (the `H S% L%` channel triplet shadcn CSS vars require — computed exactly via
 * colorsys from the hex, never hand-eyeballed, per §11.4.6 no-guessing).
 */

export interface ColorToken {
  /** Authored hex (audit §3.1). */
  readonly hex: string
  /** shadcn HSL channel triplet `H S% L%` for `hsl(var(--x))`. */
  readonly hsl: string
}

export interface ThemedColor {
  readonly light: ColorToken
  readonly dark: ColorToken
}

const t = (hex: string, hsl: string): ColorToken => ({ hex, hsl })

/**
 * Color palette — OpenDesign `color.*` section. Light + dark variants for
 * every role. Values + HSL triplets verified against audit §3.1.
 */
export const color = {
  brand: {
    primary: {
      light: t('#1565C0', '211.9 80.3% 41.8%'),
      dark: t('#9ECAFF', '212.8 100% 81%'),
    },
    primaryHover: {
      light: t('#1976D2', '209.8 78.7% 46.1%'),
      dark: t('#7FB4F5', '213.1 85.5% 72.9%'),
    },
    onPrimary: {
      light: t('#FFFFFF', '0 0% 100%'),
      dark: t('#003258', '205.9 100% 17.3%'),
    },
    secondary: {
      light: t('#535F70', '215.2 14.9% 38.2%'),
      dark: t('#BBC7DB', '217.5 30.8% 79.6%'),
    },
    onSecondary: {
      light: t('#FFFFFF', '0 0% 100%'),
      dark: t('#253140', '212.7 26.8% 19.4%'),
    },
    accent: {
      light: t('#6B5778', '276.4 15.9% 40.6%'),
      dark: t('#D6BEE4', '277.9 41.3% 82%'),
    },
    onAccent: {
      light: t('#FFFFFF', '0 0% 100%'),
      dark: t('#251431', '275.2 42% 13.5%'),
    },
  },
  semantic: {
    success: {
      light: t('#16A34A', '142.1 76.2% 36.3%'),
      dark: t('#22C55E', '142.1 70.6% 45.3%'),
    },
    warning: {
      light: t('#D97706', '32.1 94.6% 43.7%'),
      dark: t('#F59E0B', '37.7 92.1% 50.2%'),
    },
    error: {
      light: t('#DC2626', '0 72.2% 50.6%'),
      dark: t('#EF4444', '0 84.2% 60.2%'),
    },
  },
  neutral: {
    background: {
      light: t('#FFFFFF', '0 0% 100%'),
      dark: t('#020817', '222.9 84% 4.9%'),
    },
    surface: {
      light: t('#F8FAFC', '210 40% 98%'),
      dark: t('#0F172A', '222.2 47.4% 11.2%'),
    },
    foreground: {
      light: t('#020817', '222.9 84% 4.9%'),
      dark: t('#F8FAFC', '210 40% 98%'),
    },
    muted: {
      light: t('#64748B', '215.4 16.3% 46.9%'),
      dark: t('#94A3B8', '215 20.2% 65.1%'),
    },
    border: {
      light: t('#E2E8F0', '214.3 31.8% 91.4%'),
      dark: t('#1E293B', '217.2 32.6% 17.5%'),
    },
  },
  state: {
    ring: {
      light: t('#1565C0', '211.9 80.3% 41.8%'),
      dark: t('#9ECAFF', '212.8 100% 81%'),
    },
  },
} as const

/**
 * Typography — OpenDesign `typography.*` section. Origin: catalog-web
 * tailwind.config.js fontFamily (audit §3.2).
 */
export const typography = {
  fontFamily: {
    sans: 'Inter, system-ui, sans-serif',
    mono: 'JetBrains Mono, monospace',
  },
} as const

/**
 * Spacing — OpenDesign `space.*` section (named steps beyond Tailwind's grid).
 * Origin: catalog-web tailwind.config.js spacing (audit §3.3).
 */
export const space = {
  '18': '4.5rem',
  '88': '22rem',
  '128': '32rem',
} as const

/**
 * Radius — OpenDesign `radius.*` section. Origin: index.css `--radius` +
 * tailwind.config.js borderRadius (audit §3.3).
 */
export const radius = {
  base: '0.75rem',
  '4xl': '2rem',
} as const

/**
 * Aggregate "Catalogizer Blue" token object in OpenDesign DESIGN.md shape.
 * A future real OpenDesign export can be diffed against this structure.
 */
export const catalogizerBlue = {
  color,
  typography,
  space,
  radius,
} as const

/** Convenience: light-theme CSS-var map keyed by shadcn var name → HSL triplet. */
export const lightVars: Record<string, string> = {
  '--background': color.neutral.background.light.hsl,
  '--foreground': color.neutral.foreground.light.hsl,
  '--card': color.neutral.background.light.hsl,
  '--card-foreground': color.neutral.foreground.light.hsl,
  '--popover': color.neutral.background.light.hsl,
  '--popover-foreground': color.neutral.foreground.light.hsl,
  '--primary': color.brand.primary.light.hsl,
  '--primary-foreground': color.brand.onPrimary.light.hsl,
  '--secondary': color.neutral.surface.light.hsl,
  '--secondary-foreground': color.neutral.foreground.light.hsl,
  '--muted': color.neutral.surface.light.hsl,
  '--muted-foreground': color.neutral.muted.light.hsl,
  '--accent': color.brand.accent.light.hsl,
  '--accent-foreground': color.brand.onAccent.light.hsl,
  '--destructive': color.semantic.error.light.hsl,
  '--destructive-foreground': color.brand.onPrimary.light.hsl,
  '--border': color.neutral.border.light.hsl,
  '--input': color.neutral.border.light.hsl,
  '--ring': color.state.ring.light.hsl,
  '--radius': radius.base,
}

/** Convenience: dark-theme CSS-var map keyed by shadcn var name → HSL triplet. */
export const darkVars: Record<string, string> = {
  '--background': color.neutral.background.dark.hsl,
  '--foreground': color.neutral.foreground.dark.hsl,
  '--card': color.neutral.background.dark.hsl,
  '--card-foreground': color.neutral.foreground.dark.hsl,
  '--popover': color.neutral.background.dark.hsl,
  '--popover-foreground': color.neutral.foreground.dark.hsl,
  '--primary': color.brand.primary.dark.hsl,
  '--primary-foreground': color.brand.onPrimary.dark.hsl,
  '--secondary': color.neutral.surface.dark.hsl,
  '--secondary-foreground': color.neutral.foreground.dark.hsl,
  '--muted': color.neutral.surface.dark.hsl,
  '--muted-foreground': color.neutral.muted.dark.hsl,
  '--accent': color.brand.accent.dark.hsl,
  '--accent-foreground': color.brand.onAccent.dark.hsl,
  '--destructive': color.semantic.error.dark.hsl,
  '--destructive-foreground': color.neutral.foreground.dark.hsl,
  '--border': color.neutral.border.dark.hsl,
  '--input': color.neutral.border.dark.hsl,
  '--ring': color.state.ring.dark.hsl,
}

export default catalogizerBlue
