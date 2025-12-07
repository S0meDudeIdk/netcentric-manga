class WebSocketService {
  constructor() {
    this.connections = new Map(); // roomId -> { ws, listeners }
  }

  connect(roomId, token, callbacks) {
    // Check if connection already exists
    if (this.connections.has(roomId)) {
      const existing = this.connections.get(roomId);

      // If WebSocket is already open, just add the listener
      if (existing.ws.readyState === WebSocket.OPEN) {
        console.log(`✅ Reusing existing connection for room ${roomId}`);
        this.addListener(roomId, callbacks);
        // Trigger onOpen for the new listener
        callbacks.onOpen?.();
        return existing.ws;
      }

      // If connecting, wait for it
      if (existing.ws.readyState === WebSocket.CONNECTING) {
        console.log(`⏳ Connection already in progress for room ${roomId}`);
        this.addListener(roomId, callbacks);
        return existing.ws;
      }

      // Otherwise, clean up and create new connection
      this.disconnect(roomId);
    }

    console.log(`🔌 Creating new WebSocket connection for room ${roomId}`);

    const wsUrl = `${process.env.REACT_APP_WS_URL || 'ws://localhost:8080'}/api/v1/ws/chat?room=${roomId}&token=${token}`;
    const ws = new WebSocket(wsUrl);

    const connectionData = {
      ws,
      listeners: new Set([callbacks]),
      roomId,
      token
    };

    ws.onopen = () => {
      console.log(`✅ Connected to room ${roomId}`);

      // Notify all listeners
      connectionData.listeners.forEach(cb => cb.onOpen?.());
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        // Broadcast to all listeners for this room
        connectionData.listeners.forEach(cb => cb.onMessage?.(data));
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error(`❌ WebSocket error for room ${roomId}:`, error);

      // Notify all listeners
      connectionData.listeners.forEach(cb => cb.onError?.(error));
    };

    ws.onclose = (event) => {
      console.log(`🔌 Disconnected from room ${roomId}`, event);

      // Notify all listeners
      connectionData.listeners.forEach(cb => cb.onClose?.(event));

      // Clean up connection
      console.log(`🧹 Cleaning up connection for room ${roomId}`);
      this.connections.delete(roomId);
    };

    this.connections.set(roomId, connectionData);
    return ws;
  }

  addListener(roomId, callbacks) {
    const connection = this.connections.get(roomId);
    if (connection) {
      connection.listeners.add(callbacks);
      console.log(`➕ Added listener for room ${roomId}. Total: ${connection.listeners.size}`);
    }
  }

  removeListener(roomId, callbacks) {
    const connection = this.connections.get(roomId);
    if (connection) {
      connection.listeners.delete(callbacks);
      console.log(`➖ Removed listener for room ${roomId}. Remaining: ${connection.listeners.size}`);

      // If no more listeners, disconnect
      if (connection.listeners.size === 0) {
        console.log(`🧹 No more listeners for room ${roomId}, disconnecting...`);
        this.disconnect(roomId);
      }
    }
  }

  send(roomId, message) {
    const connection = this.connections.get(roomId);
    if (connection && connection.ws.readyState === WebSocket.OPEN) {
      connection.ws.send(JSON.stringify(message));
      return true;
    }
    console.warn(`Cannot send message to room ${roomId}: not connected`);
    return false;
  }

  disconnect(roomId) {
    const connection = this.connections.get(roomId);
    if (connection) {
      // Close WebSocket
      if (connection.ws.readyState === WebSocket.OPEN ||
        connection.ws.readyState === WebSocket.CONNECTING) {
        connection.ws.close(1000, 'Intentional disconnect'); // Normal closure
      }

      this.connections.delete(roomId);
      console.log(`🔌 Disconnected from room ${roomId}`);
    }
  }

  disconnectAll() {
    console.log('🔌 Disconnecting all WebSocket connections...');
    this.connections.forEach((_, roomId) => {
      this.disconnect(roomId);
    });
  }

  getConnectionStatus(roomId) {
    const connection = this.connections.get(roomId);
    if (!connection) return 'disconnected';

    switch (connection.ws.readyState) {
      case WebSocket.CONNECTING: return 'connecting';
      case WebSocket.OPEN: return 'connected';
      case WebSocket.CLOSING: return 'closing';
      case WebSocket.CLOSED: return 'disconnected';
      default: return 'unknown';
    }
  }
}

const webSocketServiceInstance = new WebSocketService();
export default webSocketServiceInstance;