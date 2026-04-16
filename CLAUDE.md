# CLAUDE.md

## Build & Run

```bash
wails dev              # Development with hot reload
wails build            # Production build -> build/bin/travian-bot
```

Frontend only (from frontend/):
```bash
npm install            # Install frontend deps
npm run dev            # Vite dev server
npm run build          # Production build
```

## Test

```bash
go test ./...          # Run all Go tests
```

## Lint

```bash
go vet ./...           # Go static analysis
```

## Project Structure

- `main.go` — Wails entry point, wires up database, event bus, browser manager, scheduler, and app
- `internal/app/` — Wails-bound `App` struct; exported methods = frontend API
- `internal/task/` — `Scheduler` + task implementations (one file per task type)
- `internal/command/feature/` — Bot actions (upgrade, train, farm, send resources, etc.)
- `internal/command/navigate/` — Page navigation helpers
- `internal/command/update/` — Data sync from game HTML
- `internal/browser/` — go-rod headless Chrome wrapper with anti-detection
- `internal/database/` — SQLite via sqlx; `db.go` has migrations, `*_repo.go` files are repositories
- `internal/parser/` — goquery HTML parsers
- `internal/domain/model/` — Data structs
- `internal/domain/enum/` — Game enums (Building, Troop, Tribe, Status, etc.)
- `internal/event/` — Event bus bridging Go and Wails frontend
- `internal/service/` — Delays, work windows, settings
- `frontend/src/` — Svelte 5 + TypeScript + Tailwind CSS 4

## Code Style

- Go: standard library style, `slog` for structured logging
- Errors returned, not panicked — `log.Fatalf` only in `main.go`
- Tasks implement `Execute(ctx, browser, db) error` interface
- Commands are stateless functions composable by tasks
- Frontend: Svelte 5 with TypeScript, Tailwind utility classes
- Database: embedded SQL migrations in `internal/database/migrations/`
