import { useState, useCallback, useEffect } from 'react';

interface WebSocketMessage {
  type: string;
  projectId?: string;
  data?: any;
}

interface UseWebSocketReturn {
  isConnected: boolean;
  sendMessage: (message: WebSocketMessage) => void;
  lastMessage: WebSocketMessage | null;
  connectionStatus: 'connecting' | 'connected' | 'disconnected' | 'error';
}

export const useWebSocket = (projectId?: string): UseWebSocketReturn => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected');

  const sendMessage = useCallback((message: WebSocketMessage) => {
    // Simulate sending message
    console.log('Sending message:', message);
    
    // Simulate receiving response after delay
    setTimeout(() => {
      if (message.type === 'compile_request') {
        setLastMessage({
          type: 'compile_response',
          data: {
            success: true,
            errors: [],
            warnings: []
          }
        });
      }
    }, 1000);
  }, []);

  useEffect(() => {
    if (projectId) {
      setConnectionStatus('connecting');
      setTimeout(() => {
        setIsConnected(true);
        setConnectionStatus('connected');
      }, 1000);
    }

    return () => {
      setIsConnected(false);
      setConnectionStatus('disconnected');
    };
  }, [projectId]);

  return {
    isConnected,
    sendMessage,
    lastMessage,
    connectionStatus
  };
};