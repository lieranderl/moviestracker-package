# AGENTS.md

This file defines project-specific guidance for coding agents working in `moviestracker-package`.

## Project Purpose

`moviestracker-package` searches trackers (Rutor + Kinozal), normalizes torrent metadata, enriches movies from TMDB, and optionally saves data to MongoDB.

Persistence target is MongoDB only (`go.mongodb.org/mongo-driver/v2`). Firestore/Firebase paths are intentionally removed.

## Architecture

- Entry point: `cmd/main.go`
- Pipeline orchestrator: `executor/executor.go`
- Tracker adapters:
  - `internal/rutor/*`
  - `internal/kinozal/*`
- Domain models:
  - `internal/torrents/torrent.go`
  - `internal/movies/*`
- Generic pipeline utilities:
  - `pkg/pipeline/*`

## Development Rules

1. Keep orchestration in `executor/`; keep parsing logic inside tracker packages.
2. Prefer explicit error returns over `log.Fatal` in non-`main` packages.
3. Maintain cancellation-safe channel patterns:
   - handle closed channels (`value, ok := <-ch`)
   - stop producers/workers when context is canceled
4. Preserve backward compatibility where practical:
   - do not remove existing exported methods without wrappers/deprecation path
5. Avoid introducing tracker-specific logic in shared model packages.

## DRY/SOLID Expectations

- Single responsibility:
  - Parsing functions only parse.
  - Persistence methods only persist.
  - Pipeline methods only orchestrate.
- Open/closed:
  - Add new trackers via `tracker.Config` + parser function, avoid hard-coded branches when possible.
- Interface boundaries:
  - Keep HTTP scraping details within tracker packages.
  - Keep DB client lifecycle in executor-level persistence paths.
- DRY:
  - Reuse helpers for duplicate parser patterns (hash extraction, title normalization, date parsing).

## Test Strategy

- Unit tests: deterministic and offline.
- Integration tests:
  - Kinozal network/auth tests should run only when `RUN_INTEGRATION_TESTS=1`.
  - They must skip if `KZ_LOGIN`/`KZ_PASSWORD` are missing.
  - Never fail CI/local runs solely because private credentials are absent.

Run:

```bash
go test ./...
```

Quality tools:

```bash
make fmt
make lint
make vet
make test
make race
make cover
```

## Configuration

Primary env vars:

- `TMDBAPIKEY`
- `RUTOR_SEARCH_URL`
- `KZ_SEARCH_URL`
- `KZ_LOGIN` (integration only)
- `KZ_PASSWORD` (integration only)
- `MONGO_URI` (optional save path)
- `MONGO_COLLECTION` (optional, default `movies`)
- `SAVE_TO_MONGO` (`true`/`false`, optional)
- `LOG_LEVEL` (`debug|info|warn|error`, optional)

## Local Skill

A project-specific skill is available at:

- `skills/moviestracker-maintainer/SKILL.md`

Use it for maintenance/refactoring and release readiness tasks in this repository.
