# 047 迁移后验证报告

## 1. 数据完整性

### 命令
```
psql -c "SELECT COUNT(*) AS null_effective_date FROM position_assignments WHERE effective_date IS NULL;"
psql -c "SELECT COUNT(*) AS invalid_end_date FROM position_assignments WHERE end_date IS NOT NULL AND end_date <= effective_date;"
```

### 输出
 null_effective_date 
---------------------
                   0
(1 row)

 invalid_end_date 
------------------
                0
(1 row)


## 2. REST API 冒烟

### 命令
```
curl -s -H "Authorization: Bearer <TENANT_A_TOKEN>" \
     -H "X-Tenant-ID: TENANT_A_ID" \
     http://localhost:9090/api/v1/positions/P9000003/assignments
```

### 输出
```
{
  "success": true,
  "totalCount": 12,
  "sample": [
    {
      "assignmentId": "af7ce1ae-1985-4ea6-ad97-d4acf4f58eff",
      "positionCode": "P9000003",
      "positionRecordId": "131383e5-d16f-4a57-9adb-59151978bd0d",
      "employeeId": "81786f7b-a01a-4aa1-8d49-23ea828568ac",
      "employeeName": "跨租户验收 20251021T075917",
      "assignmentType": "ACTING",
      "assignmentStatus": "ENDED",
      "fte": 0.1,
      "effectiveDate": "2025-10-21T00:00:00Z",
      "endDate": "2025-10-22T00:00:00Z",
      "actingUntil": "2025-10-28T00:00:00Z",
      "autoRevert": false,
      "isCurrent": false,
      "createdAt": "2025-10-20T23:59:17.218368Z",
      "updatedAt": "2025-10-21T00:05:18.868202Z"
    }
  ]
}
```

## 3. GraphQL 冒烟

### 命令
```
curl -s -X POST http://localhost:8090/graphql \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <TENANT_A_TOKEN>" \
     -H "X-Tenant-ID: TENANT_A_ID" \
     -d '{"query":"query($code: PositionCode!) { positionAssignments(positionCode: $code) { totalCount } }","variables":{"code":"P9000003"}}'
```

### 输出
```
{"success":true,"data":{"positionAssignments":{"totalCount":12}},"message":"Query executed successfully","timestamp":"2025-10-21T02:42:45Z","requestId":"unknown"}```

## 4. 结论
- 无 NULL `effective_date`，无 `end_date <= effective_date` 异常。
- REST 与 GraphQL 接口返回成功，迁移后 API 功能正常。
