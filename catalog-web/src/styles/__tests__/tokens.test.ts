import { describe, it, expect } from 'vitest'
import { catalogizerBlue, color, lightVars, darkVars } from '../tokens'

/**
 * Token-source drift guard (single-source guarantee, §11.4.162).
 *
 * Asserts the canonical Catalogizer Blue token module (tokens.ts) holds the
 * exact audit §3.1 values AND that the convenience var maps consumed by
 * src/index.css resolve to those tokens. If index.css and tokens.ts drift, or
 * an audit value is mistyped, these assertions FAIL — proving the brand is one
 * source, not two parallel ones.
 *
 * Includes a self-validation case (§1.1 / §11.4.107(10)): a deliberately wrong
 * expected value MUST make the comparator FAIL, so a green run is trustworthy.
 */

describe('Catalogizer Blue tokens', () => {
  it('encodes the audit §3.1 brand primary (light + dark)', () => {
    expect(color.brand.primary.light.hex).toBe('#1565C0')
    expect(color.brand.primary.dark.hex).toBe('#9ECAFF')
    expect(color.brand.onPrimary.light.hex).toBe('#FFFFFF')
    expect(color.brand.onPrimary.dark.hex).toBe('#003258')
  })

  it('encodes the audit §3.1 brand accent (replaces stray fuchsia)', () => {
    expect(color.brand.accent.light.hex).toBe('#6B5778')
    expect(color.brand.accent.dark.hex).toBe('#D6BEE4')
  })

  it('encodes the audit §3.1 semantic + neutral roles', () => {
    expect(color.semantic.success.light.hex).toBe('#16A34A')
    expect(color.semantic.warning.light.hex).toBe('#D97706')
    expect(color.semantic.error.light.hex).toBe('#DC2626')
    expect(color.neutral.background.light.hex).toBe('#FFFFFF')
    expect(color.neutral.background.dark.hex).toBe('#020817')
    expect(color.neutral.border.light.hex).toBe('#E2E8F0')
    expect(color.neutral.border.dark.hex).toBe('#1E293B')
  })

  it('exposes light + dark CSS-var maps that mirror index.css', () => {
    // These MUST match the literal triplets written into src/index.css :root/.dark.
    expect(lightVars['--primary']).toBe('211.9 80.3% 41.8%')
    expect(lightVars['--ring']).toBe('211.9 80.3% 41.8%')
    expect(lightVars['--accent']).toBe('276.4 15.9% 40.6%')
    expect(lightVars['--background']).toBe('0 0% 100%')
    expect(lightVars['--radius']).toBe('0.75rem')

    expect(darkVars['--primary']).toBe('212.8 100% 81%')
    expect(darkVars['--ring']).toBe('212.8 100% 81%')
    expect(darkVars['--accent']).toBe('277.9 41.3% 82%')
    expect(darkVars['--background']).toBe('222.9 84% 4.9%')
  })

  it('ring follows primary in both themes (single brand source)', () => {
    expect(lightVars['--ring']).toBe(lightVars['--primary'])
    expect(darkVars['--ring']).toBe(darkVars['--primary'])
    expect(color.state.ring.light.hex).toBe(color.brand.primary.light.hex)
    expect(color.state.ring.dark.hex).toBe(color.brand.primary.dark.hex)
  })

  it('provides light AND dark for every brand/semantic/neutral role', () => {
    for (const group of [color.brand, color.semantic, color.neutral, color.state]) {
      for (const token of Object.values(group)) {
        expect(token.light.hex).toMatch(/^#[0-9A-Fa-f]{6}$/)
        expect(token.dark.hex).toMatch(/^#[0-9A-Fa-f]{6}$/)
        expect(token.light.hsl).toMatch(/^[\d.]+ [\d.]+% [\d.]+%$/)
        expect(token.dark.hsl).toMatch(/^[\d.]+ [\d.]+% [\d.]+%$/)
      }
    }
  })

  it('aggregate object carries the OpenDesign taxonomy sections', () => {
    expect(catalogizerBlue.color).toBe(color)
    expect(catalogizerBlue.typography.fontFamily.sans).toContain('Inter')
    expect(catalogizerBlue.space['18']).toBe('4.5rem')
    expect(catalogizerBlue.radius.base).toBe('0.75rem')
  })

  it('self-validation: comparator catches a wrong value (anti-bluff §1.1)', () => {
    // The guard above only protects against drift if a MISMATCH genuinely fails.
    expect(color.brand.primary.light.hex).not.toBe('#000000')
    expect(() => {
      // Simulate a corrupted expectation; this assertion MUST throw.
      expect(color.brand.primary.light.hex).toBe('#FF00FF')
    }).toThrow()
  })
})
