# Travian Bot

A desktop automation bot for Travian built with [Wails](https://wails.io/) (Go backend + Svelte/TypeScript frontend).

## Features

- **Multi-account support** with per-account browser profiles and proxy configuration
- **Building automation** - queue upgrades, resource buildings, special construction
- **Resource management** - inter-village transfers, NPC trading, storage tracking
- **Farming** - automated farmlist execution
- **Defense** - attack detection, troop evacuation, resource evasion
- **Hero management** - adventures, inventory, item usage
- **Quest claiming** - automatic quest completion and reward collection
- **Troop training** - coordinated training across villages
- **Anti-detection** - random delays, custom user agents, stealth browser flags
- **Work windows** - configurable active hours per account
- **Real-time UI** - live status monitoring, per-account logs, queue previews

## Tech Stack

- **Backend:** Go, go-rod (headless Chrome), SQLite, goquery
- **Frontend:** Svelte 5, TypeScript, Tailwind CSS 4, Vite 6
- **Desktop:** Wails v2

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

## Development

Run in live development mode with hot reload:

```bash
wails dev
```

The frontend dev server is available at http://localhost:34115 for browser-based development with access to Go methods.

## Building

Build a production binary:

```bash
wails build
```

The output binary will be in `build/bin/`.

## Project Structure

```
├── main.go                 # Wails entry point
├── internal/
│   ├── app/                # Wails-bound application API
│   ├── browser/            # go-rod browser pool & automation
│   ├── command/
│   │   ├── feature/        # Bot actions (upgrade, train, farm, etc.)
│   │   ├── navigate/       # Page navigation commands
│   │   └── update/         # Game data synchronization
│   ├── database/           # SQLite persistence & migrations
│   ├── domain/
│   │   ├── enum/           # Game enumerations (buildings, troops, tribes)
│   │   ├── model/          # Data structures
│   │   ├── errs/           # Error types
│   │   └── gamedata/       # Static game data
│   ├── event/              # Event bus (Go <-> frontend)
│   ├── parser/             # HTML page parsers
│   ├── service/            # Business logic (delays, work windows, settings)
│   └── task/               # Task scheduler & implementations
└── frontend/
    └── src/
        ├── lib/components/ # Svelte UI components
        ├── lib/stores/     # Reactive state
        └── lib/i18n/       # Internationalization
```
