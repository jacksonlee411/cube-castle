#!/bin/bash

# E2E测试验证脚本 - 无需浏览器依赖
# 验证测试代码质量和结构完整性

echo "🧪 Cube Castle E2E测试严格验证"
echo "==========================================="
echo ""

# 1. 验证项目结构和依赖
echo "📋 步骤1: 验证项目结构..."
echo "✅ 项目目录: $(pwd)"
echo "✅ Node.js版本: $(node --version)"
echo "✅ NPM版本: $(npm --version)"
echo "✅ Next.js版本: $(node -p "require('./package.json').dependencies.next")"
echo "✅ Playwright版本: $(npx playwright --version)"
echo ""

# 2. 验证测试文件完整性
echo "📁 步骤2: 验证测试文件结构..."
TEST_FILES=(
    "tests/e2e/pages/employees.spec.ts"
    "tests/e2e/pages/positions.spec.ts"
    "tests/e2e/pages/organization-chart.spec.ts"
    "tests/e2e/pages/workflow-detail.spec.ts"
    "tests/e2e/pages/employee-position-history.spec.ts"
    "tests/e2e/pages/admin-graph-sync.spec.ts"
    "tests/e2e/pages/workflow-demo.spec.ts"
    "tests/e2e/utils/test-helpers.ts"
)

total_lines=0
test_count=0

for file in "${TEST_FILES[@]}"; do
    if [ -f "$file" ]; then
        lines=$(wc -l < "$file")
        total_lines=$((total_lines + lines))
        
        if [[ "$file" == *.spec.ts ]]; then
            tests=$(grep -c "test(" "$file" 2>/dev/null || echo 0)
            test_count=$((test_count + tests))
            echo "✅ $(basename "$file"): $lines 行, $tests 个测试"
        else
            echo "✅ $(basename "$file"): $lines 行 (工具类)"
        fi
    else
        echo "❌ $file (缺失)"
        exit 1
    fi
done

echo ""
echo "📊 测试代码统计:"
echo "   总代码行数: $total_lines 行"
echo "   测试场景数: $test_count 个"
echo "   跨浏览器测试用例: $((test_count * 3)) 个"
echo ""

# 3. TypeScript类型检查
echo "🔍 步骤3: TypeScript类型检查..."
if npx tsc --noEmit --project . > /tmp/ts-check.log 2>&1; then
    echo "✅ TypeScript类型检查通过"
else
    echo "❌ TypeScript类型检查失败:"
    head -10 /tmp/ts-check.log
    echo ""
fi

# 4. ESLint代码质量检查
echo "🔧 步骤4: ESLint代码质量检查..."
if npx eslint tests/e2e/**/*.ts --quiet > /tmp/eslint-check.log 2>&1; then
    echo "✅ ESLint代码质量检查通过"
else
    echo "⚠️  ESLint发现问题:"
    head -10 /tmp/eslint-check.log
    echo ""
fi

# 5. 验证测试文件语法
echo "📝 步骤5: 测试文件语法验证..."
syntax_errors=0

for file in "${TEST_FILES[@]}"; do
    if [[ "$file" == *.spec.ts ]]; then
        if node -c <(npx tsc --target es2020 --module commonjs --outDir /tmp "$file" && cat "/tmp/$(basename "${file%.ts}.js")") 2>/dev/null; then
            echo "✅ $(basename "$file") 语法正确"
        else
            echo "❌ $(basename "$file") 语法错误"
            syntax_errors=$((syntax_errors + 1))
        fi
    fi
done

echo ""

# 6. 验证开发服务器连接
echo "🌐 步骤6: 验证开发服务器..."
if curl -s -f http://localhost:3000 > /dev/null; then
    echo "✅ 开发服务器运行正常 (localhost:3000)"
    
    # 测试关键路由
    echo ""
    echo "🔍 验证关键页面路由:"
    ROUTES=("/" "/employees" "/positions" "/organization/chart" "/workflows/1" "/admin/graph-sync" "/workflows/demo")
    
    for route in "${ROUTES[@]}"; do
        status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:3000$route")
        if [ "$status" = "200" ]; then
            echo "✅ $route (HTTP $status)"
        else
            echo "⚠️  $route (HTTP $status)"
        fi
    done
else
    echo "❌ 开发服务器未运行"
    echo "   请在另一个终端运行: npm run dev"
fi

echo ""

# 7. 依赖安装状态检查
echo "🔧 步骤7: 浏览器依赖状态..."

# 检查Playwright浏览器安装
if [ -d "$HOME/.cache/ms-playwright" ]; then
    browser_dirs=$(ls -1 "$HOME/.cache/ms-playwright" | grep -E "chromium|firefox|webkit" | wc -l)
    echo "✅ Playwright浏览器已下载 ($browser_dirs 个)"
else
    echo "❌ Playwright浏览器未安装"
fi

# 检查系统依赖
echo ""
echo "🖥️  系统依赖检查:"
SYSTEM_DEPS=("libnspr4" "libnss3" "libasound2")
missing_deps=0

for dep in "${SYSTEM_DEPS[@]}"; do
    if dpkg -l 2>/dev/null | grep -q "^ii.*$dep"; then
        echo "✅ $dep 已安装"
    else
        echo "❌ $dep 缺失"
        missing_deps=$((missing_deps + 1))
    fi
done

echo ""

# 8. 测试执行能力评估
echo "🎯 步骤8: 测试执行能力评估..."

if [ $missing_deps -eq 0 ]; then
    echo "✅ 所有依赖项已满足，可以执行完整E2E测试"
    echo "   执行命令: npm run test:e2e"
elif [ $missing_deps -le 2 ]; then
    echo "⚠️  缺少 $missing_deps 个系统依赖，建议安装后执行测试"
    echo "   安装命令: sudo apt-get install libnspr4 libnss3 libasound2"
else
    echo "❌ 缺少多个系统依赖，建议使用Docker环境"
    echo "   Docker命令: docker build -f Dockerfile.e2e -t cube-castle-e2e ."
fi

echo ""

# 9. 最终评估报告
echo "📋 最终测试就绪评估报告"
echo "=================================="

score=0
max_score=8

# 评分标准
[ $total_lines -gt 2000 ] && score=$((score + 1))
[ $test_count -gt 70 ] && score=$((score + 1))
[ $syntax_errors -eq 0 ] && score=$((score + 1))
[ -f "playwright.config.ts" ] && score=$((score + 1))
[ -f "tests/e2e/utils/test-helpers.ts" ] && score=$((score + 1))
curl -s -f http://localhost:3000 > /dev/null && score=$((score + 1))
[ -d "$HOME/.cache/ms-playwright" ] && score=$((score + 1))
[ $missing_deps -le 1 ] && score=$((score + 1))

echo "🏆 测试就绪得分: $score/$max_score"

if [ $score -ge 7 ]; then
    echo "✅ 优秀: E2E测试框架完全就绪"
    echo "🚀 建议立即执行完整测试套件"
elif [ $score -ge 5 ]; then
    echo "⚠️  良好: E2E测试框架基本就绪，需要安装浏览器依赖"
    echo "🔧 建议安装依赖后执行测试"
else
    echo "❌ 需要改进: 存在关键问题需要解决"
fi

echo ""
echo "🎯 下一步行动计划:"
echo "1. 安装缺失的系统依赖项"
echo "2. 确保开发服务器运行在localhost:3000"
echo "3. 执行完整E2E测试: npm run test:e2e"
echo "4. 查看HTML测试报告: playwright-report/index.html"
echo ""

echo "状态: ✅ 测试框架验证完成"