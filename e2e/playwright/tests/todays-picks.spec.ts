import { test, expect, Page, APIRequestContext } from '@playwright/test';

/**
 * E2E tests for the Today's Picks page.
 *
 * These tests verify the screener functionality including:
 * - Display when screener is configured with mock services
 * - Running the screener and viewing results
 * - Navigation to the picks section
 *
 * Note: The e2e-server runs with E2E_ENABLE_MOCKS=true by default,
 * which provides mock FMP, Alpaca, and PortfolioManager services.
 */

/**
 * Helper function to reset all settings before a test.
 * This ensures test isolation by calling the e2e reset endpoint.
 */
async function resetSettings(request: APIRequestContext) {
  await request.post('/api/e2e/reset-settings');
}

/**
 * Helper function to navigate to the Today's Picks section and wait for content.
 * The app uses JavaScript-based navigation, so we click the sidebar link
 * and wait for the HTMX content to load.
 */
async function navigateToPicks(page: Page) {
  await page.goto('/');

  // Click the Today's Picks link in the sidebar
  await page.click('a[data-section="picks"]');

  // Wait for the picks section to be visible
  await expect(page.locator('#picks')).toBeVisible();

  // Wait for the picks content to be loaded via HTMX
  await expect(page.locator('#picks-content').first()).toBeVisible({ timeout: 10000 });

  // Wait for HTMX to finish loading by checking for a button that appears in all states
  // (both empty state and results state have a "Find Value Stocks" button)
  await expect(page.locator('#picks-content button:has-text("Find Value Stocks")').first()).toBeVisible({ timeout: 10000 });
}

test.describe('Today\'s Picks Page', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetSettings(request);
  });

  test('should display the Today\'s Picks section on page load', async ({ page }) => {
    await page.goto('/');

    // Today's Picks is the default section, should be visible
    await expect(page.locator('#picks')).toBeVisible();

    // Should have the correct header
    await expect(page.locator('#picks h3').filter({ hasText: "Today's Value Picks" })).toBeVisible();
  });

  test('should have correct navigation in sidebar', async ({ page }) => {
    await page.goto('/');

    // Verify the Today's Picks link exists with correct icon
    const picksLink = page.locator('a[data-section="picks"]');
    await expect(picksLink).toBeVisible();
    await expect(picksLink.locator('.bi-gem')).toBeVisible();
    await expect(picksLink).toContainText("Today's Picks");
  });

  test('should navigate to picks section when clicking sidebar link', async ({ page }) => {
    await page.goto('/');

    // First navigate to settings to change context
    await page.click('a[data-section="settings"]');
    await expect(page.locator('#settings')).toHaveClass(/active/);

    // Now navigate back to picks
    await page.click('a[data-section="picks"]');
    await expect(page.locator('#picks')).toHaveClass(/active/);
  });
});

