import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  MessageCircle, Users, TrendingUp, Send, User, Circle, Bell, ArrowLeft, Info
} from 'lucide-react';
import authService from '../services/authService';
import mangaService from '../services/mangaService';
import websocketService from '../services/websocketService';
import LoadingSpinner from '../components/LoadingSpinner';

const ChatHub = () => {
  const { mangaId } = useParams();
  const navigate = useNavigate();
  
  // State management
  const [manga, setManga] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  // Chat state
  const [messages, setMessages] = useState([]);
  const [messageInput, setMessageInput] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  
  // User list state
  const [onlineUsers, setOnlineUsers] = useState([]);
  
  // Hub statistics - keeping for potential future use
  const [, setHubStats] = useState({
    totalUsers: 0,
    activeChats: 0,
    lastUpdate: null
  });
  
  // Notifications
  const [notifications, setNotifications] = useState([]);
  
  // Refs
  const messagesEndRef = useRef(null);
  const wsInitializedRef = useRef(false);
  
  const currentUser = authService.getUser();

  // Fetch manga details
  useEffect(() => {
    const fetchManga = async () => {
      try {
        setLoading(true);
        
        // Check if this is a MAL ID (format: mal-123)
        let data;
        const idStr = mangaId.toString();
        
        if (idStr.startsWith('mal-')) {
          // Extract MAL ID and fetch from MAL API
          const malId = idStr.replace('mal-', '');
          console.log('Fetching MAL manga with ID:', malId);
          data = await mangaService.getMALMangaById(malId);
        } else {
          // Fetch from local database
          console.log('Fetching local manga with ID:', mangaId);
          data = await mangaService.getMangaById(mangaId);
        }
        
        setManga(data);
      } catch (err) {
        setError('Failed to load manga information');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    if (!authService.isAuthenticated()) {
      navigate('/login');
      return;
    }

    fetchManga();
  }, [mangaId, navigate]);

  // Add notification helper (moved up to be used in useEffect)
  const addNotification = useCallback((message) => {
    const id = Date.now();
    setNotifications(prev => [...prev, { id, message }]);
    
    // Remove notification after 3 seconds
    setTimeout(() => {
      setNotifications(prev => prev.filter(n => n.id !== id));
    }, 3000);
  }, [setNotifications]);

  // Handle incoming WebSocket messages
  const handleWebSocketMessage = useCallback((data) => {
    console.log('Received WebSocket message:', data);
    
    switch (data.type) {
      case 'message':
        setMessages(prev => [...prev, {
          id: `${data.user_id}-${data.timestamp}`,
          user_id: data.user_id,
          username: data.username,
          message: data.message,
          timestamp: data.timestamp * 1000,
          type: 'message'
        }]);
        
        setHubStats(prev => ({
          ...prev,
          activeChats: prev.activeChats + 1,
          lastUpdate: new Date()
        }));
        break;
        
      case 'join':
        setMessages(prev => [...prev, {
          id: `join-${data.user_id}-${Date.now()}`,
          message: data.message,
          timestamp: Date.now(),
          type: 'system'
        }]);
        break;
        
      case 'leave':
        setMessages(prev => [...prev, {
          id: `leave-${data.user_id}-${Date.now()}`,
          message: data.message,
          timestamp: Date.now(),
          type: 'system'
        }]);
        break;
        
      case 'user_list':
        if (data.users && Array.isArray(data.users)) {
          setOnlineUsers(data.users);
          setHubStats(prev => ({
            ...prev,
            totalUsers: data.users.length
          }));
        }
        break;
        
      default:
        console.log('Unknown message type:', data.type);
    }
  }, []);

  // WebSocket connection - runs once when manga and user are ready
  useEffect(() => {
    // Skip if dependencies not ready
    if (!manga || !currentUser) {
      console.log('⏳ Waiting for manga and user data...');
      return;
    }
    
    // Skip if already initialized
    if (wsInitializedRef.current) {
      console.log('⏭️ WebSocket already initialized');
      return;
    }
    
    // Mark as initialized
    wsInitializedRef.current = true;
    console.log('🔌 Initializing WebSocket connection...');
    
    const token = authService.getToken();
    const callbacks = {
      onOpen: () => {
        console.log('✅ Connected to chat');
        setIsConnected(true);
        addNotification('Connected to chat hub');
      },
      onMessage: handleWebSocketMessage,
      onError: (error) => {
        console.error('❌ WebSocket error:', error);
        setIsConnected(false);
      },
      onClose: (event) => {
        console.log('🔌 Disconnected from chat', event.code, event.reason);
        setIsConnected(false);
        wsInitializedRef.current = false; // Allow reconnection
      }
    };
    
    // Connect - singleton service handles deduplication
    websocketService.connect(mangaId, token, callbacks);
    
    // NO cleanup - let connection persist
  }, [manga, currentUser, mangaId, handleWebSocketMessage, addNotification]);

  // Only disconnect when truly leaving the page
  useEffect(() => {
    return () => {
      console.log('🚪 Leaving ChatHub - disconnecting WebSocket');
      wsInitializedRef.current = false;
      websocketService.disconnect(mangaId);
    };
  }, [mangaId]);

  // Send chat message
  const handleSendMessage = useCallback(() => {
    if (!messageInput.trim() || !isConnected) return;
    
    const message = {
      type: 'message',
      message: messageInput.trim(),
      room: mangaId
    };
    
    const sent = websocketService.send(mangaId, message);
    if (sent) {
      setMessageInput('');
    } else {
      addNotification('Failed to send message');
    }
  }, [messageInput, isConnected, mangaId]);

  // Handle Enter key in message input
  const handleKeyPress = useCallback((e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  }, [handleSendMessage]);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <LoadingSpinner />
      </div>
    );
  }

  if (error || !manga) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error || 'Manga not found'}</p>
          <button
            onClick={() => navigate(-1)}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark">
      {/* Notifications Toast */}
      <AnimatePresence>
        {notifications.map((notif) => (
          <motion.div
            key={notif.id}
            initial={{ opacity: 0, y: -50, x: '-50%' }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -50 }}
            className="fixed top-4 left-1/2 transform -translate-x-1/2 bg-primary text-white px-6 py-3 rounded-lg shadow-lg z-50 flex items-center gap-2"
          >
            <Bell className="w-5 h-5" />
            <span>{notif.message}</span>
          </motion.div>
        ))}
      </AnimatePresence>

      {/* Header */}
      <div className="bg-white dark:bg-[#191022] shadow-sm border-b border-zinc-200 dark:border-zinc-800">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button
                onClick={() => navigate(`/manga/${mangaId}`)}
                className="p-2 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded-lg transition"
              >
                <ArrowLeft className="w-5 h-5 text-zinc-600 dark:text-zinc-400" />
              </button>
              
              <div>
                <h1 className="text-xl font-bold text-zinc-900 dark:text-white flex items-center gap-2">
                  {manga.title} - Discussion
                </h1>
                <div className="flex items-center gap-2 mt-1">
                  <div className="flex items-center gap-1">
                    <Users className="w-4 h-4 text-zinc-400" />
                    <span className="text-sm text-zinc-600 dark:text-zinc-400">+5</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Connection Status */}
            <div className="flex items-center gap-2">
              <div className={`flex items-center gap-2 px-3 py-1.5 rounded-lg ${
                isConnected 
                  ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400' 
                  : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400'
              }`}>
                <Circle className={`w-2 h-2 ${isConnected ? 'fill-green-500 text-green-500' : 'fill-zinc-400 text-zinc-400'}`} />
                <span className="text-xs font-medium">{isConnected ? 'Connected' : 'Disconnected'}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="container mx-auto px-4 py-6" style={{ height: 'calc(100vh - 140px)' }}>
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 h-full">
          {/* Left Sidebar - Hub Info */}
          <div className="lg:col-span-3 space-y-4 overflow-y-auto" style={{ maxHeight: '100%' }}>
            {/* Manga Info Card */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 p-6"
            >
              <div className="flex items-center gap-2 mb-4">
                <Info className="w-5 h-5 text-primary" />
                <h3 className="font-bold text-zinc-900 dark:text-white">Manga Info</h3>
              </div>
              
              <div className="space-y-3 text-sm">
                <div>
                  <p className="text-zinc-500 dark:text-zinc-400 text-xs mb-1">Status:</p>
                  <p className="font-semibold text-zinc-900 dark:text-white">{manga.status || 'ongoing'}</p>
                </div>
                <div>
                  <p className="text-zinc-500 dark:text-zinc-400 text-xs mb-1">Chapters:</p>
                  <p className="font-semibold text-zinc-900 dark:text-white">N/A</p>
                </div>
                <div>
                  <p className="text-zinc-500 dark:text-zinc-400 text-xs mb-1">Rating:</p>
                  <p className="font-semibold text-zinc-900 dark:text-white">⭐ N/A</p>
                </div>
              </div>
            </motion.div>

            {/* Hub Statistics */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.1 }}
              className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 p-6"
            >
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp className="w-5 h-5 text-green-600" />
                <h3 className="font-bold text-zinc-900 dark:text-white">Hub Stats</h3>
              </div>
              
              <div className="space-y-4">
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-zinc-600 dark:text-zinc-400">Online Users:</span>
                    <span className="font-bold text-zinc-900 dark:text-white">{onlineUsers.length}</span>
                  </div>
                </div>
                
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-zinc-600 dark:text-zinc-400">Messages:</span>
                    <span className="font-bold text-zinc-900 dark:text-white">{messages.filter(m => m.type === 'message').length}</span>
                  </div>
                </div>
              </div>
            </motion.div>
          </div>

          {/* Center - Chat Area */}
          <div className="lg:col-span-6 h-full">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 flex flex-col h-full overflow-hidden"
            >
              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-6 space-y-4" style={{ minHeight: 0 }}>
                {messages.length === 0 ? (
                  <div className="text-center text-zinc-500 dark:text-zinc-400 py-8">
                    <MessageCircle className="w-12 h-12 mx-auto mb-2 text-zinc-300 dark:text-zinc-600" />
                    <p>No messages yet. Start the conversation!</p>
                  </div>
                ) : (
                  messages.map((msg) => {
                    if (msg.type === 'system' || msg.type === 'notification') {
                      return (
                        <div key={msg.id} className="text-center">
                          <span className="inline-block px-3 py-1 bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400 text-sm rounded-full">
                            {msg.type === 'notification' && <Bell className="w-3 h-3 inline mr-1" />}
                            {msg.message}
                          </span>
                        </div>
                      );
                    }

                    const isOwn = msg.user_id === currentUser?.user_id;
                    return (
                      <motion.div
                        key={msg.id}
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        className={`flex items-start gap-3 ${isOwn ? 'justify-end' : 'justify-start'}`}
                      >
                        {!isOwn && (
                          <div className="w-10 h-10 bg-gradient-to-br from-primary to-purple-600 rounded-full flex items-center justify-center flex-shrink-0">
                            <User className="w-5 h-5 text-white" />
                          </div>
                        )}
                        <div className={`max-w-[70%] ${isOwn ? 'bg-primary text-white' : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-white'} rounded-2xl px-4 py-3`}>
                          <div className="flex items-center gap-2 mb-1">
                            <span className={`text-sm font-bold ${isOwn ? 'text-white' : 'text-zinc-900 dark:text-white'}`}>
                              {msg.username}
                            </span>
                            <span className={`text-xs ${isOwn ? 'text-white/70' : 'text-zinc-500 dark:text-zinc-400'}`}>
                              {new Date(msg.timestamp).toLocaleTimeString()}
                            </span>
                          </div>
                          <p className="text-sm break-words">{msg.message}</p>
                        </div>
                        {isOwn && (
                          <div className="w-10 h-10 bg-gradient-to-br from-primary to-purple-600 rounded-full flex items-center justify-center flex-shrink-0">
                            <User className="w-5 h-5 text-white" />
                          </div>
                        )}
                      </motion.div>
                    );
                  })
                )}
                <div ref={messagesEndRef} />
              </div>

              {/* Input Area */}
              <div className="p-4 border-t border-zinc-200 dark:border-zinc-800 flex-shrink-0">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    onKeyPress={handleKeyPress}
                    placeholder={isConnected ? "Type a message..." : "Connecting..."}
                    disabled={!isConnected}
                    className="flex-1 px-4 py-3 border border-zinc-300 dark:border-zinc-700 rounded-xl bg-white dark:bg-zinc-900 text-zinc-900 dark:text-white focus:ring-2 focus:ring-primary/50 focus:border-primary outline-none disabled:bg-zinc-100 dark:disabled:bg-zinc-800"
                  />
                  <button
                    onClick={handleSendMessage}
                    disabled={!isConnected || !messageInput.trim()}
                    className="px-6 py-3 bg-primary text-white rounded-xl hover:bg-primary/90 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 font-bold shadow-lg shadow-primary/25"
                  >
                    <Send className="w-4 h-4" />
                  </button>
                </div>
              </div>
            </motion.div>
          </div>

          {/* Right Sidebar - User List */}
          <div className="lg:col-span-3 h-full">
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 h-full flex flex-col overflow-hidden"
            >
              {/* User List Header */}
              <div className="px-6 py-4 border-b border-zinc-200 dark:border-zinc-800 flex-shrink-0">
                <div className="flex items-center gap-2">
                  <Users className="w-5 h-5 text-primary" />
                  <h3 className="font-bold text-zinc-900 dark:text-white">Users in Chat</h3>
                </div>
              </div>

              {/* User List */}
              <div className="flex-1 overflow-y-auto p-4 space-y-3" style={{ minHeight: 0 }}>
                {onlineUsers.length === 0 ? (
                  <div className="text-center text-zinc-500 dark:text-zinc-400 py-8">
                    <Users className="w-12 h-12 mx-auto mb-2 text-zinc-300 dark:text-zinc-600" />
                    <p className="text-sm">No users online</p>
                  </div>
                ) : (
                  onlineUsers.map((user, index) => {
                    return (
                      <motion.div
                        key={user.user_id || index}
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.05 }}
                        className="flex items-center gap-3 p-3 rounded-lg hover:bg-zinc-50 dark:hover:bg-zinc-800 transition"
                      >
                        <div className="relative">
                          <div className="w-10 h-10 bg-gradient-to-br from-primary to-purple-600 rounded-full flex items-center justify-center">
                            <User className="w-5 h-5 text-white" />
                          </div>
                          <Circle className="absolute bottom-0 right-0 w-3 h-3 fill-green-500 text-green-500 border-2 border-white dark:border-zinc-900 rounded-full" />
                        </div>
                        
                        <div className="flex-1 min-w-0">
                          <p className="font-semibold text-sm text-zinc-900 dark:text-white truncate">
                            {user.username}
                          </p>
                          <p className="text-xs text-zinc-500 dark:text-zinc-400">Reading</p>
                        </div>
                      </motion.div>
                    );
                  })
                )}
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatHub;
