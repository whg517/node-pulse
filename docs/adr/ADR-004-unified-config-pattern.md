# ADR-004: Unified configuration pattern across components

- **Status**: Accepted
- **Date:** 2026-07-04
- **Owners:** Kevin
- **Related:** supersedes the ad-hoc per-component config conventions; followed by `refactor-pulse-config-viper`, `refactor-beacon-config-env-coverage`, `refactor-frontend-config`

## Context

Node-Pulse has three independently configured components, and a gap analysis
(2026-07-04) found that each grew its own config conventions with material
inconsistencies:

- **Pulse** (`pulse/internal/config/config.go`, 928 lines): parses `pulse.yaml`
  with `gopkg.in/yaml.v3` and binds ~46 environment variables **by hand** via
  `os.Getenv` in `mergeFromEnv`. No external config library. `sync.Once`
  singleton, frozen after load. Thorough per-subsystem `Validate()` and secret
  auto-generation.
- **Beacon** (`beacon/internal/config/`): uses **Viper** + `mapstructure/v2`,
  `BEACON_` env prefix, fsnotify + SIGHUP **hot reload** with atomic swap.
  Config instance is passed around explicitly (not a singleton). Env overrides
  exist for most but not all fields (`probes`, `tags`, `resource_monitor`,
  `telemetry` lack env coverage).
- **Frontend** (`frontend/src/config/`): Vite build-time `import.meta.env` with
  a single `VITE_API_BASE_URL` variable. No runtime config bootstrap.

Concrete problems found:

1. **Library divergence** — Pulse hand-rolls env binding; Beacon uses Viper.
   Adding a new Pulse config field requires editing the struct *and* writing a
   new `os.Getenv` branch, which is easy to forget (the `Notify`/`SMTP` section
   has zero env binding today, making pure-env deployments unable to send mail).
2. **Env coverage gaps** — Pulse `notify.smtp.*` and several Beacon subtrees
   cannot be set via env vars.
3. **Example files incomplete** — `pulse.yaml.example` omits the `notify:`
   section; `beacon.yaml.example` omits `log_*`, `mode`, `compression`, `resume`,
   `resource_monitor`.
4. **Docs vs. implementation drift** — `pulse/.env.example` lists "legacy" env
   names (`DATABASE_URL`, `JWT_SECRET`) that the code never reads;
   `notify/smtp_sender.go` comments reference `PULSE_NOTIFY_SMTP_*` that did not
   exist before this refactor; `AGENTS.md` points frontend contributors at a
   `designTokens.ts` that is dead code.
5. **A latent frontend bug** — `src/api/export.ts` reads
   `import.meta.env.VITE_API_BASE_URL` directly instead of the centralized
   `API_BASE_URL` constant, producing `undefined/...` URLs in dev mode.

## Decision

Adopt a single **config contract** that all three components honor. The contract
is library-agnostic on the frontend (Vite is the only sane choice) but mandates
**Viper for both Go components** so env binding is automatic and consistent.

### Contract 1 — Resolution priority (all components)

```
built-in defaults  <  config file (pulse.yaml / beacon.yaml)  <  environment variables
```

Higher precedence overrides lower. Secret auto-generation and validation run
last (Pulse only). Beacon's hot reload re-runs the *entire* pipeline on change.

### Contract 2 — Environment variable naming

| Component | Prefix              | Example                      |
|-----------|---------------------|------------------------------|
| Pulse     | `PULSE_`            | `PULSE_DATABASE_URL`         |
| Beacon    | `BEACON_`           | `BEACON_METRICS_PORT`        |
| Frontend  | `VITE_` (build-time)| `VITE_API_BASE_URL`          |

Viper's `SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` maps nested keys
deterministically: `notify.smtp.host` → `PULSE_NOTIFY_SMTP_HOST`.

### Contract 3 — Config file lookup order (Go components)

1. Explicit path via env var (`PULSE_CONFIG_PATH` / `BEACON_CONFIG_PATH`)
2. Current working directory (`./pulse.yaml` / `./beacon.yaml`)
3. System directory (`/etc/node-pulse/pulse.yaml` / `/etc/beacon/beacon.yaml`)

