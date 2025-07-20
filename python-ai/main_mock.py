import asyncio
import logging
from concurrent import futures
import json
import uuid

import grpc

# 导入我们生成的gRPC代码
import intelligence_pb2
import intelligence_pb2_grpc

# --- gRPC Service Implementation ---
class IntelligenceServiceImpl(intelligence_pb2_grpc.IntelligenceServiceServicer):
    def InterpretText(self, request: intelligence_pb2.InterpretRequest, context):
        logging.info(f"Received text: '{request.user_text}' from session '{request.session_id}'")

        # 模拟 AI 响应逻辑
        user_text = request.user_text.lower()
        
        # 检查是否包含电话号码更新相关的关键词
        if any(keyword in user_text for keyword in ['电话', '手机', '号码', 'phone', 'update']):
            # 模拟更新电话号码的意图
            mock_response = {
                "employee_id": str(uuid.uuid4()),
                "new_phone_number": "13800138000"
            }
            
            return intelligence_pb2.InterpretResponse(
                intent="update_phone_number",
                structured_data_json=json.dumps(mock_response)
            )
        
        # 检查是否包含查询员工相关的关键词
        elif any(keyword in user_text for keyword in ['员工', '查询', '查找', 'employee', 'search']):
            # 模拟查询员工的意图
            mock_response = {
                "employee_id": str(uuid.uuid4())
            }
            
            return intelligence_pb2.InterpretResponse(
                intent="get_employee_manager",
                structured_data_json=json.dumps(mock_response)
            )
        
        # 默认返回无意图检测
        return intelligence_pb2.InterpretResponse(
            intent="no_intent_detected", 
            structured_data_json="{}"
        )

async def serve() -> None:
    port = "50051"
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    intelligence_pb2_grpc.add_IntelligenceServiceServicer_to_server(
        IntelligenceServiceImpl(), server
    )
    server.add_insecure_port(f"[::]:{port}")
    logging.info(f"🧙 Python AI Service 'The Wizard Tower' (MOCK) is listening on gRPC port {port}")
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(serve()) 