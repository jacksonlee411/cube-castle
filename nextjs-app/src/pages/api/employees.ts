// Next.js API route to proxy requests to Go backend
import { NextApiRequest, NextApiResponse } from 'next';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const { method, query } = req;
  
  try {
    // 构建后端URL
    const backendUrl = `http://localhost:8080/api/v1/corehr/employees`;
    const searchParams = new URLSearchParams();
    
    // 传递查询参数
    if (query.page) searchParams.append('page', String(query.page));
    if (query.page_size) searchParams.append('page_size', String(query.page_size));
    if (query.search) searchParams.append('search', String(query.search));
    
    const fullUrl = `${backendUrl}?${searchParams.toString()}`;
    
    console.log('🔗 API代理请求:', fullUrl);
    
    // 发送请求到Go后端
    const response = await fetch(fullUrl, {
      method: method || 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      throw new Error(`Backend responded with ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    console.log('✅ 后端响应:', data.total_count, '个员工');
    
    // 返回数据给前端
    res.status(200).json(data);
    
  } catch (error: any) {
    console.error('❌ API代理错误:', error.message);
    res.status(500).json({ 
      error: error.message || 'Internal server error',
      success: false 
    });
  }
}