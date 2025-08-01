import { NextApiRequest, NextApiResponse } from 'next';

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  console.log('🧪 Error Test API: Triggering server error for error boundary test');
  
  // Return a 500 error to test error handling
  res.status(500).json({
    error: 'Internal Server Error',
    message: '服务器内部错误，用于测试错误边界机制',
    timestamp: new Date().toISOString(),
    type: 'network'
  });
}