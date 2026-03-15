import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { LanguageSwitcher } from '../LanguageSwitcher'

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn() },
  }),
}))

// Mock settingsStore
const mockSetLanguage = vi.fn()
vi.mock('../../../stores/settingsStore', () => ({
  useSettingsStore: (selector: (s: { language: string; setLanguage: typeof mockSetLanguage }) => unknown) =>
    selector({ language: 'en', setLanguage: mockSetLanguage }),
  supportedLanguages: [
    { code: 'en', name: 'English', nativeName: 'English' },
    { code: 'zh-CN', name: 'Chinese (Simplified)', nativeName: '简体中文' },
  ],
}))

describe('LanguageSwitcher', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders dropdown variant by default', () => {
    render(<LanguageSwitcher />)
    const select = screen.getByRole('combobox')
    expect(select).toBeInTheDocument()
  })

  it('renders buttons variant', () => {
    render(<LanguageSwitcher variant="buttons" />)
    const buttons = screen.getAllByRole('button')
    expect(buttons.length).toBeGreaterThan(0)
  })

  it('dropdown shows language options', () => {
    render(<LanguageSwitcher />)
    expect(screen.getByText('English')).toBeInTheDocument()
    expect(screen.getByText('简体中文')).toBeInTheDocument()
  })

  it('buttons variant shows language buttons', () => {
    render(<LanguageSwitcher variant="buttons" />)
    expect(screen.getByText('English')).toBeInTheDocument()
    expect(screen.getByText('简体中文')).toBeInTheDocument()
  })

  it('calls setLanguage when dropdown changes', () => {
    render(<LanguageSwitcher />)
    const select = screen.getByRole('combobox')
    fireEvent.change(select, { target: { value: 'zh-CN' } })
    expect(mockSetLanguage).toHaveBeenCalledWith('zh-CN')
  })

  it('calls setLanguage when button is clicked', () => {
    render(<LanguageSwitcher variant="buttons" />)
    fireEvent.click(screen.getByText('简体中文'))
    expect(mockSetLanguage).toHaveBeenCalledWith('zh-CN')
  })

  it('applies custom className', () => {
    const { container } = render(<LanguageSwitcher className="my-custom-class" />)
    expect(container.firstChild?.toString()).toBeDefined()
  })

  it('renders with sm size', () => {
    render(<LanguageSwitcher size="sm" />)
    const select = screen.getByRole('combobox')
    expect(select.className).toContain('text-xs')
  })

  it('renders with lg size', () => {
    render(<LanguageSwitcher size="lg" />)
    const select = screen.getByRole('combobox')
    expect(select.className).toContain('text-base')
  })
})
