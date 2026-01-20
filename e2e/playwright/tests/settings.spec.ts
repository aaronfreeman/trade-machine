import { test, expect, Page, APIRequestContext } from '@playwright/test';

// Run settings tests serially since they share mutable state (the settings store)
test.describe.configure({ mode: 'serial' });

/**
 * E2E tests for the Settings page.
 *
 * These tests verify the full user workflow for configuring API keys
 * through the browser UI, including HTMX interactions.
 *
 * The app is an SPA where navigation happens via JavaScript. To view the
 * settings section, we need to click the Settings link in the sidebar.
 */

/**
 * Helper function to reset all settings before a test.
 * This ensures test isolation by calling the e2e reset endpoint.
 */
async function resetSettings(request: APIRequestContext) {
  await request.post('/api/e2e/reset-settings');
}

/**
 * Helper function to navigate to the settings section.
 * The app uses JavaScript-based navigation, so we click the sidebar link
 * and wait for the HTMX content to load.
 */
async function navigateToSettings(page: Page) {
  await page.goto('/');

  // Click the Settings link in the sidebar
  await page.click('a[data-section="settings"]');

  // Wait for the settings section to be visible
  await expect(page.locator('#settings')).toBeVisible();

  // Wait for the settings content to be loaded via HTMX
  // The HTMX request fires on page load (hx-trigger="load"), so we wait for the form content
  await expect(page.locator('#settings-content .card').first()).toBeVisible({ timeout: 10000 });
}

/**
 * Helper to submit a settings form and wait for HTMX response.
 * Sets up the response listener before clicking to ensure we capture it.
 */
async function submitAndWaitForResponse(page: Page, submitButton: ReturnType<typeof page.locator>) {
  const responsePromise = page.waitForResponse(
    response => response.url().includes('/api/settings/api-keys') && response.status() === 200,
    { timeout: 10000 }
  );
  await submitButton.click();
  await responsePromise;
}

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetSettings(request);
    await navigateToSettings(page);
  });

  test('should display the settings page with all service cards', async ({ page }) => {
    // Verify the page title/header is visible
    await expect(page.locator('#settings h3').filter({ hasText: 'API Configuration' })).toBeVisible();

    // Verify all service cards are present
    const services = ['OpenAI', 'Alpaca Markets', 'Alpha Vantage', 'NewsAPI', 'Financial Modeling Prep'];

    for (const service of services) {
      await expect(page.locator('#settings .card-header h5').filter({ hasText: service })).toBeVisible();
    }
  });

  test('should show "Not Set" badge for unconfigured services', async ({ page }) => {
    // All services should show "Not Set" initially
    const notSetBadges = page.locator('#settings .badge-pending').filter({ hasText: 'Not Set' });
    await expect(notSetBadges).toHaveCount(5);
  });

  test('should display the About API Keys information section', async ({ page }) => {
    await expect(page.locator('#settings').getByText('About API Keys')).toBeVisible();
    await expect(page.locator('#settings').getByText('API keys are stored locally in an encrypted file')).toBeVisible();
  });
});

