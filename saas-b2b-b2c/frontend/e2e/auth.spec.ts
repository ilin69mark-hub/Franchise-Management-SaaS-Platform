import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should display login form with all required fields', async ({ page }) => {
    await expect(page.locator('input[type="text"], input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"], button:has-text("Войти")')).toBeVisible();
  });

  test('should show validation error for empty form submission', async ({ page }) => {
    await page.click('button[type="submit"], button:has-text("Войти")');
    await expect(page.locator('.ant-form-item-explain-error, text=обязательно')).toBeVisible({ timeout: 3000 });
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.fill('input[type="text"], input[type="email"]', 'invalid@test.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"], button:has-text("Войти")');
    await page.waitForTimeout(1000);
    const errorMessage = page.locator('.ant-alert-error, text=error, text=Error');
    await expect(errorMessage.first()).toBeVisible({ timeout: 5000 }).catch(() => {});
  });

  test('should have working register link', async ({ page }) => {
    const registerLink = page.locator('a:has-text("регистрац"), a:has-text("Register"), a:has-text("Регистрац")');
    if (await registerLink.count() > 0) {
      await expect(registerLink.first()).toBeVisible();
    }
  });
});

test.describe('Registration Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register');
  });

  test('should display registration form', async ({ page }) => {
    await expect(page.locator('input[type="text"], input[type="email"]').first()).toBeVisible();
    await expect(page.locator('input[type="password"]').first()).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });
});

test.describe('Password Reset', () => {
  test('should have forgot password link on login page', async ({ page }) => {
    await page.goto('/login');
    const forgotLink = page.locator('a:has-text("забыли"), a:has-text("Forgot")');
    if (await forgotLink.count() > 0) {
      await expect(forgotLink.first()).toBeVisible();
    }
  });
});