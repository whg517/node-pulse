# NodePulse Frontend

Network monitoring dashboard built with React, TypeScript, Vite, and shadcn/ui.

## Tech Stack

- **React 19** + TypeScript
- **Vite** for build tooling
- **Tailwind CSS 4** with oklch color space
- **shadcn/ui v4** component library
- **Recharts** for data visualization
- **react-i18next** for internationalization (zh-CN / en)

## Styling Architecture

This project follows the shadcn/ui + Tailwind CSS community best practices for theming and styling.

### Three-Layer System

```
CSS Variables (:root / .dark)  →  @theme inline  →  Tailwind Utilities  →  Components
```

1. **CSS Variables** (`src/index.css` `:root` / `.dark`): Define colors in oklch format
2. **`@theme inline`** (`src/index.css`): Maps CSS variables to Tailwind color tokens
3. **Tailwind Utilities**: Use semantic classes like `bg-card`, `text-foreground`, `border`
4. **Components**: Compose utilities; never reference CSS variables directly

### Rules

1. **Use Tailwind semantic utilities, not CSS variable references**
   ```tsx
   // ✅ Correct
   <div className="bg-card text-foreground border">
   
   // ❌ Wrong
   <div className="bg-[var(--color-bg-surface)] text-[var(--color-text-primary)] border-[var(--color-border)]">
   ```

2. **Use theme tokens, not hardcoded colors**
   ```tsx
   // ✅ Correct
   <span className="text-muted-foreground">
   <div className="bg-muted">
   
   // ❌ Wrong
   <span className="text-gray-500">
   <div className="bg-gray-100">
   ```

3. **Block containers use the standard pattern**
   ```tsx
   <div className="rounded-lg border bg-card p-4">
     <h3 className="text-sm font-semibold">Title</h3>
   </div>
   ```

4. **Status colors use semantic tokens**
   - Healthy: `bg-healthy`, `text-healthy`, `bg-healthy-bg`, `text-healthy-text`
   - Warning: `bg-warning`, `text-warning`, `bg-warning-bg`, `text-warning-text`
   - Critical/Error: `text-destructive`, `bg-destructive`, `bg-destructive/10`
   - Unknown/Offline: `text-muted-foreground`, `bg-muted`

5. **Dark mode is handled via CSS variables** — no `dark:` prefix needed for theme colors. The `:root` and `.dark` selectors automatically switch all values.

### Color Token Reference

| Token | Utility | Usage |
|-------|---------|-------|
| `foreground` | `text-foreground` | Primary text |
| `muted-foreground` | `text-muted-foreground` | Secondary text, labels |
| `primary` | `bg-primary`, `text-primary` | Brand color, CTAs |
| `card` | `bg-card` | Card/surface backgrounds |
| `background` | `bg-background` | Page background |
| `muted` | `bg-muted` | Muted backgrounds, badges |
| `border` | `border-border` (or just `border`) | Borders (set in base layer) |
| `input` | `border-input` | Form input borders |
| `destructive` | `text-destructive`, `bg-destructive` | Error states |
| `accent` | `bg-accent` | Hover states, highlights |

## Development

```bash
npm install
npm run dev       # Start dev server
npm run build     # Production build
npm test          # Run tests
npm run lint      # Lint check
```

## Project Structure

```
src/
├── api/              # API client and types
├── components/
│   ├── charts/       # Recharts visualizations
│   ├── common/       # Shared components
│   ├── dashboard/    # Dashboard-specific components
│   ├── layout/       # App shell, sidebar, header
│   ├── nodes/        # Node management
│   ├── reports/      # Report generation
│   └── ui/           # shadcn/ui primitives
├── hooks/            # Custom React hooks
├── locales/          # i18n JSON files (zh-CN, en)
├── pages/            # Route page components
├── stores/           # Zustand stores
└── utils/            # Utilities
```
