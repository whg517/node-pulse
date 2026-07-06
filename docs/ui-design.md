# NodePulse UI Design System

**Version:** 4.0
**Date:** 2026-06-14
**Status:** Current implementation

This document reflects the current React 19 + Tailwind CSS 4 + shadcn/ui frontend. It replaces the older legacy charting-era design document.

---

## 1. Design Direction

NodePulse is an operations dashboard for repeated monitoring and incident response work. The interface should feel dense, calm, and precise rather than promotional.

Design goals:

- High scanability for network health, alerts, and node state.
- Consistent light/dark behavior through semantic theme tokens.
- Predictable layout and controls across all management pages.
- Accessible keyboard and screen-reader behavior through Radix-based primitives.
- Clear separation between reusable UI primitives and domain components.

---

## 2. Current Frontend Stack

- React 19 + TypeScript 5
- Vite 7
- React Router 7
- Tailwind CSS 4 using `@theme inline`
- shadcn/ui v4-style primitives in `frontend/src/components/ui`
- Radix UI primitives through the `radix-ui` package
- Recharts for time-series and gauge visualizations
- react-simple-maps for map visualizations
- Zustand and TanStack Query for state/data workflows
- i18next / react-i18next for `en` and `zh-CN`

---

## 3. Theming Architecture

The frontend uses the shadcn/Tailwind semantic-token model:

```
src/index.css
  :root / .dark CSS variables
    -> @theme inline token mapping
      -> Tailwind utilities
        -> shadcn/ui primitives
          -> feature components
```

### 3.1 Styling Rules

Use semantic utilities:

```tsx
<div className="rounded-lg border bg-card text-card-foreground">
  <span className="text-muted-foreground">Latency</span>
</div>
```

Avoid direct CSS variable styling in components:

```tsx
// Avoid for ordinary UI styling
<div className="bg-[var(--color-bg-surface)] text-[var(--color-text-primary)]" />
```

Avoid hardcoded palette classes for theme-sensitive UI:

```tsx
// Avoid
<span className="text-gray-500 dark:text-gray-300" />
```

Use status tokens:

| Purpose | Utilities |
|---------|-----------|
| Healthy | `text-healthy`, `bg-healthy-bg`, `text-healthy-text` |
| Warning | `text-warning`, `bg-warning-bg`, `text-warning-text` |
| Critical / destructive | `text-destructive`, `bg-destructive`, `bg-destructive/10` |
| Muted / unknown | `text-muted-foreground`, `bg-muted` |
| Brand / primary action | `bg-primary`, `text-primary`, `text-primary-foreground` |

Dark mode is handled by `:root` and `.dark` variables. Use `dark:` only for structural exceptions, not for routine color mirroring.

---

## 4. UI Primitive Layer

All reusable primitives live in `frontend/src/components/ui/`.

Current primitive set:

```
alert-dialog.tsx
avatar.tsx
badge.tsx
breadcrumb.tsx
button.tsx
card.tsx
chart.tsx
collapsible.tsx
dialog.tsx
dropdown-menu.tsx
input.tsx
label.tsx
scroll-area.tsx
separator.tsx
sheet.tsx
sidebar.tsx
skeleton.tsx
switch.tsx
textarea.tsx
tooltip.tsx
```

Rules:

- Use `Button` for commands and actions.
- Use `Dialog` for create/edit/detail forms.
- Use `AlertDialog` for destructive confirmations.
- Use `Card` for repeated items, tables, and contained dashboard regions.
- Use `Badge` for status and compact labels.
- Use `Label` with `htmlFor` and a matching form-control `id`.
- Use `Switch` for binary state.
- Use `Skeleton` for structured loading states.

---

## 5. Dialog and Modal Standards

Dialog consistency is important because NodePulse has many CRUD workflows.

### 5.1 Standard Dialog

Use:

```tsx
<Dialog open={open} onOpenChange={(nextOpen) => { if (!nextOpen) onClose() }}>
  <DialogContent className="sm:max-w-md">
    <DialogHeader>
      <DialogTitle>{title}</DialogTitle>
    </DialogHeader>
    {/* form or content */}
    <DialogFooter>
      <Button variant="outline" onClick={onClose}>Cancel</Button>
      <Button type="submit">Save</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

Size guidance:

| Use case | Width |
|----------|-------|
| Small confirmation or compact form | `sm:max-w-md` |
| Standard create/edit form | `sm:max-w-lg` |
| Detail panel or wider form | `sm:max-w-2xl` |
| JSON/template editor | `sm:max-w-3xl max-h-[85vh] overflow-y-auto` |
| Preview/report dialog | `max-w-4xl max-h-[90vh] overflow-y-auto` |

### 5.2 AlertDialog

Use `AlertDialog` for destructive confirmations and set the action variant semantically:

```tsx
<AlertDialogAction variant="destructive">Delete</AlertDialogAction>
```

Do not repeat destructive button styles with ad hoc `className="bg-destructive ..."`.

### 5.3 Accessibility Requirements

- Dialog content must have a `DialogTitle`.
- Forms must use `Label htmlFor` with a matching input/select/textarea `id`.
- Loading submit buttons must be disabled.
- Errors should appear close to the failing field or above the form for form-level errors.
- Do not hand-roll fixed overlays for normal dialogs; use `Dialog`/`AlertDialog`.

---

## 6. Layout Architecture

Authenticated pages render under `AppLayout`:

```
App
  -> QueryClientProvider
  -> BrowserRouter
  -> ProtectedLayout
      -> ProtectedRoute
      -> AppLayout
          -> Header
          -> Sidebar
          -> Page content
