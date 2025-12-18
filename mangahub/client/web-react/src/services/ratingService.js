/**
 * Unified Rating Service
 * Automatically routes requests to either REST API or gRPC based on configuration
 */

import grpcService from './grpcService';
import mangaService from './mangaService';
import serviceConfig from './serviceConfig';

const ratingService = {
  /**
   * Rate a manga
   * Routes to gRPC or REST based on configuration
   */
  rateManga: async (mangaId, rating) => {
    if (serviceConfig.useGrpcForRating) {
      console.log('📡 Using gRPC for rateManga');
      return await grpcService.rateManga(mangaId, rating);
    } else {
      console.log('🌐 Using REST for rateManga');
      return await mangaService.rateManga(mangaId, rating);
    }
  },

  /**
   * Get manga ratings
   * Routes to gRPC or REST based on configuration
   */
  getMangaRatings: async (mangaId) => {
    if (serviceConfig.useGrpcForRating) {
      console.log('📡 Using gRPC for getMangaRatings');
      return await grpcService.getMangaRatings(mangaId);
    } else {
      console.log('🌐 Using REST for getMangaRatings');
      return await mangaService.getMangaRatings(mangaId);
    }
  },

  /**
   * Delete a rating
   * Routes to gRPC or REST based on configuration
   */
  deleteRating: async (mangaId) => {
    if (serviceConfig.useGrpcForRating) {
      console.log('📡 Using gRPC for deleteRating');
      return await grpcService.deleteRating(mangaId);
    } else {
      console.log('🌐 Using REST for deleteRating');
      return await mangaService.deleteRating(mangaId);
    }
  }
};

export default ratingService;
