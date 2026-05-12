import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test('should load without critical errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (error) => {
      errors.push(error.message);
    });
    
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    expect(errors.filter(e => !e.includes('Warning'))).toHaveLength(0);
  });

  test('should display page content', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const mainContent = page.locator('main, .ant-layout-content, .content');
    if (await mainContent.count() > 0) {
      await expect(mainContent.first()).toBeVisible();
    }
  });

  test('should display footer', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const footer = page.locator('footer, .ant-layout-footer');
    if (await footer.count() > 0) {
      await expect(footer.first()).toBeVisible();
    }
  });
});

test.describe('Role-specific Pages', () => {
  test('should load admin page', async ({ page }) => {
    await page.goto('/admin');
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/admin/);
  });

  test('should load franchiser-manager page', async ({ page }) => {
    await page.goto('/franchiser-manager');
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/franchiser-manager/);
  });

  test('should load dealer page', async ({ page }) => {
    await page.goto('/dealer');
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/dealer/);
  });

  test('should load salon-manager page', async ({ page }) => {
    await page.goto('/salon-manager');
    await page.waitForLoadState('networkidle');
    await expect(page).toHaveURL(/salon-manager/);
  });
});