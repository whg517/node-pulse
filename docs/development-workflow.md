# Development Workflow

This document defines the standard development workflow for Node-Pulse. All feature, fix, documentation, refactor, and test work must follow this process.

## 1. Worktree-Based Development

All development work must happen in a Git worktree under the repository-local `.worktree/` directory.

Branch names must use:

```text
<type>-<name>
```

Examples:

```text
feat-alert-timeline
fix-mtls-release-mode
docs-development-workflow
refactor-export-service
test-webhook-delivery
```

Use lowercase words separated by hyphens. Keep names short and descriptive.

## 2. Allowed Types

Allowed branch and commit types:

- `feat` - user-facing feature or capability
- `fix` - bug fix
- `docs` - documentation-only change
- `refactor` - code restructuring with no behavior change
- `test` - test-only change
- `perf` - performance improvement
- `build` - build system or dependency change
- `ci` - CI or automation change
- `revert` - revert a previous change

Do not use `chore`. If the work does not fit, choose the closest meaningful type from the list above.

## 3. Start Work

Always start from the main checkout and verify that it is on `main`.

```bash
git branch --show-current
git status --short
```

If the current branch is not `main`, switch to `main` before creating a development worktree.

```bash
git switch main
git pull --ff-only
mkdir -p .worktree
git worktree add -b <type>-<name> .worktree/<type>-<name> main
cd .worktree/<type>-<name>
```

Example:

```bash
git switch main
git pull --ff-only
mkdir -p .worktree
git worktree add -b feat-alert-timeline .worktree/feat-alert-timeline main
cd .worktree/feat-alert-timeline
```

## 4. Develop And Commit

Keep commits focused. Prefer small commits that describe one logical step.

Commit messages must use Conventional Commit style:

```text
<type>(optional-scope): <summary>
```

Examples:

```text
feat(alerts): add timeline endpoint
fix(auth): reject expired websocket token
docs: add development workflow
test(webhooks): cover manual delivery failure
```

Rules:

- Use one of the allowed types from this document.
- Do not use `chore`.
- Use imperative summaries.
- Keep the subject concise.
- Include a body when the reason or migration impact is not obvious.

## 5. Completion Gates

Before work can be merged back to `main`, all completion gates in this section must pass.

### Go Lint

Run `golangci-lint` through the project Makefiles.

```bash
(cd pulse && make lint)
(cd beacon && make lint)
```

### Go Build

```bash
(cd pulse && make build)
(cd beacon && make build-local)
```

### Frontend Lint

```bash
(cd frontend && npm run lint)
```

### Frontend Build

```bash
(cd frontend && npm run build)
```

## 6. Squash Merge Back To Main

Development branches must be squash-merged into `main`.

From the main checkout:

```bash
git switch main
git pull --ff-only
git merge --squash <type>-<name>
git commit -m "<type>(optional-scope): <summary>"
```

Example:

```bash
git switch main
git pull --ff-only
git merge --squash feat-alert-timeline
git commit -m "feat(alerts): add alert timeline"
```

After the squash merge is complete and verified:

```bash
git worktree remove .worktree/<type>-<name>
git branch -d <type>-<name>
```

## 7. Quick Checklist

Use this checklist before merging:

- Work was done in `.worktree/<type>-<name>`.
- Development branch name follows `<type>-<name>`.
- Base branch was `main`.
- No `chore` branch or commit type was used.
- Go lint passed for Pulse and Beacon.
- Go build passed for Pulse and Beacon.
- Frontend lint passed.
- Frontend build passed.
- Changes were squash-merged back to `main`.
