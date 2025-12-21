class NotificationService {
  constructor() {
    this.notifications = [];
    this.listeners = new Set();
    this.ws = null;
    this.tcpEventSource = null;
    this.udpEventSource = null;
    this.wsReconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 3000;
    this.expiryTime = 6 * 60 * 60 * 1000; // 6 hours in milliseconds
    
    // Start cleanup timer to remove expired notifications
    this.startExpiryCleanup();
  }

  // Start periodic cleanup of expired notifications (every 5 minutes)
  startExpiryCleanup() {
    setInterval(() => {
      const now = Date.now();
      const initialCount = this.notifications.length;
      
      this.notifications = this.notifications.filter(
        n => (now - n.timestamp) < this.expiryTime
      );
      
      if (this.notifications.length < initialCount) {
        const removed = initialCount - this.notifications.length;
        console.log(`🗑️ Cleaned up ${removed} expired notification(s)`);
        this.notifyListeners();
      }
    }, 5 * 60 * 1000); // Check every 5 minutes
  }

  // Initialize WebSocket connection for notifications
  connect(token) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('Notification service already connected');
      return;
    }

    const wsUrl = process.env.REACT_APP_WS_URL
      ? `${process.env.REACT_APP_WS_URL}/api/v1/ws/chat?room=global-notifications&token=${token}`
      : `ws://${window.location.hostname}:8080/api/v1/ws/chat?room=global-notifications&token=${token}`;

    console.log('Connecting to notification WebSocket...');
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('✅ Connected to notification service');
      this.wsReconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Handle different types of notifications
        if (data.type === 'notification' || data.type === 'progress_update') {
          const notification = {
            id: Date.now() + Math.random(),
            type: data.type,
            message: data.message,
            mangaId: data.manga_id,
            mangaTitle: data.manga_title,
            timestamp: Date.now(),
            read: false
          };

          this.addToNotifications(notification);
        }
      } catch (error) {
        console.error('Error parsing notification:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('❌ Notification WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('🔌 Disconnected from notification service');
      this.ws = null;
      
      // Attempt to reconnect
      if (this.wsReconnectAttempts < this.maxReconnectAttempts) {
        this.wsReconnectAttempts++;
        console.log(`Attempting to reconnect... (${this.wsReconnectAttempts}/${this.maxReconnectAttempts})`);
        setTimeout(() => this.connect(token), this.reconnectDelay);
      }
    };

    // Also connect to TCP progress updates via SSE
    this.connectTCPProgress(token);
    
    // Connect to UDP notifications via SSE
    this.connectUDPNotifications(token);
  }

  // Connect to TCP Progress Updates (SSE)
  connectTCPProgress(token) {
    if (this.tcpEventSource) {
      console.log('TCP progress stream already connected');
      return;
    }

    const apiUrl = process.env.REACT_APP_API_URL || `http://${window.location.hostname}:8080`;
    const url = `${apiUrl}/api/v1/sse/progress?token=${token}`;

    console.log('🔗 Connecting to TCP Progress Updates...');
    this.tcpEventSource = new EventSource(url);

    this.tcpEventSource.addEventListener('connected', () => {
      console.log('✅ TCP Progress Updates connected');
    });

    this.tcpEventSource.addEventListener('message', (event) => {
      try {
        const data = JSON.parse(event.data);
        const notification = {
          id: Date.now() + Math.random(),
          type: 'progress_update',
          message: `${data.username || 'Someone'} is reading "${data.manga_title}" - Chapter ${data.chapter}`,
          mangaId: null,
          mangaTitle: data.manga_title,
          timestamp: data.timestamp * 1000 || Date.now(),
          read: false,
          userId: data.user_id,
          chapter: data.chapter
        };

        this.addToNotifications(notification);
      } catch (error) {
        console.error('Error parsing TCP progress update:', error);
      }
    });

    this.tcpEventSource.onerror = (error) => {
      console.error('❌ TCP Progress SSE error:', error);
      this.tcpEventSource?.close();
      this.tcpEventSource = null;
    };
  }

  // Connect to UDP Notifications (SSE)
  connectUDPNotifications(token) {
    if (this.udpEventSource) {
      console.log('UDP notifications stream already connected');
      return;
    }

    const apiUrl = process.env.REACT_APP_API_URL || `http://${window.location.hostname}:8080`;
    const url = `${apiUrl}/api/v1/sse/notifications?token=${token}`;

    console.log('🔗 Connecting to UDP Notifications...');
    this.udpEventSource = new EventSource(url);

    this.udpEventSource.addEventListener('connected', () => {
      console.log('✅ UDP Notifications connected');
    });

    this.udpEventSource.addEventListener('message', (event) => {
      try {
        const data = JSON.parse(event.data);
        const notification = {
          id: Date.now() + Math.random(),
          type: data.type || 'notification',
          message: data.message,
          mangaId: null,
          mangaTitle: null,
          timestamp: data.timestamp * 1000 || Date.now(),
          read: false
        };

        this.addToNotifications(notification);
      } catch (error) {
        console.error('Error parsing UDP notification:', error);
      }
    });

    this.udpEventSource.onerror = (error) => {
      console.error('❌ UDP Notification SSE error:', error);
      this.udpEventSource?.close();
      this.udpEventSource = null;
    };
  }

  // Helper method to add notification and trigger updates
  addToNotifications(notification) {
    this.notifications.unshift(notification);
    
    // Keep only last 100 notifications
    if (this.notifications.length > 100) {
      this.notifications = this.notifications.slice(0, 100);
    }

    this.notifyListeners();
    this.showBrowserNotification(notification);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    if (this.tcpEventSource) {
      this.tcpEventSource.close();
      this.tcpEventSource = null;
    }
    if (this.udpEventSource) {
      this.udpEventSource.close();
      this.udpEventSource = null;
    }
  }

  // Subscribe to notification updates
  subscribe(callback) {
    this.listeners.add(callback);
    
    // Return unsubscribe function
    return () => {
      this.listeners.delete(callback);
    };
  }

  notifyListeners() {
    this.listeners.forEach(callback => {
      try {
        callback(this.notifications);
      } catch (error) {
        console.error('Error in notification listener:', error);
      }
    });
  }

  // Show browser notification
  async showBrowserNotification(notification) {
    if (!('Notification' in window)) {
      return;
    }

    if (Notification.permission === 'granted') {
      try {
        new Notification('MangaHub Notification', {
          body: notification.message,
          icon: '/favicon.ico',
          badge: '/favicon.ico',
          tag: notification.id
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

  // Get all notifications
  getNotifications() {
    return this.notifications;
  }

  // Get unread count
  getUnreadCount() {
    return this.notifications.filter(n => !n.read).length;
  }

  // Mark notification as read
  markAsRead(notificationId) {
    const notification = this.notifications.find(n => n.id === notificationId);
    if (notification) {
      notification.read = true;
      this.notifyListeners();
    }
  }

  // Mark all as read
  markAllAsRead() {
    this.notifications.forEach(n => n.read = true);
    this.notifyListeners();
  }

  // Clear all notifications
  clearAll() {
    this.notifications = [];
    this.notifyListeners();
  }

  // Add a manual notification (for testing or manual triggers)
  addNotification(type, message, mangaId = null, mangaTitle = null) {
    const notification = {
      id: Date.now() + Math.random(),
      type,
      message,
      mangaId,
      mangaTitle,
      timestamp: Date.now(),
      read: false
    };

    this.notifications.unshift(notification);
    if (this.notifications.length > 50) {
      this.notifications = this.notifications.slice(0, 50);
    }

    this.notifyListeners();
    this.showBrowserNotification(notification);
  }
}

const notificationService = new NotificationService();
export default notificationService;
