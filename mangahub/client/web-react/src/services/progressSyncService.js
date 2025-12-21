/**
 * Progress Sync Service
 * Handles real-time TCP progress updates via Server-Sent Events (SSE)
 */

class ProgressSyncService {
  constructor() {
    this.eventSource = null;
    this.listeners = new Set();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 3000;
    this.isConnected = false;
  }

  /**
   * Connect to SSE progress stream
   * @param {string} token - JWT authentication token
   * @param {function} onProgress - Callback for progress updates
   * @param {function} onError - Callback for errors (optional)
   * @param {function} onConnect - Callback when connected (optional)
   */
  connect(token, onProgress, onError, onConnect) {
    if (this.eventSource) {
      console.warn('Progress sync already connected');
      return;
    }

    // Add listener
    const listener = { onProgress, onError, onConnect };
    this.listeners.add(listener);

    // Get API URL
    const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080';
    const url = `${apiUrl}/api/v1/sse/progress`;

    console.log('🔗 Connecting to TCP Progress Sync via SSE...');

    // Create EventSource with authentication
    // Note: EventSource doesn't support custom headers, so we use query param
    this.eventSource = new EventSource(`${url}?token=${token}`);

    // Handle connection
    this.eventSource.addEventListener('connected', (event) => {
      console.log('✅ TCP Progress Sync connected:', event.data);
      this.isConnected = true;
      this.reconnectAttempts = 0;
      
      // Notify connection listeners
      this.listeners.forEach(listener => {
        if (listener.onConnect) {
          listener.onConnect();
        }
      });
    });

    // Handle progress updates
    this.eventSource.addEventListener('message', (event) => {
      try {
        const update = JSON.parse(event.data);
        console.log('📡 TCP Progress Update:', update);
        
        // Notify all listeners
        this.listeners.forEach(listener => {
          if (listener.onProgress) {
            listener.onProgress(update);
          }
        });
      } catch (error) {
        console.error('Error parsing progress update:', error);
      }
    });

    // Handle ping (keep-alive)
    this.eventSource.addEventListener('ping', (event) => {
      // Keep-alive ping received
    });

    // Handle errors
    this.eventSource.onerror = (error) => {
      console.error('❌ TCP Progress Sync error:', error);
      this.isConnected = false;

      // Notify error listeners
      this.listeners.forEach(listener => {
        if (listener.onError) {
          listener.onError(error);
        }
      });

      // Attempt reconnection
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        console.log(`Reconnecting to TCP Progress Sync (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        
        setTimeout(() => {
          this.disconnect();
          this.connect(token, onProgress, onError);
        }, this.reconnectDelay);
      } else {
        console.error('Max reconnection attempts reached. Please refresh the page.');
        this.disconnect();
      }
    };
  }

  /**
   * Disconnect from SSE progress stream
   */
  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
      this.isConnected = false;
      this.listeners.clear();
      console.log('🔌 TCP Progress Sync disconnected');
    }
  }

  /**
   * Add a listener for progress updates
   * @param {function} onProgress - Callback for progress updates
   * @param {function} onError - Callback for errors (optional)
   * @param {function} onConnect - Callback when connected (optional)
   */
  addListener(onProgress, onError, onConnect) {
    const listener = { onProgress, onError, onConnect };
    this.listeners.add(listener);
    return listener;
  }

  /**
   * Remove a specific listener
   * @param {object} listener - The listener object to remove
   */
  removeListener(listener) {
    this.listeners.delete(listener);
  }

  /**
   * Check if service is connected
   * @returns {boolean}
   */
  isServiceConnected() {
    return this.isConnected;
  }
}

// Export singleton instance
const progressSyncService = new ProgressSyncService();
export default progressSyncService;
