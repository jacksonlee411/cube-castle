// Neo4j 业务ID系统迁移脚本
// 文件: neo4j_business_id_migration.cypher  
// 日期: 2025-08-04
// 描述: 为Neo4j图数据库中的员工和组织节点添加业务ID支持

// ==========================================
// 第一部分: 数据备份和准备工作
// ==========================================

// 1. 创建备份 - 导出现有数据
// CALL apoc.export.json.all("business_id_migration_backup.json", {})

// 2. 显示现有数据统计
MATCH (e:Employee) 
WITH count(e) as employee_count
MATCH (o:Organization)
WITH employee_count, count(o) as org_count
RETURN employee_count, org_count, 
       'Migration will process ' + employee_count + ' employees and ' + org_count + ' organizations' as message;

// ==========================================
// 第二部分: 员工节点业务ID迁移  
// ==========================================

// 2.1 为员工节点添加业务ID (使用节点内部ID + 1作为起始值)
MATCH (e:Employee)
WHERE e.business_id IS NULL
WITH e, id(e) as internal_id
SET e.business_id = toString(internal_id + 1)
RETURN count(e) as employees_updated;

// 2.2 验证员工业务ID格式和唯一性
MATCH (e:Employee)
WHERE e.business_id IS NOT NULL
WITH e.business_id as bid, count(*) as cnt
WHERE cnt > 1
RETURN bid, cnt as duplicate_count
ORDER BY cnt DESC;

// 2.3 检查业务ID格式是否符合规范 (1-99999999)
MATCH (e:Employee)
WHERE e.business_id IS NOT NULL
  AND NOT e.business_id =~ '^[1-9][0-9]{0,7}$'
RETURN e.business_id as invalid_business_id, 
       'Invalid format for employee business_id' as issue
LIMIT 10;

// 2.4 为超出范围的员工重新分配业务ID
MATCH (e:Employee)
WHERE e.business_id IS NOT NULL
  AND (toInteger(e.business_id) = 0 OR toInteger(e.business_id) > 99999999)
WITH e, row_number() OVER () as row_num
SET e.business_id = toString(row_num)
RETURN count(e) as employees_reassigned;

// ==========================================
// 第三部分: 组织节点业务ID迁移
// ==========================================

// 3.1 为组织节点添加业务ID (100000 + 内部ID)
MATCH (o:Organization)
WHERE o.business_id IS NULL
WITH o, id(o) as internal_id
SET o.business_id = toString(100000 + internal_id)
RETURN count(o) as organizations_updated;

// 3.2 验证组织业务ID格式和唯一性
MATCH (o:Organization)
WHERE o.business_id IS NOT NULL
WITH o.business_id as bid, count(*) as cnt
WHERE cnt > 1
RETURN bid, cnt as duplicate_count
ORDER BY cnt DESC;

// 3.3 检查组织业务ID格式是否符合规范 (100000-999999)
MATCH (o:Organization)
WHERE o.business_id IS NOT NULL
  AND NOT o.business_id =~ '^[1-9][0-9]{5}$'
RETURN o.business_id as invalid_business_id,
       'Invalid format for organization business_id' as issue
LIMIT 10;

// 3.4 为超出范围的组织重新分配业务ID
MATCH (o:Organization)
WHERE o.business_id IS NOT NULL
  AND (toInteger(o.business_id) < 100000 OR toInteger(o.business_id) > 999999)
WITH o, row_number() OVER () as row_num
SET o.business_id = toString(100000 + row_num - 1)
RETURN count(o) as organizations_reassigned;

// ==========================================
// 第四部分: 索引创建
// ==========================================

// 4.1 为员工业务ID创建索引
CREATE INDEX employee_business_id_index IF NOT EXISTS
FOR (e:Employee) ON (e.business_id);

// 4.2 为组织业务ID创建索引  
CREATE INDEX organization_business_id_index IF NOT EXISTS
FOR (o:Organization) ON (o.business_id);

// 4.3 为员工UUID创建索引 (保持向后兼容)
CREATE INDEX employee_uuid_index IF NOT EXISTS
FOR (e:Employee) ON (e.id);

// 4.4 为组织UUID创建索引 (保持向后兼容)
CREATE INDEX organization_uuid_index IF NOT EXISTS
FOR (o:Organization) ON (o.id);

// 4.5 创建复合索引 (tenant_id + business_id)
CREATE INDEX employee_tenant_business_id_index IF NOT EXISTS  
FOR (e:Employee) ON (e.tenant_id, e.business_id);

CREATE INDEX organization_tenant_business_id_index IF NOT EXISTS
FOR (o:Organization) ON (o.tenant_id, o.business_id);

