FROM node:22-alpine AS builder

WORKDIR /app

# Accept build args for Next.js public env vars
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Copy package files first for better caching
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production=false

# Copy source code
COPY . .

# Create empty .env.local (env vars will come from docker-compose)
RUN touch .env.local

# Create public folder if not exists
RUN mkdir -p public

# Build the application
RUN npm run build

# Production stage
FROM node:22-alpine AS production

WORKDIR /app

# Copy built assets from builder
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/public ./public
COPY --from=builder /app/suppress-deprecation.js ./suppress-deprecation.js

# Set environment
ENV NODE_ENV=production

EXPOSE 3000

CMD ["npm", "run", "start"]
