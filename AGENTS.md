# Repository Guidelines

## Project Structure & Module Organization
- Backend Go services reside in `cmd/organization-command-service/` and `cmd/organization-query-service/`; shared middleware, auth, cache, and GraphQL helpers live under `internal/`.
- Database migrations live in `database/migrations/`, and reusable SQL helpers are under `sql/`.
- The React + Vite frontend sits in `frontend/`; feature slices are in `frontend/src/features/` with shared types in `frontend/src/shared/`.
- Tests span Go suites in `tests/` and `cmd/*`, frontend specs in `frontend/src/**/__tests__` and `frontend/tests/`, and Playwright E2E specs under `tests/e2e/` with configuration in `frontend/`.

## Build, Test, and Development Commands
- `make run-dev` starts Postgres, Redis, and both Go services on ports 9090/8090.
- `make build` compiles binaries into `bin/`; `make clean` resets build artifacts.
- `make test`, `make test-integration`, and `make coverage` run Go suites and produce coverage reports.
- Frontend commands: `cd frontend && npm run dev | test | build`; Playwright E2E lives under `npm run test:e2e` and spins up its own server.
- Auth utilities: `make run-auth-rs256-sim` serves JWKS; `make jwt-dev-mint` refreshes `.cache/dev.jwt`.

## Coding Style & Naming Conventions
- Go code follows idiomatic camelCase for internals and PascalCase for exports; run `make fmt` before committing.
- TypeScript uses ESLint, two-space indentation, and functional React components; share types through `frontend/src/shared/`.
- Preserve camelCase for API payloads and keep service logic isolated inside `cmd/*/internal/` packages.

## Testing Guidelines
- Go tests end with `_test.go`; tag integration suites appropriately and ensure `make test` passes locally.
- Frontend testing uses Vitest and Playwright; keep specs close to features under `frontend/src/**/__tests__` or `frontend/tests/`.
- Run `frontend/scripts/validate-field-naming*.js` and `node scripts/quality/architecture-validator.js` before pushing to match CI checks.

## Commit & Pull Request Guidelines
- Use Conventional Commits (e.g., `feat: add temporal validation`); keep each change focused.
- Pull requests should reference related issues, outline behavior changes, and attach test artifacts (logs or screenshots). Update `docs/reference/` if behavior shifts and document any deviations from established conventions.
