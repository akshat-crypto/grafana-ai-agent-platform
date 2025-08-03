#!/bin/bash

echo "🚀 Setting up Kubernetes AI Agent Platform..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Start database services
echo "📦 Starting database services..."
docker-compose up -d postgres redis

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
sleep 10

# Setup backend
echo "🔧 Setting up backend..."
cd backend

# Initialize Go module
go mod init grafana-ai-agent-platform/backend

# Download dependencies
go mod tidy

# Create .env file for backend
cat > .env << EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=kubernetes_ai_platform
JWT_SECRET=your-secret-key-change-in-production
OPENROUTER_KEY=your-openrouter-api-key
EOF

echo "✅ Backend setup completed"

# Setup frontend
echo "🔧 Setting up frontend..."
cd ../frontend

# Install dependencies
npm install

# Create .env.local file for frontend
cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
EOF

echo "✅ Frontend setup completed"

# Go back to root
cd ..

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Update the OPENROUTER_KEY in backend/.env with your API key"
echo "2. Start the backend: cd backend && go run cmd/main.go"
echo "3. Start the frontend: cd frontend && npm run dev"
echo "4. Access the application at http://localhost:3000"
echo ""
echo "🐳 Or use Docker Compose to run everything:"
echo "docker-compose up -d"
echo "" 