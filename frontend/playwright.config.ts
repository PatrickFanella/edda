import { defineConfig } from '@playwright/test';
const port = Number(process.env.EDDA_FIXTURE_PORT ?? 4175);
const baseURL = `http://127.0.0.1:${port}`;
export default defineConfig({
  testDir: './tests/browser', fullyParallel: true, forbidOnly: !!process.env.CI,
  retries: 0, use: { baseURL, viewport: { width: 1440, height: 1000 }, trace: 'retain-on-failure' },
  webServer: { command: `pnpm dev --host 127.0.0.1 --port ${port} --strictPort`, url: baseURL, reuseExistingServer: false },
});
