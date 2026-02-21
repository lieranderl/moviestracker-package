# Contributing

## Prerequisites

- Go `1.25.x` or newer
- `golangci-lint` (for local linting)

Install lint tool:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Local Workflow

1. Sync dependencies:

```bash
make tidy
```

2. Install pre-commit hooks (recommended):

```bash
make hooks
```

3. Format and lint:

```bash
make fmt
make lint
```

4. Run quality checks:

```bash
make vet
make test
make race
make cover
```

5. Build executable:

```bash
make build
```

Optional: run all pre-commit checks manually:

```bash
make precommit
```

## Testing Rules

- Keep unit tests deterministic and offline.
- Integration tests are opt-in:
  - set `RUN_INTEGRATION_TESTS=1`
  - provide required credentials (`KZ_LOGIN`, `KZ_PASSWORD`)
- New business logic should include table-driven tests when practical.

## Coding Standards

- Keep package responsibilities narrow.
- Return wrapped errors with `%w`.
- Respect context cancellation and timeouts.
- Avoid `log.Fatal` in non-`main` packages.
- Run `make fmt vet test` before opening PRs.
