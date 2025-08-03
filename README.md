# Kubernetes AI Agent Platform

A modern web application that combines Kubernetes cluster management with AI-powered automation for deploying and managing cloud-native stacks.

## Features

- 🔐 **Authentication System**: Secure login/register with PostgreSQL
- 🎯 **Kubernetes Dashboard**: Interactive cluster management interface
- 🤖 **AI Agent Integration**: GPT-powered automation for stack deployment
- 📊 **Real-time Monitoring**: Live cluster status and metrics
- 🚀 **One-click Deployments**: Deploy Grafana, ELK, and other stacks
- 🔧 **Cluster Validation**: Automatic kubeconfig validation and connection testing

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT tokens
- **Kubernetes**: client-go
- **AI Integration**: OpenRouter API

### Frontend
- **Framework**: Next.js 14
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **UI Components**: Headless UI + Heroicons
- **State Management**: Zustand
- **HTTP Client**: Axios

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd grafana-ai-agent-platform
   ```

2. **Start the development environment**
   ```bash
   docker-compose up -d
   ```

3. **Setup Backend**
   ```bash
   cd backend
   go mod init grafana-ai-agent-platform/backend
   go mod tidy
   go run main.go
   ```

4. **Setup Frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

5. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Database: localhost:5432

## Environment Variables

Create `.env` files in both `backend/` and `frontend/` directories:

### Backend (.env)
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=kubernetes_ai_platform
JWT_SECRET=your-secret-key
OPENROUTER_KEY=your-openrouter-api-key
```

### Frontend (.env.local)
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout

### Kubernetes
- `POST /api/kubernetes/validate` - Validate kubeconfig
- `POST /api/kubernetes/clusters` - Add new cluster
- `GET /api/kubernetes/clusters` - List user clusters
- `DELETE /api/kubernetes/clusters/:id` - Remove cluster

### AI Agent
- `POST /api/agent/query` - Send prompt to AI agent
- `POST /api/agent/deploy` - Deploy stack via AI

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Kubernetes    │
│   (Next.js)     │◄──►│   (Go/Gin)      │◄──►│   Clusters      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   PostgreSQL    │
                       │   Database      │
                       └─────────────────┘
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details 