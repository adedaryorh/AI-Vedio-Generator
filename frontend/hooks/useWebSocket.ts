import { useEffect, useState, useCallback } from 'react';

interface UseWebSocketOptions {
  onMessage?: (data: any) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
}

/**
 * Custom hook for WebSocket connection.
 * @param url - WebSocket server URL
 * @param options - Callback functions for WebSocket events
 * @returns [sendMessage, isConnected, error]
 */
export function useWebSocket(
  url: string,
  options: UseWebSocketOptions = {}
): [(data: any) => void, boolean, Error | null] {
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const { onMessage, onOpen, onClose, onError } = options;

  const sendMessage = useCallback(
    (data: any) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(data));
      } else {
        console.warn('WebSocket is not connected. Message not sent.');
      }
    },
    [ws]
  );

  useEffect(() => {
    // Prevent creating multiple WebSocket connections in strict mode
    if (ws) {
      return;
    }

    try {
      const websocket = new WebSocket(url);
      setWs(websocket);

      websocket.onopen = () => {
        setIsConnected(true);
        setError(null);
        onOpen?.();
      };

      websocket.onmessage = (event) => {
        let data: any;
        try {
          data = JSON.parse(event.data);
        } catch {
          data = event.data;
        }
        onMessage?.(data);
      };

      websocket.onclose = () => {
        setIsConnected(false);
        setWs(null);
        onClose?.();
      };

      websocket.onerror = (event) => {
        setError(new Error('WebSocket error'));
        onError?.(event);
      };
    } catch (err) {
      setError(err as Error);
    }

    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, [url, onMessage, onOpen, onClose, onError]);

  return [sendMessage, isConnected, error];
}