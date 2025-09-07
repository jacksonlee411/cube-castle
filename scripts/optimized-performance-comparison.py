#!/usr/bin/env python3

"""
[已废弃 - 2025-09-07]
PostgreSQL vs Neo4j 性能对比历史脚本。现行为 PostgreSQL 单一数据源，不再比较 Neo4j。
"""

import psycopg2
from neo4j import GraphDatabase
import time
import statistics
import json
from datetime import datetime
from typing import List, Dict, Tuple

class OptimizedPerformanceComparison:
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
    
    def postgresql_optimized_query(self, org_code: str) -> Tuple[List[Dict], float]:
        """PostgreSQL优化递归CTE查询"""
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
    
    def neo4j_optimized_query(self, org_code: str) -> Tuple[List[Dict], float]:
        """Neo4j优化图遍历查询"""
        start_time = time.perf_counter()
        
        with self.neo4j_driver.session() as session:
            # 使用优化后的路径解析查询
            query = """
            MATCH (org:Organization {code: $org_code})
            WHERE org.is_current = true
            WITH org, 
                 [segment IN split(org.path, '/') WHERE segment <> ''] as path_segments
            UNWIND range(0, size(path_segments)-1) as idx
            WITH org, path_segments[idx] as ancestor_code, idx
            WHERE ancestor_code <> ''
            MATCH (ancestor:Organization {code: ancestor_code, tenant_id: org.tenant_id})
            WHERE ancestor.is_current = true
            RETURN DISTINCT
              ancestor.code as code,
              ancestor.name as name,
              ancestor.level as level,
              ancestor.path as path,
              idx as hierarchy_depth
            ORDER BY ancestor.level
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
    
    def run_comprehensive_performance_test(self, test_org_codes: List[str], iterations: int = 10) -> Dict:
        """运行综合性能测试"""
        print("🏃 开始优化后的综合性能测试...")
        
        results = {
            "test_summary": {
                "iterations": iterations,
                "test_org_codes": test_org_codes,
                "timestamp": datetime.now().isoformat(),
                "test_type": "optimized_clean_data"
            },
            "postgresql_results": {},
            "neo4j_results": {},
            "performance_comparison": {}
        }
        
        for org_code in test_org_codes:
            print(f"\n🧪 测试组织代码: {org_code}")
            
            # PostgreSQL测试
            print("📊 测试PostgreSQL优化递归CTE...")
            pg_times = []
            pg_last_result = None
            
            for i in range(iterations):
                try:
                    result, exec_time = self.postgresql_optimized_query(org_code)
                    pg_times.append(exec_time)
                    if i == 0:  # 保存第一次结果用于比较
                        pg_last_result = result
                    print(f"  第{i+1}次: {exec_time:.3f}ms, 结果数: {len(result)}")
                except Exception as e:
                    print(f"  ❌ PostgreSQL查询失败: {e}")
                    continue
            
            # Neo4j测试  
            print("📊 测试Neo4j优化图遍历...")
            neo4j_times = []
            neo4j_last_result = None
            
            for i in range(iterations):
                try:
                    result, exec_time = self.neo4j_optimized_query(org_code)
                    neo4j_times.append(exec_time)
                    if i == 0:  # 保存第一次结果用于比较
                        neo4j_last_result = result
                    print(f"  第{i+1}次: {exec_time:.3f}ms, 结果数: {len(result)}")
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
                
                # 数据一致性检查
                pg_count = len(pg_last_result) if pg_last_result else 0
                neo4j_count = len(neo4j_last_result) if neo4j_last_result else 0
                data_consistent = pg_count == neo4j_count
                
                if pg_avg > 0:
                    speedup_ratio = neo4j_avg / pg_avg
                    winner = "PostgreSQL" if pg_avg < neo4j_avg else "Neo4j"
                else:
                    speedup_ratio = 0
                    winner = "Unknown"
                
                results["performance_comparison"][org_code] = {
                    "postgresql_avg_ms": round(pg_avg, 3),
                    "neo4j_avg_ms": round(neo4j_avg, 3),
                    "speedup_ratio": round(speedup_ratio, 2),
                    "winner": winner,
                    "postgresql_result_count": pg_count,
                    "neo4j_result_count": neo4j_count,
                    "data_consistent": data_consistent,
                    "performance_improvement": abs(1 - speedup_ratio) * 100
                }
                
                print(f"📈 性能对比结果:")
                print(f"  PostgreSQL平均: {pg_avg:.3f}ms (结果数: {pg_count})")
                print(f"  Neo4j平均: {neo4j_avg:.3f}ms (结果数: {neo4j_count})")
                print(f"  数据一致性: {'✅ 一致' if data_consistent else '❌ 不一致'}")
                print(f"  性能优势: {winner} ({abs(1-speedup_ratio)*100:.1f}% faster)")
        
        return results
    
    def analyze_overall_performance(self, results: Dict) -> Dict:
        """分析整体性能表现"""
        print("\n📊 整体性能分析...")
        
        pg_all_times = []
        neo4j_all_times = []
        consistency_scores = []
        
        for org_code, comparison in results["performance_comparison"].items():
            pg_all_times.append(comparison["postgresql_avg_ms"])
            neo4j_all_times.append(comparison["neo4j_avg_ms"])
            consistency_scores.append(1 if comparison["data_consistent"] else 0)
        
        if pg_all_times and neo4j_all_times:
            overall_analysis = {
                "postgresql_overall": {
                    "avg_response_time": round(statistics.mean(pg_all_times), 3),
                    "min_response_time": round(min(pg_all_times), 3),
                    "max_response_time": round(max(pg_all_times), 3),
                    "std_dev": round(statistics.stdev(pg_all_times), 3) if len(pg_all_times) > 1 else 0
                },
                "neo4j_overall": {
                    "avg_response_time": round(statistics.mean(neo4j_all_times), 3),
                    "min_response_time": round(min(neo4j_all_times), 3),
                    "max_response_time": round(max(neo4j_all_times), 3),
                    "std_dev": round(statistics.stdev(neo4j_all_times), 3) if len(neo4j_all_times) > 1 else 0
                },
                "overall_winner": "PostgreSQL" if statistics.mean(pg_all_times) < statistics.mean(neo4j_all_times) else "Neo4j",
                "data_consistency_rate": round(statistics.mean(consistency_scores) * 100, 1),
                "performance_gap": round(abs(statistics.mean(neo4j_all_times) - statistics.mean(pg_all_times)), 3)
            }
            
            print(f"🏆 整体性能冠军: {overall_analysis['overall_winner']}")
            print(f"📊 PostgreSQL平均: {overall_analysis['postgresql_overall']['avg_response_time']}ms")
            print(f"📊 Neo4j平均: {overall_analysis['neo4j_overall']['avg_response_time']}ms")
            print(f"📊 数据一致性: {overall_analysis['data_consistency_rate']}%")
            print(f"📊 性能差距: {overall_analysis['performance_gap']}ms")
            
            return overall_analysis
        
        return {}
    
    def generate_optimization_recommendations(self, results: Dict, overall_analysis: Dict) -> Dict:
        """生成优化建议"""
        print("\n💡 生成优化建议...")
        
        recommendations = {
            "performance_summary": {
                "test_completed": True,
                "data_quality": "优秀" if overall_analysis.get("data_consistency_rate", 0) >= 90 else "需要改进",
                "recommended_approach": overall_analysis.get("overall_winner", "PostgreSQL")
            },
            "postgresql_recommendations": [
                "PostgreSQL在层级查询方面表现出色",
                "递归CTE查询性能稳定可靠",
                "适合作为CQRS架构的主要层级计算引擎",
                "建议添加适当的索引优化查询性能"
            ],
            "neo4j_recommendations": [],
            "architecture_recommendations": []
        }
        
        if overall_analysis.get("overall_winner") == "PostgreSQL":
            recommendations["neo4j_recommendations"] = [
                "Neo4j在当前场景下性能不如PostgreSQL",
                "建议暂时保持PostgreSQL作为主要层级计算引擎",
                "Neo4j可用于其他图关系查询场景",
                "如需使用Neo4j，考虑进一步优化查询算法"
            ]
            recommendations["architecture_recommendations"] = [
                "建议采用PostgreSQL为主的CQRS架构",
                "使用Redis缓存提升查询性能",
                "保持Neo4j作为辅助图查询引擎",
                "优先投入PostgreSQL查询优化"
            ]
        else:
            recommendations["neo4j_recommendations"] = [
                "Neo4j在优化后表现良好",
                "可以考虑作为主要层级查询引擎",
                "继续优化图查询算法",
                "确保数据同步机制的稳定性"
            ]
            recommendations["architecture_recommendations"] = [
                "可以考虑Neo4j为主的图查询架构",
                "保持PostgreSQL作为事务性操作引擎",
                "加强CDC数据同步机制",
                "投入Neo4j查询性能优化"
            ]
        
        return recommendations
    
    def run_complete_optimized_test(self):
        """运行完整的优化测试"""
        print("🚀 开始完整的优化后性能对比测试...")
        
        try:
            # 测试组织代码（基于清理后的数据）
            test_org_codes = ["1000000", "1000001", "1000002", "1000003"]
            
            # 1. 综合性能测试
            performance_results = self.run_comprehensive_performance_test(test_org_codes, iterations=8)
            
            # 2. 整体性能分析
            overall_analysis = self.analyze_overall_performance(performance_results)
            
            # 3. 生成优化建议
            recommendations = self.generate_optimization_recommendations(performance_results, overall_analysis)
            
            # 4. 生成最终报告
            final_report = {
                "optimization_test_completed": True,
                "performance_results": performance_results,
                "overall_analysis": overall_analysis,
                "recommendations": recommendations,
                "test_environment": {
                    "postgresql_optimized": True,
                    "neo4j_optimized": True,
                    "data_cleaned": True,
                    "test_date": datetime.now().isoformat()
                }
            }
            
            # 保存详细报告
            with open("/home/shangmeilin/cube-castle/optimized-performance-comparison-report.json", "w", encoding="utf-8") as f:
                json.dump(final_report, f, indent=2, ensure_ascii=False)
            
            print("\n✅ 优化后性能对比测试完成！")
            print(f"📋 详细报告已保存到: optimized-performance-comparison-report.json")
            
            # 输出关键结论
            if overall_analysis:
                print(f"\n🏆 最终结论:")
                print(f"  性能冠军: {overall_analysis.get('overall_winner', 'Unknown')}")
                print(f"  数据一致性: {overall_analysis.get('data_consistency_rate', 0)}%")
                print(f"  推荐方案: {recommendations['performance_summary']['recommended_approach']}")
            
            return final_report
            
        except Exception as e:
            print(f"❌ 优化测试失败: {e}")
            import traceback
            traceback.print_exc()
            return None
    
    def close(self):
        """关闭连接"""
        self.pg_conn.close()
        self.neo4j_driver.close()

if __name__ == "__main__":
    tester = OptimizedPerformanceComparison()
    try:
        result = tester.run_complete_optimized_test()
        if result:
            print("✅ 优化后性能测试成功完成")
        else:
            print("❌ 优化后性能测试失败")
    finally:
        tester.close()
