import { defineConfig, devices } from "@playwright/test";

/**
 * Visual-regression scaffold for catalogizer-desktop (§11.4.162 mandates
 * visual-regression coverage for UI changes). Drives the Vite frontend
 * (`npm run dev`, :1420) headless with Chromium and snapshots the main view
 * in BOTH light and dark themes (Catalogizer Blue).
 *
 * Tauri-window VR is operator-attended where headless is infeasible
 * (SKIP-with-reason per §11.4.3) — this scaffold covers the React frontend,
 * which is where the token/theme change lives.
 *
 * First run (baseline): `npx playwright test --update-snapshots`.
 * Subsequent runs perceptual-diff against the committed baseline; PASS only
 * within `maxDiffPixelRatio` tolerance.
 */
export default defineConfig({
  testDir: "./e2e",
  // Deterministic snapshots: single worker, no parallelism, no retries.
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: "http://localhost:1420",
    // Fixed viewport so screenshots are reproducible across machines.
    viewport: { width: 1280, height: 800 },
  },
  expect: {
    // Allow tiny anti-aliasing/font-rendering differences across hosts.
    toHaveScreenshot: { maxDiffPixelRatio: 0.02 },
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  // Boot the Vite dev server on demand for the VR run.
  webServer: {
    command: "npm run dev",
    url: "http://localhost:1420",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
