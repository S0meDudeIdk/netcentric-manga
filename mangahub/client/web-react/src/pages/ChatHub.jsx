import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  MessageCircle, Users, TrendingUp, Send, Settings, 
  User, Circle, Book, Bell, Wifi, Activity, ArrowLeft,
  Info, Hash, Loader
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
  const [userProgress, setUserProgress] = useState({});
  
  // Hub statistics
  const [hubStats, setHubStats] = useState({
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
  }, []);

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
    <div className="min-h-screen bg-gray-50">
      {/* Notifications Toast */}
      <AnimatePresence>
        {notifications.map((notif) => (
          <motion.div
            key={notif.id}
            initial={{ opacity: 0, y: -50, x: '-50%' }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -50 }}
            className="fixed top-4 left-1/2 transform -translate-x-1/2 bg-blue-600 text-white px-6 py-3 rounded-lg shadow-lg z-50 flex items-center gap-2"
          >
            <Bell className="w-5 h-5" />
            <span>{notif.message}</span>
          </motion.div>
        ))}
      </AnimatePresence>

      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button
                onClick={() => navigate(`/manga/${mangaId}`)}
                className="p-2 hover:bg-gray-100 rounded-lg transition"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              
              <div className="flex items-center gap-3">
                <div className="w-12 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded flex items-center justify-center">
                  <Book className="w-6 h-6 text-white" />
                </div>
                <div>
                  <h1 className="text-xl font-bold text-gray-900 flex items-center gap-2">
                    <Hash className="w-5 h-5 text-gray-400" />
                    {manga.title}
                  </h1>
                  <p className="text-sm text-gray-600">Chat Hub</p>
                </div>
              </div>
            </div>

            {/* Connection Status */}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2 px-3 py-1 bg-gray-100 rounded-lg">
                <div className="flex items-center gap-1">
                  <Wifi className={`w-4 h-4 ${isConnected ? 'text-green-500' : 'text-gray-400'}`} />
                  <span className="text-xs font-medium">WebSocket</span>
                </div>
                <Circle className={`w-2 h-2 ${isConnected ? 'fill-green-500 text-green-500' : 'fill-gray-400 text-gray-400'}`} />
              </div>
              
              {/* TCP/UDP Disabled - WebSocket only mode */}
              {/* <div className="flex items-center gap-2 px-3 py-1 bg-gray-100 rounded-lg">
                <Activity className="w-4 h-4 text-blue-500" />
                <span className="text-xs font-medium">TCP/UDP Active</span>
              </div> */}
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="container mx-auto px-4 py-6" style={{ height: 'calc(100vh - 140px)' }}>
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 h-full">
          {/* Left Sidebar - Hub Info */}
          <div className="lg:col-span-1 space-y-4 overflow-y-auto" style={{ maxHeight: '100%' }}>
            {/* Manga Info Card */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              className="bg-white rounded-lg shadow-sm p-4"
            >
              <div className="flex items-center gap-2 mb-3">
                <Info className="w-5 h-5 text-blue-600" />
                <h3 className="font-semibold text-gray-900">Manga Info</h3>
              </div>
              
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Status:</span>
                  <span className="font-medium">{manga.status || 'Ongoing'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Chapters:</span>
                  <span className="font-medium">{manga.chapters || 'N/A'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Rating:</span>
                  <span className="font-medium">⭐ {manga.score || 'N/A'}</span>
                </div>
              </div>
            </motion.div>

            {/* Hub Statistics */}
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.1 }}
              className="bg-white rounded-lg shadow-sm p-4"
            >
              <div className="flex items-center gap-2 mb-3">
                <TrendingUp className="w-5 h-5 text-green-600" />
                <h3 className="font-semibold text-gray-900">Hub Stats</h3>
              </div>
              
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Users className="w-4 h-4 text-gray-400" />
                    <span className="text-sm text-gray-600">Online Users</span>
                  </div>
                  <span className="font-bold text-lg text-blue-600">{onlineUsers.length}</span>
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <MessageCircle className="w-4 h-4 text-gray-400" />
                    <span className="text-sm text-gray-600">Messages</span>
                  </div>
                  <span className="font-bold text-lg text-green-600">{messages.filter(m => m.type === 'message').length}</span>
                </div>
              </div>
            </motion.div>
          </div>

          {/* Center - Chat Area */}
          <div className="lg:col-span-2 h-full">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-white rounded-lg shadow-sm flex flex-col h-full overflow-hidden"
            >
              {/* Chat Header */}
              <div className="px-4 py-3 border-b flex items-center justify-between flex-shrink-0">
                <div className="flex items-center gap-2">
                  <MessageCircle className="w-5 h-5 text-blue-600" />
                  <h3 className="font-semibold text-gray-900">Chat Room</h3>
                </div>
                <span className="text-sm text-gray-500">
                  {onlineUsers.length} online
                </span>
              </div>

              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-4 space-y-3" style={{ minHeight: 0 }}>
                {messages.length === 0 ? (
                  <div className="text-center text-gray-500 py-8">
                    <MessageCircle className="w-12 h-12 mx-auto mb-2 text-gray-300" />
                    <p>No messages yet. Start the conversation!</p>
                  </div>
                ) : (
                  messages.map((msg) => {
                    if (msg.type === 'system' || msg.type === 'notification') {
                      return (
                        <div key={msg.id} className="text-center">
                          <span className="inline-block px-3 py-1 bg-gray-100 text-gray-600 text-sm rounded-full">
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
                        className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}
                      >
                        <div className={`max-w-[70%] ${isOwn ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-900'} rounded-lg px-4 py-2`}>
                          <div className="flex items-center gap-2 mb-1">
                            <span className={`text-xs font-semibold ${isOwn ? 'text-blue-100' : 'text-gray-600'}`}>
                              {msg.username}
                            </span>
                            <span className={`text-xs ${isOwn ? 'text-blue-200' : 'text-gray-400'}`}>
                              {new Date(msg.timestamp).toLocaleTimeString()}
                            </span>
                          </div>
                          <p className="text-sm break-words">{msg.message}</p>
                        </div>
                      </motion.div>
                    );
                  })
                )}
                <div ref={messagesEndRef} />
              </div>

              {/* Input Area */}
              <div className="p-4 border-t flex-shrink-0">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    onKeyPress={handleKeyPress}
                    placeholder={isConnected ? "Type a message..." : "Connecting..."}
                    disabled={!isConnected}
                    className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100"
                  />
                  <button
                    onClick={handleSendMessage}
                    disabled={!isConnected || !messageInput.trim()}
                    className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                  >
                    <Send className="w-4 h-4" />
                    Send
                  </button>
                </div>
              </div>
            </motion.div>
          </div>

          {/* Right Sidebar - User List */}
          <div className="lg:col-span-1 h-full">
            <motion.div
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              className="bg-white rounded-lg shadow-sm h-full flex flex-col overflow-hidden"
            >
              {/* User List Header */}
              <div className="px-4 py-3 border-b flex-shrink-0">
                <div className="flex items-center gap-2">
                  <Users className="w-5 h-5 text-blue-600" />
                  <h3 className="font-semibold text-gray-900">Users</h3>
                  <span className="ml-auto text-sm text-gray-500">
                    {onlineUsers.length}
                  </span>
                </div>
              </div>

              {/* User List */}
              <div className="flex-1 overflow-y-auto p-3 space-y-2" style={{ minHeight: 0 }}>
                {onlineUsers.length === 0 ? (
                  <div className="text-center text-gray-500 py-8">
                    <Users className="w-12 h-12 mx-auto mb-2 text-gray-300" />
                    <p className="text-sm">No users online</p>
                  </div>
                ) : (
                  onlineUsers.map((user, index) => {
                    const progress = userProgress[user.user_id];
                    return (
                      <motion.div
                        key={user.user_id || index}
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.05 }}
                        className="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 transition"
                      >
                        <div className="relative">
                          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                            <User className="w-5 h-5 text-white" />
                          </div>
                          <Circle className="absolute bottom-0 right-0 w-3 h-3 fill-green-500 text-green-500 border-2 border-white rounded-full" />
                        </div>
                        
                        <div className="flex-1 min-w-0">
                          <p className="font-medium text-sm text-gray-900 truncate">
                            {user.username}
                          </p>
                          {/* Progress display disabled - TCP not active */}
                          {/* {progress && (
                            <p className="text-xs text-gray-500 flex items-center gap-1">
                              <Book className="w-3 h-3" />
                              Ch. {progress.current_chapter || 0}/{manga.chapters || '?'}
                            </p>
                          )} */}
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
