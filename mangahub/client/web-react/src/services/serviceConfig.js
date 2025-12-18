/**
 * Service Configuration
 * Toggle between REST API and gRPC for different features
 * 
 * Set environment variable REACT_APP_USE_GRPC=true to enable gRPC globally
 * Or configure individual features below
 */

const USE_GRPC_GLOBAL = process.env.REACT_APP_USE_GRPC === 'true';

export const serviceConfig = {
  // Library Management
  useGrpcForLibrary: USE_GRPC_GLOBAL || process.env.REACT_APP_USE_GRPC_LIBRARY === 'true',
  
  // Progress Updates
  useGrpcForProgress: USE_GRPC_GLOBAL || process.env.REACT_APP_USE_GRPC_PROGRESS === 'true',
  
  // Rating System
  useGrpcForRating: USE_GRPC_GLOBAL || process.env.REACT_APP_USE_GRPC_RATING === 'true',
  
  // Search (local manga only)
  useGrpcForSearch: USE_GRPC_GLOBAL || process.env.REACT_APP_USE_GRPC_SEARCH === 'true',
};

// Helper function to check if gRPC should be used for a feature
export const shouldUseGrpc = (feature) => {
  return serviceConfig[`useGrpcFor${feature}`] === true;
};

export default serviceConfig;
