# Use Node.js 18 Alpine
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Set environment variable
ENV PORT 3000
ENV NODE_ENV development

# Start the application in development mode
CMD ["npm", "run", "dev"] 