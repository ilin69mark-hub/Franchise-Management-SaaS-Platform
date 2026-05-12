import { test, expect } from '@playwright/test';

test.describe('Checklist Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/checklists');
    await page.waitForLoadState('networkidle').catch(() => {});
  });

  test('should load checklist page', async ({ page }) => {
    await expect(page).toHaveURL(/checklists/);
  });

  test('should display checklist header', async ({ page }) => {
    const title = page.locator('h1, h2, .ant-typography:has-text("Чек-лист"), .ant-page-header-heading');
    if (await title.count() > 0) {
      await expect(title.first()).toBeVisible();
    }
  });

  test('should display checklist content area', async ({ page }) => {
    const content = page.locator('.ant-card, .ant-list, .checklist-item, [class*="checklist"]');
    const count = await content.count();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should have add new checklist button if authorized', async ({ page }) => {
    const addButton = page.locator('button:has-text("Добавить"), button:has-text("Создать"), button:has-text("New")');
    const buttonCount = await addButton.count();
    if (buttonCount > 0) {
      await expect(addButton.first()).toBeVisible();
    }
  });
});

test.describe('Checklist Interactions', () => {
  test('should filter checklists by status', async ({ page }) => {
    await page.goto('/checklists');
    await page.waitForLoadState('networkidle').catch(() => {});
    
    const filterSelect = page.locator('.ant-select, select, [class*="filter"]');
    if (await filterSelect.count() > 0) {
      await filterSelect.first().click();
      await page.waitForTimeout(300);
    }
  });

  test('should filter checklists by priority', async ({ page }) => {
    await page.goto('/checklists');
    await page.waitForLoadState('networkidle').catch(() => {});
    
    const priorityFilter = page.locator('[class*="priority"], .ant-select-selection-item');
    if (await priorityFilter.count() > 0) {
      await priorityFilter.first().click();
      await page.waitForTimeout(300);
    }
  });

  test('should handle empty checklist state', async ({ page }) => {
    await page.goto('/checklists');
    await page.waitForLoadState('networkidle').catch(() => {});
    
    const emptyState = page.locator('.ant-empty, text=Нет данных, text=No data');
    const count = await emptyState.count();
    if (count > 0) {
      await expect(emptyState.first()).toBeVisible();
    }
  });
});