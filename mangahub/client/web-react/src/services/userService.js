import axios from 'axios';
import authService from './authService';

const getBaseUrl = () => {
  const port = '8080';
  if (window.location.hostname === 'localhost') {
    return `http://localhost:${port}/api/v1/users`;
  }
  return `${process.env.REACT_APP_BACKEND_URL}/api/v1/users`;
};

const BASE_URL = getBaseUrl();

const getAuthHeaders = () => {
  const token = authService.getToken();
  if (!token) {
    throw new Error('No authentication token found. Please log in.');
  }
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };
};

const userService = {
  getProfile: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/profile`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching profile:', error);
      throw error.response?.data || error;
    }
  },

  getLibrary: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/library`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      // Don't log 401 errors (unauthenticated users)
      if (error.response?.status !== 401) {
        console.error('Error fetching library:', error);
      }
      throw error.response?.data || error;
    }
  },

  addToLibrary: async (mangaId, status = 'reading') => {
    try {
      const response = await axios.post(`${BASE_URL}/library`, {
        manga_id: mangaId,
        status: status
      }, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error adding to library:', error);
      throw error.response?.data || error;
    }
  },

  updateProgress: async (mangaId, currentChapter, status) => {
    try {
      const response = await axios.put(`${BASE_URL}/progress`, {
        manga_id: mangaId,
        current_chapter: currentChapter,
        status: status
      }, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error updating progress:', error);
      throw error.response?.data || error;
    }
  },

  getLibraryStats: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/library/stats`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching library stats:', error);
      throw error.response?.data || error;
    }
  },

  getRecommendations: async (limit = 10) => {
    try {
      const response = await axios.get(`${BASE_URL}/recommendations?limit=${limit}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching recommendations:', error);
      throw error.response?.data || error;
    }
  },

  removeFromLibrary: async (mangaId) => {
    try {
      const response = await axios.delete(`${BASE_URL}/library/${mangaId}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error removing from library:', error);
      throw error.response?.data || error;
    }
  }
};

export default userService;
