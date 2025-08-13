#!/usr/bin/env python3

"""
精确时态过滤逻辑实施脚本
修复数据完整性问题并优化查询性能
"""

import psycopg2
from neo4j import GraphDatabase
import json
from datetime import datetime

class PreciseTemporalFilter:
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
    
    def fix_orphan_organizations(self):
        """修复孤儿组织问题"""
        print("🔧 修复孤儿组织问题...")
        
        with self.neo4j_driver.session() as session:
            # 查找孤儿组织
            orphan_query = """
            MATCH (org:Organization)
            WHERE org.parent_code IS NOT NULL
            OPTIONAL MATCH (parent:Organization {code: org.parent_code, tenant_id: org.tenant_id})
            WITH org, parent
            WHERE parent IS NULL
            RETURN org.code as orphan_code, org.parent_code as missing_parent
            """
            
            result = session.run(orphan_query)
            orphans = list(result)
            
            if orphans:
                print(f"  🔍 发现 {len(orphans)} 个孤儿组织:")
                for orphan in orphans:
                    print(f"    - {orphan['orphan_code']} -> 缺失父组织: {orphan['missing_parent']}")
                
                # 从PostgreSQL补充缺失的父组织
                for orphan in orphans:
                    missing_parent_code = orphan['missing_parent']
                    self.add_missing_parent_from_postgres(missing_parent_code)
            else:
                print("  ✅ 没有发现孤儿组织")
    
    def add_missing_parent_from_postgres(self, parent_code):
        """从PostgreSQL添加缺失的父组织"""
        cursor = self.pg_conn.cursor()
        
        # 查找缺失的父组织
        query = """
        SELECT code, parent_code, name, unit_type, status,
               level, path, sort_order, description,
               effective_date, end_date, is_current,
               tenant_id, created_at, updated_at
        FROM organization_units 
        WHERE code = %s AND is_current = true
        """
        
        cursor.execute(query, (parent_code,))
        parent_data = cursor.fetchone()
        
        if parent_data:
            with self.neo4j_driver.session() as session:
                # 添加缺失的父组织
                create_parent_query = """
                CREATE (org:Organization {
                    tenant_id: $tenant_id,
                    code: $code,
                    parent_code: $parent_code,
                    name: $name,
                    unit_type: $unit_type,
                    status: $status,
                    level: $level,
                    path: $path,
                    sort_order: $sort_order,
                    description: $description,
                    effective_date: date($effective_date),
                    end_date: CASE WHEN $end_date IS NOT NULL THEN date($end_date) ELSE null END,
                    is_current: $is_current,
                    created_at: datetime($created_at),
                    updated_at: datetime($updated_at),
                    synced_at: datetime()
                })
                RETURN org.code as created_code
                """
                
                result = session.run(create_parent_query, {
                    "tenant_id": str(parent_data[11]),
                    "code": parent_data[0],
                    "parent_code": parent_data[1],
                    "name": parent_data[2],
                    "unit_type": parent_data[3],
                    "status": parent_data[4],
                    "level": parent_data[5],
                    "path": parent_data[6],
                    "sort_order": parent_data[7],
                    "description": parent_data[8],
                    "effective_date": parent_data[9].strftime("%Y-%m-%d"),
                    "end_date": parent_data[10].strftime("%Y-%m-%d") if parent_data[10] else None,
                    "is_current": parent_data[11],
                    "created_at": parent_data[12].isoformat(),
                    "updated_at": parent_data[13].isoformat()
                })
                
                created = list(result)
                if created:
                    print(f"    ✅ 添加缺失父组织: {parent_code}")
                    
                    # 重新创建层级关系
                    self.rebuild_hierarchy_relationships()
    
    def rebuild_hierarchy_relationships(self):
        """重建层级关系"""
        print("🔗 重建层级关系...")
        
        with self.neo4j_driver.session() as session:
            # 删除现有关系
            session.run("MATCH ()-[r:HAS_CHILD]->() DELETE r")
            session.run("MATCH ()-[r:PARENT_OF]->() DELETE r")
            session.run("MATCH ()-[r:ANCESTOR_OF]->() DELETE r")
            
            # 重新创建直接父子关系
            parent_child_query = """
            MATCH (child:Organization), (parent:Organization)
            WHERE child.parent_code = parent.code
              AND child.tenant_id = parent.tenant_id
              AND child.parent_code IS NOT NULL
            MERGE (parent)-[:HAS_CHILD {
                created_at: datetime(),
                relationship_level: 1
            }]->(child)
            MERGE (child)-[:PARENT_OF {
                created_at: datetime(),
                relationship_level: 1
            }]->(parent)
            RETURN count(*) as direct_relationships
            """
            
            result = session.run(parent_child_query)
            direct_count = list(result)[0]["direct_relationships"]
            print(f"  ✅ 重建了 {direct_count} 个直接层级关系")
    
    def create_precise_temporal_queries(self):
        """创建精确的时态查询"""
        print("⚡ 创建精确的时态查询函数...")
        
        with self.neo4j_driver.session() as session:
            # 1. 精确的层级路径查询（不使用APOC）
            hierarchy_test_query = """
            // 测试组织 1000002 的层级路径
            MATCH (org:Organization {code: '1000002'})
            WHERE org.is_current = true
            OPTIONAL MATCH path = (org)-[:PARENT_OF*]->(ancestors:Organization)
            WHERE ancestors.is_current = true
            WITH org, path,
                 CASE WHEN path IS NULL THEN [org] 
                      ELSE nodes(path) END as hierarchy_nodes
            UNWIND hierarchy_nodes as node
            RETURN DISTINCT
              node.code as code,
              node.name as name,
              node.level as level,
              node.path as path,
              node.unit_type as unit_type,
              node.status as status,
              length(path) as hierarchy_depth
            ORDER BY node.level
            """
            
            result = session.run(hierarchy_test_query)
            test_results = list(result)
            
            print(f"  📊 测试查询返回 {len(test_results)} 条记录")
            for record in test_results[:3]:  # 显示前3条
                print(f"    - {record['code']}: {record['name']} (级别: {record['level']})")
    
    def create_optimized_hierarchy_function(self):
        """创建优化的层级查询函数（不依赖APOC）"""
        print("⚡ 创建优化的层级查询.....")
        
        with self.neo4j_driver.session() as session:
            # 测试优化查询性能
            optimized_query = """
            // 优化的层级查询：使用路径字符串解析
            WITH '1000002' as target_code
            MATCH (org:Organization {code: target_code})
            WHERE org.is_current = true
            WITH org, 
                 split(substring(org.path, 1), '/') as path_segments
            UNWIND range(0, size(path_segments)-1) as idx
            WITH org, path_segments[idx] as ancestor_code, idx
            WHERE ancestor_code <> ''
            MATCH (ancestor:Organization {code: ancestor_code})
            WHERE ancestor.is_current = true
            RETURN 
              ancestor.code as code,
              ancestor.name as name,
              ancestor.level as level,
              ancestor.path as path,
              idx as hierarchy_depth
            ORDER BY ancestor.level
            """
            
            result = session.run(optimized_query)
            optimized_results = list(result)
            
            print(f"  ⚡ 优化查询返回 {len(optimized_results)} 条记录")
            return len(optimized_results)
    
    def benchmark_query_performance(self):
        """基准测试查询性能"""
        print("🏃 基准测试优化后的查询性能...")
        
        import time
        
        with self.neo4j_driver.session() as session:
            test_codes = ['1000000', '1000001', '1000002']
            
            for code in test_codes:
                # 测试优化查询
                start_time = time.perf_counter()
                
                query = f"""
                MATCH (org:Organization {{code: '{code}'}})
                WHERE org.is_current = true
                WITH org, 
                     split(substring(org.path, 1), '/') as path_segments
                UNWIND range(0, size(path_segments)-1) as idx
                WITH org, path_segments[idx] as ancestor_code, idx
                WHERE ancestor_code <> ''
                MATCH (ancestor:Organization {{code: ancestor_code}})
                WHERE ancestor.is_current = true
                RETURN count(ancestor) as hierarchy_count
                """
                
                result = session.run(query)
                count = list(result)[0]["hierarchy_count"]
                
                end_time = time.perf_counter()
                execution_time = (end_time - start_time) * 1000
                
                print(f"  📊 {code}: {count} 层级节点, 耗时: {execution_time:.3f}ms")
    
    def validate_data_consistency(self):
        """验证数据一致性"""
        print("🔍 验证修复后的数据一致性...")
        
        with self.neo4j_driver.session() as session:
            # 重新检查孤儿组织
            orphan_check = """
            MATCH (org:Organization)
            WHERE org.parent_code IS NOT NULL
            OPTIONAL MATCH (parent:Organization {code: org.parent_code, tenant_id: org.tenant_id})
            RETURN 
              count(org) as orgs_with_parent,
              count(parent) as valid_parents,
              count(org) - count(parent) as orphan_count
            """
            
            result = session.run(orphan_check)
            stats = list(result)[0]
            
            # 统计信息
            summary_query = """
            MATCH (org:Organization)
            RETURN 
              count(org) as total_organizations,
              count(CASE WHEN org.parent_code IS NULL THEN 1 END) as root_organizations,
              max(org.level) as max_level,
              avg(org.level) as avg_level
            """
            
            result = session.run(summary_query)
            summary = list(result)[0]
            
            validation_report = {
                "organizations_with_parent": stats["orgs_with_parent"],
                "valid_parent_references": stats["valid_parents"],
                "orphan_organizations": stats["orphan_count"],
                "data_integrity": stats["orphan_count"] == 0,
                "total_organizations": summary["total_organizations"],
                "root_organizations": summary["root_organizations"],
                "max_hierarchy_level": summary["max_level"],
                "avg_hierarchy_level": round(summary["avg_level"], 2),
                "validated_at": datetime.now().isoformat()
            }
            
            print(f"  📊 总组织数: {validation_report['total_organizations']}")
            print(f"  📊 根组织数: {validation_report['root_organizations']}")
            print(f"  📊 最大层级: {validation_report['max_hierarchy_level']}")
            print(f"  📊 数据完整性: {'✅ 完整' if validation_report['data_integrity'] else '❌ 有问题'}")
            
            return validation_report
    
    def run_temporal_optimization(self):
        """运行完整的时态优化"""
        print("🚀 开始精确时态过滤逻辑实施...")
        
        try:
            # 1. 修复孤儿组织
            self.fix_orphan_organizations()
            
            # 2. 创建精确查询
            self.create_precise_temporal_queries()
            
            # 3. 创建优化函数
            result_count = self.create_optimized_hierarchy_function()
            
            # 4. 性能测试
            self.benchmark_query_performance()
            
            # 5. 验证一致性
            validation_report = self.validate_data_consistency()
            
            # 6. 生成报告
            optimization_report = {
                "temporal_optimization_completed": True,
                "optimized_query_result_count": result_count,
                "validation": validation_report,
                "completed_at": datetime.now().isoformat()
            }
            
            # 保存报告
            with open("/home/shangmeilin/cube-castle/temporal-optimization-report.json", "w", encoding="utf-8") as f:
                json.dump(optimization_report, f, indent=2, ensure_ascii=False)
            
            print("\n✅ 精确时态过滤逻辑实施完成！")
            print(f"📋 详细报告已保存到: temporal-optimization-report.json")
            
            return optimization_report
            
        except Exception as e:
            print(f"❌ 时态优化失败: {e}")
            import traceback
            traceback.print_exc()
            return None
    
    def close(self):
        """关闭连接"""
        self.pg_conn.close()
        self.neo4j_driver.close()

if __name__ == "__main__":
    optimizer = PreciseTemporalFilter()
    try:
        result = optimizer.run_temporal_optimization()
        if result:
            print("✅ 时态过滤优化成功完成")
        else:
            print("❌ 时态过滤优化失败")
    finally:
        optimizer.close()