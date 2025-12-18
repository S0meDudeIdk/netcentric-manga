import axios from 'axios';
import authService from './authService';

const getBaseUrl = () => {
  const port = '8080';
  if (window.location.hostname === 'localhost') {
    return `http://localhost:${port}/api/v1/grpc`;
  }
  return `${process.env.REACT_APP_BACKEND_URL}/api/v1/grpc`;
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

/**
 * gRPC Service - Provides access to gRPC-backed endpoints
 * Implements UC-014, UC-015, and UC-016
 */
const grpcService = {
  /**
   * UC-014: Retrieve Manga via gRPC
   * Fetches manga data through gRPC interface
   * @param {string} mangaId - The manga ID to retrieve
   * @returns {Promise<Object>} Manga data
   */
  getManga: async (mangaId) => {
    try {
      const response = await axios.get(`${BASE_URL}/manga/${mangaId}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error fetching manga via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * UC-015: Search Manga via gRPC
   * Searches for manga using gRPC interface
   * @param {string} query - Search query
   * @param {number} limit - Maximum number of results (default: 20)
   * @returns {Promise<Object>} Search results with manga list and total count
   */
  searchManga: async (query, limit = 20) => {
    try {
      const response = await axios.get(`${BASE_URL}/manga/search`, {
        params: { q: query, limit },
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error searching manga via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * UC-016: Update Progress via gRPC
   * Updates user reading progress through gRPC
   * This also triggers TCP broadcast for real-time sync
   * @param {string} mangaId - The manga ID
   * @param {number} currentChapter - Current chapter number
   * @param {string} status - Reading status (reading, completed, etc.)
   * @returns {Promise<Object>} Success message
   */
  updateProgress: async (mangaId, currentChapter, status) => {
    try {
      const response = await axios.put(`${BASE_URL}/progress/update`, {
        manga_id: mangaId,
        current_chapter: currentChapter,
        status: status
      }, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error updating progress via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Check if gRPC service is available
   * @returns {Promise<boolean>} True if available, false otherwise
   */
  isAvailable: async () => {
    try {
      // Try to search with empty query to check availability
      await grpcService.searchManga('', 1);
      return true;
    } catch (error) {
      if (error.error === 'gRPC service unavailable') {
        return false;
      }
      // If it's a different error, service is available
      return true;
    }
  }
};

export default grpcService;
