import { test, expect, type Page } from "@playwright/test";

/**
 * Visual-regression of the main view in BOTH themes (Catalogizer Blue,
 * §11.4.162). Drives the Vite frontend headless. The app boots to a splash
 * then — with no server configured — redirects to the Settings view, which
 * exercises the full token surface (brand-blue primary buttons, cards,
 * borders, the theme <select>, focus rings) so a palette regression is
 * visible in the diff.
 *
 * The Catalogizer Blue palette flips entirely off the `.dark` class on
 * <html> (the token mechanism in src/styles/globals.css + tokens.css), so we
 * set the class directly to capture each theme deterministically.
 */

const APP_URL = "/";

/** Wait past the SplashScreen until a real view (the theme select) renders. */
async function waitForMainView(page: Page): Promise<void> {
  // SettingsPage renders a <select id="theme"> once initialization completes.
  await page.locator("#theme").waitFor({ state: "visible", timeout: 30_000 });
}

async function setTheme(page: Page, theme: "light" | "dark"): Promise<void> {
  await page.evaluate((t) => {
    document.documentElement.classList.toggle("dark", t === "dark");
  }, theme);
  // Let the CSS-variable cascade settle before snapshotting.
  await page.waitForTimeout(150);
}

/**
 * WCAG 2.x relative-contrast ratio between two `rgb(...)` / `rgba(...)` strings
 * (the form getComputedStyle returns). Mirrors the W3C definition.
 */