test.describe('Settings - Configure API Key', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetSettings(request);
    await navigateToSettings(page);
  });

  test('should configure OpenAI API key', async ({ page }) => {
    // Find the OpenAI card within the settings section
    const openaiCard = page.locator('#settings .card').filter({ hasText: 'OpenAI' });

    // Fill in the API key
    await openaiCard.locator('input[name="api_key"]').fill('sk-test-key-1234567890');

    // Submit and wait for response
    await submitAndWaitForResponse(page, openaiCard.locator('button[type="submit"]'));

    // Verify the status badge changed to "Configured"
    const openaiStatus = page.locator('#status-openai');
    await expect(openaiStatus.locator('.badge-approved')).toBeVisible();
    await expect(openaiStatus.getByText('Configured')).toBeVisible();

    // Verify the masked key is displayed
    await expect(page.locator('#settings .card').filter({ hasText: 'OpenAI' }).getByText('Current: ****7890')).toBeVisible();
  });

  test('should configure Alpaca with API key and secret', async ({ page }) => {
    // Find the Alpaca card
    const alpacaCard = page.locator('#settings .card').filter({ hasText: 'Alpaca Markets' });

    // Fill in API key and secret
    await alpacaCard.locator('input[name="api_key"]').fill('PKTEST123456789');
    await alpacaCard.locator('input[name="api_secret"]').fill('supersecretkey123');
    await alpacaCard.locator('input[name="base_url"]').fill('https://paper-api.alpaca.markets');

    // Submit and wait for response
    await submitAndWaitForResponse(page, alpacaCard.locator('button[type="submit"]'));

    // Verify configured status
    const alpacaStatus = page.locator('#status-alpaca');
    await expect(alpacaStatus.locator('.badge-approved')).toBeVisible();

    // Verify masked values are displayed
    const refreshedAlpacaCard = page.locator('#settings .card').filter({ hasText: 'Alpaca Markets' });
    await expect(refreshedAlpacaCard.getByText('Current: ****6789')).toBeVisible();
    await expect(refreshedAlpacaCard.getByText('Current: ****y123')).toBeVisible();
  });

  test('should update an existing API key', async ({ page }) => {
    // First, configure the NewsAPI key
    const newsCard = page.locator('#settings .card').filter({ hasText: 'NewsAPI' });
    await newsCard.locator('input[name="api_key"]').fill('initial-key-1234');
    await submitAndWaitForResponse(page, newsCard.locator('button[type="submit"]'));

    // Verify initial key is set
    await expect(page.locator('#settings .card').filter({ hasText: 'NewsAPI' }).getByText('Current: ****1234')).toBeVisible();

    // Update to a new key - need to re-select the card after HTMX refresh
    const refreshedNewsCard = page.locator('#settings .card').filter({ hasText: 'NewsAPI' });
    await refreshedNewsCard.locator('input[name="api_key"]').fill('updated-key-5678');
    await submitAndWaitForResponse(page, refreshedNewsCard.locator('button[type="submit"]'));

    // Verify the key was updated
    await expect(page.locator('#settings .card').filter({ hasText: 'NewsAPI' }).getByText('Current: ****5678')).toBeVisible();
  });
});

test.describe('Settings - Delete API Key', () => {
  test.beforeEach(async ({ page, request }) => {
    await resetSettings(request);
    await navigateToSettings(page);

    // First configure a key so we can delete it
    // Use specific locator to avoid matching "About API Keys" section
    const fmpCard = page.locator('#settings .card').filter({ has: page.locator('.card-header h5', { hasText: 'Financial Modeling Prep' }) });
    await fmpCard.locator('input[name="api_key"]').fill('fmp-test-key-abcd');
    await submitAndWaitForResponse(page, fmpCard.locator('button[type="submit"]'));

    // Verify it's configured
    const fmpStatus = page.locator('#status-fmp');
    await expect(fmpStatus.locator('.badge-approved')).toBeVisible();
  });

  test('should delete an API key with confirmation', async ({ page }) => {
    const fmpCard = page.locator('#settings .card').filter({ has: page.locator('.card-header h5', { hasText: 'Financial Modeling Prep' }) });
    const deleteButton = fmpCard.locator('button.btn-outline-danger');
    await expect(deleteButton).toBeVisible();

    // Override window.confirm to always return true (auto-accept)
    await page.evaluate(() => {
      window.confirm = () => true;
    });

    // Set up response listener before clicking
    const responsePromise = page.waitForResponse(
      response => response.url().includes('/api/settings/api-keys/fmp') && response.status() === 200,
      { timeout: 10000 }
    );

    // Click the delete button - use force to bypass any potential click interceptors
    await deleteButton.click();

    // Wait for response
    await responsePromise;

    // Verify the status changed back to "Not Set"
    const fmpStatus = page.locator('#status-fmp');
    await expect(fmpStatus.locator('.badge-pending')).toBeVisible();
    await expect(fmpStatus.getByText('Not Set')).toBeVisible();

    // Verify the delete button is no longer visible (only shown when configured)
    await expect(page.locator('#settings .card').filter({ has: page.locator('.card-header h5', { hasText: 'Financial Modeling Prep' }) }).locator('button.btn-outline-danger')).not.toBeVisible();
  });

  test('should cancel deletion when dialog is dismissed', async ({ page }) => {
    const fmpCard = page.locator('#settings .card').filter({ has: page.locator('.card-header h5', { hasText: 'Financial Modeling Prep' }) });

    // Override window.confirm to return false (cancel/dismiss)
    await page.evaluate(() => {
      window.confirm = () => false;
    });

    // Click the delete button
    await fmpCard.locator('button.btn-outline-danger').click();

    // Wait a moment to ensure nothing happens
    await page.waitForTimeout(500);

    // Verify the key is still configured
    const fmpStatus = page.locator('#status-fmp');
    await expect(fmpStatus.locator('.badge-approved')).toBeVisible();
  });
});

