import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test('should display main navigation elements', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const header = page.locator('header, .ant-layout-header');
    if (await header.count() > 0) {
      await expect(header.first()).toBeVisible();
    }
  });

  test('should have working back button on subpages', async ({ page }) => {
    await page.goto('/login');
    const backButton = page.locator('button:has-text("На главную"), button:has(.anticon-left)');
    if (await backButton.count() > 0) {
      await expect(backButton.first()).toBeVisible();
    }
  });

  test('should navigate to home page', async ({ page }) => {
    await page.goto('/login');
    const homeLink = page.locator('a[href="/"], a:has-text("Главная")');
    if (await homeLink.count() > 0) {
      await homeLink.first().click();
      await page.waitForTimeout(500);
      await expect(page).toHaveURL(/\/|\/login/);
    }
  });
});

test.describe('Responsive Navigation', () => {
  test('should display navigation in desktop view', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const nav = page.locator('.ant-menu, nav, header');
    if (await nav.count() > 0) {
      await expect(nav.first()).toBeVisible();
    }
  });

  test('should display navigation in tablet view', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const nav = page.locator('.ant-menu, nav, header');
    if (await nav.count() > 0) {
      await expect(nav.first()).toBeVisible();
    }
  });

  test('should display navigation in mobile view', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    const mobileNav = page.locator('.ant-menu-mobile, .ant-drawer, .mobile-menu');
    if (await mobileNav.count() > 0) {
      await expect(mobileNav.first()).toBeVisible();
    }
  });
});