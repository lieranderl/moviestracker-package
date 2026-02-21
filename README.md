# moviestracker-package

[![CI](https://github.com/lieranderl/moviestracker-package/actions/workflows/ci.yml/badge.svg)](https://github.com/lieranderl/moviestracker-package/actions/workflows/ci.yml)

Concurrent Go pipeline for finding torrents on Rutor/Kinozal, enriching metadata from TMDB, and saving results to MongoDB.

## Features

- Parallel tracker scraping (Rutor + Kinozal)
- Torrent deduplication by magnet hash
- Movie-level grouping and TMDB enrichment
- Optional persistence to MongoDB (mongo-driver v2)
- Context-aware pipeline stages with aggregated errors

## Requirements

- Go `1.25+` (target in this repo)
- TMDB API key
- Network access to trackers/TMDB

Optional:

- Kinozal credentials (`KZ_LOGIN`, `KZ_PASSWORD`) for magnet enrichment
- MongoDB credentials for persistence

## Setup

1. Install dependencies:

```bash
make tidy
```

2. Create local env file:

```bash
cp .env.example .env
```

3. Fill required values in `.env`:

- `TMDBAPIKEY`
- `RUTOR_SEARCH_URL`
- `KZ_SEARCH_URL`
- `LOG_LEVEL` (`debug|info|warn|error`, default `info`)

## Run

```bash
go run ./cmd -query "Bad Boys" -year 2024 -movie=true
```

Persist to MongoDB:

```bash
go run ./cmd -query "Bad Boys" -year 2024 -movie=true -save=true -collection=movies
```

Flags:

- `-query`: movie/series title
- `-year`: release year
- `-movie`: `true` (movies) or `false` (series)
- `-save`: enable MongoDB persistence (`MONGO_URI` required)
- `-collection`: MongoDB collection (default `movies`)

## Quality Commands

```bash
make fmt
make lint
make vet
make test
make race
make cover
make build
```

## Pre-commit Checks

Install and enable git hooks:

```bash
pipx install pre-commit
make hooks
```

Run all configured checks manually:

```bash
make precommit
```

Configured hooks include:

- basic file hygiene (`trailing-whitespace`, EOF, yaml, merge conflicts, large files)
- `gofmt` on staged Go files
- `go vet` on changed Go packages
- `go test -short` on changed Go packages
- `golangci-lint` on changed Go packages

## Testing Notes

- Integration tests are opt-in:
  - set `RUN_INTEGRATION_TESTS=1`
  - configure `KZ_LOGIN` and `KZ_PASSWORD`
- By default, integration tests are skipped to keep local/CI deterministic.
- Coverage policy:
  - short-term target: overall `40%+`
  - critical packages (`executor`, `internal/torrents`, `pkg/pipeline`): `80%+`
  - avoid chasing coverage on brittle network/auth integration code; prioritize deterministic business logic.

## Architecture (Short)

1. `cmd/main.go`: CLI + environment bootstrap.
2. `executor/`: orchestration pipeline and persistence stages.
3. `internal/rutor`, `internal/kinozal`: tracker-specific parsing/adapters.
4. `internal/movies`, `internal/torrents`: domain models + enrichment/persistence helpers.
5. `pkg/pipeline`: generic producer/worker/merge primitives.

## Docker

Build and run:

```bash
docker build -t moviestracker:local .
docker run --rm --env-file .env moviestracker:local
```

## Backward Compatibility

Legacy methods with typos are preserved as wrappers:

- `RunTrackersSearchPipilene(...)`
- `RunRutorPipiline(...)`

Preferred methods:

- `RunTrackersSearchPipeline(...)`
- `RunRutorPipeline()`

## Contributing

See `CONTRIBUTING.md`.

## Security & Ops Notes

- Do not commit `.env`; use `.env.example` as the template.
- Prefer secret managers (GitHub Secrets, Vault, GCP Secret Manager, etc.) in CI/prod.
- Avoid logging credentials, connection strings, and tracker auth values.
- DB writes should always run with context timeouts (already enforced in executor save paths).
- If HTTP endpoints are added later:
  - configure server/read/write/idle timeouts
  - enforce request body size limits
  - implement graceful shutdown with context cancellation
  - expose `/healthz` and `/readyz`
  - include request IDs in structured logs

## License

No license file is currently present in this repository.
