#!/usr/bin/env python3
"""验证契约文件是否与基线一致"""
import json, hashlib, sys
from pathlib import Path

BASELINE_PATH = Path(__file__).parent / 'inventory.baseline.json'
TARGET_FILES = [
    Path('docs/api/openapi.yaml'),
    Path('docs/api/schema.graphql'),
    Path('shared/contracts/organization.json'),
]

def normalized_sha(path: Path) -> str:
    text = path.read_text()
    return hashlib.sha256(text.encode('utf-8')).hexdigest()

def load_baseline():
    if not BASELINE_PATH.exists():
        print('Baseline not found:', BASELINE_PATH)
        sys.exit(1)
    return json.loads(BASELINE_PATH.read_text())['files']

def main() -> int:
    baseline = load_baseline()
    mismatches = []
    for target in TARGET_FILES:
        sha = normalized_sha(target)
        entry = baseline.get(str(target))
        if not entry or entry['sha256'] != sha:
            mismatches.append(str(target))
    if mismatches:
        print('Contract snapshot mismatch detected for:')
        for path in mismatches:
            print('  -', path)
        return 2
    print('Contract snapshot verified successfully.')
    return 0

if __name__ == '__main__':
    sys.exit(main())
