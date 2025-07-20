import asyncio
import logging
from concurrent import futures
import os

import grpc
import openai  # 引入openai库
from dotenv import load_dotenv

# 导入我们生成的gRPC代码
import intelligence_pb2
import intelligence_pb2_grpc

# --- 从 .env 文件加载环境变量 ---
load_dotenv()

# --- 这里是唯一的修改点 ---
# 创建一个OpenAI客户端，并明确告诉它使用我们的代理地址
client = openai.OpenAI(
    api_key=os.getenv("OPENAI_API_KEY"),
    base_url=os.getenv("OPENAI_API_BASE_URL"), # 使用我们新配置的代理URL
)
# -------------------------

# --- gRPC Service Implementation ---
# ... main.py 文件前面的部分保持不变 ...

# --- gRPC Service Implementation ---
class IntelligenceServiceImpl(intelligence_pb2_grpc.IntelligenceServiceServicer):
    def InterpretText(self, request: intelligence_pb2.InterpretRequest, context):
        logging.info(f"Received text: '{request.user_text}' from session '{request.session_id}'")

        tools = [
            {
                "type": "function",
                "function": {
                    "name": "get_employee_manager",
                    "description": "Get the manager of a specified employee",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "employee_id": { "type": "string", "description": "The UUID of the employee" }
                        },
                        "required": ["employee_id"],
                    },
                },
            },
            # --- 这里是新增的部分 ---
            {
                "type": "function",
                "function": {
                    "name": "update_phone_number",
                    "description": "Update an employee's phone number",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "employee_id": {
                                "type": "string",
                                "description": "The UUID of the employee whose phone number is to be updated",
                            },
                            "new_phone_number": {
                                "type": "string",
                                "description": "The new phone number",
                            }
                        },
                        "required": ["employee_id", "new_phone_number"],
                    },
                },
            }
            # -----------------------
        ]

        try:
            # 使用我们新创建的 client 对象来调用API
            response = client.chat.completions.create(
                model="deepseek-chat",
                messages=[{"role": "user", "content": request.user_text}],
                tools=tools,
                tool_choice="auto",
            )

            response_message = response.choices[0].message
            tool_calls = response_message.tool_calls

            if tool_calls:
                function_name = tool_calls[0].function.name
                function_args_json = tool_calls[0].function.arguments

                logging.info(f"LLM wants to call function: {function_name} with args: {function_args_json}")

                return intelligence_pb2.InterpretResponse(
                    intent=function_name,
                    structured_data_json=function_args_json
                )

        except Exception as e:
            logging.error(f"Error calling OpenAI: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error communicating with LLM: {e}")
            return intelligence_pb2.InterpretResponse()

        return intelligence_pb2.InterpretResponse(intent="no_intent_detected", structured_data_json="{}")

# ... 文件后面的 serve() 和 main() 函数保持不变 ...
async def serve() -> None:
    port = "50051"
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    intelligence_pb2_grpc.add_IntelligenceServiceServicer_to_server(
        IntelligenceServiceImpl(), server
    )
    server.add_insecure_port(f"[::]:{port}")
    logging.info(f"🧙 Python AI Service 'The Wizard Tower' is listening on gRPC port {port}")
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(serve())