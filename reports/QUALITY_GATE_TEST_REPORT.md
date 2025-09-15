# Quality Gate Test Report

## Contract Tests (Vitest)
- Status: PASSED
- Files: 3
- Tests: 33 (32 passed, 1 skipped)
- Output: reports/contract-test-output.txt

## Simplified E2E Smoke (Shell + curl)
- Target endpoints:
  - Command REST: http://localhost:9090
  - Query GraphQL: http://localhost:8090
  - Frontend: http://localhost:3000
- Summary:
  - Passed: 4
  - Failed: 6
  - Total: 10
- Output: reports/e2e-test-output.txt

## Notes
- Backend services not reachable; health and GraphQL checks failed as expected without services running.
- Frontend is reachable and serves content.

## Next Steps
- Start local services via docker-compose or Makefile targets, then rerun simplified E2E.
- For CI: run E2E job with service containers (Postgres + backend binaries) to validate full path.