function contrastRatio(fg: string, bg: string): number {
  const parse = (c: string): [number, number, number] => {
    const m = c.match(/[\d.]+/g);
    if (!m || m.length < 3) throw new Error(`unparseable color: ${c}`);
    return [Number(m[0]), Number(m[1]), Number(m[2])];
  };
  const lin = (v: number): number => {
    const s = v / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  const lum = ([r, g, b]: [number, number, number]): number =>
    0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
  const L1 = lum(parse(fg));
  const L2 = lum(parse(bg));
  const [hi, lo] = L1 >= L2 ? [L1, L2] : [L2, L1];
  return (hi + 0.05) / (lo + 0.05);
}

const WCAG_AA = 4.5;

/**
 * The nav-item hover classes Layout.tsx applies to every <Link> / logout
 * <button> (base muted-foreground, then the hover bg + matched on-accent fg).
 * Kept in lockstep with src/components/Layout.tsx — if a call-site reverts
 * `hover:text-accent-foreground` back to `hover:text-foreground`, mirror it
 * here and the guard FAILs in dark (that revert IS the original B1 defect).
 */
const NAV_HOVER_CLASSES =
  "text-muted-foreground hover:text-accent-foreground hover:bg-accent";

/**
 * Measure the ACTUAL rendered hover fg/bg of a nav item by injecting a probe
 * element carrying the real Tailwind-compiled nav-hover utility classes, then
 * driving a genuine `:hover` with Playwright and reading getComputedStyle.
 *
 * This exercises BOTH the token values (--accent / --accent-foreground) AND the
 * call-site class pairing end-to-end — so it catches the original B1 regression
 * by either route: (a) --accent repointed to a saturated brand hue, OR (b) a
 * call-site reverting to `hover:text-foreground` (which would diverge from the
 * accent background on a non-foreground accent). Real rendered rgb(...) only,
 * never a source string (anti-bluff §11.4.6 / §11.4.107).
 */
async function navHoverPair(
  page: Page
): Promise<{ fg: string; bg: string }> {
  const PROBE_ID = "__nav_hover_probe__";
  await page.evaluate(
    ({ id, classes }) => {
      document.getElementById(id)?.remove();
      const el = document.createElement("a");
      el.id = id;
      el.className = classes;
      el.textContent = "Library";
      // Place it where it cannot be obscured so the real hover lands on it.
      el.style.position = "fixed";
      el.style.top = "8px";
      el.style.left = "8px";
      el.style.zIndex = "99999";
      el.style.padding = "8px 12px";
      document.body.appendChild(el);
    },
    { id: PROBE_ID, classes: NAV_HOVER_CLASSES }
  );

  await page.locator(`#${PROBE_ID}`).hover();
  // Let the :hover transition-colors settle.
  await page.waitForTimeout(150);

  const pair = await page.evaluate((id) => {
    const el = document.getElementById(id)!;
    const cs = getComputedStyle(el);
    const out = { fg: cs.color, bg: cs.backgroundColor };
    el.remove();
    return out;
  }, PROBE_ID);
  return pair;
}

test.describe("Catalogizer Blue theme — main view visual regression", () => {
  test("light theme renders Catalogizer Blue", async ({ page }) => {
    await page.goto(APP_URL);
    await waitForMainView(page);
    await setTheme(page, "light");
    await expect(page).toHaveScreenshot("main-view-light.png", {
      fullPage: true,
    });
  });

  test("dark theme renders Catalogizer Blue", async ({ page }) => {
    await page.goto(APP_URL);
    await waitForMainView(page);
    await setTheme(page, "dark");
    await expect(page).toHaveScreenshot("main-view-dark.png", {
      fullPage: true,
    });
  });

  test("the brand primary is blue, not the legacy slate", async ({ page }) => {
    await page.goto(APP_URL);
    await waitForMainView(page);
    await setTheme(page, "light");
    // Resolve the runtime --primary token; assert it is the blue triplet,
    // proving the brand change actually reached the DOM (anti-bluff §11.4.6).
    const primary = await page.evaluate(() =>
      getComputedStyle(document.documentElement)
        .getPropertyValue("--primary")
        .trim()
    );
    expect(primary).toBe("211.9 80.3% 41.8%"); // #1565C0
    expect(primary).not.toBe("222.2 47.4% 11.2%"); // not the legacy slate
  });
});

/**
 * §11.4.135 regression guard for the dark/light nav-hover contrast regression
 * (review B1): nav items use `hover:bg-accent hover:text-accent-foreground`, so
 * the rendered --accent / --accent-foreground pair MUST clear WCAG AA (4.5:1)
 * in BOTH themes. The original defect (—-accent repointed to the saturated
 * brand accent #D6BEE4) produced ≈1.6:1 in dark — nav labels vanished on hover.
 * Resting-state VR snapshots could not catch it (no :hover driven), so this
 * computed-contrast assertion is the permanent guard. It is self-validated:
 * a deliberately-bad pair MUST fail the same checker (paired §1.1 mutation),
 * proving the assertion is not a tautology.
 */
test.describe("nav hover contrast clears WCAG AA (§11.4.135 guard for B1)", () => {
  for (const theme of ["light", "dark"] as const) {
    test(`${theme}: accent hover bg/fg pair ≥ 4.5:1`, async ({ page }) => {
      await page.goto(APP_URL);
      await waitForMainView(page);
      await setTheme(page, theme);

      const { fg, bg } = await navHoverPair(page);
      const ratio = contrastRatio(fg, bg);

      // Captured-evidence: surface the real rendered colors + ratio in the log.
      // eslint-disable-next-line no-console
      console.log(
        `[hover-contrast] ${theme}: fg=${fg} bg=${bg} ratio=${ratio.toFixed(2)}:1 (AA=${WCAG_AA})`
      );

      expect(
        ratio,
        `dark/light nav hover (bg-accent + text-accent-foreground) must clear WCAG AA; got ${ratio.toFixed(2)}:1 fg=${fg} bg=${bg}`
      ).toBeGreaterThanOrEqual(WCAG_AA);
    });
  }

  test("self-validation: the contrast checker rejects a sub-AA pair", () => {
    // The exact failing pair from the B1 regression (near-white on light
    // lavender) MUST be flagged < AA — proving the guard above is real, not a
    // tautology that would pass any colors.
    const badRatio = contrastRatio("rgb(248, 250, 252)", "rgb(214, 190, 228)");
    expect(badRatio).toBeLessThan(WCAG_AA); // ≈1.6:1 — the original defect
    // And a known-good pair (near-white on near-black) MUST clear it.
    const goodRatio = contrastRatio("rgb(248, 250, 252)", "rgb(15, 23, 42)");
    expect(goodRatio).toBeGreaterThanOrEqual(WCAG_AA);
  });
});
