
# Multi-stage build for Vue 3 Frontend
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /build

# Copy package files
COPY package.json pnpm-lock.yaml ./

# install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Install dependencies (use regular install if lock file doesn't exist)
RUN if [ -f pnpm-lock.yaml ]; then \
        pnpm install --frozen-lockfile; \
    else \
        pnpm install; \
    fi

# Copy source code
COPY . .

# Build the application
RUN pnpm build-only

# Production stage - serve with Node (simple HTTP server)
FROM node:22-alpine

# Install a simple HTTP server
RUN npm install -g http-server

# Copy built application from builder
COPY --from=builder /build/dist /app

# Change ownership to node user (already exists in Alpine)
RUN chown -R node:node /app

# Switch to node user
USER node

WORKDIR /app

# Expose port
EXPOSE 3000

# Start simple HTTP server
CMD ["http-server", "-p", "3000", "--gzip", "-c-1"]
