import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  MessageCircle, Users, Send, User, Circle, ArrowLeft
} from 'lucide-react';
import authService from '../services/authService';
import websocketService from '../services/websocketService';

const GeneralChat = () => {
  const navigate = useNavigate();
  
  // Chat state
  const [messages, setMessages] = useState([]);
  const [messageInput, setMessageInput] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  
  // User list state
  const [onlineUsers, setOnlineUsers] = useState([]);
  
  const messagesEndRef = useRef(null);
  const wsInitializedRef = useRef(false);
  const isMountedRef = useRef(true);
  const currentUser = authService.getUser();

  // Auto-scroll to bottom when new messages arrive
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

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
          // Extract usernames from user objects
          const usernames = data.users.map(u => typeof u === 'string' ? u : u.username);
          setOnlineUsers(usernames);
        }
        break;
        
      default:
        console.log('Unknown message type:', data.type);
    }
  }, []);

  // WebSocket connection management
  useEffect(() => {
    if (!authService.isAuthenticated()) {
      navigate('/login');
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
        console.log('✅ Connected to general chat');
        setIsConnected(true);
      },
      onMessage: handleWebSocketMessage,
      onError: (error) => {
        console.error('❌ WebSocket error:', error);
        setIsConnected(false);
      },
      onClose: (event) => {
        console.log('🔌 Disconnected from general chat', event.code, event.reason);
        setIsConnected(false);
        wsInitializedRef.current = false; // Allow reconnection
      }
    };

    // Connect to general chat room - singleton service handles deduplication
    websocketService.connect('general', token, callbacks);

    // NO cleanup - let connection persist
  }, [navigate, handleWebSocketMessage]);

  // Only disconnect when truly leaving the page
  useEffect(() => {
    isMountedRef.current = true;
    
    return () => {
      // Use a timeout to detect if this is a real unmount or just strict mode
      setTimeout(() => {
        if (!isMountedRef.current) {
          console.log('🚪 Leaving General Chat - disconnecting WebSocket');
          wsInitializedRef.current = false;
          websocketService.disconnect('general');
        }
      }, 0);
      
      isMountedRef.current = false;
    };
  }, []);

  const handleSendMessage = useCallback((e) => {
    e.preventDefault();
    
    if (!messageInput.trim() || !isConnected) return;

    const message = {
      type: 'message',
      message: messageInput.trim(),
      room: 'general'
    };

    const sent = websocketService.send('general', message);
    if (sent) {
      // Optimistically add message to UI immediately
      setMessages(prev => [...prev, {
        id: `${currentUser?.user_id || 'temp'}-${Date.now()}`,
        user_id: currentUser?.user_id,
        username: currentUser?.username,
        message: messageInput.trim(),
        timestamp: Date.now(),
        type: 'message'
      }]);
      setMessageInput('');
    } else {
      console.error('Failed to send message');
    }
  }, [messageInput, isConnected, currentUser]);

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit'
    });
  };

  return (
    <div className="min-h-screen bg-background-light dark:bg-background-dark">
      <div className="container mx-auto px-4 py-6">
        {/* Header */}
        <div className="mb-6">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 text-zinc-600 dark:text-zinc-400 hover:text-primary dark:hover:text-primary transition-colors mb-4"
          >
            <ArrowLeft className="w-5 h-5" />
            <span>Back</span>
          </button>

          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-black text-zinc-900 dark:text-white mb-2">
                General Chat
              </h1>
              <p className="text-zinc-600 dark:text-zinc-400">
                Chat with manga enthusiasts from around the world
              </p>
            </div>
            
            <div className="flex items-center gap-2">
              <div className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-[#211c27] rounded-lg border border-zinc-200 dark:border-zinc-700">
                <Circle className={`w-2 h-2 ${isConnected ? 'fill-green-500 text-green-500' : 'fill-zinc-400 text-zinc-400'}`} />
                <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
                  {isConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Chat Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Main Chat Area */}
          <div className="lg:col-span-3">
            <div className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 overflow-hidden flex flex-col h-[calc(100vh-280px)]">
              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-6 space-y-4">
                <AnimatePresence mode="popLayout">
                  {messages.map((message, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0 }}
                      className={`${
                        message.type === 'system'
                          ? 'flex justify-center'
                          : message.username === currentUser?.username
                          ? 'flex justify-end'
                          : 'flex justify-start'
                      }`}
                    >
                      {message.type === 'system' ? (
                        <div className="text-xs text-zinc-400 dark:text-zinc-500 bg-zinc-100 dark:bg-zinc-800/50 px-3 py-1 rounded-full">
                          {message.message}
                        </div>
                      ) : (
                        <div className={`max-w-[70%] ${
                          message.username === currentUser?.username ? 'items-end' : 'items-start'
                        } flex flex-col gap-1`}>
                          <div className="flex items-center gap-2">
                            <span className="text-xs font-semibold text-zinc-600 dark:text-zinc-400">
                              {message.username}
                            </span>
                            <span className="text-xs text-zinc-400 dark:text-zinc-500">
                              {formatTimestamp(message.timestamp)}
                            </span>
                          </div>
                          <div className={`px-4 py-2 rounded-2xl ${
                            message.username === currentUser?.username
                              ? 'bg-primary text-white'
                              : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-white'
                          }`}>
                            <p className="text-sm break-words">{message.message}</p>
                          </div>
                        </div>
                      )}
                    </motion.div>
                  ))}
                </AnimatePresence>
                <div ref={messagesEndRef} />
              </div>

              {/* Message Input */}
              <div className="border-t border-zinc-200 dark:border-zinc-800 p-4">
                <form onSubmit={handleSendMessage} className="flex gap-2">
                  <input
                    type="text"
                    value={messageInput}
                    onChange={(e) => setMessageInput(e.target.value)}
                    placeholder="Type your message..."
                    disabled={!isConnected}
                    className="flex-1 px-4 py-3 bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-700 rounded-xl focus:ring-2 focus:ring-primary/50 focus:border-primary text-zinc-900 dark:text-white placeholder:text-zinc-400 outline-none transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  />
                  <button
                    type="submit"
                    disabled={!messageInput.trim() || !isConnected}
                    className="px-6 py-3 bg-primary hover:bg-primary/90 text-white rounded-xl transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 font-semibold"
                  >
                    <Send className="w-5 h-5" />
                    <span className="hidden sm:inline">Send</span>
                  </button>
                </form>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Online Users */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 p-6">
              <div className="flex items-center gap-2 mb-4">
                <Users className="w-5 h-5 text-primary" />
                <h3 className="font-bold text-zinc-900 dark:text-white">
                  Online Users ({onlineUsers.length})
                </h3>
              </div>
              <div className="space-y-2 max-h-[400px] overflow-y-auto">
                {onlineUsers.map((user, index) => (
                  <div
                    key={index}
                    className="flex items-center gap-3 p-2 rounded-lg hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors"
                  >
                    <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                      <User className="w-4 h-4 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-zinc-900 dark:text-white truncate">
                        {user}
                      </p>
                    </div>
                    <Circle className="w-2 h-2 fill-green-500 text-green-500 flex-shrink-0" />
                  </div>
                ))}
                {onlineUsers.length === 0 && (
                  <p className="text-sm text-zinc-500 dark:text-zinc-400 text-center py-4">
                    No users online
                  </p>
                )}
              </div>
            </div>

            {/* Chat Info */}
            <div className="bg-white dark:bg-[#191022] rounded-2xl border border-zinc-200 dark:border-zinc-800 p-6">
              <div className="flex items-center gap-2 mb-4">
                <MessageCircle className="w-5 h-5 text-primary" />
                <h3 className="font-bold text-zinc-900 dark:text-white">About</h3>
              </div>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                Welcome to the General Chat! This is a space for all manga fans to discuss their favorite series, share recommendations, and connect with fellow enthusiasts.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default GeneralChat;
