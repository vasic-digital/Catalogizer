import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

/**
 * WCAG-AA contrast oracle for the shadcn `destructive` button pair
 * (anti-bluff §11.4.115 RED-baseline / §11.4.107(10) self-validated analyzer).
 *
 * The pre-existing tokens.test.ts only proves tokens.ts ↔ tokens.css do not
 * drift — it never computes a CONTRAST RATIO, which is exactly the gap that let
 * the dark `bg-destructive text-destructive-foreground` pair ship at ~3.60:1
 * (below the WCAG-AA 4.5:1 floor for normal text). This oracle closes that gap.
 *
 * It is computed end-to-end on the ACTUAL shipped tokens — it parses
 * globals.css for the `--destructive` / `--destructive-foreground` mappings,
 * follows each `var(--cz-*)` reference into tokens.css, converts the HSL
 * triplet to sRGB, linearises, computes relative luminance
 * (L = 0.2126·R + 0.7152·G + 0.0722·B) and the WCAG ratio (L1+0.05)/(L2+0.05),
 * then asserts >= 4.5 for BOTH the light (:root) and dark (.dark) themes.
 * Nothing is hardcoded; editing a token changes the computed result.
 */

const here = dirname(fileURLToPath(import.meta.url));
const globalsText = readFileSync(join(here, "globals.css"), "utf8");
const tokensText = readFileSync(join(here, "tokens.css"), "utf8");

const WCAG_AA_NORMAL = 4.5;

/** Body of the FIRST `<selector> { … }` block (no nested braces in these files). */
function block(css: string, selector: string): string {
  const re = new RegExp(
    `${selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\s*\\{([^}]*)\\}`
  );
  const m = css.match(re);
  if (!m) throw new Error(`selector block not found: ${selector}`);
  return m[1];
}

/** Read `--var: <value>;` from a block body, stripping inline comments. */
function cssVar(blockBody: string, varName: string): string {
  const re = new RegExp(`${varName}\\s*:\\s*([^;]+);`);
  const m = blockBody.match(re);
  if (!m) throw new Error(`css var not found: ${varName}`);
  return m[1].replace(/\/\*.*$/, "").trim();
}

/**
 * Resolve a shadcn variable (e.g. `--destructive`) declared in globals.css for
 * a given theme to its final `H S% L%` triplet by following the single
 * `var(--cz-*)` indirection into tokens.css.
 */
function resolveTriplet(themeSelector: string, shadcnVar: string): string {
  const globalsBlock = block(globalsText, themeSelector);
  const raw = cssVar(globalsBlock, shadcnVar); // e.g. "var(--cz-semantic-error)"
  const refMatch = raw.match(/var\(\s*(--[a-z0-9-]+)\s*\)/i);
  if (!refMatch) {
    // A direct triplet (no indirection) is allowed too.
    return raw;
  }
  const czVar = refMatch[1];
  const tokensBlock = block(tokensText, themeSelector);
  return cssVar(tokensBlock, czVar);
}

/** Parse a bare `H S% L%` HSL triplet into [h, s, l] numbers (s,l in 0..1). */
function parseHsl(triplet: string): [number, number, number] {
  const m = triplet.trim().match(/^([\d.]+)\s+([\d.]+)%\s+([\d.]+)%$/);
  if (!m) throw new Error(`not an HSL triplet: "${triplet}"`);
  return [parseFloat(m[1]), parseFloat(m[2]) / 100, parseFloat(m[3]) / 100];
}

/** HSL → sRGB (each channel 0..1). Standard CSS Color 3 conversion. */
function hslToRgb(h: number, s: number, l: number): [number, number, number] {
  const c = (1 - Math.abs(2 * l - 1)) * s;
  const hp = h / 60;
  const x = c * (1 - Math.abs((hp % 2) - 1));
  let r = 0,
    g = 0,
    b = 0;
  if (hp >= 0 && hp < 1) [r, g, b] = [c, x, 0];
  else if (hp < 2) [r, g, b] = [x, c, 0];
  else if (hp < 3) [r, g, b] = [0, c, x];
  else if (hp < 4) [r, g, b] = [0, x, c];
  else if (hp < 5) [r, g, b] = [x, 0, c];
  else [r, g, b] = [c, 0, x];
  const mlt = l - c / 2;
  return [r + mlt, g + mlt, b + mlt];
}

/** WCAG relative luminance from sRGB channels (0..1). */
function relativeLuminance([r, g, b]: [number, number, number]): number {
  const lin = (ch: number) =>
    ch <= 0.03928 ? ch / 12.92 : Math.pow((ch + 0.055) / 1.055, 2.4);
  const [R, G, B] = [lin(r), lin(g), lin(b)];
  return 0.2126 * R + 0.7152 * G + 0.0722 * B;
}

