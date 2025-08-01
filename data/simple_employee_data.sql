-- 简化的50个员工数据直接插入脚本
-- 删除现有的测试数据
DELETE FROM corehr.employees WHERE tenant_id = '00000000-0000-0000-0000-000000000000'::uuid;

-- 插入50个软件公司员工
INSERT INTO corehr.employees (
    id, tenant_id, employee_number, first_name, last_name, email, 
    position, department, hire_date, status, created_at, updated_at
) VALUES
-- 高管层 (5人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP001', '张', '伟强', 'zhang.weiqiang@techcorp.com', 'CTO', '技术部', '2020-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP002', '李', '芳芳', 'li.fangfang@techcorp.com', 'CPO', '产品部', '2020-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP003', '王', '建国', 'wang.jianguo@techcorp.com', 'VP Engineering', '技术部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP004', '刘', '美丽', 'liu.meili@techcorp.com', 'VP Sales', '销售部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP005', '陈', '志华', 'chen.zhihua@techcorp.com', 'CFO', '财务部', '2020-01-01', 'active', NOW(), NOW()),

-- 技术团队 - 前端开发部 (8人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP006', '赵', '晓明', 'zhao.xiaoming@techcorp.com', '前端开发总监', '前端开发部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP007', '吴', '小丽', 'wu.xiaoli@techcorp.com', '高级前端工程师', '前端开发部', '2021-03-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP008', '周', '大伟', 'zhou.dawei@techcorp.com', '高级前端工程师', '前端开发部', '2021-08-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP009', '郑', '晓红', 'zheng.xiaohong@techcorp.com', '前端工程师', '前端开发部', '2022-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP010', '孙', '志强', 'sun.zhiqiang@techcorp.com', '前端工程师', '前端开发部', '2022-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP011', '朱', '小芳', 'zhu.xiaofang@techcorp.com', '前端工程师', '前端开发部', '2023-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP012', '胡', '建华', 'hu.jianhua@techcorp.com', '初级前端工程师', '前端开发部', '2023-09-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP013', '高', '小雨', 'gao.xiaoyu@techcorp.com', '前端实习生', '前端开发部', '2024-09-01', 'active', NOW(), NOW()),

-- 技术团队 - 后端开发部 (10人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP014', '许', '文博', 'xu.wenbo@techcorp.com', '后端开发总监', '后端开发部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP015', '何', '志远', 'he.zhiyuan@techcorp.com', '架构师', '后端开发部', '2021-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP016', '韩', '小强', 'han.xiaoqiang@techcorp.com', '高级后端工程师', '后端开发部', '2021-03-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP017', '冯', '大明', 'feng.daming@techcorp.com', '高级后端工程师', '后端开发部', '2021-08-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP018', '邓', '晓丽', 'deng.xiaoli@techcorp.com', '后端工程师', '后端开发部', '2022-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP019', '曹', '志华', 'cao.zhihua@techcorp.com', '后端工程师', '后端开发部', '2022-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP020', '彭', '小芳', 'peng.xiaofang@techcorp.com', '后端工程师', '后端开发部', '2023-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP021', '吕', '建国', 'lv.jianguo@techcorp.com', '后端工程师', '后端开发部', '2023-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP022', '苏', '小雨', 'su.xiaoyu@techcorp.com', '初级后端工程师', '后端开发部', '2023-09-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP023', '丁', '志强', 'ding.zhiqiang@techcorp.com', '后端实习生', '后端开发部', '2024-09-01', 'active', NOW(), NOW()),

-- 技术团队 - 移动开发部 (6人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP024', '任', '小明', 'ren.xiaoming@techcorp.com', '移动开发总监', '移动开发部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP025', '姜', '大伟', 'jiang.dawei@techcorp.com', '高级iOS工程师', '移动开发部', '2021-08-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP026', '谢', '晓红', 'xie.xiaohong@techcorp.com', '高级Android工程师', '移动开发部', '2022-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP027', '沈', '志华', 'shen.zhihua@techcorp.com', 'React Native工程师', '移动开发部', '2022-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP028', '韦', '小芳', 'wei.xiaofang@techcorp.com', 'Flutter工程师', '移动开发部', '2023-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP029', '段', '建华', 'duan.jianhua@techcorp.com', '移动开发实习生', '移动开发部', '2024-09-01', 'active', NOW(), NOW()),

-- 技术团队 - 数据工程部 (5人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP030', '毛', '建华', 'mao.jianhua@techcorp.com', '数据工程总监', '数据工程部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP031', '薛', '小雨', 'xue.xiaoyu@techcorp.com', '数据架构师', '数据工程部', '2021-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP032', '白', '志强', 'bai.zhiqiang@techcorp.com', '大数据工程师', '数据工程部', '2021-08-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP033', '崔', '小明', 'cui.xiaoming@techcorp.com', '数据分析师', '数据工程部', '2022-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP034', '田', '大伟', 'tian.dawei@techcorp.com', '机器学习工程师', '数据工程部', '2022-06-01', 'active', NOW(), NOW()),

-- 技术团队 - DevOps部 (4人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP035', '侯', '伟光', 'hou.weiguang@techcorp.com', 'DevOps总监', 'DevOps部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP036', '邹', '晓红', 'zou.xiaohong@techcorp.com', '高级DevOps工程师', 'DevOps部', '2021-03-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP037', '石', '志华', 'shi.zhihua@techcorp.com', 'DevOps工程师', 'DevOps部', '2021-08-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP038', '龙', '小芳', 'long.xiaofang@techcorp.com', '云平台工程师', 'DevOps部', '2022-01-01', 'active', NOW(), NOW()),

-- 技术团队 - 测试部 (6人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP039', '谭', '建平', 'tan.jianping@techcorp.com', '测试总监', '测试部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP040', '黎', '小雨', 'li.xiaoyu@techcorp.com', '高级测试工程师', '测试部', '2021-08-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP041', '严', '志强', 'yan.zhiqiang@techcorp.com', '自动化测试工程师', '测试部', '2022-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP042', '文', '小明', 'wen.xiaoming@techcorp.com', '性能测试工程师', '测试部', '2022-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP043', '尹', '大伟', 'yin.dawei@techcorp.com', '测试工程师', '测试部', '2023-01-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP044', '卢', '晓红', 'lu.xiaohong@techcorp.com', '测试实习生', '测试部', '2024-09-01', 'active', NOW(), NOW()),

-- 产品团队 - 产品部 (3人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP045', '常', '晓东', 'chang.xiaodong@techcorp.com', '产品总监', '产品部', '2020-06-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP046', '马', '志华', 'ma.zhihua@techcorp.com', '高级产品经理', '产品部', '2021-03-01', 'active', NOW(), NOW()),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP047', '方', '小芳', 'fang.xiaofang@techcorp.com', '产品经理', '产品部', '2021-08-01', 'active', NOW(), NOW()),

-- 支持团队 - 人力资源部 (1人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP048', '夏', '志华', 'xia.zhihua@techcorp.com', '人力资源经理', '人力资源部', '2021-08-01', 'active', NOW(), NOW()),

-- 支持团队 - UX设计部 (1人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP049', '华', '小明', 'hua.xiaoming@techcorp.com', 'UX设计师', 'UX设计部', '2022-06-01', 'active', NOW(), NOW()),

-- 支持团队 - 销售部 (1人)
(gen_random_uuid(), '00000000-0000-0000-0000-000000000000'::uuid, 'EMP050', '金', '晓丽', 'jin.xiaoli@techcorp.com', '销售经理', '销售部', '2020-06-01', 'active', NOW(), NOW());

-- 验证插入结果
SELECT COUNT(*) as total_employees FROM corehr.employees WHERE tenant_id = '00000000-0000-0000-0000-000000000000'::uuid;
SELECT DISTINCT department, COUNT(*) as count FROM corehr.employees WHERE tenant_id = '00000000-0000-0000-0000-000000000000'::uuid GROUP BY department ORDER BY count DESC;