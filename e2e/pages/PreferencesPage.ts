/**
 * Preferences Page Object Model
 *
 * Handles user preferences:
 * - Language selection
 * - Timezone selection
 * - Theme toggle
 * - Notification settings
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './common/BasePage'

export interface PreferencesSelectors extends PageSelectors {
  languageSelect?: string
  timezoneSelect?: string
  themeToggle?: string
  notificationsToggle?: string
  saveButton?: string
  resetButton?: string
}

export const DEFAULT_PREFERENCES_SELECTORS: PreferencesSelectors = {
  ...DEFAULT_SELECTORS,
  languageSelect: '[data-testid="language-select"], select[name="language"], select[name="lang"]',
  timezoneSelect: '[data-testid="timezone-select"], select[name="timezone"]',
  themeToggle: '[data-testid="theme-toggle"], button:has-text("Dark"), button:has-text("Light"), [role="switch"]',
  notificationsToggle: '[data-testid="notifications-toggle"], input[type="checkbox"][name="notifications"]',
  saveButton: '[data-testid="save-button"], button:has-text("Save"), button[type="submit"]',
  resetButton: '[data-testid="reset-button"], button:has-text("Reset"), button:has-text("Restore Defaults")',
}

export class PreferencesPage extends BasePage {
  readonly languageSelect: Locator
  readonly timezoneSelect: Locator
  readonly themeToggle: Locator
  readonly notificationsToggle: Locator
  readonly saveButton: Locator
  readonly resetButton: Locator

  constructor(page: Page, selectors: PreferencesSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_PREFERENCES_SELECTORS, ...selectors }

    this.languageSelect = page.locator(mergedSelectors.languageSelect!)
    this.timezoneSelect = page.locator(mergedSelectors.timezoneSelect!)
    this.themeToggle = page.locator(mergedSelectors.themeToggle!)
    this.notificationsToggle = page.locator(mergedSelectors.notificationsToggle!)
    this.saveButton = page.locator(mergedSelectors.saveButton!)
    this.resetButton = page.locator(mergedSelectors.resetButton!)
  }

  /**
   * Navigate to preferences page
   */
  async goto(): Promise<void> {
    await super.goto('/settings/preferences')
    await this.waitForReady()
  }

  /**
   * Select language
   */
  async selectLanguage(language: string): Promise<void> {
    await this.languageSelect.selectOption(language)
  }

  /**
   * Select language by label
   */
  async selectLanguageByLabel(label: string): Promise<void> {
    await this.languageSelect.selectOption({ label })
  }

  /**
   * Select timezone
   */
  async selectTimezone(timezone: string): Promise<void> {
    await this.timezoneSelect.selectOption(timezone)
  }

  /**
   * Select timezone by label
   */
  async selectTimezoneByLabel(label: string): Promise<void> {
    await this.timezoneSelect.selectOption({ label })
  }

  /**
   * Toggle theme
   */
  async toggleTheme(): Promise<void> {
    await this.themeToggle.click()
    await this.page.waitForTimeout(500)
  }

  /**
   * Set theme to dark
   */
  async setDarkTheme(): Promise<void> {
    const currentTheme = await this.getCurrentTheme()
    if (currentTheme === 'light') {
      await this.toggleTheme()
    }
  }

  /**
   * Set theme to light
   */
  async setLightTheme(): Promise<void> {
    const currentTheme = await this.getCurrentTheme()
    if (currentTheme === 'dark') {
      await this.toggleTheme()
    }
  }

  /**
   * Get current theme
   */
  async getCurrentTheme(): Promise<string> {
    const html = this.page.locator('html')
    const classAttr = await html.getAttribute('class')
    if (classAttr?.includes('dark')) {
      return 'dark'
    }

    // Check data-theme attribute
    const dataTheme = await html.getAttribute('data-theme')
    if (dataTheme === 'dark') {
      return 'dark'
    }

    // Check prefers-color-scheme
    const isDark = await this.page.evaluate(() =>
      window.matchMedia('(prefers-color-scheme: dark)').matches
    )
    return isDark ? 'dark' : 'light'
  }

  /**
   * Toggle notifications
   */
  async toggleNotifications(): Promise<void> {
    await this.notificationsToggle.click()
  }

  /**
   * Enable notifications
   */
  async enableNotifications(): Promise<void> {
    const isChecked = await this.notificationsToggle.isChecked()
    if (!isChecked) {
      await this.notificationsToggle.check()
    }
  }

  /**
   * Disable notifications
   */
  async disableNotifications(): Promise<void> {
    const isChecked = await this.notificationsToggle.isChecked()
    if (isChecked) {
      await this.notificationsToggle.uncheck()
    }
  }

  /**
   * Check if notifications are enabled
   */
  async areNotificationsEnabled(): Promise<boolean> {
    return await this.notificationsToggle.isChecked()
  }

  /**
   * Save preferences
   */
  async save(): Promise<void> {
    await this.saveButton.click()
    await this.waitForReady()
  }

  /**
   * Save and wait for success message
   */
  async saveAndWaitForSuccess(message = 'saved'): Promise<void> {
    await this.save()
    await this.waitForToast(message)
  }

  /**
   * Reset preferences to defaults
   */
  async reset(): Promise<void> {
    await this.resetButton.click()
    await this.waitForReady()
  }

  /**
   * Get current language
   */
  async getCurrentLanguage(): Promise<string> {
    return await this.languageSelect.inputValue()
  }

  /**
   * Get current timezone
   */
  async getCurrentTimezone(): Promise<string> {
    return await this.timezoneSelect.inputValue()
  }

  /**
   * Expect language options available
   */
  async expectLanguageOptions(languages: string[]): Promise<void> {
    for (const language of languages) {
      const optionLocator = this.languageSelect.locator(`option:has-text("${language}")`)
      await expect(optionLocator).toBeVisible()
    }
  }

  /**
   * Expect timezone options available
   */
  async expectTimezoneOptions(timezones: string[]): Promise<void> {
    for (const timezone of timezones) {
      const optionLocator = this.timezoneSelect.locator(`option:has-text("${timezone}")`)
      await expect(optionLocator).toBeVisible()
    }
  }

  /**
   * Expect save button visible
   */
  async expectSaveButtonVisible(): Promise<void> {
    await expect(this.saveButton).toBeVisible()
  }

  /**
   * Expect preferences form visible
   */
  async expectFormVisible(): Promise<void> {
    await expect(this.languageSelect).toBeVisible()
    await expect(this.timezoneSelect).toBeVisible()
    await expect(this.themeToggle).toBeVisible()
  }

  /**
   * Verify theme changed
   */
  async expectThemeChanged(expectedTheme: string): Promise<void> {
    const actualTheme = await this.getCurrentTheme()
    expect(actualTheme).toBe(expectedTheme)
  }
}
