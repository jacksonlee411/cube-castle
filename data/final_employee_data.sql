-- Final Employee Data Creation Script using UUID field
-- 使用UUID字段的最终员工数据创建脚本

-- 设置租户ID和管理员ID
\set tenant_id '00000000-0000-0000-0000-000000000000'
\set admin_user_id '11111111-1111-1111-1111-111111111111'

-- 1. 清空position_history表中现有的数据，避免冲突
DELETE FROM position_history WHERE tenant_id = :'tenant_id'::uuid;

-- 2. 验证员工表结构
SELECT column_name, data_type FROM information_schema.columns 
WHERE table_name = 'employees' ORDER BY ordinal_position;

-- 3. 为所有员工创建职位历史记录，使用uuid_id字段
INSERT INTO position_history (
  id, tenant_id, employee_id, position_title, department, job_level, location, 
  employment_type, reports_to_employee_id, effective_date, end_date, change_reason, 
  is_retroactive, created_by, created_at, min_salary, max_salary, currency
) 
SELECT 
  gen_random_uuid(),
  :'tenant_id'::uuid,
  e.uuid_id,  -- 使用新的UUID字段
  CASE 
    WHEN e.position LIKE '%CTO%' THEN 'CTO & 联合创始人'
    WHEN e.position LIKE '%CPO%' THEN 'CPO & 产品副总裁'
    WHEN e.position LIKE '%VP Engineering%' THEN 'VP Engineering'
    WHEN e.position LIKE '%VP Sales%' THEN 'VP Sales & Marketing'
    WHEN e.position LIKE '%CFO%' THEN 'CFO & 运营副总裁'
    ELSE SPLIT_PART(e.position, '&', 1)
  END,
  CASE 
    WHEN e.position LIKE '%总监%' OR e.position LIKE '%VP%' OR e.position LIKE '%CTO%' OR e.position LIKE '%CPO%' OR e.position LIKE '%CFO%' THEN 
      CASE 
        WHEN e.position LIKE '%前端%' THEN '前端开发部'
        WHEN e.position LIKE '%后端%' THEN '后端开发部'
        WHEN e.position LIKE '%移动%' THEN '移动开发部'
        WHEN e.position LIKE '%数据%' THEN '数据工程部'
        WHEN e.position LIKE '%DevOps%' THEN 'DevOps部'
        WHEN e.position LIKE '%测试%' THEN '测试部'
        WHEN e.position LIKE '%产品%' OR e.position LIKE '%CPO%' THEN '产品管理部'
        WHEN e.position LIKE '%CTO%' THEN '架构部'
        WHEN e.position LIKE '%销售%' OR e.position LIKE '%VP Sales%' THEN '销售部'
        WHEN e.position LIKE '%CFO%' OR e.position LIKE '%财务%' THEN '财务部'
        ELSE '架构部'
      END
    WHEN e.position LIKE '%前端%' THEN '前端开发部'
    WHEN e.position LIKE '%后端%' THEN '后端开发部'
    WHEN e.position LIKE '%移动%' THEN '移动开发部'
    WHEN e.position LIKE '%数据%' THEN '数据工程部'
    WHEN e.position LIKE '%DevOps%' THEN 'DevOps部'
    WHEN e.position LIKE '%测试%' THEN '测试部'
    WHEN e.position LIKE '%产品%' THEN '产品管理部'
    WHEN e.position LIKE '%UX%' THEN 'UX设计部'
    WHEN e.position LIKE '%销售%' THEN '销售部'
    WHEN e.position LIKE '%人力%' THEN '人力资源部'
    WHEN e.position LIKE '%财务%' THEN '财务部'
    ELSE '产品管理部'
  END,
  CASE 
    WHEN e.position LIKE '%CTO%' OR e.position LIKE '%CPO%' OR e.position LIKE '%CFO%' OR e.position LIKE '%VP%' THEN 'EXECUTIVE'
    WHEN e.position LIKE '%总监%' THEN 'DIRECTOR'
    WHEN e.position LIKE '%首席%' THEN 'PRINCIPAL'
    WHEN e.position LIKE '%高级%' THEN 'SENIOR'
    WHEN e.position LIKE '%经理%' THEN 'MANAGER'
    WHEN e.position LIKE '%初级%' THEN 'JUNIOR'
    WHEN e.position LIKE '%实习%' THEN 'INTERN'
    ELSE 'REGULAR'
  END,
  '上海总部',
  CASE 
    WHEN e.position LIKE '%实习%' THEN 'INTERN'
    ELSE 'FULL_TIME'
  END,
  NULL, -- reports_to_employee_id，稍后设置
  CASE 
    WHEN e.position LIKE '%CTO%' OR e.position LIKE '%CPO%' OR e.position LIKE '%CFO%' THEN '2020-01-01'::timestamp
    WHEN e.position LIKE '%VP%' OR e.position LIKE '%总监%' THEN '2020-06-01'::timestamp
    WHEN e.position LIKE '%高级%' THEN '2021-03-01'::timestamp
    WHEN e.position LIKE '%经理%' THEN '2021-08-01'::timestamp
    WHEN e.position LIKE '%工程师%' AND e.position NOT LIKE '%高级%' AND e.position NOT LIKE '%初级%' THEN '2022-01-01'::timestamp
    WHEN e.position LIKE '%初级%' THEN '2023-07-01'::timestamp
    WHEN e.position LIKE '%实习%' THEN '2024-09-01'::timestamp
    ELSE '2022-06-01'::timestamp
  END,
  NULL, -- end_date
  CASE 
    WHEN e.position LIKE '%CTO%' OR e.position LIKE '%CPO%' OR e.position LIKE '%CFO%' THEN '公司创立'
    WHEN e.position LIKE '%VP%' THEN '高级管理层加入'
    WHEN e.position LIKE '%总监%' THEN '部门负责人任命'
    WHEN e.position LIKE '%实习%' THEN '实习项目'
    ELSE '团队扩张'
  END,
  false,
  :'admin_user_id'::uuid,
  NOW(),
  CASE 
    WHEN e.position LIKE '%CTO%' THEN 800000
    WHEN e.position LIKE '%CPO%' OR e.position LIKE '%CFO%' THEN 600000
    WHEN e.position LIKE '%VP%' THEN 500000
    WHEN e.position LIKE '%总监%' THEN 350000
    WHEN e.position LIKE '%首席%' THEN 500000
    WHEN e.position LIKE '%高级%' THEN 250000
    WHEN e.position LIKE '%经理%' THEN 180000
    WHEN e.position LIKE '%工程师%' AND e.position NOT LIKE '%高级%' AND e.position NOT LIKE '%初级%' THEN 180000
    WHEN e.position LIKE '%初级%' THEN 120000
    WHEN e.position LIKE '%实习%' THEN 8000
    ELSE 160000
  END,
  CASE 
    WHEN e.position LIKE '%CTO%' THEN 1200000
    WHEN e.position LIKE '%CPO%' OR e.position LIKE '%CFO%' THEN 900000
    WHEN e.position LIKE '%VP%' THEN 800000
    WHEN e.position LIKE '%总监%' THEN 600000
    WHEN e.position LIKE '%首席%' THEN 700000
    WHEN e.position LIKE '%高级%' THEN 420000
    WHEN e.position LIKE '%经理%' THEN 350000
    WHEN e.position LIKE '%工程师%' AND e.position NOT LIKE '%高级%' AND e.position NOT LIKE '%初级%' THEN 320000
    WHEN e.position LIKE '%初级%' THEN 200000
    WHEN e.position LIKE '%实习%' THEN 15000
    ELSE 300000
  END,
  'CNY'