/** WCAG contrast ratio between two `H S% L%` triplets. */
function contrastRatio(tripletA: string, tripletB: string): number {
  const la = relativeLuminance(hslToRgb(...parseHsl(tripletA)));
  const lb = relativeLuminance(hslToRgb(...parseHsl(tripletB)));
  const [hi, lo] = la >= lb ? [la, lb] : [lb, la];
  return (hi + 0.05) / (lo + 0.05);
}

/** Convenience: resolved destructive ↔ destructive-foreground ratio per theme. */
function destructiveRatio(themeSelector: string): number {
  const bg = resolveTriplet(themeSelector, "--destructive");
  const fg = resolveTriplet(themeSelector, "--destructive-foreground");
  return contrastRatio(bg, fg);
}

/**
 * Generic resolved surface↔foreground ratio for ANY shadcn pair in a theme —
 * the §11.4.146 STEP-3 extend of `destructiveRatio` to the whole palette. Each
 * var is resolved through its `var(--cz-*)` indirection in the SAME theme, so
 * light/dark each pick up their own tokens.
 */
function pairRatio(
  themeSelector: string,
  surfaceVar: string,
  fgVar: string
): number {
  const bg = resolveTriplet(themeSelector, surfaceVar);
  const fg = resolveTriplet(themeSelector, fgVar);
  return contrastRatio(bg, fg);
}

/**
 * §11.4.118 discovery-pressure / §11.4.146 STEP-3 extend-to-all-cases.
 *
 * Every shadcn foreground/background semantic pair that globals.css defines AND
 * that renders NORMAL BODY TEXT (`text-<x>-foreground` on `bg-<x>`) → WCAG-AA
 * normal-text 4.5:1. globals.css declares no `--popover` pair (so it is absent
 * here), and unlike the web client maps `--accent` to the NEUTRAL hover surface
 * (`--cz-neutral-surface`/`--cz-neutral-foreground`) rather than the saturated
 * brand accent — the audited ratio reflects that real shipped mapping. This
 * palette has NO large-text-only pairs, so every pair below is 4.5:1 (§11.4.6 —
 * classification stated, not blanket-applied). Decorative `--border` / `--input`
 * / `--ring` carry no text and are EXCLUDED (N/A: WCAG 1.4.11 non-text is a
 * separate, out-of-scope concern). `destructive` retains its dedicated tests
 * above and is re-covered here for whole-palette completeness.
 */
const AA_TEXT_PAIRS: ReadonlyArray<[string, string]> = [
  ["--background", "--foreground"],
  ["--card", "--card-foreground"],
  ["--primary", "--primary-foreground"],
  ["--secondary", "--secondary-foreground"],
  ["--muted", "--muted-foreground"],
  ["--accent", "--accent-foreground"],
  ["--destructive", "--destructive-foreground"],
];

describe("WCAG-AA destructive contrast (real shipped tokens)", () => {
  it("light theme (:root) destructive pair clears 4.5:1", () => {
    const ratio = destructiveRatio(":root");
    expect(ratio).toBeGreaterThanOrEqual(WCAG_AA_NORMAL);
  });

  it("dark theme (.dark) destructive pair clears 4.5:1", () => {
    const ratio = destructiveRatio(".dark");
    expect(ratio).toBeGreaterThanOrEqual(WCAG_AA_NORMAL);
  });
});

describe("WCAG-AA whole semantic palette (real shipped tokens, §11.4.146/§11.4.118)", () => {
  for (const [theme, themeName] of [
    [":root", "light"],
    [".dark", "dark"],
  ] as const) {
    for (const [surfaceVar, fgVar] of AA_TEXT_PAIRS) {
      it(`${themeName} ${surfaceVar}↔${fgVar} clears AA normal-text 4.5:1`, () => {
        const ratio = pairRatio(theme, surfaceVar, fgVar);
        expect(ratio).toBeGreaterThanOrEqual(WCAG_AA_NORMAL);
      });
    }
  }
});

describe("self-validation of the contrast analyzer (§11.4.107(10))", () => {
  it("golden-good: black vs white computes ~21:1", () => {
    const ratio = contrastRatio("0 0% 0%", "0 0% 100%");
    expect(ratio).toBeCloseTo(21, 1);
  });

  it("golden-good: identical colors compute exactly 1:1", () => {
    expect(contrastRatio("211.9 80.3% 41.8%", "211.9 80.3% 41.8%")).toBeCloseTo(
      1,
      5
    );
  });

  it("golden-bad: the historical broken dark pair (#EF4444 on #F8FAFC) is ~3.60:1 and FAILS the AA floor", () => {
    // #EF4444 = 0 84.2% 60.2%   ;  #F8FAFC = 210 40% 98%
    const ratio = contrastRatio("0 84.2% 60.2%", "210 40% 98%");
    expect(ratio).toBeCloseTo(3.6, 1);
    expect(ratio).toBeLessThan(WCAG_AA_NORMAL); // proves the oracle would reject it
  });
});
