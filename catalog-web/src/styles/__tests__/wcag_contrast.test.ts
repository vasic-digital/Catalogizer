import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

/**
 * WCAG 2.1 AA contrast oracle for the destructive ↔ destructive-foreground
 * token pair (§11.4.115 RED→GREEN, §11.4.107(10) self-validated analyzer,
 * §11.4.162 OpenDesign light+dark + no-overlap/legibility mandate).
 *
 * The PASS comes from a REAL computed contrast ratio on the ACTUAL shipped
 * tokens parsed out of `src/index.css` — never a hardcoded assertion. The
 * standard sRGB relative-luminance formula is used:
 *   linearize each channel, L = 0.2126·R + 0.7152·G + 0.0722·B,
 *   ratio = (L_light + 0.05) / (L_dark + 0.05).
 * AA for normal text requires ratio >= 4.5:1.
 *
 * RED baseline (pre-fix): the dark `.dark` pair #EF4444 / #F8FAFC computes
 * ~3.60:1 and this test FAILS on it — the proof the defect is real.
 * GREEN (post-fix): the darkened destructive surface reaches >= 4.5:1.
 */

const AA_NORMAL_TEXT = 4.5

const here = dirname(fileURLToPath(import.meta.url))
const cssPath = resolve(here, '../../index.css')
const css = readFileSync(cssPath, 'utf8')

/** Extract a single non-nested `selector { ... }` rule body from the CSS. */
function ruleBody(selector: string): string {
  const re = new RegExp(`${selector.replace('.', '\\.')}\\s*\\{([^}]*)\\}`)
  const m = css.match(re)
  if (!m) throw new Error(`Could not find rule "${selector}" in index.css`)
  return m[1]
}

/** Read a CSS custom-property value (the `H S% L%` triplet) from a rule body. */
function cssVar(body: string, name: string): string {
  // `--destructive:` must NOT match `--destructive-foreground:` — the `:`
  // immediately follows the exact name, so the prefix collision is excluded.
  const m = body.match(new RegExp(`--${name}:\\s*([^;]+);`))
  if (!m) throw new Error(`Could not find --${name} in rule body`)
  return m[1].trim()
}

/** Convert a shadcn `H S% L%` triplet to sRGB [r,g,b] in 0..255. */
function hslTripletToRgb(triplet: string): [number, number, number] {
  const m = triplet.match(/^([\d.]+)\s+([\d.]+)%\s+([\d.]+)%$/)
  if (!m) throw new Error(`Cannot parse HSL triplet: "${triplet}"`)
  const h = parseFloat(m[1])
  const s = parseFloat(m[2]) / 100
  const l = parseFloat(m[3]) / 100
  const c = (1 - Math.abs(2 * l - 1)) * s
  const hp = h / 60
  const x = c * (1 - Math.abs((hp % 2) - 1))
  let r1 = 0
  let g1 = 0
  let b1 = 0
  if (hp >= 0 && hp < 1) [r1, g1, b1] = [c, x, 0]
  else if (hp < 2) [r1, g1, b1] = [x, c, 0]
  else if (hp < 3) [r1, g1, b1] = [0, c, x]
  else if (hp < 4) [r1, g1, b1] = [0, x, c]
  else if (hp < 5) [r1, g1, b1] = [x, 0, c]
  else [r1, g1, b1] = [c, 0, x]
  const mm = l - c / 2
  return [(r1 + mm) * 255, (g1 + mm) * 255, (b1 + mm) * 255]
}

/** WCAG relative luminance of an sRGB [r,g,b] (0..255) colour. */
function relativeLuminance([r, g, b]: [number, number, number]): number {
  const lin = (v: number): number => {
    const cc = v / 255
    return cc <= 0.03928 ? cc / 12.92 : Math.pow((cc + 0.055) / 1.055, 2.4)
  }
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b)
}

/** WCAG contrast ratio between two sRGB colours. */
function contrastRatio(
  a: [number, number, number],
  b: [number, number, number]
): number {
  const la = relativeLuminance(a)
  const lb = relativeLuminance(b)
  const [hi, lo] = la >= lb ? [la, lb] : [lb, la]
  return (hi + 0.05) / (lo + 0.05)
}

/** Compute destructive↔foreground contrast for a theme rule in index.css. */
function destructiveContrast(selector: string): number {
  const body = ruleBody(selector)
  const surface = hslTripletToRgb(cssVar(body, 'destructive'))
  const fg = hslTripletToRgb(cssVar(body, 'destructive-foreground'))
  return contrastRatio(surface, fg)
}

describe('WCAG AA contrast — destructive token pair (index.css)', () => {
  it('self-validation: the comparator FAILs a known sub-AA pair (anti-bluff §1.1)', () => {
    // golden-bad: #EF4444 on #F8FAFC is the historical defect ~3.60:1 — the
    // oracle MUST compute < 4.5 for it, or a green run is untrustworthy.
    const bad = contrastRatio(
      hslTripletToRgb('0 84.2% 60.2%'), // #EF4444
      hslTripletToRgb('210 40% 98%') // #F8FAFC
    )
    expect(bad).toBeLessThan(AA_NORMAL_TEXT)
    expect(bad).toBeGreaterThan(3.4)
    expect(bad).toBeLessThan(3.8)
    // golden-good: #DC2626 on #FFFFFF is the known-passing light pair ~4.83:1.
    const good = contrastRatio(
      hslTripletToRgb('0 72.2% 50.6%'), // #DC2626
      hslTripletToRgb('0 0% 100%') // #FFFFFF
    )
    expect(good).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
  })

  it('light theme destructive pair meets AA (>= 4.5:1)', () => {
    const ratio = destructiveContrast(':root')
    expect(ratio).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
  })

  it('dark theme destructive pair meets AA (>= 4.5:1)', () => {
    const ratio = destructiveContrast('.dark')
    expect(ratio).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
  })
})
