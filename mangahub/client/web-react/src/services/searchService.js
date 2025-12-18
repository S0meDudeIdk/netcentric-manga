/**
 * Unified Search Service
 * Automatically routes local manga search to either REST API or gRPC based on configuration
 * External API searches (MAL, MangaDex, MangaPlus) always use REST
 */

import grpcService from './grpcService';
import mangaService from './mangaService';
import serviceConfig from './serviceConfig';

const searchService = {
  /**
   * Search local manga database
   * Routes to gRPC or REST based on configuration
   * @param {string} query - Search query
   * @param {number} limit - Maximum number of results (default: 20)
   * @returns {Promise<Object>} Search results with manga list and total count
   */
  searchLocalManga: async (query, limit = 20) => {
    if (serviceConfig.useGrpcForSearch) {
      console.log('📡 Using gRPC for local manga search');
      return await grpcService.searchManga(query, limit);
    } else {
      console.log('🌐 Using REST for local manga search');
      return await mangaService.searchManga(query);
    }
  },

  /**
   * Search MyAnimeList (always uses REST API)
   * @param {string} query - Search query
   * @param {number} page - Page number (default: 1)
   * @param {number} limit - Results per page (default: 20)
   * @param {string} orderBy - Order field (optional)
   * @param {string} sort - Sort direction (optional)
   * @returns {Promise<Object>} MAL search results
   */
  searchMAL: async (query, page = 1, limit = 20, orderBy = '', sort = '') => {
    console.log('🌐 Searching MyAnimeList (REST API)');
    return await mangaService.searchMAL(query, page, limit, orderBy, sort);
  },

  /**
   * Search MangaDex (always uses REST API)
   * @param {string} title - Manga title to search
   * @param {number} limit - Maximum number of results (default: 10)
   * @returns {Promise<Object>} MangaDex search results
   */
  searchMangaDex: async (title, limit = 10) => {
    console.log('🌐 Searching MangaDex (REST API)');
    return await mangaService.searchMangaDex(title, limit);
  },

  /**
   * Search MangaPlus (always uses REST API)
   * @param {string} query - Search query
   * @returns {Promise<Object>} MangaPlus search results
   */
  searchMangaPlus: async (query) => {
    console.log('🌐 Searching MangaPlus (REST API)');
    return await mangaService.searchMangaPlus(query);
  }
};

export default searchService;
