import { test, expect } from '@playwright/test';

/**
 * Basic health check tests to verify the E2E test server is running correctly.
 */

test.describe('Health Check', () => {
  test('API health endpoint returns ok status', async ({ request }) => {
    const response = await request.get('/api/health');

    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    // Health can be "ok" or "degraded" depending on database connection
    expect(['ok', 'degraded']).toContain(body.status);
  });

  test('home page loads successfully', async ({ page }) => {
    await page.goto('/');

    // Verify the page loads with the main content
    // The app is an SPA, so we check for the sidebar navigation
    await expect(page.locator('text=Trade Machine')).toBeVisible();
    await expect(page.locator('text=AI-Powered Trading')).toBeVisible();
  });

  test('settings section is accessible via navigation', async ({ page }) => {
    await page.goto('/');

    // Click the Settings link in the sidebar
    await page.click('a[data-section="settings"]');

    // Wait for the settings section to become visible
    // The content is loaded via HTMX after clicking
    await expect(page.locator('#settings')).toBeVisible();
    // The header says "API Configuration" not "Settings"
    await expect(page.locator('#settings h3')).toContainText('API Configuration');
  });
});
