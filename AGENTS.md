# Repository Guidelines

## Project Structure & Module Organization
- Backend Go services live in `cmd/organization-command-service/` and `cmd/organization-query-service/`; shared logic (auth, GraphQL, cache, middleware) is in `internal/`.
- Database migrations sit in `database/migrations/`, while reusable SQL helpers stay under `sql/`.
- The React+Vite frontend resides in `frontend/`; feature slices live in `frontend/src/features/`, and cross-cutting types are in `frontend/src/shared/`.
- Tests are distributed: Go unit/integration suites in `tests/` and `cmd/*`, frontend unit/integration specs in `frontend/src/**/__tests__` and `frontend/tests/`, and Playwright E2E specs in `tests/e2e/` with config in `frontend/`.
- Contract and reference material is captured in `docs/api/` and `docs/reference/` respectively.

## Build, Test, and Development Commands
- `make run-dev`: boots Postgres, Redis, and both Go services (ports 9090/8090).
- `make build` / `make clean`: compile binaries to `bin/` or reset artifacts.
- `make test`, `make test-integration`, `make coverage`: execute Go suites and produce coverage.
- `make lint`, `make security`: run `golangci-lint` and `gosec`.
- Frontend: `cd frontend && npm run dev | test | test:e2e | build`; Playwright starts its own web server.
- Auth helpers: `make run-auth-rs256-sim` to expose JWKS; `make jwt-dev-mint` to refresh `.cache/dev.jwt`.

## Coding Style & Naming Conventions
- Run `make fmt` before committing; follow Go idioms (camelCase internals, PascalCase exports) and keep handlers/services under `cmd/*/internal/`.
- TypeScript uses ESLint, two-space indentation, and functional React components; share types through `frontend/src/shared/`.
- API contracts prefer camelCase fields, `{code}` path params, GraphQL for queries, REST for commands.

## Testing Guidelines
- Go tests end with `_test.go`; tag integration suites appropriately and rely on `make test` before submitting.
- Frontend uses Vitest for unit/integration and Playwright for E2E via `npm run test:e2e`.
- Run `frontend/scripts/validate-field-naming*.js` and `node scripts/quality/architecture-validator.js` to satisfy CI validations.

## Commit & Pull Request Guidelines
- Prefer focused Conventional Commits (e.g., `feat: add temporal validation`).
- PRs should describe changes, reference related issues, include test artifacts (logs or screenshots), and update `docs/reference/` when behavior shifts.
- Document deviations from established conventions and provide migration notes for reviewers.