// ==========================================
// 第五部分: 关系更新 (如果需要)
// ==========================================

// 5.1 检查员工-组织关系是否需要更新
MATCH (e:Employee)-[r:WORKS_IN]->(o:Organization)
WHERE e.business_id IS NOT NULL AND o.business_id IS NOT NULL
RETURN count(r) as employee_org_relationships,
       'Employee-Organization relationships found' as message;

// 5.2 检查员工-经理关系是否需要更新
MATCH (e:Employee)-[r:REPORTS_TO]->(m:Employee)
WHERE e.business_id IS NOT NULL AND m.business_id IS NOT NULL
RETURN count(r) as manager_relationships,
       'Employee-Manager relationships found' as message;

// 5.3 检查组织层级关系是否需要更新
MATCH (child:Organization)-[r:BELONGS_TO]->(parent:Organization)
WHERE child.business_id IS NOT NULL AND parent.business_id IS NOT NULL
RETURN count(r) as org_hierarchy_relationships,
       'Organization hierarchy relationships found' as message;

// ==========================================
// 第六部分: 约束创建
// ==========================================

// 6.1 创建员工业务ID唯一性约束
CREATE CONSTRAINT employee_business_id_unique IF NOT EXISTS
FOR (e:Employee) REQUIRE e.business_id IS UNIQUE;

// 6.2 创建组织业务ID唯一性约束
CREATE CONSTRAINT organization_business_id_unique IF NOT EXISTS  
FOR (o:Organization) REQUIRE o.business_id IS UNIQUE;

// 6.3 验证约束是否生效
SHOW CONSTRAINTS;

// ==========================================
// 第七部分: 数据完整性验证
// ==========================================

// 7.1 员工数据完整性检查
MATCH (e:Employee)
RETURN 
    count(e) as total_employees,
    count(e.business_id) as employees_with_business_id,
    count(DISTINCT e.business_id) as unique_business_ids,
    count(e) - count(DISTINCT e.business_id) as potential_duplicates,
    'Employee business_id validation' as entity_type;

// 7.2 组织数据完整性检查  
MATCH (o:Organization)
RETURN 
    count(o) as total_organizations,
    count(o.business_id) as organizations_with_business_id, 
    count(DISTINCT o.business_id) as unique_business_ids,
    count(o) - count(DISTINCT o.business_id) as potential_duplicates,
    'Organization business_id validation' as entity_type;

// 7.3 业务ID格式验证 (员工)
MATCH (e:Employee)
WHERE e.business_id IS NOT NULL
  AND NOT e.business_id =~ '^[1-9][0-9]{0,7}$'
RETURN 
    count(e) as invalid_employee_business_ids,
    collect(e.business_id)[0..5] as sample_invalid_ids,
    'Employee business_id format validation' as validation_type;

// 7.4 业务ID格式验证 (组织)
MATCH (o:Organization)  
WHERE o.business_id IS NOT NULL
  AND NOT o.business_id =~ '^[1-9][0-9]{5}$'
RETURN 
    count(o) as invalid_organization_business_ids,
    collect(o.business_id)[0..5] as sample_invalid_ids,
    'Organization business_id format validation' as validation_type;

// 7.5 业务ID范围验证
MATCH (e:Employee)
WHERE e.business_id IS NOT NULL
WITH toInteger(e.business_id) as bid
RETURN 
    min(bid) as min_employee_id,
    max(bid) as max_employee_id,
    count(*) as total_employees,
    'Employee business_id range' as check_type

UNION ALL

MATCH (o:Organization)
WHERE o.business_id IS NOT NULL  
WITH toInteger(o.business_id) as bid
RETURN 
    min(bid) as min_organization_id,
    max(bid) as max_organization_id,
    count(*) as total_organizations,
    'Organization business_id range' as check_type;

// ==========================================
// 第八部分: 查询性能测试
// ==========================================

// 8.1 测试业务ID查询性能 (员工)
PROFILE MATCH (e:Employee {business_id: '1'})
RETURN e.first_name, e.last_name, e.email;

// 8.2 测试UUID查询性能 (员工) - 对比
PROFILE MATCH (e:Employee)
WHERE e.id = 'e60891dc-7d20-444b-9002-22419238d499'  // 示例UUID
RETURN e.first_name, e.last_name, e.email;

// 8.3 测试业务ID查询性能 (组织)
PROFILE MATCH (o:Organization {business_id: '100000'})
RETURN o.name, o.unit_type;

// 8.4 测试关联查询性能
PROFILE MATCH (e:Employee {business_id: '1'})-[:WORKS_IN]->(o:Organization)
RETURN e.first_name, e.last_name, o.name;

