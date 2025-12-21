/**
 * UDP Notification Service
 * Handles real-time UDP notifications via Server-Sent Events (SSE)
 * For chapter releases, manga updates, and system notifications
 */

class UDPNotificationService {
  constructor() {
    this.eventSource = null;
    this.listeners = new Set();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 3000;
    this.isConnected = false;
    this.notifications = [];
  }

  /**
   * Connect to SSE notification stream
   * @param {string} token - JWT authentication token
   * @param {function} onNotification - Callback for notifications
   * @param {function} onError - Callback for errors (optional)
   * @param {function} onConnect - Callback when connected (optional)
   */
  connect(token, onNotification, onError, onConnect) {
    if (this.eventSource) {
      console.warn('UDP notification service already connected');
      return;
    }

    // Add listener
    const listener = { onNotification, onError, onConnect };
    this.listeners.add(listener);

    // Get API URL
    const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080';
    const url = `${apiUrl}/api/v1/sse/notifications`;

    console.log('🔗 Connecting to UDP Notification Server via SSE...');

    // Create EventSource with authentication
    this.eventSource = new EventSource(`${url}?token=${token}`);

    // Handle connection
    this.eventSource.addEventListener('connected', (event) => {
      console.log('✅ UDP Notification Service connected:', event.data);
      this.isConnected = true;
      this.reconnectAttempts = 0;
      
      // Notify connection listeners
      this.listeners.forEach(listener => {
        if (listener.onConnect) {
          listener.onConnect();
        }
      });
    });

    // Handle notifications
    this.eventSource.addEventListener('message', (event) => {
      try {
        const notification = JSON.parse(event.data);
        console.log('🔔 UDP Notification:', notification);
        
        // Add to notification list
        this.notifications.unshift({
          ...notification,
          id: Date.now() + Math.random(),
          read: false,
          receivedAt: Date.now()
        });

        // Keep only last 50 notifications
        if (this.notifications.length > 50) {
          this.notifications = this.notifications.slice(0, 50);
        }
        
        // Notify all listeners
        this.listeners.forEach(listener => {
          if (listener.onNotification) {
            listener.onNotification(notification);
          }
        });

        // Show browser notification
        this.showBrowserNotification(notification);
      } catch (error) {
        console.error('Error parsing UDP notification:', error);
      }
    });

    // Handle ping (keep-alive)
    this.eventSource.addEventListener('ping', (event) => {
      // Keep-alive ping received
    });

    // Handle errors
    this.eventSource.onerror = (error) => {
      console.error('❌ UDP Notification Service error:', error);
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
        console.log(`Reconnecting to UDP Notification Service (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        
        setTimeout(() => {
          this.disconnect();
          this.connect(token, onNotification, onError);
        }, this.reconnectDelay);
      } else {
        console.error('Max reconnection attempts reached. Please refresh the page.');
        this.disconnect();
      }
    };
  }

  /**
   * Disconnect from SSE notification stream
   */
  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
      this.isConnected = false;
      this.listeners.clear();
      console.log('🔌 UDP Notification Service disconnected');
    }
  }

  /**
   * Add a listener for notifications
   * @param {function} onNotification - Callback for notifications
   * @param {function} onError - Callback for errors (optional)
   * @param {function} onConnect - Callback when connected (optional)
   */
  addListener(onNotification, onError, onConnect) {
    const listener = { onNotification, onError, onConnect };
    this.listeners.add(listener);
    
    // If already connected, notify immediately
    if (this.isConnected && onConnect) {
      onConnect();
    }
    
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

  /**
   * Get all notifications
   * @returns {Array}
   */
  getNotifications() {
    return this.notifications;
  }

  /**
   * Get unread notification count
   * @returns {number}
   */
  getUnreadCount() {
    return this.notifications.filter(n => !n.read).length;
  }

  /**
   * Mark notification as read
   * @param {string} notificationId
   */
  markAsRead(notificationId) {
    const notification = this.notifications.find(n => n.id === notificationId);
    if (notification) {
      notification.read = true;
    }
  }

  /**
   * Mark all notifications as read
   */
  markAllAsRead() {
    this.notifications.forEach(n => n.read = true);
  }

  /**
   * Clear all notifications
   */
  clearAll() {
    this.notifications = [];
  }

  /**
   * Show browser notification
   * @param {object} notification
   */
  async showBrowserNotification(notification) {
    if (!('Notification' in window)) {
      return;
    }

    if (Notification.permission === 'granted') {
      try {
        new Notification('MangaHub Update', {
          body: notification.message,
          icon: '/favicon.ico',
          badge: '/favicon.ico',
          tag: notification.type
        });
      } catch (error) {
        console.error('Error showing browser notification:', error);
      }
    } else if (Notification.permission !== 'denied') {
      const permission = await Notification.requestPermission();
      if (permission === 'granted') {
        this.showBrowserNotification(notification);
      }
    }
  }
}

// Export singleton instance
const udpNotificationService = new UDPNotificationService();
export default udpNotificationService;
