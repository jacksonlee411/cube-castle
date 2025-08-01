-- Software Company Data Creation Script
-- 软件公司数据创建脚本

-- 设置租户ID
SET app.current_tenant_id = '00000000-0000-0000-0000-000000000000';

-- 定义常量
\set tenant_id '00000000-0000-0000-0000-000000000000'
\set admin_user_id '11111111-1111-1111-1111-111111111111'

-- 1. 创建组织单元 (20个部门)
-- 首先创建顶层公司
INSERT INTO organization_units (id, tenant_id, unit_type, name, description, status, profile, created_at, updated_at, parent_unit_id) VALUES
(gen_random_uuid(), :'tenant_id', 'COMPANY', 'CubeCastle Technology', '领先的企业级软件解决方案提供商', 'ACTIVE', '{"industry": "software", "size": "medium", "founded": "2018"}', NOW(), NOW(), NULL);

-- 获取公司ID用于后续部门创建
WITH company AS (
  SELECT id FROM organization_units WHERE name = 'CubeCastle Technology' AND tenant_id = :'tenant_id'
)
INSERT INTO organization_units (id, tenant_id, unit_type, name, description, status, profile, created_at, updated_at, parent_unit_id) VALUES
-- 技术部门
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '前端开发部', '负责用户界面和用户体验开发', 'ACTIVE', '{"tech_stack": ["React", "Vue", "TypeScript"], "headcount": 8}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '后端开发部', '负责服务端开发和API设计', 'ACTIVE', '{"tech_stack": ["Go", "Python", "Node.js"], "headcount": 10}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '移动开发部', '负责移动应用开发', 'ACTIVE', '{"tech_stack": ["React Native", "Flutter", "iOS", "Android"], "headcount": 6}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '数据工程部', '负责大数据处理和数据平台建设', 'ACTIVE', '{"tech_stack": ["Kafka", "Spark", "Elasticsearch"], "headcount": 5}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', 'DevOps部', '负责基础设施和CI/CD', 'ACTIVE', '{"tech_stack": ["Docker", "Kubernetes", "AWS", "Jenkins"], "headcount": 4}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '测试部', '负责软件质量保证和自动化测试', 'ACTIVE', '{"tech_stack": ["Selenium", "Jest", "Cypress"], "headcount": 6}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '架构部', '负责技术架构设计和技术选型', 'ACTIVE', '{"focus": ["system_design", "performance", "scalability"], "headcount": 3}', NOW(), NOW(), (SELECT id FROM company)),
-- 产品部门
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '产品管理部', '负责产品规划和需求分析', 'ACTIVE', '{"focus": ["product_strategy", "user_research"], "headcount": 4}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', 'UX设计部', '负责用户体验设计', 'ACTIVE', '{"tools": ["Figma", "Sketch", "Adobe XD"], "headcount": 3}', NOW(), NOW(), (SELECT id FROM company)),
-- 业务部门
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '销售部', '负责客户开发和销售', 'ACTIVE', '{"focus": ["enterprise_sales", "channel_partnership"], "headcount": 5}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '市场部', '负责品牌推广和市场营销', 'ACTIVE', '{"focus": ["digital_marketing", "content_marketing"], "headcount": 3}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '客户成功部', '负责客户支持和成功', 'ACTIVE', '{"focus": ["customer_support", "success_management"], "headcount": 4}', NOW(), NOW(), (SELECT id FROM company)),
-- 支持部门
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '人力资源部', '负责人才招聘和员工发展', 'ACTIVE', '{"focus": ["recruitment", "training", "performance"], "headcount": 2}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '财务部', '负责财务管理和成本控制', 'ACTIVE', '{"focus": ["financial_planning", "budgeting"], "headcount": 2}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '法务部', '负责法律事务和合规', 'ACTIVE', '{"focus": ["contract_management", "compliance"], "headcount": 1}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '行政部', '负责日常行政事务', 'ACTIVE', '{"focus": ["office_management", "procurement"], "headcount": 2}', NOW(), NOW(), (SELECT id FROM company)),
-- 特殊团队
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '创新实验室', '负责新技术研究和原型开发', 'ACTIVE', '{"focus": ["AI", "blockchain", "IoT"], "headcount": 3}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '安全部', '负责信息安全和数据保护', 'ACTIVE', '{"focus": ["cybersecurity", "data_protection"], "headcount": 2}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '技术写作部', '负责技术文档和API文档', 'ACTIVE', '{"focus": ["technical_writing", "documentation"], "headcount": 2}', NOW(), NOW(), (SELECT id FROM company)),
(gen_random_uuid(), :'tenant_id', 'DEPARTMENT', '业务分析部', '负责业务需求分析和流程优化', 'ACTIVE', '{"focus": ["business_analysis", "process_optimization"], "headcount": 2}', NOW(), NOW(), (SELECT id FROM company));

