# 测试 CoreHR API 路由
Write-Host "🧪 测试 CoreHR API 路由" -ForegroundColor Cyan
Write-Host "======================" -ForegroundColor Cyan

$BASE_URL = "http://localhost:8080"

# 测试函数
function Test-Endpoint {
    param(
        [string]$Endpoint,
        [string]$Method = "GET",
        [string]$Data = ""
    )
    
    Write-Host "测试 $Method $Endpoint ... " -NoNewline
    
    try {
        if ($Method -eq "POST" -and $Data -ne "") {
            $response = Invoke-RestMethod -Uri "$BASE_URL$Endpoint" -Method $Method -ContentType "application/json" -Body $Data -ErrorAction Stop
            $statusCode = 200
        } else {
            $response = Invoke-RestMethod -Uri "$BASE_URL$Endpoint" -Method $Method -ErrorAction Stop
            $statusCode = 200
        }
        
        Write-Host "✅ 成功 ($statusCode)" -ForegroundColor Green
        if ($response) {
            $responseJson = $response | ConvertTo-Json -Depth 3
            Write-Host "   响应: $($responseJson.Substring(0, [Math]::Min(100, $responseJson.Length)))..." -ForegroundColor Gray
        }
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "❌ 失败 ($statusCode)" -ForegroundColor Red
        Write-Host "   错误: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "📍 服务器地址: $BASE_URL" -ForegroundColor Yellow
Write-Host ""

# 测试健康检查
Test-Endpoint "/health"

# 测试调试路由
Test-Endpoint "/debug/routes"

# 测试静态文件
Test-Endpoint "/test.html"

# 测试 CoreHR API
Test-Endpoint "/api/v1/corehr/employees"

# 测试组织 API
Test-Endpoint "/api/v1/corehr/organizations"

# 测试组织树 API
Test-Endpoint "/api/v1/corehr/organizations/tree"

# 测试创建员工（POST 请求）
$employeeData = @{
    employee_number = "EMP003"
    first_name = "王"
    last_name = "五"
    email = "wangwu@example.com"
    hire_date = "2023-03-15"
} | ConvertTo-Json

Test-Endpoint "/api/v1/corehr/employees" "POST" $employeeData

Write-Host ""
Write-Host "🎉 路由测试完成！" -ForegroundColor Green
Write-Host ""
Write-Host "📋 如果所有测试都通过，您可以访问：" -ForegroundColor Cyan
Write-Host "   🌐 测试页面: $BASE_URL/test.html" -ForegroundColor White
Write-Host "   📊 调试路由: $BASE_URL/debug/routes" -ForegroundColor White
Write-Host "   🏥 健康检查: $BASE_URL/health" -ForegroundColor White
Write-Host "   👥 员工列表: $BASE_URL/api/v1/corehr/employees" -ForegroundColor White 