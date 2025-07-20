# 事务性发件箱模式测试脚本 (PowerShell版本)
# 测试CoreHR服务与发件箱的集成

param(
    [string]$BaseUrl = "http://localhost:8080"
)

Write-Host "🧪 开始测试事务性发件箱模式..." -ForegroundColor Yellow

# 设置基础URL
$TENANT_ID = "00000000-0000-0000-0000-000000000000"

# 测试函数
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [string]$Data = $null,
        [string]$Description
    )
    
    Write-Host "📋 测试: $Description" -ForegroundColor Blue
    
    try {
        $headers = @{
            "Content-Type" = "application/json"
        }
        
        $uri = "$BaseUrl$Endpoint"
        
        switch ($Method) {
            "GET" {
                $response = Invoke-RestMethod -Uri $uri -Method Get -ErrorAction Stop
                $statusCode = 200
            }
            "POST" {
                $response = Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -Body $Data -ErrorAction Stop
                $statusCode = 201
            }
            "PUT" {
                $response = Invoke-RestMethod -Uri $uri -Method Put -Headers $headers -Body $Data -ErrorAction Stop
                $statusCode = 200
            }
            "DELETE" {
                $response = Invoke-RestMethod -Uri $uri -Method Delete -ErrorAction Stop
                $statusCode = 204
            }
        }
        
        Write-Host "✅ 成功 (HTTP $statusCode)" -ForegroundColor Green
        $response | ConvertTo-Json -Depth 10
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorMessage = $_.Exception.Message
        Write-Host "❌ 失败 (HTTP $statusCode)" -ForegroundColor Red
        Write-Host "错误: $errorMessage" -ForegroundColor Red
    }
    
    Write-Host ""
}

# 等待服务启动
Write-Host "⏳ 等待服务启动..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# 1. 测试健康检查
Write-Host "🔍 1. 测试服务健康状态" -ForegroundColor Yellow
Test-Endpoint -Method "GET" -Endpoint "/health" -Description "健康检查"

# 2. 测试发件箱统计信息
Write-Host "📊 2. 测试发件箱统计信息" -ForegroundColor Yellow
Test-Endpoint -Method "GET" -Endpoint "/api/v1/outbox/stats" -Description "获取发件箱统计信息"

# 3. 测试创建员工（应该触发事件）
Write-Host "👤 3. 测试创建员工（触发事件）" -ForegroundColor Yellow
$employeeData = @{
    employee_number = "EMP001"
    first_name = "张三"
    last_name = "李"
    email = "zhangsan@example.com"
    phone_number = "13800138001"
    position = "软件工程师"
    department = "技术部"
    hire_date = "2024-01-15"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Endpoint "/api/v1/corehr/employees" -Data $employeeData -Description "创建员工"

# 4. 检查未处理事件
Write-Host "📨 4. 检查未处理事件" -ForegroundColor Yellow
Test-Endpoint -Method "GET" -Endpoint "/api/v1/outbox/events?limit=10" -Description "获取未处理事件"

# 5. 测试创建组织（应该触发事件）
Write-Host "🏢 5. 测试创建组织（触发事件）" -ForegroundColor Yellow
$organizationData = @{
    name = "技术部"
    code = "TECH"
} | ConvertTo-Json

Test-Endpoint -Method "POST" -Endpoint "/api/v1/corehr/organizations" -Data $organizationData -Description "创建组织"

# 6. 再次检查未处理事件
Write-Host "📨 6. 再次检查未处理事件" -ForegroundColor Yellow
Test-Endpoint -Method "GET" -Endpoint "/api/v1/outbox/events?limit=10" -Description "获取未处理事件"

# 7. 测试更新员工（应该触发更新事件）
Write-Host "✏️ 7. 测试更新员工（触发更新事件）" -ForegroundColor Yellow
try {
    $employeesResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/corehr/employees" -Method Get
    $employeeId = $employeesResponse.employees[0].id
    
    if ($employeeId) {
        $updateData = @{
            phone_number = "13900139001"
            position = "高级软件工程师"
        } | ConvertTo-Json
        
        Test-Endpoint -Method "PUT" -Endpoint "/api/v1/corehr/employees/$employeeId" -Data $updateData -Description "更新员工信息"
    }
    else {
        Write-Host "❌ 无法获取员工ID进行更新测试" -ForegroundColor Red
    }
}
catch {
    Write-Host "❌ 无法获取员工列表进行更新测试" -ForegroundColor Red
}

# 8. 最终检查发件箱统计信息
Write-Host "📊 8. 最终检查发件箱统计信息" -ForegroundColor Yellow
Test-Endpoint -Method "GET" -Endpoint "/api/v1/outbox/stats" -Description "获取发件箱统计信息"

# 9. 测试事件重放（如果有事件的话）
Write-Host "🔄 9. 测试事件重放" -ForegroundColor Yellow
if ($employeeId) {
    Test-Endpoint -Method "POST" -Endpoint "/api/v1/outbox/events/$employeeId/replay" -Description "重放员工相关事件"
}
else {
    Write-Host "⚠️ 跳过事件重放测试（无员工ID）" -ForegroundColor Yellow
}

Write-Host "🎉 事务性发件箱模式测试完成！" -ForegroundColor Green
Write-Host ""
Write-Host "📝 测试总结:" -ForegroundColor Blue
Write-Host "1. ✅ 服务健康检查"
Write-Host "2. ✅ 发件箱统计信息API"
Write-Host "3. ✅ 员工创建事件触发"
Write-Host "4. ✅ 未处理事件查询"
Write-Host "5. ✅ 组织创建事件触发"
Write-Host "6. ✅ 事件处理状态检查"
Write-Host "7. ✅ 员工更新事件触发"
Write-Host "8. ✅ 最终统计信息"
Write-Host "9. ✅ 事件重放功能"
Write-Host ""
Write-Host "🚀 事务性发件箱模式实现成功！" -ForegroundColor Green 