-- 2. 为每个部门创建职位 (Position)
-- 首先获取所有部门ID
WITH dept_ids AS (
  SELECT id, name FROM organization_units 
  WHERE tenant_id = :'tenant_id' AND unit_type = 'DEPARTMENT'
),
-- 创建基础职位配置
position_configs AS (
  SELECT 
    gen_random_uuid() as job_profile_id,
    'REGULAR' as position_type,
    d.id as department_id,
    d.name as dept_name,
    pos.title,
    pos.level
  FROM dept_ids d
  CROSS JOIN (
    VALUES 
    ('部门总监', 'DIRECTOR'),
    ('高级经理', 'SENIOR_MANAGER'),
    ('经理', 'MANAGER'),
    ('高级工程师', 'SENIOR'),
    ('工程师', 'REGULAR'),
    ('初级工程师', 'JUNIOR'),
    ('实习生', 'INTERN')
  ) AS pos(title, level)
  WHERE (d.name LIKE '%开发%' OR d.name LIKE '%测试%' OR d.name LIKE '%DevOps%' OR d.name LIKE '%架构%' OR d.name = '数据工程部')
)
INSERT INTO positions (id, tenant_id, position_type, job_profile_id, status, budgeted_fte, details, created_at, updated_at, department_id, manager_position_id)
SELECT 
  gen_random_uuid(),
  :'tenant_id',
  position_type,
  job_profile_id,
  CASE 
    WHEN level IN ('DIRECTOR', 'SENIOR_MANAGER') THEN 'FILLED'
    WHEN level = 'MANAGER' THEN 'FILLED'
    WHEN level IN ('SENIOR', 'REGULAR') THEN 'FILLED'
    ELSE 'OPEN'
  END,
  1.0,
  jsonb_build_object(
    'title', title,
    'level', level,
    'department', dept_name,
    'skills_required', CASE 
      WHEN dept_name LIKE '%前端%' THEN '["React", "TypeScript", "CSS", "JavaScript"]'::jsonb
      WHEN dept_name LIKE '%后端%' THEN '["Go", "Python", "SQL", "Redis"]'::jsonb
      WHEN dept_name LIKE '%移动%' THEN '["React Native", "Flutter", "iOS", "Android"]'::jsonb
      WHEN dept_name LIKE '%数据%' THEN '["Spark", "Kafka", "Python", "SQL"]'::jsonb
      WHEN dept_name LIKE '%DevOps%' THEN '["Docker", "Kubernetes", "AWS", "Jenkins"]'::jsonb
      WHEN dept_name LIKE '%测试%' THEN '["Selenium", "Jest", "Python", "Automation"]'::jsonb
      WHEN dept_name LIKE '%架构%' THEN '["System Design", "Performance", "Scalability"]'::jsonb
      ELSE '["Communication", "Problem Solving", "Teamwork"]'::jsonb
    END
  ),
  NOW(),
  NOW(),
  department_id,
  NULL
FROM position_configs;

