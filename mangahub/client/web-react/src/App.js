import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import Header from './components/Header';
import Footer from './components/Footer';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Browse from './pages/Browse';
import Library from './pages/Library';
import MangaDetail from './pages/MangaDetail';
import ChatHub from './pages/ChatHub';
import GeneralChat from './pages/GeneralChat';
import authService from './services/authService';
import './App.css';

// Protected Route wrapper
const ProtectedRoute = ({ children }) => {
  return authService.isAuthenticated() ? children : <Navigate to="/login" />;
};

// Layout wrapper to use useLocation
const AppLayout = () => {
  const location = useLocation();
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register';

  return (
    <div className="App min-h-screen flex flex-col bg-gray-50">
      <Header />
      <main className="flex-grow">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          {/* Public routes - no login required */}
          <Route path="/browse" element={<Browse />} />
          <Route path="/manga/:id" element={<MangaDetail />} />
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
        </Routes>
      </main>
      {!isAuthPage && <Footer />}
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
