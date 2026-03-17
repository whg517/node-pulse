import { render } from '@testing-library/react'
import App from './App'

vi.mock('./i18n', () => ({
  default: {},
  i18nInitPromise: Promise.resolve(),
}))

describe('App', () => {
  it('renders without crashing', () => {
    render(<App />)
  })
})