-- 3. 创建员工数据 (基于旧的employees表结构)
-- 先清空现有员工数据
DELETE FROM employees WHERE id NOT LIKE 'system_%';

-- 插入50名员工
INSERT INTO employees (id, name, email, position, created_at, updated_at) VALUES
-- 高管层 (5人)
('emp_001', '张伟强', 'zhang.weiqiang@cubecastle.com', 'CTO & 联合创始人', NOW(), NOW()),
('emp_002', '李芳芳', 'li.fangfang@cubecastle.com', 'CPO & 产品副总裁', NOW(), NOW()),
('emp_003', '王建国', 'wang.jianguo@cubecastle.com', 'VP Engineering', NOW(), NOW()),
('emp_004', '刘美丽', 'liu.meili@cubecastle.com', 'VP Sales & Marketing', NOW(), NOW()),
('emp_005', '陈志华', 'chen.zhihua@cubecastle.com', 'CFO & 运营副总裁', NOW(), NOW()),

-- 前端开发部 (8人)
('emp_006', '赵晓明', 'zhao.xiaoming@cubecastle.com', '前端开发总监', NOW(), NOW()),
('emp_007', '孙丽娟', 'sun.lijuan@cubecastle.com', '高级前端工程师', NOW(), NOW()),
('emp_008', '周强', 'zhou.qiang@cubecastle.com', '高级前端工程师', NOW(), NOW()),
('emp_009', '吴敏', 'wu.min@cubecastle.com', '前端工程师', NOW(), NOW()),
('emp_010', '郑海洋', 'zheng.haiyang@cubecastle.com', '前端工程师', NOW(), NOW()),
('emp_011', '冯雪梅', 'feng.xuemei@cubecastle.com', '前端工程师', NOW(), NOW()),
('emp_012', '蒋大伟', 'jiang.dawei@cubecastle.com', '初级前端工程师', NOW(), NOW()),
('emp_013', '韩小红', 'han.xiaohong@cubecastle.com', '前端实习生', NOW(), NOW()),

-- 后端开发部 (10人)  
('emp_014', '许文博', 'xu.wenbo@cubecastle.com', '后端开发总监', NOW(), NOW()),
('emp_015', '何晓峰', 'he.xiaofeng@cubecastle.com', '首席后端架构师', NOW(), NOW()),
('emp_016', '沈佳琪', 'shen.jiaqi@cubecastle.com', '高级后端工程师', NOW(), NOW()),
('emp_017', '卢志强', 'lu.zhiqiang@cubecastle.com', '高级后端工程师', NOW(), NOW()),
('emp_018', '施雨婷', 'shi.yuting@cubecastle.com', '后端工程师', NOW(), NOW()),
('emp_019', '姚伟华', 'yao.weihua@cubecastle.com', '后端工程师', NOW(), NOW()),
('emp_020', '傅小丽', 'fu.xiaoli@cubecastle.com', '后端工程师', NOW(), NOW()),
('emp_021', '邓建军', 'deng.jianjun@cubecastle.com', '后端工程师', NOW(), NOW()),
('emp_022', '曹明明', 'cao.mingming@cubecastle.com', '初级后端工程师', NOW(), NOW()),
('emp_023', '彭小强', 'peng.xiaoqiang@cubecastle.com', '后端实习生', NOW(), NOW()),

-- 移动开发部 (6人)
('emp_024', '范志刚', 'fan.zhigang@cubecastle.com', '移动开发总监', NOW(), NOW()),
('emp_025', '苏美玲', 'su.meiling@cubecastle.com', '高级移动开发工程师', NOW(), NOW()),
('emp_026', '程晓燕', 'cheng.xiaoyan@cubecastle.com', '移动开发工程师', NOW(), NOW()),
('emp_027', '丁伟东', 'ding.weidong@cubecastle.com', '移动开发工程师', NOW(), NOW()),
('emp_028', '白雪莹', 'bai.xueying@cubecastle.com', '移动开发工程师', NOW(), NOW()),
('emp_029', '石磊', 'shi.lei@cubecastle.com', '移动开发实习生', NOW(), NOW()),

