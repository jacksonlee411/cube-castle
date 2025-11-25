#!/usr/bin/env python3
"""
Utility to locate the most recent Playwright results JSON for Plan 254 gates.

GitHub Actions previously inlined this logic inside the workflow file, which
made the YAML fragile and led to parse failures. Moving the search into a
tiny Python helper keeps the workflow readable and prevents indentation
mistakes from breaking the entire gate.
"""
from __future__ import annotations

import pathlib
import sys
from typing import Iterable


def iter_result_files(base: pathlib.Path) -> Iterable[pathlib.Path]:
  """Yield candidate result files under the provided base directory."""
  if not base.exists():
    return
  yield from (
      path for path in base.glob("results-*.json") if path.is_file()
  )


def main() -> int:
  repo_root = pathlib.Path(__file__).resolve().parent.parent.parent
  results_dir = repo_root / "logs" / "plan254"
  candidates = sorted(
      iter_result_files(results_dir),
      key=lambda path: path.stat().st_mtime,
      reverse=True,
  )

  if not candidates:
    return 0

  # Print the latest file path for the caller shell script.
  print(candidates[0])
  return 0


if __name__ == "__main__":
  sys.exit(main())
