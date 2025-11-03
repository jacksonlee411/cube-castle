#!/usr/bin/env python3
"""
Plan 211 Phase1 import审核/迁移辅助脚本

功能：
 1. 扫描 Go 源码中的 legacy 模块导入（module cube-castle-deployment-test 等）。
 2. 输出命中文件及替换建议，汇总统计。
 3. 可选 `--apply` 模式自动执行字符串替换（谨慎使用，文件会就地修改）。

用法：
  scripts/tools/phase1-import-audit.py             # 仅检测
  scripts/tools/phase1-import-audit.py --apply     # 检测 + 自动替换

默认扫描仓库根目录下的 `.go` 文件（排除 `vendor/`、`node_modules/`）。
"""

from __future__ import annotations

import argparse
import pathlib
import re
import sys
from typing import Dict, List, Tuple

REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]

LEGACY_PATTERNS: Dict[str, str] = {
    "organization-command-service/": "cube-castle/cmd/hrms-server/command/",
    "organization-query-service/": "cube-castle/cmd/hrms-server/query/",
    "cube-castle-deployment-test/internal": "cube-castle/internal",
    "cube-castle-deployment-test/pkg/health": "cube-castle/pkg/health",
    "cube-castle-deployment-test/cmd/hrms-server/query": "cube-castle/cmd/hrms-server/query",
    "cube-castle-deployment-test/cmd/hrms-server/command": "cube-castle/cmd/hrms-server/command",
    "cube-castle-deployment-test": "cube-castle",
}

IGNORE_DIRS = {"vendor", "node_modules", ".git", "frontend/node_modules"}


def iter_go_files(base: pathlib.Path) -> List[pathlib.Path]:
    files: List[pathlib.Path] = []
    for path in base.rglob("*.go"):
        if any(part in IGNORE_DIRS for part in path.parts):
            continue
        files.append(path)
    return files


def audit_file(path: pathlib.Path) -> List[Tuple[str, str]]:
    content = path.read_text(encoding="utf-8")
    findings: List[Tuple[str, str]] = []
    for legacy, target in LEGACY_PATTERNS.items():
        if re.search(re.escape(legacy), content):
            findings.append((legacy, target))
    return findings


def apply_replacements(path: pathlib.Path, replacements: List[Tuple[str, str]]) -> None:
    content = path.read_text(encoding="utf-8")
    for legacy, target in replacements:
        content = content.replace(legacy, target)
    path.write_text(content, encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser(description="Plan 211 Phase1 import audit helper")
    parser.add_argument(
        "--apply",
        action="store_true",
        help="自动执行替换（默认仅检测）",
    )
    args = parser.parse_args()

    go_files = iter_go_files(REPO_ROOT)
    if not go_files:
        print("未找到 Go 源文件，终止。")
        return 1

    total_matches: Dict[str, int] = {legacy: 0 for legacy in LEGACY_PATTERNS}
    affected_files: List[pathlib.Path] = []

    for file_path in go_files:
        findings = audit_file(file_path)
        if not findings:
            continue
        affected_files.append(file_path)
        print(f"\n[{file_path.relative_to(REPO_ROOT)}]")
        for legacy, target in findings:
            print(f"  - {legacy}  →  {target}")
            total_matches[legacy] += 1
        if args.apply:
            apply_replacements(file_path, findings)

    if not affected_files:
        print("✅ 未检测到 legacy 导入，当前仓库已完全迁移。")
    else:
        print("\n=== 汇总统计 ===")
        for legacy, count in total_matches.items():
            if count:
                print(f"{legacy:<50} {count:>4} 次")

        print(f"\n涉及文件总数：{len(affected_files)}")
        print("⚠️ 请人工复审自动替换结果，确认后提交。")

    return 0


if __name__ == "__main__":
    sys.exit(main())
