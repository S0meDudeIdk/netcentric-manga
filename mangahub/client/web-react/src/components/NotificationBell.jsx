import React, { useState, useEffect, useRef } from 'react';
import { Bell, CheckCheck, Trash2, X, BookOpen, Plus, TrendingUp } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import notificationService from '../services/notificationService';
import authService from '../services/authService';

const NotificationBell = () => {
  const [notifications, setNotifications] = useState([]);
  const [showDropdown, setShowDropdown] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const dropdownRef = useRef(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Initialize notification service connection
    const token = authService.getToken();
    if (token) {
      notificationService.connect(token);
    }

    // Subscribe to notification updates
    const unsubscribe = notificationService.subscribe((updatedNotifications) => {
      setNotifications(updatedNotifications);
      setUnreadCount(notificationService.getUnreadCount());
    });

    // Load initial notifications
    setNotifications(notificationService.getNotifications());
    setUnreadCount(notificationService.getUnreadCount());

    // Close dropdown when clicking outside
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setShowDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);

    return () => {
      unsubscribe();
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleNotificationClick = (notification) => {
    // Mark as read
    notificationService.markAsRead(notification.id);
    
    // Navigate to manga detail if mangaId is present
    if (notification.mangaId) {
      navigate(`/manga/${notification.mangaId}`);
      setShowDropdown(false);
    }
  };

  const handleMarkAllAsRead = () => {
    notificationService.markAllAsRead();
  };

  const handleClearAll = () => {
    notificationService.clearAll();
  };

  const getNotificationIcon = (type) => {
    switch (type) {
      case 'new_manga':
        return <Plus className="w-4 h-4 text-green-500" />;
      case 'new_chapter':
        return <BookOpen className="w-4 h-4 text-blue-500" />;
      case 'progress_update':
        return <TrendingUp className="w-4 h-4 text-purple-500" />;
      default:
        return <Bell className="w-4 h-4 text-zinc-500" />;
    }
  };

  const formatTimestamp = (timestamp) => {
    const now = Date.now();
    const diff = now - timestamp;
    
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (seconds < 60) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    
    return new Date(timestamp).toLocaleDateString();
  };

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Bell Button */}
      <button
        onClick={() => setShowDropdown(!showDropdown)}
        className="relative p-2 text-zinc-600 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800"
      >
        <Bell className="w-5 h-5" />
        
        {/* Unread Badge */}
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs font-bold rounded-full flex items-center justify-center animate-pulse">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </button>

      {/* Dropdown */}
      {showDropdown && (
        <div className="absolute right-0 mt-2 w-96 max-h-[500px] bg-white dark:bg-[#211c27] rounded-xl shadow-xl border border-zinc-200 dark:border-zinc-700 overflow-hidden animate-in fade-in zoom-in-95 duration-200 z-50">
          {/* Header */}
          <div className="sticky top-0 bg-white dark:bg-[#211c27] border-b border-zinc-200 dark:border-zinc-700 px-4 py-3 flex items-center justify-between">
            <h3 className="text-sm font-bold text-zinc-900 dark:text-white">
              Notifications {unreadCount > 0 && `(${unreadCount})`}
            </h3>
            <div className="flex items-center gap-2">
              {notifications.length > 0 && (
                <>
                  <button
                    onClick={handleMarkAllAsRead}
                    className="p-1 text-zinc-500 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors"
                    title="Mark all as read"
                  >
                    <CheckCheck className="w-4 h-4" />
                  </button>
                  <button
                    onClick={handleClearAll}
                    className="p-1 text-zinc-500 dark:text-zinc-400 hover:text-red-500 dark:hover:text-red-400 transition-colors"
                    title="Clear all"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </>
              )}
              <button
                onClick={() => setShowDropdown(false)}
                className="p-1 text-zinc-500 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white transition-colors"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* Notification List */}
          <div className="max-h-[420px] overflow-y-auto">
            {notifications.length > 0 ? (
              notifications.map((notification) => (
                <button
                  key={notification.id}
                  onClick={() => handleNotificationClick(notification)}
                  className={`w-full px-4 py-3 text-left border-b border-zinc-100 dark:border-zinc-800 hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors ${
                    !notification.read ? 'bg-primary/5 dark:bg-primary/10' : ''
                  }`}
                >
                  <div className="flex items-start gap-3">
                    {/* Icon */}
                    <div className="mt-0.5 flex-shrink-0">
                      {getNotificationIcon(notification.type)}
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-zinc-900 dark:text-white line-clamp-2">
                        {notification.message}
                      </p>
                      {notification.mangaTitle && (
                        <p className="text-xs text-primary font-medium mt-1 truncate">
                          {notification.mangaTitle}
                        </p>
                      )}
                      <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-1">
                        {formatTimestamp(notification.timestamp)}
                      </p>
                    </div>

                    {/* Unread Indicator */}
                    {!notification.read && (
                      <div className="w-2 h-2 bg-primary rounded-full flex-shrink-0 mt-1.5" />
                    )}
                  </div>
                </button>
              ))
            ) : (
              <div className="px-4 py-12 text-center">
                <Bell className="w-12 h-12 mx-auto mb-3 text-zinc-300 dark:text-zinc-700" />
                <p className="text-sm text-zinc-500 dark:text-zinc-400 font-medium">
                  No notifications yet
                </p>
                <p className="text-xs text-zinc-400 dark:text-zinc-500 mt-1">
                  You'll be notified about new manga and chapters
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default NotificationBell;
