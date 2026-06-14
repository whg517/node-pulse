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
    brand: get('--primary') || 'var(--primary)',
    brandHover: get('--primary') || 'var(--primary)',
    brandMuted: get('--muted') || 'var(--muted)',
    brandSubtle: get('--accent') || 'var(--accent)',
    healthy: get('--chart-2') || 'var(--chart-2)',
    healthyBg: get('--color-healthy-bg') || 'var(--color-healthy-bg)',
    healthyText: get('--chart-2') || 'var(--chart-2)',
    warning: get('--chart-4') || 'var(--chart-4)',
    warningBg: get('--color-warning-bg') || 'var(--color-warning-bg)',
    warningText: get('--chart-4') || 'var(--chart-4)',
    critical: get('--destructive') || 'var(--destructive)',
    criticalBg: get('--destructive') || 'var(--destructive)',
    criticalText: get('--destructive') || 'var(--destructive)',
    unknown: get('--muted-foreground') || 'var(--muted-foreground)',
    unknownBg: get('--muted') || 'var(--muted)',
    unknownText: get('--muted-foreground') || 'var(--muted-foreground)',
  }
}
