import axios from 'axios';
import authService from './authService';

const getBaseUrl = () => {
  const port = '8080';
  if (window.location.hostname === 'localhost') {
    return `http://localhost:${port}/api/v1/manga`;
  }
  return `${process.env.REACT_APP_BACKEND_URL}/api/v1/manga`;
};

const BASE_URL = getBaseUrl();

// Add auth header to requests (optional - only if user is logged in)
const getAuthHeaders = () => {
  const token = authService.getToken();
  if (token) {
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    };
  }
  return {
    'Content-Type': 'application/json'
  };
};

const mangaService = {
  searchManga: async (query) => {
    try {
      const response = await axios.get(`${BASE_URL}?query=${encodeURIComponent(query)}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error searching manga:', error);
      throw error.response?.data || error;
    }
  },

  getPopularManga: async (limit = 20) => {
    try {
      const response = await axios.get(`${BASE_URL}/popular?limit=${limit}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching popular manga:', error);
      throw error.response?.data || error;
    }
  },

  getMangaById: async (id) => {
    try {
      const response = await axios.get(`${BASE_URL}/${id}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching manga:', error);
      throw error.response?.data || error;
    }
  },

  getGenres: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/genres`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching genres:', error);
      throw error.response?.data || error;
    }
  },

  getMangaStats: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/stats`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching manga stats:', error);
      throw error.response?.data || error;
    }
  },

  // MyAnimeList API methods (via Jikan)
  searchMAL: async (query, page = 1, limit = 20, orderBy = '', sort = '') => {
    try {
      const params = { q: query, page, limit };
      if (orderBy) params.order_by = orderBy;
      if (sort) params.sort = sort;
      
      const response = await axios.get(`${BASE_URL}/mal/search`, {
        params,
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error searching MyAnimeList:', error);
      throw error.response?.data || error;
    }
  },

  getTopMAL: async (page = 1, limit = 20, orderBy = '', sort = '') => {
    try {
      const params = { page, limit };
      if (orderBy) params.order_by = orderBy;
      if (sort) params.sort = sort;
      
      const response = await axios.get(`${BASE_URL}/mal/top`, {
        params,
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching top manga from MyAnimeList:', error);
      throw error.response?.data || error;
    }
  },

  getMALMangaById: async (malId) => {
    try {
      const response = await axios.get(`${BASE_URL}/mal/${malId}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching manga from MyAnimeList:', error);
      throw error.response?.data || error;
    }
  },

  getMALRecommendations: async (malId) => {
    try {
      const response = await axios.get(`${BASE_URL}/mal/${malId}/recommendations`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching recommendations from MyAnimeList:', error);
      throw error.response?.data || error;
    }
  }
};

export default mangaService;
