import axios from 'axios';

const getBaseUrl = () => {
  const port = '8080';
  // Use environment variable if set, otherwise use current hostname
  if (process.env.REACT_APP_BACKEND_URL) {
    return `${process.env.REACT_APP_BACKEND_URL}/api/v1/auth`;
  }
  // Use the same hostname as the frontend (works for localhost and LAN access)
  return `http://${window.location.hostname}:${port}/api/v1/auth`;
};

const BASE_URL = getBaseUrl();

const authService = {
  register: async (userData) => {
    try {
      const response = await axios.post(`${BASE_URL}/register`, userData);
      if (response.data.token) {
        localStorage.setItem('token', response.data.token);
        localStorage.setItem('user', JSON.stringify(response.data.user));
      }
      return response.data;
    } catch (error) {
      console.error('Error registering user:', error);
      throw error.response?.data || error;
    }
  },

  login: async (credentials) => {
    try {
      const response = await axios.post(`${BASE_URL}/login`, {
        email: credentials.email,
        password: credentials.password
      });
      
      if (response.data.token) {
        localStorage.setItem('token', response.data.token);
        localStorage.setItem('user', JSON.stringify(response.data.user));
      }
      
      return response.data;
    } catch (error) {
      console.error('Error logging in:', error);
      throw error.response?.data || error;
    }
  },

  logout: async () => {
    const token = localStorage.getItem('token');
    
    // Call backend logout endpoint to disconnect TCP
    if (token) {
      try {
        await axios.post(`${BASE_URL}/logout`, {}, {
          headers: { Authorization: `Bearer ${token}` }
        });
        console.log('✅ Logout successful - TCP connection closed');
      } catch (error) {
        console.error('Error during logout:', error);
        // Continue with local logout even if backend call fails
      }
    }
    
    // Clear local storage
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  },

  isTokenExpired: (token) => {
    if (!token) return true;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const expirationTime = payload.exp * 1000; // Convert to milliseconds
      return Date.now() >= expirationTime;
    } catch (error) {
      console.error('Error parsing token:', error);
      return true;
    }
  },

  isAuthenticated: () => {
    const token = localStorage.getItem('token');
    if (!token) return false;
    
    // Check if token is expired
    if (authService.isTokenExpired(token)) {
      console.warn('Token expired, logging out...');
      authService.logout();
      return false;
    }
    
    return true;
  },

  getToken: () => {
    const token = localStorage.getItem('token');
    
    // Validate token before returning
    if (token && authService.isTokenExpired(token)) {
      console.warn('Token expired, clearing authentication...');
      authService.logout();
      return null;
    }
    
    return token;
  },

  getUser: () => {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  }
};

export default authService;
