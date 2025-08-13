#!/usr/bin/env python3

"""
修复后的PostgreSQL vs Neo4j层级计算性能对比测试
包含数据同步修复后的完整验证
"""

import psycopg2
from neo4j import GraphDatabase
import time
import statistics
from typing import List, Dict, Tuple
import json

class FixedHierarchyPerformanceTest:
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
    
    def postgresql_hierarchy_query(self, org_code: str) -> Tuple[List[Dict], float]:
        """PostgreSQL递归CTE层级查询"""
        start_time = time.perf_counter()
        
        cursor = self.pg_conn.cursor()
        query = """
        WITH RECURSIVE org_hierarchy AS (
            -- 基础情况：查找目标组织
            SELECT code, parent_code, name, level, path, 0 as hierarchy_depth
            FROM organization_units 
            WHERE code = %s AND is_current = true
            
            UNION ALL
            
            -- 递归情况：查找父组织
            SELECT p.code, p.parent_code, p.name, p.level, p.path, h.hierarchy_depth + 1
            FROM organization_units p
            INNER JOIN org_hierarchy h ON p.code = h.parent_code
            WHERE p.is_current = true AND h.hierarchy_depth < 10
        )
        SELECT code, name, level, hierarchy_depth, path
        FROM org_hierarchy 
        ORDER BY hierarchy_depth;
        """
        
        cursor.execute(query, (org_code,))
        results = cursor.fetchall()
        
        execution_time = (time.perf_counter() - start_time) * 1000  # 转换为毫秒
        
        formatted_results = []
        for row in results:
            formatted_results.append({
                'code': row[0],
                'name': row[1], 
                'level': row[2],
                'hierarchy_depth': row[3],
                'path': row[4]
            })
        
        return formatted_results, execution_time
    
    def neo4j_hierarchy_query_fixed(self, org_code: str) -> Tuple[List[Dict], float]:
        """修复后的Neo4j图遍历层级查询"""
        start_time = time.perf_counter()
        
        with self.neo4j_driver.session() as session:
            # 使用正确的PARENT_OF关系
            query = """
            MATCH (org:OrganizationUnit {code: $org_code})
            WHERE org.path IS NOT NULL
            OPTIONAL MATCH path = (org)-[:PARENT_OF*0..10]->(ancestors)
            WITH org, path, length(path) as depth,
                 CASE WHEN path IS NULL THEN [org] ELSE nodes(path) END as hierarchy_nodes
            UNWIND range(0, size(hierarchy_nodes)-1) as idx
            WITH hierarchy_nodes[idx] as node, idx as hierarchy_depth
            RETURN 
              node.code as code,
              node.name as name, 
              node.level as level,
              hierarchy_depth,
              node.path as path
            ORDER BY hierarchy_depth
            """
            
            result = session.run(query, {"org_code": org_code})
            records = list(result)
            
        execution_time = (time.perf_counter() - start_time) * 1000  # 转换为毫秒
        
        formatted_results = []
        for record in records:
            formatted_results.append({
                'code': record['code'],
                'name': record['name'],
                'level': record['level'], 
                'hierarchy_depth': record['hierarchy_depth'],
                'path': record['path']
            })
        
        return formatted_results, execution_time
    
    def run_performance_comparison(self, test_org_codes: List[str], iterations: int = 10) -> Dict:
        """运行性能对比测试"""
        results = {
            "test_summary": {
                "iterations": iterations,
                "test_org_codes": test_org_codes,
                "timestamp": time.strftime("%Y-%m-%d %H:%M:%S")
            },
            "postgresql_results": {},
            "neo4j_results": {},
            "performance_comparison": {}
        }
        
        for org_code in test_org_codes:
            print(f"\n🧪 测试组织代码: {org_code}")
            
            # PostgreSQL测试
            print("📊 测试PostgreSQL递归CTE...")
            pg_times = []
            pg_last_result = None
            
            for i in range(iterations):
                try:
                    result, exec_time = self.postgresql_hierarchy_query(org_code)
                    pg_times.append(exec_time)
                    if i == 0:  # 保存第一次结果用于比较
                        pg_last_result = result
                    print(f"  第{i+1}次: {exec_time:.3f}ms")
                except Exception as e:
                    print(f"  ❌ PostgreSQL查询失败: {e}")
                    continue
            
            # Neo4j测试  
            print("📊 测试Neo4j图遍历...")
            neo4j_times = []
            neo4j_last_result = None
            
            for i in range(iterations):
                try:
                    result, exec_time = self.neo4j_hierarchy_query_fixed(org_code)
                    neo4j_times.append(exec_time)
                    if i == 0:  # 保存第一次结果用于比较
                        neo4j_last_result = result
                    print(f"  第{i+1}次: {exec_time:.3f}ms")
                except Exception as e:
                    print(f"  ❌ Neo4j查询失败: {e}")
                    continue
            
            # 计算统计数据
            if pg_times:
                results["postgresql_results"][org_code] = {
                    "times": pg_times,
                    "avg_time": statistics.mean(pg_times),
                    "min_time": min(pg_times),
                    "max_time": max(pg_times),
                    "std_dev": statistics.stdev(pg_times) if len(pg_times) > 1 else 0,
                    "result_count": len(pg_last_result) if pg_last_result else 0,
                    "sample_result": pg_last_result[:3] if pg_last_result else []
                }
            
            if neo4j_times:
                results["neo4j_results"][org_code] = {
                    "times": neo4j_times,
                    "avg_time": statistics.mean(neo4j_times),
                    "min_time": min(neo4j_times),
                    "max_time": max(neo4j_times),
                    "std_dev": statistics.stdev(neo4j_times) if len(neo4j_times) > 1 else 0,
                    "result_count": len(neo4j_last_result) if neo4j_last_result else 0,
                    "sample_result": neo4j_last_result[:3] if neo4j_last_result else []
                }
            
            # 性能比较
            if pg_times and neo4j_times:
                pg_avg = statistics.mean(pg_times)
                neo4j_avg = statistics.mean(neo4j_times)
                speedup = neo4j_avg / pg_avg if pg_avg > 0 else 0
                
                results["performance_comparison"][org_code] = {
                    "postgresql_avg_ms": round(pg_avg, 3),
                    "neo4j_avg_ms": round(neo4j_avg, 3),
                    "postgresql_faster_by": round(speedup, 2),
                    "winner": "PostgreSQL" if pg_avg < neo4j_avg else "Neo4j"
                }
                
                print(f"📈 性能对比结果:")
                print(f"  PostgreSQL平均: {pg_avg:.3f}ms")
                print(f"  Neo4j平均: {neo4j_avg:.3f}ms") 
                print(f"  PostgreSQL比Neo4j快: {speedup:.2f}倍" if speedup > 1 else f"  Neo4j比PostgreSQL快: {1/speedup:.2f}倍")
        
        return results
    
    def verify_data_consistency(self, org_code: str) -> Dict:
        """验证数据一致性"""
        print(f"\n🔍 验证 {org_code} 的数据一致性...")
        
        # PostgreSQL数据
        pg_result, _ = self.postgresql_hierarchy_query(org_code)
        
        # Neo4j数据
        neo4j_result, _ = self.neo4j_hierarchy_query_fixed(org_code)
        
        consistency_check = {
            "org_code": org_code,
            "postgresql_count": len(pg_result),
            "neo4j_count": len(neo4j_result),
            "data_consistent": len(pg_result) == len(neo4j_result),
            "postgresql_sample": pg_result[:2] if pg_result else [],
            "neo4j_sample": neo4j_result[:2] if neo4j_result else []
        }
        
        print(f"  PostgreSQL结果数量: {consistency_check['postgresql_count']}")
        print(f"  Neo4j结果数量: {consistency_check['neo4j_count']}")
        print(f"  数据一致性: {'✅ 一致' if consistency_check['data_consistent'] else '❌ 不一致'}")
        
        return consistency_check
    
    def close(self):
        """关闭连接"""
        self.pg_conn.close()
        self.neo4j_driver.close()