FROM employees e
WHERE e.uuid_id IS NOT NULL;

-- 4. 更新汇报关系，使用员工的UUID
WITH reporting_structure AS (
  SELECT 
    ph.employee_id,
    ph.position_title,
    ph.department,
    ph.job_level,
    CASE 
      -- 设置汇报关系
      WHEN ph.job_level = 'EXECUTIVE' AND ph.position_title NOT LIKE '%CTO%' THEN 
        (SELECT employee_id FROM position_history WHERE position_title LIKE '%CTO%' AND tenant_id = :'tenant_id'::uuid LIMIT 1)
      WHEN ph.job_level = 'DIRECTOR' THEN
        CASE 
          WHEN ph.department IN ('前端开发部', '后端开发部', '移动开发部', '数据工程部', 'DevOps部', '测试部') THEN
            (SELECT employee_id FROM position_history WHERE position_title LIKE '%VP Engineering%' AND tenant_id = :'tenant_id'::uuid LIMIT 1)
          WHEN ph.department = '产品管理部' THEN
            (SELECT employee_id FROM position_history WHERE position_title LIKE '%CPO%' AND tenant_id = :'tenant_id'::uuid LIMIT 1)
          WHEN ph.department = '销售部' THEN
            (SELECT employee_id FROM position_history WHERE position_title LIKE '%VP Sales%' AND tenant_id = :'tenant_id'::uuid LIMIT 1)
          WHEN ph.department IN ('财务部', '人力资源部') THEN
            (SELECT employee_id FROM position_history WHERE position_title LIKE '%CFO%' AND tenant_id = :'tenant_id'::uuid LIMIT 1)
          ELSE (SELECT employee_id FROM position_history WHERE position_title LIKE '%CTO%' AND tenant_id = :'tenant_id'::uuid LIMIT 1)
        END
      WHEN ph.job_level IN ('PRINCIPAL', 'SENIOR', 'MANAGER') THEN
        (SELECT employee_id FROM position_history WHERE job_level = 'DIRECTOR' AND department = ph.department AND tenant_id = :'tenant_id'::uuid LIMIT 1)
      WHEN ph.job_level IN ('REGULAR', 'JUNIOR', 'INTERN') THEN
        COALESCE(
          (SELECT employee_id FROM position_history WHERE job_level IN ('SENIOR', 'MANAGER') AND department = ph.department AND tenant_id = :'tenant_id'::uuid LIMIT 1),
          (SELECT employee_id FROM position_history WHERE job_level = 'DIRECTOR' AND department = ph.department AND tenant_id = :'tenant_id'::uuid LIMIT 1)
        )
      ELSE NULL
    END as manager_id
  FROM position_history ph
  WHERE ph.tenant_id = :'tenant_id'::uuid AND ph.end_date IS NULL
)
UPDATE position_history 
SET reports_to_employee_id = rs.manager_id
FROM reporting_structure rs
WHERE position_history.employee_id = rs.employee_id 
  AND position_history.tenant_id = :'tenant_id'::uuid
  AND position_history.end_date IS NULL
  AND rs.manager_id IS NOT NULL;

