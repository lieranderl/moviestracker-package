---
name: moviestracker-maintainer
description: Maintain and evolve moviestracker-package with safe parser changes, pipeline correctness, and production-focused Go practices.
---

# MoviesTracker Maintainer Skill

Use this skill when changing parser logic, pipeline orchestration, enrichment flow, or persistence behavior in this repository.

## Goals

1. Keep tracker parsing resilient to malformed HTML.
2. Keep concurrency/cancellation behavior correct and leak-free.
3. Improve code quality without breaking existing API usage.
4. Keep docs and tests aligned with behavior changes.

## Workflow

1. Inspect affected packages:
   - `executor/`
   - `internal/rutor/`
   - `internal/kinozal/`
   - `internal/movies/`
   - `pkg/pipeline/`
2. Implement smallest safe change set.
3. Add/adjust tests:
   - unit tests for logic changes
   - integration tests must skip when credentials are missing
4. Verify:
   - `go test ./...`
5. Update docs when behavior/config changes:
   - `README.md`
   - `AGENTS.md`

## Guardrails

- Do not use `log.Fatal` in library code.
- Prefer returning `error` and aggregating at executor/main boundaries.
- Never assume channels remain open; always handle closure.
- Avoid panics from index access after regex/split operations.
- Keep parser heuristics deterministic and easy to reason about.