def main():
    """主测试函数"""
    print("🚀 开始修复后的PostgreSQL vs Neo4j层级计算性能对比测试")
    
    tester = FixedHierarchyPerformanceTest()
    
    try:
        # 测试组织代码
        test_org_codes = ["1000056", "1000001", "1000002"]
        
        # 数据一致性验证
        print("\n📋 步骤1: 数据一致性验证")
        consistency_results = []
        for org_code in test_org_codes:
            result = tester.verify_data_consistency(org_code)
            consistency_results.append(result)
        
        # 性能对比测试
        print("\n📋 步骤2: 性能对比测试")
        performance_results = tester.run_performance_comparison(test_org_codes, iterations=5)
        
        # 保存结果
        final_report = {
            "consistency_verification": consistency_results,
            "performance_comparison": performance_results
        }
        
        with open("/home/shangmeilin/cube-castle/fixed-hierarchy-performance-report.json", "w", encoding="utf-8") as f:
            json.dump(final_report, f, indent=2, ensure_ascii=False)
        
        print("\n✅ 测试完成！详细报告已保存到: fixed-hierarchy-performance-report.json")
        
        # 输出总结
        print("\n📊 总结:")
        for org_code in test_org_codes:
            if org_code in performance_results["performance_comparison"]:
                comp = performance_results["performance_comparison"][org_code]
                print(f"  {org_code}: {comp['winner']} 获胜 (PostgreSQL: {comp['postgresql_avg_ms']}ms, Neo4j: {comp['neo4j_avg_ms']}ms)")
    
    except Exception as e:
        print(f"❌ 测试失败: {e}")
        return 1
    finally:
        tester.close()
    
    return 0

if __name__ == "__main__":
    exit(main())