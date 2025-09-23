#!/usr/bin/env node

/**
 * 20250921-replace-temporal-validation.ts
 *
 * 自动将遗留的 temporalValidation 引用迁移至统一适配层。
 *
 * 能力：
 * - 默认执行真实替换，并保存文件。
 * - 通过 --check 仅输出将受影响的文件，不写回，返回码 1 表示需要处理。
 */

import { Project, SourceFile, ImportDeclaration, ExportDeclaration } from 'ts-morph';
import { resolve } from 'node:path';

interface Options {
  check: boolean;
}

const TARGET_REGEX = /temporalValidation(\.ts|\.js)?$/;
const NEW_SPECIFIER = '@/shared/utils/temporal-validation-adapter';

function parseOptions(): Options {
  const args = process.argv.slice(2);
  return {
    check: args.includes('--check') || args.includes('--dry-run'),
  };
}

function shouldReplace(moduleSpecifier: string): boolean {
  if (!moduleSpecifier) return false;
  if (moduleSpecifier.includes('temporal-validation-adapter')) return false;
  return TARGET_REGEX.test(moduleSpecifier);
}

function replaceModuleSpecifier(node: ImportDeclaration | ExportDeclaration): boolean {
  const moduleSpecifierValue = node.getModuleSpecifierValue();
  if (!shouldReplace(moduleSpecifierValue)) {
    return false;
  }

  node.setModuleSpecifier(NEW_SPECIFIER);
  return true;
}

function migrateSourceFile(sourceFile: SourceFile): boolean {
  let touched = false;

  sourceFile.getImportDeclarations().forEach(declaration => {
    if (replaceModuleSpecifier(declaration)) {
      touched = true;
    }
  });

  sourceFile.getExportDeclarations().forEach(declaration => {
    if (replaceModuleSpecifier(declaration)) {
      touched = true;
    }
  });

  return touched;
}

async function main() {
  const options = parseOptions();
  const tsConfigPath = resolve(process.cwd(), 'tsconfig.app.json');

  const project = new Project({
    tsConfigFilePath: tsConfigPath,
  });

  const sourceFiles = project.getSourceFiles(['src/**/*.ts', 'src/**/*.tsx']);
  const changedFiles: string[] = [];

  sourceFiles.forEach(sourceFile => {
    if (migrateSourceFile(sourceFile)) {
      changedFiles.push(sourceFile.getFilePath());
    }
  });

  if (changedFiles.length === 0) {
    const msg = options.check ? '[CHECK] 未检测到需要替换的 temporalValidation 引用。' : '[MIGRATE] 未找到需要替换的 temporalValidation 引用。';
    console.log(msg);
    process.exit(0);
  }

  if (options.check) {
    console.log('[CHECK] 以下文件需要替换 temporalValidation 引用为统一适配层：');
    changedFiles.forEach(filePath => console.log(` - ${filePath}`));
    process.exit(1);
  }

  await project.save();
  console.log(`[MIGRATE] 已更新 ${changedFiles.length} 个文件：`);
  changedFiles.forEach(filePath => console.log(` - ${filePath}`));
}

main().catch(error => {
  console.error('[MIGRATE] 迁移脚本执行失败:', error);
  process.exit(1);
});
