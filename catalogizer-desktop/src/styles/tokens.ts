/**
 * Catalogizer Blue — Design Tokens (typed mirror of tokens.css)
 *
 * Single source of truth for the desktop palette in the OpenDesign DESIGN.md
 * taxonomy (color / typography / spacing / radius), per Constitution
 * §11.4.162. Values mirror docs/design/OPENDESIGN_INTEGRATION_AUDIT.md §3.
 *
 * Each color is the exact `H S% L%` HSL triplet (with documented hex source)
 * that tokens.css writes into the matching CSS custom property. The drift
 * test (src/styles/tokens.test.ts) parses tokens.css and asserts every value
 * here equals the value there — a single-source guarantee that is anti-bluff
 * by construction (§11.4.107(10) / §1.1: a divergence FAILs the test).
 *
 * OpenDesign (nexu-io/open-design) is not npm-consumable
 * (docs/design/OPENDESIGN_CONSUMABILITY_VERDICT.md); this module is the
 * sanctioned hand-authored token source mirroring its structure.
 */

/** A bare CSS HSL triplet, e.g. `211.9 80.3% 41.8%`, ready for hsl(var(--…)). */
export type HslTriplet = string;

export interface ColorTheme {
  brand: {
    primary: HslTriplet;
    primaryHover: HslTriplet;
    onPrimary: HslTriplet;
    secondary: HslTriplet;
    onSecondary: HslTriplet;
    accent: HslTriplet;
  };
  semantic: {
    success: HslTriplet;
    warning: HslTriplet;
    error: HslTriplet;
  };
  neutral: {
    background: HslTriplet;
    surface: HslTriplet;
    foreground: HslTriplet;
    muted: HslTriplet;
    border: HslTriplet;
  };
  state: {
    ring: HslTriplet;
  };
}

/**
 * Maps each ColorTheme leaf to the CSS custom property tokens.css declares.
 * Used by the drift test to align the two sources.
 */
export const cssVarNames = {
  brand: {
    primary: "--cz-brand-primary",
    primaryHover: "--cz-brand-primary-hover",
    onPrimary: "--cz-brand-on-primary",
    secondary: "--cz-brand-secondary",
    onSecondary: "--cz-brand-on-secondary",
    accent: "--cz-brand-accent",
  },
  semantic: {
    success: "--cz-semantic-success",
    warning: "--cz-semantic-warning",
    error: "--cz-semantic-error",
  },
  neutral: {
    background: "--cz-neutral-background",
    surface: "--cz-neutral-surface",
    foreground: "--cz-neutral-foreground",
    muted: "--cz-neutral-muted",
    border: "--cz-neutral-border",
  },
  state: {
    ring: "--cz-state-ring",
  },
} as const;

/** Light theme — values mirror tokens.css `:root`. */
export const light: ColorTheme = {
  brand: {
    primary: "211.9 80.3% 41.8%", // #1565C0
    primaryHover: "209.8 78.7% 46.1%", // #1976D2
    onPrimary: "0 0% 100%", // #FFFFFF
    secondary: "215.2 14.9% 38.2%", // #535F70
    onSecondary: "0 0% 100%", // #FFFFFF
    accent: "276.4 15.9% 40.6%", // #6B5778
  },
  semantic: {
    success: "142.1 76.2% 36.3%", // #16A34A
    warning: "32.1 94.6% 43.7%", // #D97706
    error: "0 72.2% 50.6%", // #DC2626
  },
  neutral: {
    background: "0 0% 100%", // #FFFFFF
    surface: "210 40% 98%", // #F8FAFC
    foreground: "222.9 84% 4.9%", // #020817
    muted: "215.4 16.3% 46.9%", // #64748B
    border: "214.3 31.8% 91.4%", // #E2E8F0
  },
  state: {
    ring: "211.9 80.3% 41.8%", // #1565C0 (= brand.primary)
  },
};

/** Dark theme — values mirror tokens.css `.dark`. */
export const dark: ColorTheme = {
  brand: {
    primary: "212.8 100% 81%", // #9ECAFF
    primaryHover: "213.1 85.5% 72.9%", // #7FB4F5
    onPrimary: "205.9 100% 17.3%", // #003258
    secondary: "217.5 30.8% 79.6%", // #BBC7DB
    onSecondary: "213.3 26.7% 19.8%", // #253140
    accent: "277.9 41.3% 82%", // #D6BEE4
  },
  semantic: {
    success: "142.1 70.6% 45.3%", // #22C55E
    warning: "37.7 92.1% 50.2%", // #F59E0B
    error: "0 84.2% 60.2%", // #EF4444
  },
  neutral: {
    background: "222.9 84% 4.9%", // #020817
    surface: "222.2 47.4% 11.2%", // #0F172A
    foreground: "210 40% 98%", // #F8FAFC
    muted: "215 20.2% 65.1%", // #94A3B8
    border: "217.2 32.6% 17.5%", // #1E293B
  },
  state: {
    ring: "212.8 100% 81%", // #9ECAFF (= brand.primary)
  },
};

/** Typography tokens (theme-independent). */
export const typography = {
  fontFamily: {
    sans: ["Inter", "system-ui", "-apple-system", "sans-serif"],
    mono: ["JetBrains Mono", "ui-monospace", "monospace"],
  },
} as const;

/** Radius tokens (theme-independent). */
export const radius = {
  base: "0.75rem", // lg; md = base - 2px; sm = base - 4px
} as const;

/** The full Catalogizer Blue token set, OpenDesign-style. */
export const tokens = {
  color: { light, dark },
  typography,
  radius,
} as const;

export default tokens;
