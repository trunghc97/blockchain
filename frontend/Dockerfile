# Build stage
FROM node:18-alpine as build
WORKDIR /app

# Copy package files and config files
COPY package*.json ./
COPY package-lock.json* ./
COPY angular.json ./
COPY tsconfig*.json ./

# Install dependencies with caching
RUN --mount=type=cache,target=/root/.npm \
    npm install

# Copy source code
COPY src ./src

# Build application
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=build /app/dist/blockchain-frontend /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80