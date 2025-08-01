import { NextApiRequest, NextApiResponse } from 'next';

// Test endpoint to verify logging
export default function handler(req: NextApiRequest, res: NextApiResponse) {
  console.log('🧪 Test API endpoint called');
  res.status(200).json({ 
    message: 'Test successful',
    timestamp: new Date().toISOString()
  });
}