test.describe('Settings - Full CRUD Workflow', () => {
  test('should complete full create-read-update cycle', async ({ page, request }) => {
    await resetSettings(request);
    await navigateToSettings(page);

    // 1. Verify initially not configured
    await expect(page.locator('#status-alpha_vantage .badge-pending')).toBeVisible();

    // 2. Create: Configure the API key
    await page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).locator('input[name="api_key"]').fill('AV-KEY-12345678');
    await submitAndWaitForResponse(page, page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).locator('button[type="submit"]'));

    // 3. Read: Verify it's configured
    await expect(page.locator('#status-alpha_vantage .badge-approved')).toBeVisible();
    await expect(page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).getByText('Current: ****5678')).toBeVisible();

    // 4. Update: Change the API key
    await page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).locator('input[name="api_key"]').fill('AV-NEW-KEY-WXYZ');
    await submitAndWaitForResponse(page, page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).locator('button[type="submit"]'));

    // Verify updated
    await expect(page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).getByText('Current: ****WXYZ')).toBeVisible();

    // Delete functionality is already tested in the "Delete API Key" tests
    // Verify delete button is available for configured service
    const deleteButton = page.locator('#settings .card').filter({ hasText: 'Alpha Vantage' }).locator('button.btn-outline-danger');
    await expect(deleteButton).toBeVisible();
  });
});

test.describe('Settings - Multiple Services', () => {
  test('should configure multiple services independently', async ({ page, request }) => {
    await resetSettings(request);
    await navigateToSettings(page);

    // Configure OpenAI - use fresh locators to handle HTMX DOM swaps
    const openaiCard = page.locator('#settings .card').filter({ hasText: 'OpenAI' });
    await openaiCard.locator('input[name="api_key"]').fill('sk-openai-test-key');
    await submitAndWaitForResponse(page, openaiCard.locator('button[type="submit"]'));

    // Verify OpenAI is configured before proceeding
    await expect(page.locator('#status-openai .badge-approved')).toBeVisible();

    // Wait for HTMX swap to settle and re-query NewsAPI card
    const newsapiCard = page.locator('#settings .card').filter({ hasText: 'NewsAPI' });
    await expect(newsapiCard.locator('input[name="api_key"]')).toBeVisible();
    await newsapiCard.locator('input[name="api_key"]').fill('news-api-test-key');
    await submitAndWaitForResponse(page, newsapiCard.locator('button[type="submit"]'));

    // Verify both services are configured
    await expect(page.locator('#status-openai .badge-approved')).toBeVisible();
    await expect(page.locator('#status-newsapi .badge-approved')).toBeVisible();

    // Verify the other services remain unconfigured
    await expect(page.locator('#status-alpaca .badge-pending')).toBeVisible();
    await expect(page.locator('#status-alpha_vantage .badge-pending')).toBeVisible();
    await expect(page.locator('#status-fmp .badge-pending')).toBeVisible();
  });
});

test.describe('Settings - Persistence', () => {
  test('should persist settings across page reloads', async ({ page, request }) => {
    await resetSettings(request);
    await navigateToSettings(page);

    // Configure a key
    const openaiCard = page.locator('#settings .card').filter({ hasText: 'OpenAI' });
    await openaiCard.locator('input[name="api_key"]').fill('sk-persist-test-key');
    await submitAndWaitForResponse(page, openaiCard.locator('button[type="submit"]'));

    // Verify configured
    await expect(page.locator('#status-openai .badge-approved')).toBeVisible();

    // Reload the page and navigate back to settings
    await page.reload();
    await page.click('a[data-section="settings"]');
    await expect(page.locator('#settings-content .card').first()).toBeVisible({ timeout: 10000 });

    // Verify the key is still configured after reload
    await expect(page.locator('#status-openai .badge-approved')).toBeVisible();
    await expect(page.locator('#settings .card').filter({ hasText: 'OpenAI' }).getByText('Current: ****-key')).toBeVisible();
  });
});
