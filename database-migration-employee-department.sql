-- 员工-部门直接关联优化迁移脚本
-- 解决报告中提到的数据模型设计缺陷

-- 1. 添加员工直接关联部门字段
ALTER TABLE employees ADD COLUMN department_id uuid REFERENCES organization_units(id);

-- 2. 创建索引提高查询性能
CREATE INDEX idx_employees_department_id ON employees(department_id);
CREATE INDEX idx_employees_position_department ON employees(position_id, department_id);

-- 3. 数据迁移：填充现有员工的部门信息
UPDATE employees 
SET department_id = p.department_id 
FROM positions p 
WHERE employees.position_id = p.id 
AND employees.department_id IS NULL;

-- 4. 添加数据一致性约束 (使用触发器实现，因为PostgreSQL不支持子查询约束)
-- 将在后续通过应用层或触发器实现一致性检查

-- 5. 创建视图简化查询
CREATE OR REPLACE VIEW employee_details AS
SELECT 
  e.id,
  e.business_id,
  e.employee_number,
  CONCAT(e.first_name, ' ', e.last_name) as person_name,
  e.first_name,
  e.last_name,
  e.email,
  e.phone_number,
  e.hire_date,
  e.status,
  e.position_id,
  e.department_id,
  p.title as position_title,
  ou.name as department_name,
  ou.parent_id as parent_department_id
FROM employees e
LEFT JOIN positions p ON e.position_id = p.id
LEFT JOIN organization_units ou ON e.department_id = ou.id;

-- 6. 验证数据完整性
SELECT 
  COUNT(*) as total_employees,
  COUNT(department_id) as employees_with_department,
  COUNT(position_id) as employees_with_position,
  COUNT(CASE WHEN department_id IS NOT NULL AND position_id IS NOT NULL THEN 1 END) as employees_with_both
FROM employees;