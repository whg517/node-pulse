/**
 * Common Page Object Models
 *
 * Reusable base classes for building page objects:
 * - BasePage: Core functionality for all pages
 * - TablePage: Table-specific operations
 * - ModalPage: Modal dialog operations
 * - FormPage: Form handling operations
 */

export { BasePage, DEFAULT_SELECTORS, type PageSelectors } from './BasePage'
export { TablePage, DEFAULT_TABLE_SELECTORS, type TableSelectors } from './TablePage'
export { ModalPage, DEFAULT_MODAL_SELECTORS, type ModalSelectors } from './ModalPage'
export { FormPage, DEFAULT_FORM_SELECTORS, type FormSelectors, type FormField } from './FormPage'
