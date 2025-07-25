#!/usr/bin/env python3
"""
AI服务全面性能测试工具
"""

import time
import statistics
import grpc
import sys
import os
import asyncio
from concurrent.futures import ThreadPoolExecutor, as_completed

# 添加当前目录到Python路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import intelligence_pb2
import intelligence_pb2_grpc

def measure_single_request_performance():
    """测量单个请求的性能"""
    print("🔍 单请求性能测试")
    print("=" * 50)
    
    channel = grpc.insecure_channel('localhost:50051')
    stub = intelligence_pb2_grpc.IntelligenceServiceStub(channel)
    
    test_cases = [
        "更新我的电话号码为13800138000",
        "谁是我的经理？",
        "我想查看我的个人信息",
        "帮我申请年假",
        "查询公司组织架构",
        "查看我的工资单",
        "申请调岗到技术部",
        "查询培训课程"
    ]
    
    response_times = []
    
    for i, text in enumerate(test_cases, 1):
        print(f"测试 {i}/{len(test_cases)}: {text[:20]}...")
        
        start_time = time.time()
        try:
            request = intelligence_pb2.InterpretRequest()
            request.user_text = text
            request.session_id = f"single-test-{i}"
            
            response = stub.InterpretText(request, timeout=30)
            end_time = time.time()
            
            response_time = (end_time - start_time) * 1000
            response_times.append(response_time)
            
            print(f"  ✅ 响应时间: {response_time:.0f}ms")
            
        except Exception as e:
            print(f"  ❌ 测试失败: {e}")
    
    channel.close()
    
    if response_times:
        print(f"\n📊 单请求性能统计:")
        print(f"  平均响应时间: {statistics.mean(response_times):.0f}ms")
        print(f"  最短响应时间: {min(response_times):.0f}ms")
        print(f"  最长响应时间: {max(response_times):.0f}ms")
        print(f"  响应时间中位数: {statistics.median(response_times):.0f}ms")
        if len(response_times) > 1:
            print(f"  标准差: {statistics.stdev(response_times):.0f}ms")
    
    return response_times

def make_concurrent_request(session_id, text):
    """发起单个并发请求"""
    try:
        channel = grpc.insecure_channel('localhost:50051')
        stub = intelligence_pb2_grpc.IntelligenceServiceStub(channel)
        
        start_time = time.time()
        
        request = intelligence_pb2.InterpretRequest()
        request.user_text = text
        request.session_id = f"concurrent-{session_id}"
        
        response = stub.InterpretText(request, timeout=30)
        end_time = time.time()
        
        response_time = (end_time - start_time) * 1000
        channel.close()
        
        return {
            'session_id': session_id,
            'response_time': response_time,
            'success': True,
            'intent': response.intent
        }
    except Exception as e:
        return {
            'session_id': session_id, 
            'response_time': None,
            'success': False,
            'error': str(e)
        }

def measure_concurrent_performance():
    """测量并发性能"""
    print("\n🚀 并发性能测试")
    print("=" * 50)
    
    # 测试不同并发级别
    concurrency_levels = [5, 10, 20]
    
    test_texts = [
        "更新我的电话号码为13800138000",
        "谁是我的经理？",
        "我想查看我的个人信息",
        "帮我申请年假",
        "查询公司组织架构"
    ]
    
    for concurrency in concurrency_levels:
        print(f"\n📈 并发级别: {concurrency}")
        print("-" * 30)
        
        # 准备测试数据
        test_data = []
        for i in range(concurrency):
            text = test_texts[i % len(test_texts)]
            test_data.append((i, text))
        
        start_total = time.time()
        
        # 使用线程池执行并发请求
        with ThreadPoolExecutor(max_workers=concurrency) as executor:
            future_to_session = {
                executor.submit(make_concurrent_request, session_id, text): session_id 
                for session_id, text in test_data
            }
            
            results = []
            for future in as_completed(future_to_session):
                result = future.result()
                results.append(result)
        
        end_total = time.time()
        total_time = (end_total - start_total) * 1000
        
        # 统计结果
        successful_results = [r for r in results if r['success']]
        failed_count = len(results) - len(successful_results)
        
        if successful_results:
            response_times = [r['response_time'] for r in successful_results]
            
            print(f"  总用时: {total_time:.0f}ms")
            print(f"  成功请求: {len(successful_results)}/{concurrency}")
            print(f"  失败请求: {failed_count}")
            print(f"  平均响应时间: {statistics.mean(response_times):.0f}ms")
            print(f"  最大响应时间: {max(response_times):.0f}ms")
            print(f"  吞吐量: {len(successful_results) / (total_time / 1000):.1f} req/s")
        else:
            print(f"  ❌ 所有请求都失败了")

def measure_cache_performance():
    """测量缓存性能"""
    print("\n💾 缓存性能测试")
    print("=" * 50)
    
    channel = grpc.insecure_channel('localhost:50051')
    stub = intelligence_pb2_grpc.IntelligenceServiceStub(channel)
    
    test_text = "缓存性能测试请求"
    
    # 第一次请求 - 应该调用AI模型
    print("第一次请求（无缓存）:")
    start_time = time.time()
    
    request = intelligence_pb2.InterpretRequest()
    request.user_text = test_text
    request.session_id = "cache-test-1"
    
    response = stub.InterpretText(request, timeout=30)
    end_time = time.time()
    
    first_time = (end_time - start_time) * 1000
    print(f"  响应时间: {first_time:.0f}ms")
    
    # 第二次请求 - 应该命中缓存
    print("第二次请求（命中缓存）:")
    start_time = time.time()
    
    request = intelligence_pb2.InterpretRequest()
    request.user_text = test_text
    request.session_id = "cache-test-2"
    
    response = stub.InterpretText(request, timeout=30)
    end_time = time.time()
    
    second_time = (end_time - start_time) * 1000
    print(f"  响应时间: {second_time:.0f}ms")
    
    # 计算缓存效果
    if first_time > 0:
        improvement = ((first_time - second_time) / first_time) * 100
        print(f"\n📊 缓存效果分析:")
        print(f"  性能提升: {improvement:.1f}%")
        print(f"  加速比: {first_time / second_time:.1f}x")
    
    channel.close()

def run_comprehensive_test():
    """运行全面性能测试"""
    print("🏰 Cube Castle - AI服务全面性能测试")
    print("=" * 60)
    
    # 1. 单请求性能测试
    single_response_times = measure_single_request_performance()
    
    # 2. 并发性能测试
    measure_concurrent_performance()
    
    # 3. 缓存性能测试
    measure_cache_performance()
    
    # 4. 总结报告
    print("\n" + "=" * 60)
    print("📋 测试总结报告")
    print("=" * 60)
    
    if single_response_times:
        avg_time = statistics.mean(single_response_times)
        print(f"平均单请求响应时间: {avg_time:.0f}ms")
        
        if avg_time < 2000:
            print("✅ 性能目标达成：平均响应时间 < 2000ms")
        else:
            needed_improvement = ((avg_time - 2000) / avg_time) * 100
            print(f"❌ 需要改进 {needed_improvement:.1f}% 以达到目标")
    
    print("✅ 全面性能测试完成！")

if __name__ == "__main__":
    try:
        run_comprehensive_test()
    except Exception as e:
        print(f"❌ 测试失败: {e}")
        sys.exit(1)