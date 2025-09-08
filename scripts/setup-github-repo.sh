#!/bin/bash

# GitHub仓库创建和推送脚本
# 使用方法: ./scripts/setup-github-repo.sh YOUR_GITHUB_USERNAME

set -e

GITHUB_USERNAME=${1}
REPO_NAME="cube-castle"

if [ -z "$GITHUB_USERNAME" ]; then
    echo "❌ 请提供GitHub用户名: ./scripts/setup-github-repo.sh YOUR_USERNAME"
    exit 1
fi

echo "🚀 开始设置GitHub仓库..."
echo "📂 仓库: $GITHUB_USERNAME/$REPO_NAME"

# 1. 检查Git状态
echo "📋 检查Git状态..."
git status

# 2. 提交当前更改（如果有）
if ! git diff-index --quiet HEAD --; then
    echo "💾 提交当前更改..."
    git add .
    git commit -m "📚 项目文档和配置更新

🎯 关键更新:
- CLAUDE.md: P3企业级防控系统完整文档
- README.md: 统一配置架构升级
- GitHub Actions: 11个工作流配置完成
- 质量门禁: 契约测试+重复代码检测+架构守护

🔧 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
fi

# 3. 添加远程仓库
echo "🔗 添加远程仓库..."
if git remote get-url origin >/dev/null 2>&1; then
    echo "⚠️  远程仓库已存在，跳过添加"
else
    git remote add origin "https://github.com/$GITHUB_USERNAME/$REPO_NAME.git"
fi

# 4. 推送主要分支
echo "📤 推送代码到GitHub..."
git push -u origin master || echo "⚠️  master分支推送失败，可能已存在"

# 推送其他重要分支
for branch in develop feature/duplicate-code-elimination; do
    if git show-ref --verify --quiet refs/heads/$branch; then
        echo "📤 推送分支: $branch"
        git push -u origin $branch || echo "⚠️  $branch分支推送失败"
    fi
done

echo ""
echo "✅ GitHub仓库设置完成！"
echo "🔗 仓库地址: https://github.com/$GITHUB_USERNAME/$REPO_NAME"
echo "⚡ Actions页面: https://github.com/$GITHUB_USERNAME/$REPO_NAME/actions"
echo ""
echo "📋 下一步操作:"
echo "   1. 访问GitHub仓库确认代码已上传"
echo "   2. 检查Actions工作流是否自动运行"
echo "   3. 配置分支保护规则（如需要）"
echo "   4. 邀请协作者（如需要）"