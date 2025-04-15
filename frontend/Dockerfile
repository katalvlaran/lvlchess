#
# Dockerfile for building and serving the React-based frontend
# 1) Build stage with node:16-alpine
# 2) Production stage with nginx:alpine
#

# 1) Build Stage
FROM node:16-alpine as build

# We define /app as the working directory
WORKDIR /app

# Copy package.json and package-lock.json (if present) first to install deps
COPY package*.json ./

# Install dependencies
# (npm ci is typical for reproducible builds if you have package-lock.json)
RUN npm install

# Copy the rest of the frontend source code
COPY . ./

# Build a production-ready static bundle of the React app
RUN npm run build


# 2) Final Stage
FROM nginx:alpine

# Copy the build artifacts from stage 'build' to the NGINX html folder
COPY --from=build /app/build /usr/share/nginx/html

# Expose port 80 so we can map it (see docker-compose)
EXPOSE 80

# By default, run the standard nginx command (daemon off)
CMD ["nginx", "-g", "daemon off;"]
