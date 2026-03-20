/**
 * useThemeColors Hook
 *
 * Reads CSS custom properties at runtime for use in contexts that
 * cannot use Tailwind classes (e.g., ECharts options, inline styles).
 * The hook reads from :root/.dark which automatically adapts on theme toggle.
 */
export function useThemeColors() {
  const style = getComputedStyle(document.documentElement)
  const get = (name: string) => style.getPropertyValue(name).trim()

  return {
    brand: get('--color-brand'),
    brandHover: get('--color-brand-hover'),
    brandMuted: get('--color-brand-muted'),
    brandSubtle: get('--color-brand-subtle'),
    healthy: get('--color-healthy'),
    healthyBg: get('--color-healthy-bg'),
    healthyText: get('--color-healthy-text'),
    warning: get('--color-warning'),
    warningBg: get('--color-warning-bg'),
    warningText: get('--color-warning-text'),
    critical: get('--color-critical'),
    criticalBg: get('--color-critical-bg'),
    criticalText: get('--color-critical-text'),
    unknown: get('--color-unknown'),
    unknownBg: get('--color-unknown-bg'),
    unknownText: get('--color-unknown-text'),
  }
}
