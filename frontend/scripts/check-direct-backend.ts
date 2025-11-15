#!/usr/bin/env tsx
/**
 * 检查源码与产物中是否存在对后端服务的直连（:9090 或 :8090）
 * 目的：确保前端仅通过“单基址代理”（/api/v1、/graphql）访问后端
 *
 * 规则：
 * - 扫描 frontend/src 与 frontend/dist（如存在）
 * - 匹配 http(s)/ws(s) + :9090|:8090，或 localhost/127.0.0.1 + :9090|:8090
 * - 排除常见非 URL 数字场景（尽量通过模式限定减少误报）
 * - 命中则退出码 1；未命中退出码 0
 */

import { readdirSync, statSync, readFileSync, existsSync } from 'fs';
import { join, extname } from 'path';

const ROOT = process.cwd();
const TARGET_DIRS = ['src', 'dist'].map((d) => join(ROOT, d)).filter(existsSync);
const EXCLUDE_DIRS = new Set(['node_modules', '.git', 'test-results', 'logs', 'playwright-report']);
const EXTS = new Set(['.ts', '.tsx', '.js', '.jsx', '.json', '.html', '.map', '.css']);

// 直连端口匹配：要求出现协议/主机关键字 + 端口
const PATTERNS = [
  /https?:\/\/[^\s"'`]+:(9090|8090)\b/ig,
  /wss?:\/\/[^\s"'`]+:(9090|8090)\b/ig,
  /\b(?:localhost|127\.0\.0\.1)\s*:(9090|8090)\b/ig,
];

type Hit = { file: string; line: number; column: number; snippet: string };
const hits: Hit[] = [];

function scanFile(filePath: string) {
  try {
    const content = readFileSync(filePath, 'utf8');
    const lines = content.split('\n');
    lines.forEach((line, i) => {
      // 跳过明显注释行，降低误报
      const trimmed = line.trim();
      if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) return;

      for (const re of PATTERNS) {
        let m: RegExpExecArray | null;
        // 复制新的正则以重置 lastIndex
        const pattern = new RegExp(re.source, re.flags);
        while ((m = pattern.exec(line)) !== null) {
          const col = m.index + 1;
          const snippet = line.length > 160 ? `${line.slice(0, 157)}…` : line;
          hits.push({ file: filePath, line: i + 1, column: col, snippet });
        }
      }
    });
  } catch {
    // 忽略不可读文件
  }
}

function scanDir(dir: string) {
  const entries = readdirSync(dir);
  for (const name of entries) {
    const full = join(dir, name);
    let s: ReturnType<typeof statSync>;
    try {
      s = statSync(full);
    } catch {
      continue;
    }
    if (s.isDirectory()) {
      if (EXCLUDE_DIRS.has(name)) continue;
      scanDir(full);
    } else if (s.isFile()) {
      const ext = extname(name).toLowerCase();
      if (!EXTS.has(ext)) continue;
      scanFile(full);
    }
  }
}

if (TARGET_DIRS.length === 0) {
  console.log('ℹ️ 未发现可扫描目录（src/dist），跳过检查');
  process.exit(0);
}

for (const d of TARGET_DIRS) {
  scanDir(d);
}

if (hits.length > 0) {
  console.error('🚫 检测到直连后端端口（应通过单基址代理访问）：');
  for (const h of hits) {
    const rel = h.file.replace(`${ROOT}/`, '');
    console.error(`  ${rel}:${h.line}:${h.column}`);
    console.error(`    ${h.snippet}`);
  }
  process.exit(1);
}

console.log('✅ 未发现直连后端端口，符合单基址代理约束');
process.exit(0);

