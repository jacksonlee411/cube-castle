#!/usr/bin/env python3
"""验证契约文件是否与基线一致"""
import json, hashlib, sys
from pathlib import Path
from typing import Optional
import re

BASELINE_PATH = Path(__file__).parent / 'inventory.baseline.json'
TARGET_FILES = [
    Path('docs/api/openapi.yaml'),
    Path('docs/api/schema.graphql'),
    Path('shared/contracts/organization.json'),
    Path('internal/types/contract_gen.go'),
    Path('frontend/src/shared/types/contract_gen.ts'),
]

def normalized_text(path: Path) -> str:
    """
    标准化文件内容以获得稳定快照：
    - 对 shared/contracts/organization.json：忽略 generatedAt 字段（OpenAPI/GraphQL）以消除每次生成时间差异
    - 其他文件：按原文计算
    """
    text = path.read_text()
    if path.as_posix() == 'shared/contracts/organization.json':
        try:
            obj = json.loads(text)
            # 清理生成时间戳字段（非业务语义）
            if isinstance(obj.get('metadata'), dict):
                obj['metadata'].pop('generatedAt', None)
            if isinstance(obj.get('graphql'), dict):
                obj['graphql'].pop('generatedAt', None)
            # 以紧凑格式序列化，避免空白差异
            return json.dumps(obj, sort_keys=True, separators=(',', ':'))
        except Exception:
            # 回退到原文
            return text
    return text

def normalized_sha(path: Path) -> str:
    # JSON 中间层做语义归一；其余文件按原始字节计算（避免换行符转换导致的误差）
    if path.as_posix() == 'shared/contracts/organization.json':
        text = normalized_text(path)
        return hashlib.sha256(text.encode('utf-8')).hexdigest()
    else:
        data = path.read_bytes()
        return hashlib.sha256(data).hexdigest()

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
