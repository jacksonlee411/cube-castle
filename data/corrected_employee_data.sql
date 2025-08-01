-- Corrected Employee Data Creation Script
-- 修正的员工数据创建脚本

-- 设置租户ID和管理员ID
\set tenant_id '00000000-0000-0000-0000-000000000000'
\set admin_user_id '11111111-1111-1111-1111-111111111111'

-- 1. 确认员工数据已经更新为UUID格式
SELECT COUNT(*) as employees_with_uuid_ids FROM employees WHERE id::text ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- 2. 创建职位历史记录，直接使用员工的UUID ID
INSERT INTO position_history (
  id, tenant_id, employee_id, position_title, department, job_level, location, 
  employment_type, reports_to_employee_id, effective_date, end_date, change_reason, 
  is_retroactive, created_by, created_at, min_salary, max_salary, currency
) 
SELECT 
  gen_random_uuid(),
  :'tenant_id'::uuid,
  e.id::uuid,  -- 直接使用员工的UUID ID
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
WHERE e.id::text ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

-- 3. 更新汇报关系
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

-- 4. 创建一些历史职位变更记录 (晋升记录)
INSERT INTO position_history (
  id, tenant_id, employee_id, position_title, department, job_level, location, 
  employment_type, reports_to_employee_id, effective_date, end_date, change_reason, 
  is_retroactive, created_by, created_at, min_salary, max_salary, currency
) 
SELECT 
  gen_random_uuid(),
  :'tenant_id'::uuid,
  ph.employee_id,
  CASE 
    WHEN ph.job_level = 'SENIOR' THEN REPLACE(ph.position_title, '高级', '')
    ELSE ph.position_title
  END,
  ph.department,
  CASE 
    WHEN ph.job_level = 'SENIOR' THEN 'REGULAR'
    ELSE ph.job_level
  END,
  ph.location,
  ph.employment_type,
  ph.reports_to_employee_id,
  ph.effective_date - INTERVAL '2 years',
  ph.effective_date,
  '晋升',
  false,
  :'admin_user_id'::uuid,
  NOW(),
  ph.min_salary * 0.8,
  ph.max_salary * 0.8,
  ph.currency
FROM position_history ph
WHERE ph.tenant_id = :'tenant_id'::uuid
  AND ph.end_date IS NULL 
  AND ph.job_level = 'SENIOR'
  AND random() < 0.6; -- 60% 的高级员工有晋升记录

-- 5. 创建改进的统计视图
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

-- 6. 创建员工层级视图
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
  JOIN employees e ON ph.employee_id = e.id::uuid
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
  JOIN employees e ON ph.employee_id = e.id::uuid
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

-- 显示创建结果
SELECT '=== 数据创建完成 ===' as status;
SELECT COUNT(*) as total_employees FROM employees;
SELECT COUNT(*) as total_position_records FROM position_history WHERE tenant_id = :'tenant_id'::uuid;
SELECT COUNT(*) as current_positions FROM position_history WHERE tenant_id = :'tenant_id'::uuid AND end_date IS NULL;

SELECT '=== 部门员工统计 ===' as status;
SELECT * FROM employee_department_summary;

SELECT '=== 员工层级结构（前20名）===' as status;
SELECT indented_name, position_title, department FROM employee_hierarchy LIMIT 20;