-- 5. 重新创建统计视图
DROP VIEW IF EXISTS employee_department_summary;
CREATE OR REPLACE VIEW employee_department_summary AS
SELECT 
  ph.department as department_name,
  COUNT(DISTINCT ph.employee_id) as employee_count,
  ROUND(AVG((ph.min_salary + ph.max_salary) / 2.0)) as avg_salary,
  MIN(ph.effective_date)::date as earliest_hire_date,
  MAX(ph.effective_date)::date as latest_hire_date,
  COUNT(CASE WHEN ph.job_level = 'DIRECTOR' THEN 1 END) as directors,
  COUNT(CASE WHEN ph.job_level = 'SENIOR' THEN 1 END) as senior_staff,
  COUNT(CASE WHEN ph.job_level = 'REGULAR' THEN 1 END) as regular_staff,
  COUNT(CASE WHEN ph.job_level = 'JUNIOR' THEN 1 END) as junior_staff,
  COUNT(CASE WHEN ph.job_level = 'INTERN' THEN 1 END) as interns
FROM position_history ph
WHERE ph.tenant_id = :'tenant_id'::uuid
  AND ph.end_date IS NULL
GROUP BY ph.department
ORDER BY employee_count DESC, avg_salary DESC;

-- 6. 创建员工层级视图，使用UUID连接
DROP VIEW IF EXISTS employee_hierarchy;
CREATE OR REPLACE VIEW employee_hierarchy AS
WITH RECURSIVE hierarchy AS (
  -- 起始点：高管层
  SELECT 
    ph.employee_id,
    e.name,
    ph.position_title,
    ph.department,
    ph.job_level,
    ph.reports_to_employee_id,
    0 as level,
    ARRAY[e.name] as path
  FROM position_history ph
  JOIN employees e ON ph.employee_id = e.uuid_id
  WHERE ph.tenant_id = :'tenant_id'::uuid
    AND ph.end_date IS NULL
    AND ph.reports_to_employee_id IS NULL
  
  UNION ALL
  
  -- 递归：下级员工
  SELECT 
    ph.employee_id,
    e.name,
    ph.position_title,
    ph.department,
    ph.job_level,
    ph.reports_to_employee_id,
    h.level + 1,
    h.path || e.name
  FROM position_history ph
  JOIN employees e ON ph.employee_id = e.uuid_id
  JOIN hierarchy h ON ph.reports_to_employee_id = h.employee_id
  WHERE ph.tenant_id = :'tenant_id'::uuid
    AND ph.end_date IS NULL
    AND h.level < 5 -- 防止无限递归
)
SELECT 
  employee_id,
  name,
  position_title,
  department,
  job_level,
  reports_to_employee_id,
  level,
  REPEAT('  ', level) || name as indented_name
FROM hierarchy
ORDER BY level, department, name;

-- 7. 测试API的员工数据查询
CREATE OR REPLACE VIEW api_employees_view AS
SELECT 
  e.uuid_id as id,
  e.id as legacy_id,
  e.name,
  e.email,
  e.position as legacy_position,
  ph.position_title,
  ph.department,
  ph.job_level,
  ph.min_salary,
  ph.max_salary,
  ph.effective_date as start_date,
  e.created_at,
  e.updated_at
FROM employees e
LEFT JOIN position_history ph ON e.uuid_id = ph.employee_id 
  AND ph.tenant_id = :'tenant_id'::uuid 
  AND ph.end_date IS NULL;

-- 显示创建结果
SELECT '=== 数据创建完成 ===' as status;
SELECT COUNT(*) as total_employees FROM employees;
SELECT COUNT(*) as total_position_records FROM position_history WHERE tenant_id = :'tenant_id'::uuid;
SELECT COUNT(*) as current_positions FROM position_history WHERE tenant_id = :'tenant_id'::uuid AND end_date IS NULL;

SELECT '=== 部门员工统计 ===' as status;
SELECT * FROM employee_department_summary;

SELECT '=== 员工层级结构（前15名）===' as status;
SELECT indented_name, position_title, department FROM employee_hierarchy LIMIT 15;

SELECT '=== API视图样本数据 ===' as status;
SELECT name, position_title, department, job_level FROM api_employees_view LIMIT 10;