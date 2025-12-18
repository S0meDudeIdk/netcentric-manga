/**
 * Unified Library Service
 * Automatically routes requests to either REST API or gRPC based on configuration
 */

import grpcService from './grpcService';
import userService from './userService';
import serviceConfig from './serviceConfig';

const libraryService = {
  /**
   * Get user's library
   * Routes to gRPC or REST based on configuration
   */
  getLibrary: async () => {
    if (serviceConfig.useGrpcForLibrary) {
      console.log('📡 Using gRPC for getLibrary');
      return await grpcService.getLibrary();
    } else {
      console.log('🌐 Using REST for getLibrary');
      return await userService.getLibrary();
    }
  },

  /**
   * Add manga to library
   * Routes to gRPC or REST based on configuration
   */
  addToLibrary: async (mangaId, status = 'plan_to_read') => {
    if (serviceConfig.useGrpcForLibrary) {
      console.log('📡 Using gRPC for addToLibrary');
      return await grpcService.addToLibrary(mangaId, status);
    } else {
      console.log('🌐 Using REST for addToLibrary');
      return await userService.addToLibrary(mangaId, status);
    }
  },

  /**
   * Remove manga from library
   * Routes to gRPC or REST based on configuration
   */
  removeFromLibrary: async (mangaId) => {
    if (serviceConfig.useGrpcForLibrary) {
      console.log('📡 Using gRPC for removeFromLibrary');
      return await grpcService.removeFromLibrary(mangaId);
    } else {
      console.log('🌐 Using REST for removeFromLibrary');
      return await userService.removeFromLibrary(mangaId);
    }
  },

  /**
   * Update reading progress
   * Routes to gRPC or REST based on configuration
   */
  updateProgress: async (mangaId, currentChapter, status) => {
    if (serviceConfig.useGrpcForProgress) {
      console.log('📡 Using gRPC for updateProgress (with TCP broadcast)');
      return await grpcService.updateProgress(mangaId, currentChapter, status);
    } else {
      console.log('🌐 Using REST for updateProgress');
      return await userService.updateProgress(mangaId, currentChapter, status);
    }
  },

  /**
   * Get library statistics
   * Routes to gRPC or REST based on configuration
   */
  getLibraryStats: async () => {
    if (serviceConfig.useGrpcForLibrary) {
      console.log('📡 Using gRPC for getLibraryStats');
      return await grpcService.getLibraryStats();
    } else {
      console.log('🌐 Using REST for getLibraryStats');
      return await userService.getLibraryStats();
    }
  }
};

export default libraryService;
