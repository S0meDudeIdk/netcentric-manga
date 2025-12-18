import React, { useState, useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import Header from './components/Header';
import Footer from './components/Footer';
import NotificationContainer from './components/NotificationToast';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Browse from './pages/Browse';
import Library from './pages/Library';
import MangaDetail from './pages/MangaDetail';
import ChapterReader from './pages/ChapterReader';
import ChatHub from './pages/ChatHub';
import GeneralChat from './pages/GeneralChat';
import GRPCTestPage from './pages/GRPCTestPage';
import authService from './services/authService';
import websocketService from './services/websocketService';
import './App.css';

// Protected Route wrapper
const ProtectedRoute = ({ children }) => {
  return authService.isAuthenticated() ? children : <Navigate to="/login" />;
};

// Layout wrapper to use useLocation
const AppLayout = () => {
  const location = useLocation();
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register';
  const isReaderPage = location.pathname.startsWith('/read');
  const [notifications, setNotifications] = useState([]);
  const wsInitializedRef = useRef(false);

  useEffect(() => {
    // Only subscribe to notifications if user is authenticated
    if (!authService.isAuthenticated() || wsInitializedRef.current) {
      return;
    }

    console.log('📢 Subscribing to global notifications...');
    wsInitializedRef.current = true;

    const token = authService.getToken();
    const notificationRoomId = 'global-notifications';

    // Connect to a special "notifications" room
    websocketService.connect(notificationRoomId, token, {
      onMessage: (data) => {
        if (data.type === 'notification') {
          console.log('🔔 Received notification:', data);
          addNotification(data);
        }
      },
      onError: (error) => {
        console.error('Notification WebSocket error:', error);
      }
    });

    // Cleanup on logout or unmount
    return () => {
      console.log('🔌 Disconnecting from notifications');
      websocketService.disconnect(notificationRoomId);
      wsInitializedRef.current = false;
    };
  }, [location.pathname]); // Re-subscribe when navigating

  const addNotification = (notificationData) => {
    const notification = {
      id: Date.now() + Math.random(),
      message: notificationData.message,
      manga_id: notificationData.manga_id,
      timestamp: notificationData.timestamp,
    };

    setNotifications((prev) => [...prev, notification]);
  };

  const removeNotification = (id) => {
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  };

  return (
    <div className="App min-h-screen flex flex-col bg-gray-50">
      {!isReaderPage && <Header />}
      <NotificationContainer 
        notifications={notifications}
        removeNotification={removeNotification}
      />
      <main className="flex-grow">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          {/* Public routes - no login required */}
          <Route path="/browse" element={<Browse />} />
          <Route path="/manga/:id" element={<MangaDetail />} />
          <Route path="/read/:mangaId" element={<ChapterReader />} />
          {/* Protected routes - login required */}
          <Route path="/library" element={
            <ProtectedRoute>
              <Library />
            </ProtectedRoute>
          } />
          <Route path="/chathub/:mangaId" element={
            <ProtectedRoute>
              <ChatHub />
            </ProtectedRoute>
          } />
          <Route path="/chat" element={
            <ProtectedRoute>
              <GeneralChat />
            </ProtectedRoute>
          } />
          <Route path="/grpc-test" element={
            <ProtectedRoute>
              <GRPCTestPage />
            </ProtectedRoute>
          } />
        </Routes>
      </main>
      {!isAuthPage && !isReaderPage && <Footer />}
    </div>
  );
};

function App() {
  return (
    <Router>
      <AppLayout />
    </Router>
  );
}

export default App;
