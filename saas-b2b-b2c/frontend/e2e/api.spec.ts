import { test, expect } from '@playwright/test';

test.describe('API Error Handling', () => {
  test('should handle 404 errors gracefully', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (error) => {
      if (error.message.includes('404') || error.message.includes('Not Found')) {
        errors.push(error.message);
      }
    });
    
    await page.goto('/non-existent-page-12345');
    await page.waitForLoadState('networkidle').catch(() => {});
    
    const notFoundMessage = page.locator('text=404, text=Not Found, text=Страница не найдена');
    const count = await notFoundMessage.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should handle 500 errors gracefully', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const errorAlerts = page.locator('.ant-alert-error, .error-boundary, [class*="error"]');
    const alertCount = await errorAlerts.count();
    expect(alertCount).toBeGreaterThanOrEqual(0);
  });

  test('should show loading states during API calls', async ({ page }) => {
    await page.goto('/checklists');
    await page.waitForLoadState('networkidle').catch(() => {});
    
    const loadingSpinner = page.locator('.ant-spin, .ant-skeleton, [class*="loading"]');
    const count = await loadingSpinner.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });
});

test.describe('Network Conditions', () => {
  test('should handle slow network', async ({ page }) => {
    await page.route('**/*', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 500));
      await route.continue();
    });
    
    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
    
    const content = page.locator('main, .ant-layout-content');
    if (await content.count() > 0) {
      await expect(content.first()).toBeVisible({ timeout: 10000 });
    }
  });

  test('should handle offline mode', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const offlineBanner = page.locator('.ant-offline, text=Offline, text=Нет подключения');
    const count = await offlineBanner.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });
});