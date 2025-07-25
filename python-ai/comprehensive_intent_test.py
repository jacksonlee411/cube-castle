#!/usr/bin/env python3
"""
智能网关意图识别优化验证测试
"""
import requests
import json
import time

def test_intent_recognition():
    """测试意图识别准确率"""
    test_cases = [
        ('更新我的电话号码为13800138000', 'update_phone_number'),
        ('谁是我的经理？', 'get_employee_manager'),
        ('我想知道我的上级是谁', 'get_employee_manager'),
        ('修改手机号为18888888888', 'update_phone_number'),
        ('查看我的经理信息', 'get_employee_manager'),
        ('更新电话号码', 'update_phone_number'),
        ('我的主管是谁？', 'get_employee_manager'),
        ('换个手机号', 'update_phone_number'),
    ]
    
    correct = 0
    total = len(test_cases)
    
    print("🧠 智能意图识别准确率测试")
    print("=" * 50)
    
    for text, expected_intent in test_cases:
        try:
            response = requests.post(
                'http://localhost:8080/api/v1/intelligence/interpret',
                json={
                    'query': text, 
                    'user_id': '11111111-1111-1111-1111-111111111111'
                },
                timeout=15
            )
            
            if response.status_code == 200:
                result = response.json()
                actual_intent = result.get('intent')
                
                if actual_intent == expected_intent:
                    correct += 1
                    status = "✅"
                else:
                    status = "❌"
                
                print(f"{status} 文本: {text}")
                print(f"   期望: {expected_intent} | 实际: {actual_intent}")
                if 'structured_data' in result:
                    print(f"   结构化数据: {result['structured_data']}")
            else:
                print(f"❌ API错误: {response.status_code} - {response.text[:100]}")
                
        except Exception as e:
            print(f"❌ 请求失败: {e}")
        
        time.sleep(0.5)  # 避免过于频繁的请求
    
    accuracy = (correct / total) * 100
    print("\n" + "=" * 50)
    print(f"📊 测试结果:")
    print(f"   总测试案例: {total}")
    print(f"   识别正确: {correct}")
    print(f"   准确率: {accuracy:.1f}%")
    
    if accuracy >= 90:
        print("🎉 意图识别准确率达标!")
    else:
        print("⚠️  意图识别准确率需要进一步优化")
    
    return accuracy

if __name__ == '__main__':
    test_intent_recognition()