#!/usr/bin/env python3
"""
AI服务性能基线测试工具
"""

import time
import statistics
import grpc
import sys
import os

# 添加当前目录到Python路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import intelligence_pb2
import intelligence_pb2_grpc

def measure_ai_performance():
    """测量AI服务当前性能基线"""
    print("🔍 AI服务性能基线测量")
    print("=" * 50)
    
    # 连接到AI服务
    channel = grpc.insecure_channel('localhost:50051')
    stub = intelligence_pb2_grpc.IntelligenceServiceStub(channel)
    
    test_cases = [
        "更新我的电话号码为13800138000",
        "谁是我的经理？",
        "我想查看我的个人信息",
        "帮我申请年假",
        "查询公司组织架构"
    ]
    
    response_times = []
    
    print(f"测试用例数量: {len(test_cases)}")
    print("开始性能测试...\n")
    
    for i, text in enumerate(test_cases, 1):
        print(f"测试 {i}/{len(test_cases)}: {text[:20]}...")
        
        start_time = time.time()
        try:
            request = intelligence_pb2.InterpretRequest()
            request.user_text = text
            request.session_id = f"perf-test-{i}"
            
            response = stub.InterpretText(request, timeout=30)
            end_time = time.time()
            
            response_time = (end_time - start_time) * 1000  # 转换为毫秒
            response_times.append(response_time)
            
            print(f"  ✅ 响应时间: {response_time:.0f}ms")
            print(f"  📋 意图: {response.intent}")
            
        except Exception as e:
            print(f"  ❌ 测试失败: {e}")
    
    # 统计分析
    if response_times:
        print("\n" + "=" * 50)
        print("📊 性能统计结果:")
        print(f"  平均响应时间: {statistics.mean(response_times):.0f}ms")
        print(f"  最短响应时间: {min(response_times):.0f}ms")
        print(f"  最长响应时间: {max(response_times):.0f}ms")
        print(f"  响应时间中位数: {statistics.median(response_times):.0f}ms")
        if len(response_times) > 1:
            print(f"  标准差: {statistics.stdev(response_times):.0f}ms")
        
        # 性能评估
        avg_time = statistics.mean(response_times)
        print(f"\n🎯 性能目标分析:")
        print(f"  当前平均响应时间: {avg_time:.0f}ms")
        print(f"  目标响应时间: <2000ms")
        if avg_time > 2000:
            improvement_needed = ((avg_time - 2000) / avg_time) * 100
            print(f"  需要改进: {improvement_needed:.1f}%")
            print(f"  状态: ❌ 需要优化")
        else:
            print(f"  状态: ✅ 已达标")
    
    channel.close()
    return response_times

if __name__ == "__main__":
    try:
        response_times = measure_ai_performance()
    except Exception as e:
        print(f"❌ 性能测试失败: {e}")
        sys.exit(1)