/**
 * i18n static configuration constants
 *
 * Extracted from i18n.ts so components can import language list
 * without pulling in the full i18next initialization.
 */

export const supportedLanguages = [
  { code: 'en', name: 'English', nativeName: 'English' },
  { code: 'zh-CN', name: 'Chinese (Simplified)', nativeName: '简体中文' },
] as const

export type LanguageCode = (typeof supportedLanguages)[number]['code']