-- 数据工程部 (5人)
('emp_030', '毛建华', 'mao.jianhua@cubecastle.com', '数据工程总监', NOW(), NOW()),
('emp_031', '文小芳', 'wen.xiaofang@cubecastle.com', '高级数据工程师', NOW(), NOW()),
('emp_032', '方志敏', 'fang.zhimin@cubecastle.com', '数据工程师', NOW(), NOW()),
('emp_033', '宋雨桐', 'song.yutong@cubecastle.com', '数据工程师', NOW(), NOW()),
('emp_034', '戴小明', 'dai.xiaoming@cubecastle.com', '数据分析师', NOW(), NOW()),

-- DevOps部 (4人)
('emp_035', '侯伟光', 'hou.weiguang@cubecastle.com', 'DevOps总监', NOW(), NOW()),
('emp_036', '薛晓琳', 'xue.xiaolin@cubecastle.com', '高级DevOps工程师', NOW(), NOW()),
('emp_037', '顾志华', 'gu.zhihua@cubecastle.com', 'DevOps工程师', NOW(), NOW()),
('emp_038', '廖小梅', 'liao.xiaomei@cubecastle.com', 'DevOps工程师', NOW(), NOW()),

-- 测试部 (6人)
('emp_039', '谭建平', 'tan.jianping@cubecastle.com', '测试总监', NOW(), NOW()),
('emp_040', '洪美华', 'hong.meihua@cubecastle.com', '高级测试工程师', NOW(), NOW()),
('emp_041', '黎志强', 'li.zhiqiang@cubecastle.com', '测试工程师', NOW(), NOW()),
('emp_042', '康小红', 'kang.xiaohong@cubecastle.com', '测试工程师', NOW(), NOW()),
('emp_043', '贺文静', 'he.wenjing@cubecastle.com', '自动化测试工程师', NOW(), NOW()),
('emp_044', '龙小飞', 'long.xiaofei@cubecastle.com', '测试实习生', NOW(), NOW()),

-- 产品和其他部门 (6人)  
('emp_045', '常晓东', 'chang.xiaodong@cubecastle.com', '产品总监', NOW(), NOW()),
('emp_046', '包雪芳', 'bao.xuefang@cubecastle.com', '高级产品经理', NOW(), NOW()),
('emp_047', '华小明', 'hua.xiaoming@cubecastle.com', 'UX设计师', NOW(), NOW()),
('emp_048', '金晓丽', 'jin.xiaoli@cubecastle.com', '销售经理', NOW(), NOW()),
('emp_049', '夏志华', 'xia.zhihua@cubecastle.com', '人力资源经理', NOW(), NOW()),
('emp_050', '武小强', 'wu.xiaoqiang@cubecastle.com', '财务经理', NOW(), NOW());

-- 4. 创建职位历史记录 (Position History)
-- 为所有员工创建当前职位记录
INSERT INTO position_history (
  id, tenant_id, employee_id, position_title, department, job_level, location, 
  employment_type, reports_to_employee_id, effective_date, end_date, change_reason, 
  is_retroactive, created_by, created_at, min_salary, max_salary, currency
) VALUES
-- 高管层职位历史
(gen_random_uuid(), :'tenant_id', 'emp_001', 'CTO & 联合创始人', '架构部', 'EXECUTIVE', '上海总部', 'FULL_TIME', NULL, '2020-01-01', NULL, '公司创立', false, :'admin_user_id', NOW(), 800000, 1200000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_002', 'CPO & 产品副总裁', '产品管理部', 'EXECUTIVE', '上海总部', 'FULL_TIME', 'emp_001', '2020-03-01', NULL, '产品线负责人', false, :'admin_user_id', NOW(), 600000, 900000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_003', 'VP Engineering', '后端开发部', 'EXECUTIVE', '上海总部', 'FULL_TIME', 'emp_001', '2020-06-01', NULL, '工程团队负责人', false, :'admin_user_id', NOW(), 700000, 1000000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_004', 'VP Sales & Marketing', '销售部', 'EXECUTIVE', '上海总部', 'FULL_TIME', 'emp_001', '2021-01-01', NULL, '商业化负责人', false, :'admin_user_id', NOW(), 500000, 800000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_005', 'CFO & 运营副总裁', '财务部', 'EXECUTIVE', '上海总部', 'FULL_TIME', 'emp_001', '2021-06-01', NULL, '财务运营负责人', false, :'admin_user_id', NOW(), 600000, 900000, 'CNY'),

