import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import progressSyncService from '../services/progressSyncService';
import udpNotificationService from '../services/udpNotificationService';

const RealtimeSyncPage = () => {
  const navigate = useNavigate();
  const [progressUpdates, setProgressUpdates] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [progressConnected, setProgressConnected] = useState(false);
  const [notificationConnected, setNotificationConnected] = useState(false);
  const token = localStorage.getItem('token');

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }

    // Connect to TCP Progress Sync
    progressSyncService.connect(
      token,
      (update) => {
        setProgressUpdates(prev => [update, ...prev].slice(0, 20));
      },
      (error) => {
        console.error('Progress sync error:', error);
        setProgressConnected(false);
      },
      () => {
        // Connection established
        setProgressConnected(true);
      }
    );

    // Connect to UDP Notifications
    udpNotificationService.connect(
      token,
      (notification) => {
        setNotifications(prev => [notification, ...prev].slice(0, 20));
      },
      (error) => {
        console.error('Notification error:', error);
        setNotificationConnected(false);
      },
      () => {
        // Connection established
        setNotificationConnected(true);
      }
    );

    // Cleanup on unmount
    return () => {
      progressSyncService.disconnect();
      udpNotificationService.disconnect();
    };
  }, [token, navigate]);

  const formatTimestamp = (timestamp) => {
    if (!timestamp) return 'N/A';
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString();
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
      <div className="container mx-auto px-4 max-w-7xl">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Real-time Sync & Notifications
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Live TCP progress updates and UDP notifications via Server-Sent Events
          </p>
        </div>

        {/* Connection Status */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
          <div className={`p-4 rounded-lg ${progressConnected ? 'bg-green-100 dark:bg-green-900' : 'bg-yellow-100 dark:bg-yellow-900'}`}>
            <div className="flex items-center">
              <div className={`w-3 h-3 rounded-full mr-3 ${progressConnected ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`}></div>
              <div>
                <h3 className="font-semibold text-gray-900 dark:text-white">TCP Progress Sync</h3>
                <p className="text-sm text-gray-600 dark:text-gray-300">
                  {progressConnected ? 'Connected - Receiving updates' : 'Connecting...'}
                </p>
              </div>
            </div>
          </div>

          <div className={`p-4 rounded-lg ${notificationConnected ? 'bg-green-100 dark:bg-green-900' : 'bg-yellow-100 dark:bg-yellow-900'}`}>
            <div className="flex items-center">
              <div className={`w-3 h-3 rounded-full mr-3 ${notificationConnected ? 'bg-green-500 animate-pulse' : 'bg-yellow-500'}`}></div>
              <div>
                <h3 className="font-semibold text-gray-900 dark:text-white">UDP Notifications</h3>
                <p className="text-sm text-gray-600 dark:text-gray-300">
                  {notificationConnected ? 'Connected - Receiving notifications' : 'Connecting...'}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* TCP Progress Updates */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                📡 Progress Updates (TCP)
              </h2>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {progressUpdates.length} updates
              </span>
            </div>
            
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {progressUpdates.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <p>Waiting for progress updates...</p>
                  <p className="text-sm mt-2">Updates appear when users read manga</p>
                </div>
              ) : (
                progressUpdates.map((update, index) => (
                  <div key={index} className="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <p className="font-semibold text-gray-900 dark:text-white">
                          {update.username || update.user_id}
                        </p>
                        <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">
                          Reading: <span className="font-medium">{update.manga_title || update.manga_id}</span>
                        </p>
                        <p className="text-sm text-blue-600 dark:text-blue-400 mt-1">
                          Chapter {update.chapter}
                        </p>
                      </div>
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {formatTimestamp(update.timestamp)}
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* UDP Notifications */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                🔔 Notifications (UDP)
              </h2>
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {notifications.length} notifications
              </span>
            </div>
            
            <div className="space-y-3 max-h-96 overflow-y-auto">
              {notifications.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <p>Waiting for notifications...</p>
                  <p className="text-sm mt-2">New chapters, updates, etc.</p>
                </div>
              ) : (
                notifications.map((notification, index) => (
                  <div key={index} className="p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg border border-purple-200 dark:border-purple-800">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <span className="text-xs px-2 py-1 bg-purple-200 dark:bg-purple-800 text-purple-800 dark:text-purple-200 rounded">
                            {notification.type}
                          </span>
                        </div>
                        <p className="text-sm text-gray-700 dark:text-gray-300 mt-2">
                          {notification.message}
                        </p>
                        {notification.manga_id && (
                          <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                            Manga ID: {notification.manga_id}
                          </p>
                        )}
                      </div>
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        {formatTimestamp(notification.timestamp)}
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Info Section */}
        <div className="mt-8 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-6">
          <h3 className="font-bold text-gray-900 dark:text-white mb-3">ℹ️ How it works</h3>
          <div className="grid md:grid-cols-2 gap-4 text-sm text-gray-700 dark:text-gray-300">
            <div>
              <h4 className="font-semibold mb-2">TCP Progress Sync</h4>
              <ul className="list-disc list-inside space-y-1">
                <li>Real-time reading progress updates</li>
                <li>Broadcasts when users read manga</li>
                <li>Shows username, manga, and chapter</li>
                <li>Uses Server-Sent Events (SSE)</li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-2">UDP Notifications</h4>
              <ul className="list-disc list-inside space-y-1">
                <li>Chapter release notifications</li>
                <li>Manga status updates</li>
                <li>System announcements</li>
                <li>Uses Server-Sent Events (SSE)</li>
              </ul>
            </div>
          </div>
          <div className="mt-4 p-3 bg-white dark:bg-gray-800 rounded">
            <p className="text-xs text-gray-600 dark:text-gray-400">
              <strong>Note:</strong> Make sure TCP Server (port 9001) and UDP Server (port 8081) are running. 
              The API server connects to these servers and bridges the data to the web client via SSE.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RealtimeSyncPage;