A missing file is not an error (defaults + env still apply).

### Contract 4 — Example file completeness

Every `*.yaml.example` MUST cover all top-level sections of its config struct,
with default values shown as comments. No section may be silently omitted. This
is enforced by review (no automated check yet).

### Contract 5 — Documentation fidelity

Code comments, `.env.example`, and `AGENTS.md` may only reference env var names
or file names that actually exist in the code. Deprecated/legacy names must be
removed, not merely labelled "deprecated", unless the code genuinely still reads
them.

### Library decision

- **Pulse migrates to Viper** (`github.com/spf13/viper`, aligned to Beacon's
  version), gaining automatic env binding and eliminating ~200 lines of
  hand-written `mergeFromEnv`. The `sync.Once` singleton, `Reset()`,
  `Validate()`, `generateSecrets()`, and `String()` redaction are preserved.
- **Beacon stays on Viper** (already there); this ADR just plugs its env
  coverage gaps and consolidates its dual default mechanisms.
- **Frontend stays on Vite `import.meta.env`**; this ADR fixes the `export.ts`
  divergence, removes dead code, and adds typed env declarations.

### KnownFields strictness (Pulse)

The current `yaml.Decoder.KnownFields(true)` rejects unknown YAML keys. Viper's
default `Unmarshal` is lenient. To preserve the existing behavior and its test
(`TestConfig_LoadFromFile_UnknownFieldFails`), Pulse keeps a strict yaml
re-decode of the file bytes alongside the Viper load specifically to reject
unknown fields. This is the single behavioral invariant that needs explicit care
in the migration.

## Consequences

**Positive**

- One mental model: defaults < file < env, with deterministic env naming.
- Adding a Pulse config field no longer requires touching `mergeFromEnv` — tag
  the struct and Viper binds the env var automatically. The `Notify`/`SMTP`
  gap is closed as a direct consequence.
- Go components share a library; Beacon's hot-reload Viper usage becomes a
  reference rather than an outlier.
- Example files and docs become trustworthy sources of truth.

**Negative**

- Pulse gains a non-trivial dependency (`viper` + transitive `mapstructure`,
  `pflag`, etc.). Accepted: Beacon already carries it, and the maintenance
  reduction (no hand-written env binding) outweighs the cost.
- The Pulse migration is the highest-risk step: ~900-line file rewrite, test
  adjustments, and the `KnownFields` invariant to preserve. Mitigated by running
  `make test-unit` + the coverage ratchet (31.0% floor) before merge.
- Beacon's `probes` / `tags` / `resource_monitor` remain file-only (env override
  is impractical for nested lists). Documented as an explicit exception here so
  it is a deliberate choice, not an oversight.

## Alternatives considered

- **"Unify the pattern, not the library"** — keep Pulse on `yaml.v3` + hand
  written env, Beacon on Viper, just align the contracts. Rejected: the
  hand-written env layer is itself the source of the `Notify`/SMTP gap and the
  maintenance burden; choosing to keep it would preserve the root cause.
- **Migrate both Go components to `envconfig` + `yaml.v3`** (lighter than Viper).
  Rejected: Beacon would lose Viper's hot-reload foundation (fsnotify wiring),
  requiring a rewrite of `watcher.go`. Higher disruption for no net benefit over
  standardizing on the library already in use.

## Migration plan

Executed as four independent, squash-merged branches (each with the full
completion-gate workflow):

1. **This ADR** (`docs-adr-config-pattern`) — establishes the baseline.
2. `refactor-frontend-config` — `export.ts` fix, `designTokens.ts` removal,
   `vite-env.d.ts` typing.
3. `refactor-pulse-config-viper` — Viper migration, SMTP env binding, example
   completion, doc fixes.
4. `refactor-beacon-config-env-coverage` — `telemetry` env coverage, default
   consolidation, example completion.

## Current state (as of 2026-07-04)

- **Proposed.** This ADR is the baseline; the four migration branches follow.
  Status will move to "implemented" as each branch lands.