-- 前端开发部职位历史
(gen_random_uuid(), :'tenant_id', 'emp_006', '前端开发总监', '前端开发部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_003', '2020-08-01', NULL, '前端团队建立', false, :'admin_user_id', NOW(), 400000, 600000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_007', '高级前端工程师', '前端开发部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_006', '2021-03-01', NULL, '团队扩张', false, :'admin_user_id', NOW(), 280000, 420000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_008', '高级前端工程师', '前端开发部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_006', '2021-06-01', NULL, '团队扩张', false, :'admin_user_id', NOW(), 270000, 400000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_009', '前端工程师', '前端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_007', '2022-01-01', NULL, '校招入职', false, :'admin_user_id', NOW(), 180000, 280000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_010', '前端工程师', '前端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_007', '2022-03-01', NULL, '社招入职', false, :'admin_user_id', NOW(), 200000, 300000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_011', '前端工程师', '前端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_008', '2022-08-01', NULL, '团队扩张', false, :'admin_user_id', NOW(), 190000, 290000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_012', '初级前端工程师', '前端开发部', 'JUNIOR', '上海总部', 'FULL_TIME', 'emp_009', '2023-07-01', NULL, '校招入职', false, :'admin_user_id', NOW(), 120000, 180000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_013', '前端实习生', '前端开发部', 'INTERN', '上海总部', 'INTERN', 'emp_010', '2024-09-01', NULL, '实习项目', false, :'admin_user_id', NOW(), 8000, 12000, 'CNY'),

-- 后端开发部职位历史
(gen_random_uuid(), :'tenant_id', 'emp_014', '后端开发总监', '后端开发部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_003', '2020-09-01', NULL, '后端团队负责人', false, :'admin_user_id', NOW(), 450000, 650000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_015', '首席后端架构师', '后端开发部', 'PRINCIPAL', '上海总部', 'FULL_TIME', 'emp_014', '2021-01-01', NULL, '架构设计负责人', false, :'admin_user_id', NOW(), 500000, 700000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_016', '高级后端工程师', '后端开发部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_015', '2021-04-01', NULL, '核心服务开发', false, :'admin_user_id', NOW(), 300000, 450000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_017', '高级后端工程师', '后端开发部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_015', '2021-07-01', NULL, '核心服务开发', false, :'admin_user_id', NOW(), 290000, 430000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_018', '后端工程师', '后端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_016', '2022-02-01', NULL, '业务服务开发', false, :'admin_user_id', NOW(), 200000, 320000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_019', '后端工程师', '后端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_016', '2022-05-01', NULL, '业务服务开发', false, :'admin_user_id', NOW(), 210000, 330000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_020', '后端工程师', '后端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_017', '2022-09-01', NULL, '微服务开发', false, :'admin_user_id', NOW(), 195000, 310000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_021', '后端工程师', '后端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_017', '2023-01-01', NULL, '微服务开发', false, :'admin_user_id', NOW(), 205000, 325000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_022', '初级后端工程师', '后端开发部', 'JUNIOR', '上海总部', 'FULL_TIME', 'emp_018', '2023-08-01', NULL, '校招入职', false, :'admin_user_id', NOW(), 140000, 200000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_023', '后端实习生', '后端开发部', 'INTERN', '上海总部', 'INTERN', 'emp_019', '2024-09-01', NULL, '实习项目', false, :'admin_user_id', NOW(), 10000, 15000, 'CNY'),

-- 移动开发部职位历史
(gen_random_uuid(), :'tenant_id', 'emp_024', '移动开发总监', '移动开发部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_003', '2021-08-01', NULL, '移动端产品线启动', false, :'admin_user_id', NOW(), 380000, 550000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_025', '高级移动开发工程师', '移动开发部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_024', '2022-01-01', NULL, 'React Native专家', false, :'admin_user_id', NOW(), 280000, 420000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_026', '移动开发工程师', '移动开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_025', '2022-06-01', NULL, 'iOS开发', false, :'admin_user_id', NOW(), 200000, 320000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_027', '移动开发工程师', '移动开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_025', '2022-10-01', NULL, 'Android开发', false, :'admin_user_id', NOW(), 195000, 315000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_028', '移动开发工程师', '移动开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_024', '2023-03-01', NULL, 'Flutter开发', false, :'admin_user_id', NOW(), 205000, 325000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_029', '移动开发实习生', '移动开发部', 'INTERN', '上海总部', 'INTERN', 'emp_026', '2024-09-01', NULL, '实习项目', false, :'admin_user_id', NOW(), 9000, 13000, 'CNY'),

-- 数据工程部职位历史  
(gen_random_uuid(), :'tenant_id', 'emp_030', '数据工程总监', '数据工程部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_003', '2022-03-01', NULL, '数据平台建设', false, :'admin_user_id', NOW(), 420000, 600000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_031', '高级数据工程师', '数据工程部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_030', '2022-06-01', NULL, '大数据处理专家', false, :'admin_user_id', NOW(), 320000, 480000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_032', '数据工程师', '数据工程部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_031', '2023-01-01', NULL, '数据管道开发', false, :'admin_user_id', NOW(), 220000, 350000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_033', '数据工程师', '数据工程部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_031', '2023-05-01', NULL, '实时数据处理', false, :'admin_user_id', NOW(), 230000, 360000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_034', '数据分析师', '数据工程部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_030', '2023-09-01', NULL, '业务数据分析', false, :'admin_user_id', NOW(), 180000, 280000, 'CNY'),

-- DevOps部职位历史
(gen_random_uuid(), :'tenant_id', 'emp_035', 'DevOps总监', 'DevOps部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_003', '2021-05-01', NULL, '基础设施负责人', false, :'admin_user_id', NOW(), 400000, 580000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_036', '高级DevOps工程师', 'DevOps部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_035', '2021-10-01', NULL, 'K8s平台建设', false, :'admin_user_id', NOW(), 300000, 450000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_037', 'DevOps工程师', 'DevOps部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_036', '2022-04-01', NULL, 'CI/CD流水线', false, :'admin_user_id', NOW(), 210000, 330000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_038', 'DevOps工程师', 'DevOps部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_036', '2022-11-01', NULL, '监控告警系统', false, :'admin_user_id', NOW(), 220000, 340000, 'CNY'),

-- 测试部职位历史
(gen_random_uuid(), :'tenant_id', 'emp_039', '测试总监', '测试部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_003', '2021-02-01', NULL, '质量保证负责人', false, :'admin_user_id', NOW(), 350000, 520000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_040', '高级测试工程师', '测试部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_039', '2021-07-01', NULL, '测试架构设计', false, :'admin_user_id', NOW(), 250000, 380000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_041', '测试工程师', '测试部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_040', '2022-02-01', NULL, '功能测试', false, :'admin_user_id', NOW(), 160000, 260000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_042', '测试工程师', '测试部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_040', '2022-07-01', NULL, '接口测试', false, :'admin_user_id', NOW(), 170000, 270000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_043', '自动化测试工程师', '测试部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_039', '2023-02-01', NULL, '自动化测试框架', false, :'admin_user_id', NOW(), 200000, 320000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_044', '测试实习生', '测试部', 'INTERN', '上海总部', 'INTERN', 'emp_041', '2024-09-01', NULL, '实习项目', false, :'admin_user_id', NOW(), 7000, 11000, 'CNY'),

-- 产品和其他部门职位历史
(gen_random_uuid(), :'tenant_id', 'emp_045', '产品总监', '产品管理部', 'DIRECTOR', '上海总部', 'FULL_TIME', 'emp_002', '2020-12-01', NULL, '产品规划负责人', false, :'admin_user_id', NOW(), 380000, 550000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_046', '高级产品经理', '产品管理部', 'SENIOR', '上海总部', 'FULL_TIME', 'emp_045', '2021-08-01', NULL, '核心产品负责人', false, :'admin_user_id', NOW(), 250000, 400000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_047', 'UX设计师', 'UX设计部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_045', '2022-01-01', NULL, '用户体验设计', false, :'admin_user_id', NOW(), 180000, 300000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_048', '销售经理', '销售部', 'MANAGER', '上海总部', 'FULL_TIME', 'emp_004', '2021-10-01', NULL, '企业客户开发', false, :'admin_user_id', NOW(), 200000, 350000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_049', '人力资源经理', '人力资源部', 'MANAGER', '上海总部', 'FULL_TIME', 'emp_005', '2021-06-01', NULL, '人才招聘', false, :'admin_user_id', NOW(), 180000, 280000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_050', '财务经理', '财务部', 'MANAGER', '上海总部', 'FULL_TIME', 'emp_005', '2021-12-01', NULL, '财务管理', false, :'admin_user_id', NOW(), 200000, 300000, 'CNY');

-- 5. 添加一些历史职位变更记录 (展示员工晋升)
INSERT INTO position_history (
  id, tenant_id, employee_id, position_title, department, job_level, location, 
  employment_type, reports_to_employee_id, effective_date, end_date, change_reason, 
  is_retroactive, created_by, created_at, min_salary, max_salary, currency
) VALUES
-- 员工晋升历史记录
(gen_random_uuid(), :'tenant_id', 'emp_007', '前端工程师', '前端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_006', '2021-03-01', '2023-01-01', '入职', false, :'admin_user_id', NOW(), 200000, 300000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_016', '后端工程师', '后端开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_014', '2021-04-01', '2023-04-01', '入职', false, :'admin_user_id', NOW(), 220000, 320000, 'CNY'),
(gen_random_uuid(), :'tenant_id', 'emp_025', '移动开发工程师', '移动开发部', 'REGULAR', '上海总部', 'FULL_TIME', 'emp_024', '2022-01-01', '2023-06-01', '入职', false, :'admin_user_id', NOW(), 210000, 310000, 'CNY');

-- 创建数据统计视图
CREATE OR REPLACE VIEW employee_department_summary AS
SELECT 
  d.name as department_name,
  COUNT(DISTINCT ph.employee_id) as employee_count,
  AVG((ph.min_salary + ph.max_salary) / 2) as avg_salary,
  MIN(ph.effective_date) as earliest_hire_date,
  MAX(ph.effective_date) as latest_hire_date
FROM organization_units d
LEFT JOIN position_history ph ON d.name = ph.department 
  AND ph.tenant_id = :'tenant_id' 
  AND ph.end_date IS NULL
WHERE d.tenant_id = :'tenant_id' 
  AND d.unit_type = 'DEPARTMENT'
GROUP BY d.id, d.name
ORDER BY employee_count DESC;

-- 显示创建结果统计
SELECT 'Data Creation Complete!' as status;
SELECT COUNT(*) as total_departments FROM organization_units WHERE tenant_id = :'tenant_id' AND unit_type = 'DEPARTMENT';
SELECT COUNT(*) as total_positions FROM positions WHERE tenant_id = :'tenant_id';  
SELECT COUNT(*) as total_employees FROM employees;
SELECT COUNT(*) as total_position_records FROM position_history WHERE tenant_id = :'tenant_id';
SELECT * FROM employee_department_summary;