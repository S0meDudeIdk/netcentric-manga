# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build arguments for environment variables
ARG REACT_APP_BACKEND_URL
ARG REACT_APP_USE_GRPC_LIBRARY
ARG REACT_APP_USE_GRPC_PROGRESS
ARG REACT_APP_USE_GRPC_RATING
ARG REACT_APP_USE_GRPC_SEARCH

# Set environment variables for build
ENV REACT_APP_BACKEND_URL=${REACT_APP_BACKEND_URL}
ENV REACT_APP_USE_GRPC_LIBRARY=${REACT_APP_USE_GRPC_LIBRARY}
ENV REACT_APP_USE_GRPC_PROGRESS=${REACT_APP_USE_GRPC_PROGRESS}
ENV REACT_APP_USE_GRPC_RATING=${REACT_APP_USE_GRPC_RATING}
ENV REACT_APP_USE_GRPC_SEARCH=${REACT_APP_USE_GRPC_SEARCH}

# Build the React app
RUN npm run build

# Runtime stage - using nginx to serve
FROM nginx:alpine

# Copy built app from builder
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
