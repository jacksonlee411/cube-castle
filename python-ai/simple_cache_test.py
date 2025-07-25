#!/usr/bin/env python3
"""
简单缓存测试
"""

import time
import grpc
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import intelligence_pb2
import intelligence_pb2_grpc

def test_cache():
    channel = grpc.insecure_channel('localhost:50051')
    stub = intelligence_pb2_grpc.IntelligenceServiceStub(channel)
    
    test_text = "测试缓存功能"
    
    print("第1次请求:")
    start = time.time()
    request = intelligence_pb2.InterpretRequest()
    request.user_text = test_text
    request.session_id = "test-1"
    response1 = stub.InterpretText(request, timeout=30)
    time1 = (time.time() - start) * 1000
    print(f"响应时间: {time1:.0f}ms, 意图: {response1.intent}")
    
    print("\n第2次请求 (应该命中缓存):")
    start = time.time()  
    request = intelligence_pb2.InterpretRequest()
    request.user_text = test_text
    request.session_id = "test-2"
    response2 = stub.InterpretText(request, timeout=30)
    time2 = (time.time() - start) * 1000
    print(f"响应时间: {time2:.0f}ms, 意图: {response2.intent}")
    
    print("\n第3次请求 (应该命中缓存):")
    start = time.time()
    request = intelligence_pb2.InterpretRequest()
    request.user_text = test_text
    request.session_id = "test-3"
    response3 = stub.InterpretText(request, timeout=30)
    time3 = (time.time() - start) * 1000
    print(f"响应时间: {time3:.0f}ms, 意图: {response3.intent}")
    
    channel.close()

if __name__ == "__main__":
    test_cache()