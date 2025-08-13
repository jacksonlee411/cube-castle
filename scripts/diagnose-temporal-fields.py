#!/usr/bin/env python3
"""
Neo4j时态字段诊断脚本
检查时态查询字段映射问题
"""

from neo4j import GraphDatabase
import json

# Neo4j连接配置
NEO4J_CONFIG = {
    'uri': 'bolt://localhost:7687',
    'user': 'neo4j',
    'password': 'password'
}

def get_neo4j_driver():
    """获取Neo4j驱动"""
    return GraphDatabase.driver(NEO4J_CONFIG['uri'], 
                               auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password']))

def diagnose_temporal_fields():
    """诊断时态字段问题"""
    driver = get_neo4j_driver()
    
    try:
        with driver.session() as session:
            # 检查组织1000056的时态字段
            result = session.run("""
                MATCH (o:OrganizationUnit {code: "1000056"})
                RETURN o.code, o.name, o.effective_date, o.end_date, o.is_current,
                       o.valid_from, o.valid_to, o.change_reason
                ORDER BY o.effective_date DESC
                LIMIT 5
            """)
            
            print("🔍 检查组织1000056的时态字段:")
            records = list(result)
            if not records:
                print("❌ 未找到组织1000056的记录")
                return
            
            for record in records:
                print(f"  代码: {record['o.code']}")
                print(f"  名称: {record['o.name']}")
                print(f"  生效日期: {record['o.effective_date']} (类型: {type(record['o.effective_date'])})")
                print(f"  结束日期: {record['o.end_date']} (类型: {type(record['o.end_date'])})")
                print(f"  当前有效: {record['o.is_current']}")
                print(f"  有效期开始: {record['o.valid_from']}")
                print(f"  有效期结束: {record['o.valid_to']}")
                print(f"  变更原因: {record['o.change_reason']}")
                print("-" * 50)
            
            # 测试时态查询条件
            print("\n🔍 测试时态查询条件:")
            
            # 测试字符串日期比较
            test_date = "2025-08-13"
            result = session.run("""
                MATCH (org:OrganizationUnit {code: "1000056"})
                WHERE toString(org.effective_date) <= $as_of_date
                  AND (org.end_date IS NULL OR toString(org.end_date) >= $as_of_date)
                RETURN org.code, org.name, org.effective_date, org.end_date, org.is_current
                ORDER BY org.effective_date DESC
                LIMIT 3
            """, {"as_of_date": test_date})
            
            records = list(result)
            print(f"使用字符串比较 (as_of_date={test_date}): 找到 {len(records)} 条记录")
            for record in records:
                print(f"  {record['org.name']} - {record['org.effective_date']} 到 {record['org.end_date']}")
            
            # 测试日期类型比较
            result = session.run("""
                MATCH (org:OrganizationUnit {code: "1000056"})
                WHERE org.effective_date <= date($as_of_date)
                  AND (org.end_date IS NULL OR org.end_date >= date($as_of_date))
                RETURN org.code, org.name, org.effective_date, org.end_date, org.is_current
                ORDER BY org.effective_date DESC
                LIMIT 3
            """, {"as_of_date": test_date})
            
            records = list(result)
            print(f"使用date()函数比较: 找到 {len(records)} 条记录")
            for record in records:
                print(f"  {record['org.name']} - {record['org.effective_date']} 到 {record['org.end_date']}")
    
    finally:
        driver.close()

def fix_temporal_queries():
    """修复时态查询字段映射"""
    driver = get_neo4j_driver()
    
    try:
        with driver.session() as session:
            print("\n🔧 修复时态字段格式...")
            
            # 确保所有effective_date和end_date都是正确的日期格式
            result = session.run("""
                MATCH (o:OrganizationUnit)
                WHERE o.effective_date IS NOT NULL AND toString(o.effective_date) <> ""
                SET o.effective_date = CASE 
                    WHEN o.effective_date CONTAINS "T" THEN date(split(o.effective_date, "T")[0])
                    ELSE date(o.effective_date)
                END
                RETURN count(o) as updated_count
            """)
            
            record = result.single()
            if record:
                print(f"✅ 更新了 {record['updated_count']} 个组织的effective_date格式")
            
            # 修复end_date格式
            result = session.run("""
                MATCH (o:OrganizationUnit)
                WHERE o.end_date IS NOT NULL AND toString(o.end_date) <> ""
                SET o.end_date = CASE 
                    WHEN o.end_date CONTAINS "T" THEN date(split(o.end_date, "T")[0])
                    ELSE date(o.end_date)
                END
                RETURN count(o) as updated_count
            """)
            
            record = result.single()
            if record:
                print(f"✅ 更新了 {record['updated_count']} 个组织的end_date格式")
            
            # 添加时态字段索引
            try:
                session.run("CREATE INDEX temporal_effective_date IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.effective_date)")
                session.run("CREATE INDEX temporal_end_date IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.end_date)")
                session.run("CREATE INDEX temporal_is_current IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.is_current)")
                print("✅ 创建时态字段索引")
            except Exception as e:
                print(f"⚠️ 索引创建可能已存在: {e}")
    
    finally:
        driver.close()

if __name__ == '__main__':
    print("🔍 开始Neo4j时态字段诊断...")
    diagnose_temporal_fields()
    fix_temporal_queries()
    print("\n✅ 时态字段诊断和修复完成")