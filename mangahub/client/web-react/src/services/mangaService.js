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
const API_BASE = BASE_URL.replace('/manga', ''); // Remove /manga for user routes

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
  },

  // Generate chapter list with volume information
  getChapterList: (manga) => {
    // For now, we'll generate chapters based on total_chapters
    // In the future, this could fetch from an API endpoint if chapter data is available
    const chapters = [];
    const totalChapters = manga.total_chapters || 0;
    
    if (totalChapters === 0) return chapters;

    // If volumes are available, organize chapters by volume
    // Assuming roughly 18 chapters per volume as average
    const chaptersPerVolume = 18;
    
    for (let i = 1; i <= totalChapters; i++) {
      const volume = Math.ceil(i / chaptersPerVolume);
      chapters.push({
        number: i,
        volume: volume,
        title: `Chapter ${i}`,
        publishDate: null, // Can be populated if data is available
        read: false
      });
    }
    
    return chapters;
  },

  // Get chapter list from backend (MangaDex/MangaPlus)
  getChapters: async (mangaID, language = ['en'], limit = 100, offset = 0) => {
    try {
      const params = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString()
      });
      
      // Add multiple language parameters
      language.forEach(lang => {
        params.append('language', lang);
      });
      
      const response = await axios.get(`${BASE_URL}/${mangaID}/chapters?${params.toString()}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching chapters:', error);
      throw error.response?.data || error;
    }
  },

  // Get chapter pages
  getChapterPages: async (chapterID, source = 'mangadex') => {
    try {
      const response = await axios.get(`${BASE_URL}/chapters/${chapterID}/pages?source=${source}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching chapter pages:', error);
      throw error.response?.data || error;
    }
  },

  // Get manga ratings
  getMangaRatings: async (mangaID) => {
    try {
      const response = await axios.get(`${BASE_URL}/${mangaID}/ratings`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching manga ratings:', error);
      throw error.response?.data || error;
    }
  },

  // Rate a manga
  rateManga: async (mangaID, rating) => {
    try {
      const response = await axios.post(`${API_BASE}/users/manga/${mangaID}/rating`, 
        { rating },
        { headers: getAuthHeaders() }
      );
      return response.data;
    } catch (error) {
      console.error('Error rating manga:', error);
      throw error.response?.data || error;
    }
  },

  // Delete user's rating
  deleteRating: async (mangaID) => {
    try {
      const response = await axios.delete(`${API_BASE}/users/manga/${mangaID}/rating`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error deleting rating:', error);
      throw error.response?.data || error;
    }
  }
};

export default mangaService;