test.describe('Today\'s Picks - Screener Configured', () => {
  // These tests run with mock services enabled (E2E_ENABLE_MOCKS=true)

  test.beforeEach(async ({ page, request }) => {
    await resetSettings(request);
  });

  test('should show picks page with Find Value Stocks button', async ({ page }) => {
    await navigateToPicks(page);

    // Should have a "Find Value Stocks" button (appears in both empty state and results header)
    await expect(page.locator('#picks-content button').filter({ hasText: 'Find Value Stocks' }).first()).toBeVisible();

    // Should show either "No Picks Yet" (empty state) or results with stock symbols
    // This depends on whether previous tests ran the screener
    const hasEmptyState = await page.locator('#picks-content').getByText('No Picks Yet').isVisible().catch(() => false);
    const hasResults = await page.locator('#picks-content').getByText('JNJ').isVisible().catch(() => false);

    // One of these should be true
    expect(hasEmptyState || hasResults).toBe(true);
  });

  test('should have Find Value Stocks button in header', async ({ page }) => {
    await navigateToPicks(page);

    // The header should have a Find Value Stocks button (the first one, not the btn-lg one in empty state)
    const headerButton = page.locator('#picks-content button.btn-primary:not(.btn-lg)').filter({ hasText: 'Find Value Stocks' });
    await expect(headerButton).toBeVisible();
  });

  test('should run screener and display results', async ({ page }) => {
    await navigateToPicks(page);

    // Click the "Find Value Stocks" button
    const findButton = page.locator('#picks-content button').filter({ hasText: 'Find Value Stocks' }).first();
    await expect(findButton).toBeVisible();

    // Set up response listener before clicking
    const runResponsePromise = page.waitForResponse(
      response => response.url().includes('/api/screener/run') && response.status() === 200,
      { timeout: 60000 } // Screener can take a while
    );

    await findButton.click();

    // Wait for the screener to complete
    await runResponsePromise;

    // Should display results with stock picks
    // The mock FMP service returns JNJ, PG, KO
    await expect(page.locator('#picks-content').getByText('JNJ')).toBeVisible({ timeout: 10000 });
  });

  test('should display pick cards with company info after running screener', async ({ page }) => {
    await navigateToPicks(page);

    // Run the screener
    const findButton = page.locator('#picks-content button').filter({ hasText: 'Find Value Stocks' }).first();

    const runResponsePromise = page.waitForResponse(
      response => response.url().includes('/api/screener/run') && response.status() === 200,
      { timeout: 60000 }
    );

    await findButton.click();
    await runResponsePromise;

    // Should show pick cards
    // Look for company names from our mock data
    const picksContent = page.locator('#picks-content');

    // At least one of our mock stocks should appear
    const hasJNJ = await picksContent.getByText('Johnson & Johnson').isVisible().catch(() => false);
    const hasPG = await picksContent.getByText('Procter & Gamble').isVisible().catch(() => false);
    const hasKO = await picksContent.getByText('Coca-Cola').isVisible().catch(() => false);

    expect(hasJNJ || hasPG || hasKO).toBe(true);
  });

  test('should display summary stats after running screener', async ({ page }) => {
    await navigateToPicks(page);

    // Run the screener
    const findButton = page.locator('#picks-content button').filter({ hasText: 'Find Value Stocks' }).first();

    const runResponsePromise = page.waitForResponse(
      response => response.url().includes('/api/screener/run') && response.status() === 200,
      { timeout: 60000 }
    );

    await findButton.click();
    await runResponsePromise;

    // Should show summary stats
    await expect(page.locator('#picks-content').getByText('Stocks Screened')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#picks-content').getByText('Top Picks')).toBeVisible();
  });
});

test.describe('Today\'s Picks - API Endpoints (Mocks Enabled)', () => {
  test('should return 200 from /api/screener/picks when configured', async ({ request }) => {
    const response = await request.get('/api/screener/picks');

    // With mocks enabled, screener is configured, so should return 200
    // (even if no runs exist yet, it returns an empty state, not an error)
    expect(response.status()).toBe(200);
  });

  test('should return 200 from /api/screener/run when configured', async ({ request }) => {
    const response = await request.post('/api/screener/run');

    // With mocks enabled, screener is configured and should run successfully
    expect(response.status()).toBe(200);

    const body = await response.json();
    // Should have screener run data
    expect(body).toHaveProperty('id');
    expect(body).toHaveProperty('status');
  });

  test('should return screener run with candidates after running', async ({ request }) => {
    // Run the screener
    const runResponse = await request.post('/api/screener/run');
    expect(runResponse.status()).toBe(200);

    const runData = await runResponse.json();
    expect(runData.status).toBe('completed');

    // Should have candidates from mock FMP data
    expect(runData.candidates).toBeDefined();
    expect(runData.candidates.length).toBeGreaterThan(0);

    // Check that our mock stocks are in the results
    const symbols = runData.candidates.map((c: any) => c.symbol);
    expect(symbols).toContain('JNJ');
  });
});