// ==========================================
// 第九部分: 数据同步验证 (与PostgreSQL对比)
// ==========================================

// 9.1 检查员工总数是否与PostgreSQL一致
// 注意: 这个查询结果需要与PostgreSQL的结果进行人工对比
MATCH (e:Employee)
RETURN count(e) as neo4j_employee_count,
       'Compare with PostgreSQL: SELECT count(*) FROM corehr.employees' as instruction;

// 9.2 检查组织总数是否与PostgreSQL一致
MATCH (o:Organization)  
RETURN count(o) as neo4j_organization_count,
       'Compare with PostgreSQL: SELECT count(*) FROM corehr.organizations' as instruction;

// 9.3 检查业务ID是否与PostgreSQL同步
// 这需要应用程序层面的同步机制来确保一致性

// ==========================================
// 第十部分: 清理和优化
// ==========================================

// 10.1 删除无效或重复的节点 (如果有)
// 注意: 这是一个危险操作，请谨慎执行
/*
MATCH (e:Employee)
WHERE e.business_id IS NULL OR e.business_id = ''
DELETE e;

MATCH (o:Organization)
WHERE o.business_id IS NULL OR o.business_id = ''  
DELETE o;
*/

// 10.2 优化内存使用
CALL gds.graph.drop('*', false);

// ==========================================
// 迁移完成报告
// ==========================================

// 最终验证报告
MATCH (e:Employee)
WITH count(e) as emp_total, count(e.business_id) as emp_with_bid
MATCH (o:Organization)
WITH emp_total, emp_with_bid, count(o) as org_total, count(o.business_id) as org_with_bid
RETURN 
    emp_total as total_employees,
    emp_with_bid as employees_with_business_id,
    round(100.0 * emp_with_bid / emp_total, 2) as employee_completion_percentage,
    org_total as total_organizations,
    org_with_bid as organizations_with_business_id,
    round(100.0 * org_with_bid / org_total, 2) as organization_completion_percentage,
    CASE 
        WHEN emp_with_bid = emp_total AND org_with_bid = org_total 
        THEN '✅ Migration completed successfully!'
        ELSE '⚠️ Migration incomplete - please check the data'
    END as migration_status;

// 显示索引状态
SHOW INDEXES;

// 显示约束状态  
SHOW CONSTRAINTS;

// ==========================================
// 使用示例和测试查询
// ==========================================

// 示例1: 通过业务ID查找员工
MATCH (e:Employee {business_id: '1'})
RETURN e.business_id, e.first_name, e.last_name, e.email;

// 示例2: 通过业务ID查找组织
MATCH (o:Organization {business_id: '100000'})  
RETURN o.business_id, o.name, o.unit_type;

// 示例3: 查找组织下的所有员工 (使用业务ID)
MATCH (o:Organization {business_id: '100000'})<-[:WORKS_IN]-(e:Employee)
RETURN o.name as organization_name, 
       collect({
           business_id: e.business_id,
           name: e.first_name + ' ' + e.last_name,
           email: e.email
       }) as employees;

// 示例4: 查找员工的直接下属 (使用业务ID)
MATCH (manager:Employee {business_id: '1'})<-[:REPORTS_TO]-(subordinate:Employee)
RETURN manager.business_id as manager_id,
       manager.first_name + ' ' + manager.last_name as manager_name,
       collect({
           business_id: subordinate.business_id,
           name: subordinate.first_name + ' ' + subordinate.last_name
       }) as subordinates;

// 示例5: 组织层级查询 (使用业务ID)
MATCH path = (root:Organization)-[:BELONGS_TO*0..3]->(parent:Organization)
WHERE NOT (parent)-[:BELONGS_TO]->()
WITH root, parent, length(path) as level
ORDER BY level, root.business_id
RETURN 
    parent.business_id as root_org_id,
    parent.name as root_org_name,
    level,
    root.business_id as child_org_id,
    root.name as child_org_name;

// ==========================================
// 迁移完成通知
// ==========================================

RETURN '==============================================',
       'Neo4j 业务ID系统迁移已完成!',
       '==============================================',
       '完成的工作:',
       '1. 为所有员工和组织节点生成了业务ID',
       '2. 创建了业务ID相关的索引和约束',
       '3. 验证了数据完整性和格式正确性',
       '4. 优化了查询性能',
       '',
       '下一步操作:',
       '1. 更新应用程序代码以使用业务ID进行Neo4j查询',
       '2. 建立PostgreSQL与Neo4j的业务ID同步机制',
       '3. 进行全面的功能和性能测试',
       '4. 监控查询性能和数据一致性',
       '==============================================';