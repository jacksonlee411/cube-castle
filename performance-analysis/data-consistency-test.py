#!/usr/bin/env python3
"""
结果一致性验证测试 - 确保PostgreSQL和Neo4j返回相同的层级结构
"""

import psycopg2
from neo4j import GraphDatabase
import json

def verify_data_consistency():
    """验证PostgreSQL和Neo4j的数据一致性"""
    
    # PostgreSQL连接
    pg_conn = psycopg2.connect(
        host="localhost", port=5432, database="cubecastle",
        user="user", password="password"
    )
    
    # Neo4j连接
    neo4j_driver = GraphDatabase.driver(
        "bolt://localhost:7687", auth=("neo4j", "password")
    )
    
    test_org_codes = ["1000056", "1000000", "1000002"]
    
    print("🔍 数据一致性验证测试")
    print("=" * 60)
    
    for org_code in test_org_codes:
        print(f"\n📋 测试组织: {org_code}")
        print("-" * 40)
        
        # PostgreSQL查询
        with pg_conn.cursor() as cursor:
            cursor.execute("""
                WITH RECURSIVE org_hierarchy AS (
                  SELECT code, name, parent_code, level, 1 as depth, code::text as path
                  FROM organization_units 
                  WHERE code = %s AND is_current = true
                  
                  UNION ALL
                  
                  SELECT p.code, p.name, p.parent_code, p.level, oh.depth + 1,
                         p.code || ' -> ' || oh.path
                  FROM organization_units p
                  INNER JOIN org_hierarchy oh ON p.code = oh.parent_code
                  WHERE p.is_current = true
                )
                SELECT code, name, level, depth, path
                FROM org_hierarchy ORDER BY depth DESC;
            """, (org_code,))
            
            pg_results = cursor.fetchall()
        
        # Neo4j查询
        with neo4j_driver.session() as session:
            query = """
            MATCH (org:Organization {code: $org_code})
            OPTIONAL MATCH path = (org)-[:PARENT*0..10]->(ancestor:Organization)
            WITH org, ancestor, length(path) as depth,
                 CASE WHEN ancestor IS NOT NULL THEN ancestor.code ELSE org.code END as ancestor_code,
                 CASE WHEN ancestor IS NOT NULL THEN ancestor.name ELSE org.name END as ancestor_name,
                 CASE WHEN ancestor IS NOT NULL THEN ancestor.level ELSE org.level END as ancestor_level
            RETURN DISTINCT ancestor_code, ancestor_name, ancestor_level, depth
            ORDER BY depth
            """
            
            result = session.run(query, org_code=org_code)
            neo4j_results = [(r['ancestor_code'], r['ancestor_name'], r['ancestor_level'], r['depth']) for r in result]
        
        # 结果对比
        print(f"📊 PostgreSQL结果数量: {len(pg_results)}")
        print(f"📊 Neo4j结果数量: {len(neo4j_results)}")
        
        if pg_results:
            print("📝 PostgreSQL层级路径:")
            for row in pg_results:
                print(f"   {row[0]} ({row[1]}) - 层级{row[2]} - 深度{row[3]}")
                
        if neo4j_results:
            print("📝 Neo4j层级路径:")
            for row in neo4j_results:
                print(f"   {row[0]} ({row[1]}) - 层级{row[2]} - 深度{row[3]}")
        
        # 一致性检查
        pg_codes = set(row[0] for row in pg_results)
        neo4j_codes = set(row[0] for row in neo4j_results)
        
        if pg_codes == neo4j_codes:
            print("✅ 数据一致性: 通过")
        else:
            print("❌ 数据一致性: 失败")
            print(f"   PostgreSQL独有: {pg_codes - neo4j_codes}")
            print(f"   Neo4j独有: {neo4j_codes - pg_codes}")
    
    pg_conn.close()
    neo4j_driver.close()

if __name__ == "__main__":
    verify_data_consistency()