# 1.1.1 CoreHR Repository层验证工具启动脚本 (PowerShell版本)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

Write-ColorOutput "🏰 Cube Castle - 1.1.1 验证工具启动器" "Blue"
Write-Host "==========================================" -ForegroundColor Gray
Write-Host ""

# 检查Go服务是否运行
function Test-GoService {
    Write-ColorOutput "🔍 检查Go服务状态..." "Blue"
    
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 5 -UseBasicParsing
        if ($response.StatusCode -eq 200) {
            Write-ColorOutput "✅ Go服务正在运行 (http://localhost:8080)" "Green"
            return $true
        }
    }
    catch {
        Write-ColorOutput "❌ Go服务未运行" "Red"
        return $false
    }
}

# 启动Go服务
function Start-GoService {
    Write-ColorOutput "🚀 启动Go服务..." "Yellow"
    
    # 检查是否在正确的目录
    if (-not (Test-Path "cmd/server/main.go")) {
        Write-ColorOutput "❌ 请在go-app目录下运行此脚本" "Red"
        exit 1
    }
    
    # 检查Go是否安装
    try {
        $goVersion = go version
        Write-ColorOutput "✅ Go已安装: $goVersion" "Green"
    }
    catch {
        Write-ColorOutput "❌ Go未安装或不在PATH中" "Red"
        exit 1
    }
    
    # 清理并重新构建
    Write-ColorOutput "📦 清理并重新构建项目..." "Blue"
    try {
        go clean -cache
        go mod tidy
        Write-ColorOutput "✅ 项目构建完成" "Green"
    }
    catch {
        Write-ColorOutput "❌ 项目构建失败" "Red"
        exit 1
    }
    
    # 启动服务
    Write-ColorOutput "🚀 启动Go服务..." "Green"
    Write-ColorOutput "💡 服务将在后台运行，按Ctrl+C停止" "Yellow"
    Write-Host ""
    
    # 启动Go服务
    Start-Process -FilePath "go" -ArgumentList "run", "cmd/server/main.go" -NoNewWindow
    
    # 等待服务启动
    Write-ColorOutput "⏳ 等待服务启动..." "Blue"
    Start-Sleep -Seconds 5
    
    # 检查服务是否成功启动
    $retryCount = 0
    $maxRetries = 10
    
    while ($retryCount -lt $maxRetries) {
        if (Test-GoService) {
            Write-ColorOutput "✅ Go服务启动成功！" "Green"
            return
        }
        
        $retryCount++
        Write-ColorOutput "⏳ 重试 $retryCount/$maxRetries..." "Yellow"
        Start-Sleep -Seconds 2
    }
    
    Write-ColorOutput "❌ Go服务启动失败" "Red"
    exit 1
}

# 打开验证网页
function Open-VerificationPage {
    Write-Host ""
    Write-ColorOutput "🌐 打开验证网页..." "Green"
    
    # 检查验证文件是否存在
    if (-not (Test-Path "verify_1.1.1.html")) {
        Write-ColorOutput "❌ 验证文件 verify_1.1.1.html 不存在" "Red"
        exit 1
    }
    
    # 获取完整路径
    $htmlPath = (Get-Item "verify_1.1.1.html").FullName
    $fileUrl = "file:///$($htmlPath.Replace('\', '/'))"
    
    try {
        # 尝试打开浏览器
        Start-Process $fileUrl
        Write-ColorOutput "✅ 验证网页已打开！" "Green"
    }
    catch {
        Write-ColorOutput "⚠️ 无法自动打开浏览器，请手动打开文件:" "Yellow"
        Write-ColorOutput "   $htmlPath" "Blue"
    }
}

# 显示使用说明
function Show-Instructions {
    Write-Host ""
    Write-ColorOutput "📋 使用说明:" "Blue"
    Write-Host "1. 在验证网页中，您可以查看1.1.1的实现状态" -ForegroundColor White
    Write-Host "2. 点击API测试按钮来验证实际功能" -ForegroundColor White
    Write-Host "3. 查看总体进度和功能覆盖度" -ForegroundColor White
    Write-Host "4. 了解下一步开发建议" -ForegroundColor White
    Write-Host ""
    Write-ColorOutput "🔗 API端点:" "Yellow"
    Write-Host "   - 员工管理: http://localhost:8080/api/v1/corehr/employees" -ForegroundColor White
    Write-Host "   - 组织管理: http://localhost:8080/api/v1/corehr/organizations" -ForegroundColor White
    Write-Host "   - 发件箱: http://localhost:8080/api/v1/outbox" -ForegroundColor White
    Write-Host ""
    Write-ColorOutput "🎯 验证目标:" "Green"
    Write-Host "   ✅ 替换所有Mock数据" -ForegroundColor White
    Write-Host "   ✅ 实现真实的数据库操作" -ForegroundColor White
    Write-Host "   ✅ 实现完整的业务逻辑" -ForegroundColor White
    Write-Host ""
}

# 主函数
function Main {
    Write-ColorOutput "🔍 检查当前环境..." "Blue"
    
    # 检查是否在go-app目录
    if (-not (Test-Path "cmd/server/main.go")) {
        Write-ColorOutput "❌ 请在go-app目录下运行此脚本" "Red"
        Write-ColorOutput "💡 运行命令: cd go-app && .\start_verification.ps1" "Yellow"
        exit 1
    }
    
    # 检查Go服务状态
    if (Test-GoService) {
        Write-ColorOutput "✅ Go服务已在运行" "Green"
    }
    else {
        Write-ColorOutput "⚠️ Go服务未运行，正在启动..." "Yellow"
        Start-GoService
    }
    
    # 打开验证网页
    Open-VerificationPage
    
    # 显示使用说明
    Show-Instructions
    
    Write-ColorOutput "🎉 验证工具启动完成！" "Green"
    Write-ColorOutput "💡 按Ctrl+C停止Go服务" "Yellow"
    
    # 等待用户中断
    try {
        while ($true) {
            Start-Sleep -Seconds 1
        }
    }
    catch {
        Write-ColorOutput "🛑 正在停止服务..." "Yellow"
        # 这里可以添加停止Go服务的逻辑
    }
}

# 执行主函数
Main 