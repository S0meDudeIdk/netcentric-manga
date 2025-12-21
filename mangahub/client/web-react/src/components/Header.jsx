import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { Book, Search, Library, Home, LogIn, UserPlus, User, LogOut, Menu, X, MessageCircle } from 'lucide-react';
import authService from '../services/authService';

const Header = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [showMobileMenu, setShowMobileMenu] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const checkAuth = () => {
      setIsAuthenticated(authService.isAuthenticated());
      setUser(authService.getUser());
    };

    checkAuth();
    const interval = setInterval(checkAuth, 1000);
    return () => clearInterval(interval);
  }, []);

  // Close menus on route change
  useEffect(() => {
    setShowUserMenu(false);
    setShowMobileMenu(false);
  }, [location]);

  const handleLogout = () => {
    authService.logout();
    setIsAuthenticated(false);
    setUser(null);
    navigate('/login');
  };

  const NavLink = ({ to, icon: Icon, label }) => {
    const isActive = location.pathname === to;
    return (
      <Link
        to={to}
        className={`flex items-center gap-2 transition-colors duration-200 ${isActive
          ? 'text-primary font-semibold'
          : 'text-zinc-600 dark:text-zinc-400 hover:text-primary dark:hover:text-primary'
          }`}
      >
        <Icon className="w-5 h-5" />
        <span>{label}</span>
      </Link>
    );
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b border-zinc-200 dark:border-zinc-800 bg-white/80 dark:bg-[#191022]/80 backdrop-blur-md">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          {/* Logo */}
          <Link to="/" className="flex items-center gap-3">
            <svg fill="none" height="32" viewBox="0 0 32 32" width="32" xmlns="http://www.w3.org/2000/svg">
              <path className="stroke-zinc-900 dark:stroke-white" d="M16 31.5C24.5604 31.5 31.5 24.5604 31.5 16C31.5 7.43959 24.5604 0.5 16 0.5C7.43959 0.5 0.5 7.43959 0.5 16C0.5 24.5604 7.43959 31.5 16 31.5Z" strokeWidth="1"></path>
              <path className="stroke-zinc-900 dark:stroke-white" d="M16 31.5C24.5604 31.5 31.5 24.5604 31.5 16C31.5 7.43959 24.5604 0.5 16 0.5C7.43959 0.5 0.5 7.43959 0.5 16C0.5 24.5604 7.43959 31.5 16 31.5Z" strokeWidth="1"></path>
              <path className="stroke-primary fill-primary/20" d="M11 11.25H21V20.75L16 26L11 20.75V11.25Z" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"></path>
            </svg>
            <span className="text-xl font-bold text-zinc-900 dark:text-white">MangaHub</span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center gap-8">
            <NavLink to="/" icon={Home} label="Home" />
            <NavLink to="/browse" icon={Book} label="Browse" />
            {isAuthenticated && (
              <>
                <NavLink to="/library" icon={Library} label="Library" />
                <NavLink to="/chat" icon={MessageCircle} label="General Chat" />
              </>
            )}
          </nav>

          {/* Search Bar */}
          <div className="hidden md:flex flex-1 max-w-md mx-8">
            <form onSubmit={(e) => { e.preventDefault(); if (searchQuery.trim()) navigate(`/browse?search=${searchQuery}`); }} className="w-full">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
                <input
                  type="text"
                  placeholder="Search manga..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-700 rounded-lg focus:ring-2 focus:ring-primary/50 focus:border-primary text-zinc-900 dark:text-white placeholder:text-zinc-400 outline-none transition-all text-sm"
                />
              </div>
            </form>
          </div>

          {/* Auth Buttons / User Menu */}
          <div className="hidden md:flex items-center gap-4">
            {isAuthenticated ? (
              <div className="relative">
                <button
                  onClick={() => setShowUserMenu(!showUserMenu)}
                  className="flex items-center gap-2 pl-2 pr-4 py-1.5 rounded-full border border-zinc-200 dark:border-zinc-700 hover:border-primary/50 dark:hover:border-primary/50 transition-colors bg-zinc-50 dark:bg-zinc-800/50"
                >
                  <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                    <User className="w-4 h-4 text-primary" />
                  </div>
                  <span className="text-sm font-medium text-zinc-700 dark:text-zinc-200 max-w-[100px] truncate">
                    {user?.username || 'User'}
                  </span>
                </button>

                {showUserMenu && (
                  <div className="absolute right-0 mt-2 w-56 bg-white dark:bg-[#211c27] rounded-xl shadow-xl border border-zinc-200 dark:border-zinc-700 py-2 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
                    <div className="px-4 py-2 border-b border-zinc-100 dark:border-zinc-700 mb-2">
                      <p className="text-sm text-zinc-500 dark:text-zinc-400">Signed in as</p>
                      <p className="text-sm font-semibold text-zinc-900 dark:text-white truncate">{user?.username}</p>
                    </div>

                    <Link
                      to="/library"
                      className="flex items-center gap-2 px-4 py-2 text-sm text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors"
                    >
                      <Library className="w-4 h-4" />
                      <span>My Library</span>
                    </Link>

                    <Link
                      to="/profile"
                      className="flex items-center gap-2 px-4 py-2 text-sm text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 transition-colors border-t border-zinc-100 dark:border-zinc-700 mt-2 pt-2"
                    >
                      <User className="w-4 h-4" />
                      <span>Profile Settings</span>
                    </Link>

                    <button
                      onClick={handleLogout}
                      className="flex items-center gap-2 px-4 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/10 w-full text-left transition-colors"
                    >
                      <LogOut className="w-4 h-4" />
                      <span>Logout</span>
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <div className="flex items-center gap-3">
                <Link
                  to="/login"
                  className="text-sm font-medium text-zinc-600 dark:text-zinc-300 hover:text-primary dark:hover:text-white transition-colors"
                >
                  Log In
                </Link>
                <Link
                  to="/register"
                  className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary/90 text-white text-sm font-medium rounded-lg transition-all shadow-lg shadow-primary/20 hover:shadow-primary/30"
                >
                  <UserPlus className="w-4 h-4" />
                  <span>Sign Up</span>
                </Link>
              </div>
            )}
          </div>

          {/* Mobile Menu Button */}
          <button
            onClick={() => setShowMobileMenu(!showMobileMenu)}
            className="md:hidden p-2 text-zinc-600 dark:text-zinc-400 hover:text-primary dark:hover:text-white transition-colors"
          >
            {showMobileMenu ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Menu */}
        {showMobileMenu && (
          <div className="md:hidden py-4 border-t border-zinc-200 dark:border-zinc-800 animate-in slide-in-from-top-2 duration-200">
            {/* Mobile Search */}
            <div className="px-4 mb-4">
              <form onSubmit={(e) => { e.preventDefault(); if (searchQuery.trim()) navigate(`/browse?search=${searchQuery}`); }}>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-400" />
                  <input
                    type="text"
                    placeholder="Search manga..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-full pl-10 pr-4 py-2 bg-zinc-50 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-700 rounded-lg focus:ring-2 focus:ring-primary/50 focus:border-primary text-zinc-900 dark:text-white placeholder:text-zinc-400 outline-none transition-all text-sm"
                  />
                </div>
              </form>
            </div>
            <nav className="flex flex-col gap-1">
              <Link to="/" className="flex items-center gap-3 px-4 py-3 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 rounded-lg">
                <Home className="w-5 h-5 text-zinc-400" />
                <span>Home</span>
              </Link>
              <Link to="/browse" className="flex items-center gap-3 px-4 py-3 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 rounded-lg">
                <Book className="w-5 h-5 text-zinc-400" />
                <span>Browse</span>
              </Link>
              {isAuthenticated && (
                <>
                  <Link to="/library" className="flex items-center gap-3 px-4 py-3 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 rounded-lg">
                    <Library className="w-5 h-5 text-zinc-400" />
                    <span>Library</span>
                  </Link>
                  <Link to="/chat" className="flex items-center gap-3 px-4 py-3 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 rounded-lg">
                    <MessageCircle className="w-5 h-5 text-zinc-400" />
                    <span>General Chat</span>
                  </Link>
                </>
              )}

              <div className="border-t border-zinc-200 dark:border-zinc-800 mt-2 pt-2 px-2">
                {isAuthenticated ? (
                  <>
                    <div className="px-2 py-2 text-sm text-zinc-500 dark:text-zinc-400">
                      Signed in as <span className="font-semibold text-zinc-900 dark:text-white">{user?.username}</span>
                    </div>
                    <button
                      onClick={handleLogout}
                      className="flex items-center gap-3 px-4 py-3 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/10 rounded-lg w-full transition-colors"
                    >
                      <LogOut className="w-5 h-5" />
                      <span>Logout</span>
                    </button>
                  </>
                ) : (
                  <div className="flex flex-col gap-2 mt-2">
                    <Link
                      to="/login"
                      className="flex items-center justify-center gap-2 px-4 py-3 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-50 dark:hover:bg-zinc-800 rounded-lg font-medium border border-zinc-200 dark:border-zinc-700"
                    >
                      <LogIn className="w-5 h-5" />
                      <span>Log In</span>
                    </Link>
                    <Link
                      to="/register"
                      className="flex items-center justify-center gap-2 px-4 py-3 bg-primary text-white rounded-lg hover:bg-primary/90 font-medium"
                    >
                      <UserPlus className="w-5 h-5" />
                      <span>create Account</span>
                    </Link>
                  </div>
                )}
              </div>
            </nav>
          </div>
        )}
      </div>
    </header>
  );
};

export default Header;