```

Page-level rules:

- Use `PageHeader` for page title, subtitle, and primary page actions.
- Keep operational pages dense and scan-friendly.
- Prefer tables for lists that operators compare repeatedly.
- Avoid marketing-style hero sections inside the application.
- Keep cards as functional containers, not nested decorative layouts.

---

## 7. Navigation and Routes

Primary navigation groups:

| Group | Routes |
|-------|--------|
| Dashboard | `/dashboard`, `/performance` |
| Nodes | `/nodes`, `/nodes/:id`, `/nodes/comparison`, `/nodes/probes`, `/beacons/config` |
| Alerts | `/alerts/rules`, `/alerts/records`, `/alerts/history` |
| Reports | `/reports`, `/reports/history` |
| Integrations | `/integrations/webhooks`, `/integrations/health` |
| Settings | `/settings/preferences`, `/settings/sessions`, `/settings/users`, `/settings/api-keys`, `/settings/audit-logs`, `/settings/system-config` |

Short aliases exist for legacy and e2e navigation: `/webhooks`, `/sessions`, `/comparison`.

---

## 8. Visualization Standards

Current visualization libraries:

- Recharts for line, area, comparison, and gauge-style charts.
- react-simple-maps for geographical views.

Chart rules:

- Use `useThemeColors()` or semantic CSS variables resolved through theme-aware helpers.
- Do not use invalid SVG color expressions such as `hsl(var(--token))` when the token is already an OKLCH/color value.
- Keep chart legends and tooltips compact.
- Use status colors sparingly; reserve destructive/warning colors for actual risk states.
- Provide a tabular or textual fallback where data is critical for operations.

Key chart components:

```
components/charts/
├── LatencyTrendChart.tsx
├── PacketLossChart.tsx
└── ProbeSuccessGauge.tsx

components/dashboard/
├── ComparisonChart.tsx
├── LatencyTrendChart.tsx
├── PerformanceTrendChart.tsx
├── TrendChart.tsx
└── WorldMap.tsx
```

---

## 9. Domain Component Conventions

### Dashboard

Dashboard components should prioritize live state and anomaly scanning:

- `MetricsSummaryCards`
- `MetricCard`
- `HealthStatusBadge`
- `WorldMap`
- `AlertStream`
- `TopAnomaliesList`
- `NodeListTable`
- `ProblemDiagnosis`

### Nodes

Node workflows should use:

- `NodeTable` for list management.
- `NodeDialog` for create/edit.
- `MTRVisualization` and `MTRPathVisualization` for route/path diagnostics.

### Alerts

Alert workflows should use:

- `AlertRuleDialog` and `AlertRuleForm` for rule editing.
- `AlertRecordsFilter` and `AlertRecordsTable` for record triage.
- `AlertRecordDetailModal` for record details.
- `AlertDetailMobile` for the mobile full-screen alert detail experience.

### Webhooks

Webhook create/edit uses:

- `WebhookDialog`
- `WebhookForm`
- `WebhooksTable`

Webhook forms may use wider dialogs because the event-format JSON editor needs more horizontal space.

---

## 10. Internationalization

Rules:

- All user-visible text should use `useTranslation()`.
- Add keys to both `frontend/src/locales/en.json` and `frontend/src/locales/zh-CN.json`.
- Test mocks use real English locale lookup in `vitest-setup.ts`; tests should assert visible English text unless a key fallback is intentional.
- Technical placeholders may remain in English where they are protocol names, metric keys, or payload examples.

---

## 11. Responsive Behavior

- Tables may scroll horizontally when column density requires it.
- Form dialogs should fit within `max-w-[calc(100%-2rem)]` inherited from `DialogContent`.
- Wider dialogs must set `max-h` and `overflow-y-auto`.
- Mobile-specific full-screen overlays are allowed only when the interaction is deliberately full-screen, such as alert detail triage.
- Touch targets should be at least 44px where practical for mobile-first interactions.

---

## 12. Verification Checklist

Before merging UI changes:

- `npm run lint`
- `npm run test -- --run`
- `npm run build`
- Confirm no hand-rolled CRUD modal overlays remain.
- Confirm dialogs use `DialogTitle`.
- Confirm destructive confirmations use `AlertDialog`.
- Confirm form labels are associated with controls.
- Confirm both light and dark theme tokens are respected.

---

## 13. Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 4.0 | 2026-06-14 | Codex | Updated to current shadcn/ui + Tailwind CSS 4 + Recharts implementation; documented dialog standards, semantic token rules, current routes, and current component architecture. |
| 3.0 | 2026-03-21 | Design Team | Older Instrument Panel design system from the previous frontend implementation. Superseded by 4.0. |
