const fs = require('fs');
const path = require('path');

function findTsFiles(dir) {
    const files = [];
    const items = fs.readdirSync(dir);
    
    for (const item of items) {
        const fullPath = path.join(dir, item);
        const stat = fs.statSync(fullPath);
        
        if (stat.isDirectory() && item !== 'node_modules') {
            files.push(...findTsFiles(fullPath));
        } else if (item.endsWith('.ts') || item.endsWith('.tsx')) {
            files.push(fullPath);
        }
    }
    return files;
}

function fixFetchCalls(filePath) {
    let content = fs.readFileSync(filePath, 'utf8');
    let modified = false;
    
    // 添加统一客户端导入（如果使用了fetch但没有导入）
    if (content.includes('fetch(') && !content.includes('unifiedRESTClient') && !content.includes('unifiedGraphQLClient')) {
        // 检查是否需要添加导入
        const hasImports = content.includes('import');
        const importLine = "import { unifiedRESTClient, unifiedGraphQLClient } from '../shared/api/unified-client';\n";
        
        if (hasImports) {
            // 在第一个import后添加
            const firstImportIndex = content.indexOf('import');
            const lineEnd = content.indexOf('\n', firstImportIndex);
            content = content.slice(0, lineEnd + 1) + importLine + content.slice(lineEnd + 1);
            modified = true;
        }
    }
    
    // 简单的fetch替换（需要手动处理复杂情况）
    const originalContent = content;
    
    // 标记需要手动处理的fetch调用
    if (content.includes('fetch(')) {
        const lines = content.split('\n');
        const updatedLines = lines.map(line => {
            if (line.includes('fetch(') && !line.includes('// TODO: 手动修复')) {
                return line + ' // TODO: 手动修复 - 使用unifiedRESTClient或unifiedGraphQLClient';
            }
            return line;
        });
        content = updatedLines.join('\n');
        modified = true;
    }
    
    if (modified) {
        fs.writeFileSync(filePath, content, 'utf8');
        return true;
    }
    return false;
}

const files = findTsFiles('src');
let fixedFiles = 0;

for (const file of files) {
    if (fixFetchCalls(file)) {
        fixedFiles++;
        console.log(`    ✅ ${file}`);
    }
}

console.log(`  → 处理了 ${fixedFiles} 个文件的fetch调用`);
