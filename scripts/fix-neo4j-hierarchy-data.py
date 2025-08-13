#!/usr/bin/env python3

"""
Neo4j数据同步修复脚本
修复PostgreSQL到Neo4j的层级关系同步问题
"""

import psycopg2
from neo4j import GraphDatabase
import json
from datetime import datetime
import sys

class HierarchyDataSync:
    def __init__(self):
        # PostgreSQL连接
        self.pg_conn = psycopg2.connect(
            host="localhost",
            port=5432,
            database="cubecastle",
            user="user",
            password="password"
        )
        
        # Neo4j连接
        self.neo4j_driver = GraphDatabase.driver(
            "bolt://localhost:7687",
            auth=("neo4j", "password")
        )
    
    def get_pg_hierarchy_data(self):
        """从PostgreSQL获取完整的层级数据"""
        cursor = self.pg_conn.cursor()
        
        query = """
        SELECT DISTINCT 
            code, parent_code, name, unit_type, status, 
            level, path, sort_order, description,
            effective_date, end_date, is_current,
            tenant_id, created_at, updated_at
        FROM organization_units 
        WHERE parent_code IS NOT NULL 
        ORDER BY level, code, effective_date DESC
        """
        
        cursor.execute(query)
        results = cursor.fetchall()
        
        columns = [desc[0] for desc in cursor.description]
        return [dict(zip(columns, row)) for row in results]
    
    def update_neo4j_hierarchy(self, org_data):
        """更新Neo4j中的层级数据"""
        with self.neo4j_driver.session() as session:
            # 更新组织节点的parent_code
            update_query = """
            MATCH (org:OrganizationUnit {
                code: $code, 
                effective_date: date($effective_date)
            })
            SET org.parent_code = $parent_code,
                org.level = $level,
                org.path = $path,
                org.hierarchy_updated = datetime()
            RETURN org.code as updated_code
            """
            
            result = session.run(update_query, {
                "code": org_data["code"],
                "effective_date": org_data["effective_date"].strftime("%Y-%m-%d"),
                "parent_code": org_data["parent_code"],
                "level": org_data["level"], 
                "path": org_data["path"]
            })
            
            return len(list(result))
    
    def create_parent_relationships(self):
        """创建正确的父子关系"""
        with self.neo4j_driver.session() as session:
            # 删除现有的错误关系
            session.run("MATCH ()-[r:HAS_CHILD]->() DELETE r")
            session.run("MATCH ()-[r:PARENT_OF]->() DELETE r")
            
            # 基于parent_code创建正确的关系
            create_relations_query = """
            MATCH (child:OrganizationUnit)
            WHERE child.parent_code IS NOT NULL
            MATCH (parent:OrganizationUnit {code: child.parent_code})
            WHERE parent.effective_date <= child.effective_date 
              AND (parent.end_date IS NULL OR parent.end_date >= child.effective_date)
            MERGE (parent)-[:HAS_CHILD]->(child)
            MERGE (child)-[:PARENT_OF]->(parent)
            RETURN count(*) as relationships_created
            """
            
            result = session.run(create_relations_query)
            return list(result)[0]["relationships_created"]
    
    def verify_data_consistency(self):
        """验证数据一致性"""
        print("📊 验证数据一致性...")
        
        # 检查PostgreSQL数据
        cursor = self.pg_conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM organization_units WHERE parent_code IS NOT NULL")
        pg_with_parent = cursor.fetchone()[0]
        
        # 检查Neo4j数据  
        with self.neo4j_driver.session() as session:
            result = session.run("""
                MATCH (org:OrganizationUnit) 
                WHERE org.parent_code IS NOT NULL 
                RETURN count(org) as count
            """)
            neo4j_with_parent = list(result)[0]["count"]
            
            result = session.run("MATCH ()-[r:HAS_CHILD]->() RETURN count(r) as count")
            relations_count = list(result)[0]["count"]
        
        print(f"PostgreSQL组织(有父组织): {pg_with_parent}")
        print(f"Neo4j组织(有父组织): {neo4j_with_parent}")
        print(f"Neo4j关系数量: {relations_count}")
        
        return {
            "pg_with_parent": pg_with_parent,
            "neo4j_with_parent": neo4j_with_parent, 
            "relations_count": relations_count
        }
    
    def run_sync(self):
        """执行完整的同步修复"""
        print("🚀 开始Neo4j层级数据同步修复...")
        
        # 1. 获取PostgreSQL数据
        print("📥 从PostgreSQL获取层级数据...")
        hierarchy_data = self.get_pg_hierarchy_data()
        print(f"获取到 {len(hierarchy_data)} 条层级数据")
        
        # 2. 更新Neo4j节点数据
        print("🔄 更新Neo4j节点数据...")
        updated_count = 0
        for org in hierarchy_data:
            try:
                result = self.update_neo4j_hierarchy(org)
                updated_count += result
            except Exception as e:
                print(f"更新失败 {org['code']}: {e}")
        
        print(f"更新了 {updated_count} 个组织节点")
        
        # 3. 重建关系
        print("🔗 重建父子关系...")
        relations_created = self.create_parent_relationships()
        print(f"创建了 {relations_created} 个层级关系")
        
        # 4. 验证结果
        consistency_result = self.verify_data_consistency()
        
        print("✅ 同步修复完成！")
        return consistency_result
    
    def close(self):
        """关闭连接"""
        self.pg_conn.close()
        self.neo4j_driver.close()

if __name__ == "__main__":
    sync = HierarchyDataSync()
    try:
        result = sync.run_sync()
        print(f"\n📋 最终结果: {json.dumps(result, indent=2)}")
    except Exception as e:
        print(f"❌ 同步失败: {e}")
        sys.exit(1)
    finally:
        sync.close()