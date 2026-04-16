# Agents

## Overview

This is a Wails v2 desktop application — a Travian game automation bot with a Go backend and Svelte/TypeScript frontend.

## Architecture

The backend follows a layered architecture:

1. **app** — Wails-bound API surface. All exported methods on `App` are callable from the frontend.
2. **task** — Scheduler runs per-account task queues. Each task implements `Execute(ctx, browser, db) error`.
3. **command** — Reusable building blocks used by tasks. Split into `feature/` (actions), `navigate/` (page nav), and `update/` (data sync).
4. **browser** — go-rod wrapper managing per-account headless Chrome instances with anti-detection.
5. **database** — SQLite via sqlx with embedded migrations. One repository per entity.
6. **parser** — goquery-based HTML parsers for extracting game state.
7. **event** — Channel-based event bus bridging backend and frontend.
8. **domain** — Models, enums, errors, and static game data. No business logic.
9. **service** — Cross-cutting concerns: random delays, work windows, settings.

## Key Patterns

- Tasks are the unit of work. To add new automation, create a new task in `internal/task/` and wire it into the scheduler.
- Commands are composable. A task typically chains navigate -> update -> feature commands.
- The event bus (`event.Bus`) emits to both Go subscribers and the Wails frontend runtime.
- Browser instances are pooled per account via `browser.Manager`.
- All persistent state lives in SQLite. Settings propagate immediately to running tasks.

## Frontend

- Svelte 5 + TypeScript + Tailwind CSS 4
- Generated Wails bindings in `frontend/src/wailsjs/` provide type-safe Go method calls.
- State managed via Svelte stores, updated by backend events.

## Development

```bash
wails dev     # Dev mode with hot reload
wails build   # Production build
```
