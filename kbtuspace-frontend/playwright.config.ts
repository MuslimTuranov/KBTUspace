import { defineConfig, devices } from '@playwright/test';

const port = Number(process.env.E2E_FRONTEND_PORT ?? 4173);
const webServerUrl = process.env.E2E_BASE_URL ?? `http://127.0.0.1:${port}`;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  timeout: 45_000,
  expect: { timeout: 10_000 },
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: webServerUrl,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: `npm.cmd run dev -- --host 127.0.0.1 --port ${port}`,
        url: webServerUrl,
        reuseExistingServer: true,
        timeout: 120_000,
        env: {
          ...process.env,
          VITE_API_URL: process.env.E2E_API_URL ?? process.env.VITE_API_URL ?? 'http://127.0.0.1:8080/api/v1',
        },
      },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile-chrome', use: { ...devices['Pixel 5'] }, grep: /@mobile|@smoke/ },
  ],
});
