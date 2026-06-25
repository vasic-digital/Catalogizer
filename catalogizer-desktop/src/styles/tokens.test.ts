import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { light, dark, typography, radius, cssVarNames, ColorTheme } from "./tokens";

/**
 * Token-source drift test (anti-bluff §11.4.107(10) / §1.1).
 *
 * tokens.css is the runtime source of truth; tokens.ts is the typed mirror the
 * app code reasons about. This test parses tokens.css and asserts every value
 * declared in tokens.ts equals the value the CSS actually ships — so the two
 * sources cannot silently drift. It is self-validating: corrupting a single
 * triplet in either file makes a concrete assertion FAIL (proven below by the
 * "self-validation" block, which feeds the comparator a deliberately wrong
 * value and asserts it is rejected).
 */

const here = dirname(fileURLToPath(import.meta.url));
const cssText = readFileSync(join(here, "tokens.css"), "utf8");

/** Extract the body of a `:root { … }` / `.dark { … }` block from the CSS. */
function selectorBlock(selector: string): string {
  // Match `:root {` or `.dark {` and capture until the first closing brace
  // that ends the rule. tokens.css has no nested braces inside these blocks.
  const re = new RegExp(
    `${selector.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\s*\\{([^}]*)\\}`
  );
  const m = cssText.match(re);
  if (!m) throw new Error(`selector block not found: ${selector}`);
  return m[1];
}

/** Read a single `--var: H S% L%;` declaration's value from a CSS block body. */
function cssVar(blockBody: string, varName: string): string {
  const re = new RegExp(`${varName}\\s*:\\s*([^;]+);`);
  const m = blockBody.match(re);
  if (!m) throw new Error(`css var not found: ${varName}`);
  // Strip any trailing inline comment and surrounding whitespace.
  return m[1].replace(/\/\*.*$/, "").trim();
}

const rootBlock = selectorBlock(":root");
const darkBlock = selectorBlock(".dark");

/** Flatten the cssVarNames tree into [path, varName] pairs. */
function flatten(
  obj: Record<string, any>,
  prefix: string[] = []
): Array<[string[], string]> {
  const out: Array<[string[], string]> = [];
  for (const [k, v] of Object.entries(obj)) {
    if (typeof v === "string") out.push([[...prefix, k], v]);
    else out.push(...flatten(v, [...prefix, k]));
  }
  return out;
}

/** Resolve a dotted path (e.g. ["brand","primary"]) against a ColorTheme. */
function resolve(theme: ColorTheme, path: string[]): string {
  return path.reduce<any>((acc, key) => acc[key], theme as any);
}

const varPairs = flatten(cssVarNames);

describe("Catalogizer Blue tokens — tokens.ts mirrors tokens.css", () => {
  it("declares 15 color token keys per theme", () => {
    expect(varPairs).toHaveLength(15);
  });

  describe("light theme (:root) values match tokens.ts.light", () => {
    for (const [path, varName] of varPairs) {
      it(`${path.join(".")} (${varName})`, () => {
        const cssValue = cssVar(rootBlock, varName);
        const tsValue = resolve(light, path);
        expect(cssValue).toBe(tsValue);
      });
    }
  });

  describe("dark theme (.dark) values match tokens.ts.dark", () => {
    for (const [path, varName] of varPairs) {
      it(`${path.join(".")} (${varName})`, () => {
        const cssValue = cssVar(darkBlock, varName);
        const tsValue = resolve(dark, path);
        expect(cssValue).toBe(tsValue);
      });
    }
  });

  it("brand primary is the Catalogizer Blue, not the legacy slate", () => {
    // Legacy shadcn slate default was `222.2 47.4% 11.2%` (#0F172A).
    // The whole point of this work is that desktop primary is now blue.
    expect(light.brand.primary).toBe("211.9 80.3% 41.8%"); // #1565C0
    expect(dark.brand.primary).toBe("212.8 100% 81%"); // #9ECAFF
    expect(light.brand.primary).not.toBe("222.2 47.4% 11.2%");
  });

  it("ring equals brand primary in both themes (focus = brand)", () => {
    expect(light.state.ring).toBe(light.brand.primary);
    expect(dark.state.ring).toBe(dark.brand.primary);
  });

  it("typography declares Inter (sans) + JetBrains Mono (mono)", () => {
    expect(typography.fontFamily.sans[0]).toBe("Inter");
    expect(typography.fontFamily.mono[0]).toBe("JetBrains Mono");
  });

  it("radius base is 0.75rem (unified with the web app)", () => {
    expect(radius.base).toBe("0.75rem");
    expect(cssVar(rootBlock, "--cz-radius-base")).toBe("0.75rem");
  });

  it("typography vars are present in tokens.css :root", () => {
    expect(cssVar(rootBlock, "--cz-font-sans")).toContain("Inter");
    expect(cssVar(rootBlock, "--cz-font-mono")).toContain("JetBrains Mono");
  });
});

describe("self-validation (anti-bluff §11.4.107(10) / §1.1)", () => {
  it("the comparator FAILs when the CSS value is mutated", () => {
    // Prove the test is not a tautology: a deliberately-wrong CSS body must
    // not match the canonical tokens.ts value.
    const mutatedBlock = rootBlock.replace(
      cssVar(rootBlock, "--cz-brand-primary"),
      "0 0% 0%"
    );
    const mutated = cssVar(mutatedBlock, "--cz-brand-primary");
    expect(mutated).toBe("0 0% 0%");
    expect(mutated).not.toBe(light.brand.primary);
  });
});
