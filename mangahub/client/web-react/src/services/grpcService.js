import axios from 'axios';
import authService from './authService';

const getBaseUrl = () => {
  const port = '8080';
  // Use environment variable if set, otherwise use current hostname
  if (process.env.REACT_APP_BACKEND_URL) {
    return `${process.env.REACT_APP_BACKEND_URL}/api/v1/grpc`;
  }
  // Use the same hostname as the frontend (works for localhost and LAN access)
  return `http://${window.location.hostname}:${port}/api/v1/grpc`;
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
   * Get User Library via gRPC
   * Retrieves user's manga library organized by status
   * @returns {Promise<Object>} Library data with categorized manga
   */
  getLibrary: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/library`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error getting library via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Add to Library via gRPC
   * Adds a manga to user's library
   * @param {string} mangaId - The manga ID
   * @param {string} status - Reading status (reading, plan_to_read, etc.)
   * @returns {Promise<Object>} Success message
   */
  addToLibrary: async (mangaId, status = 'plan_to_read') => {
    try {
      const response = await axios.post(`${BASE_URL}/library`, {
        manga_id: mangaId,
        status: status
      }, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error adding to library via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Remove from Library via gRPC
   * Removes a manga from user's library
   * @param {string} mangaId - The manga ID
   * @returns {Promise<Object>} Success message
   */
  removeFromLibrary: async (mangaId) => {
    try {
      const response = await axios.delete(`${BASE_URL}/library/${mangaId}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error removing from library via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Get Library Stats via gRPC
   * Retrieves user's library statistics
   * @returns {Promise<Object>} Statistics about user's library
   */
  getLibraryStats: async () => {
    try {
      const response = await axios.get(`${BASE_URL}/library/stats`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error getting library stats via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Rate Manga via gRPC
   * Submits or updates a rating for a manga
   * @param {string} mangaId - The manga ID
   * @param {number} rating - Rating value (1-10)
   * @returns {Promise<Object>} Updated rating statistics
   */
  rateManga: async (mangaId, rating) => {
    try {
      const response = await axios.post(`${BASE_URL}/rating`, {
        manga_id: mangaId,
        rating: rating
      }, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error rating manga via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Get Manga Ratings via gRPC
   * Retrieves rating statistics for a manga
   * @param {string} mangaId - The manga ID
   * @returns {Promise<Object>} Rating statistics including user's rating
   */
  getMangaRatings: async (mangaId) => {
    try {
      const response = await axios.get(`${BASE_URL}/rating/${mangaId}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error getting manga ratings via gRPC:', error);
      throw error.response?.data || error;
    }
  },

  /**
   * Delete Rating via gRPC
   * Removes user's rating for a manga
   * @param {string} mangaId - The manga ID
   * @returns {Promise<Object>} Success message
   */
  deleteRating: async (mangaId) => {
    try {
      const response = await axios.delete(`${BASE_URL}/rating/${mangaId}`, {
        headers: getAuthHeaders()
      });
      return response.data;
    } catch (error) {
      console.error('Error deleting rating via gRPC:', error);
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
