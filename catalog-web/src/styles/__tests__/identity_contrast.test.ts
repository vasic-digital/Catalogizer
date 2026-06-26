/**
 * WCAG 2.1 AA contrast tests for the Identity Manager + Discovered Shares UI.
 *
 * §11.4.162 — every text foreground token used against a background surface
 * must meet AA normal-text 4.5:1 in both light and dark themes. The
 * Catalogizer Blue palette is already proven AA by tokens.test.ts and
 * wcag_contrast.test.ts; this test verifies that the IDENTITY-RELATED
 * semantic token pairs we rely on also pass.
 *
 * §11.4.107(10) — self-validated analyzer: the contrast oracle is proven
 * by a golden-bad fixture that MUST FAIL.
 */

import { describe, it, expect } from 'vitest'

const AA_NORMAL_TEXT = 4.5

function relativeLuminance([r, g, b]: [number, number, number]): number {
  const lin = (v: number): number => {
    const cc = v / 255
    return cc <= 0.03928 ? cc / 12.92 : Math.pow((cc + 0.055) / 1.055, 2.4)
  }
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b)
}

function contrastRatio(
  a: [number, number, number],
  b: [number, number, number]
): number {
  const la = relativeLuminance(a)
  const lb = relativeLuminance(b)
  const [hi, lo] = la >= lb ? [la, lb] : [lb, la]
  return (hi + 0.05) / (lo + 0.05)
}

function hslToRgb(triplet: string): [number, number, number] {
  const m = triplet.match(/^([\d.]+)\s+([\d.]+)%\s+([\d.]+)%$/)
  if (!m) throw new Error(`Cannot parse HSL triplet: "${triplet}"`)
  const h = parseFloat(m[1])
  const s = parseFloat(m[2]) / 100
  const l = parseFloat(m[3]) / 100
  const c = (1 - Math.abs(2 * l - 1)) * s
  const hp = h / 60
  const x = c * (1 - Math.abs((hp % 2) - 1))
  let r1 = 0, g1 = 0, b1 = 0
  if (hp >= 0 && hp < 1) [r1, g1, b1] = [c, x, 0]
  else if (hp < 2) [r1, g1, b1] = [x, c, 0]
  else if (hp < 3) [r1, g1, b1] = [0, c, x]
  else if (hp < 4) [r1, g1, b1] = [0, x, c]
  else if (hp < 5) [r1, g1, b1] = [x, 0, c]
  else [r1, g1, b1] = [c, 0, x]
  const mm = l - c / 2
  return [(r1 + mm) * 255, (g1 + mm) * 255, (b1 + mm) * 255]
}

// Known-passing palette pairs from tokens.ts (light / dark)
const PAIRS: Array<[string, string, string, string, string, string]> = [
  // (label, light surface, light fg, dark surface, dark fg)
  ['default surface↔fg',    '0 0% 100%', '222.9 84% 4.9%',   '222.9 84% 4.9%', '210 40% 98%'],
  ['card surface↔fg',       '0 0% 100%', '222.9 84% 4.9%',   '222.9 84% 4.9%', '210 40% 98%'],
  ['primary↔on-primary',    '211.9 80.3% 41.8%', '0 0% 100%', '212.8 100% 81%', '205.9 100% 17.3%'],
  ['secondary surface↔fg',  '210 40% 98%', '222.9 84% 4.9%',  '222.2 47.4% 11.2%', '210 40% 98%'],
  ['muted surface↔fg',      '210 40% 98%', '215.4 16.3% 46.9%', '222.2 47.4% 11.2%', '215 20.2% 65.1%'],
  ['accent↔on-accent',      '276.4 15.9% 40.6%', '0 0% 100%', '277.9 41.3% 82%', '275.2 42% 13.5%'],
]

describe('Identity UI — WCAG AA contrast for semantic pairs', () => {
  for (const [label, lightSurface, lightFg, darkSurface, darkFg] of PAIRS) {
    it(`light theme ${label} meets AA >= 4.5:1`, () => {
      const ratio = contrastRatio(hslToRgb(lightSurface), hslToRgb(lightFg))
      expect(ratio).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
    })

    it(`dark theme ${label} meets AA >= 4.5:1`, () => {
      const ratio = contrastRatio(hslToRgb(darkSurface), hslToRgb(darkFg))
      expect(ratio).toBeGreaterThanOrEqual(AA_NORMAL_TEXT)
    })
  }

  it('self-validation: golden-bad pair FAILs below 4.5:1', () => {
    // #EF4444 on #F8FAFC (historical defect, ~3.60:1)
    const bad = contrastRatio(
      hslToRgb('0 84.2% 60.2%'),
      hslToRgb('210 40% 98%')
    )
    expect(bad).toBeLessThan(AA_NORMAL_TEXT)
    expect(bad).toBeGreaterThan(3.4)
    expect(bad).toBeLessThan(3.8)
  })
})
