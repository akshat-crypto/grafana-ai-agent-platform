# Kubernetes AI Agent Platform - Complete Development Prompt

## Project Overview
Create a complete Kubernetes management platform with AI-powered automation. This is a full-stack application with a Go backend and Next.js frontend that allows users to manage Kubernetes clusters and deploy stacks using AI assistance.

## Technical Requirements

### Backend (Go)
- **Framework**: Gin HTTP router
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens with bcrypt password hashing
- **Kubernetes Integration**: client-go for cluster operations
- **AI Integration**: OpenRouter API for GPT-powered responses
- **Architecture**: Clean architecture with handlers, services, models, and middleware

### Frontend (Next.js)
- **Framework**: Next.js 14 with TypeScript
- **Styling**: Tailwind CSS with custom components
- **State Management**: Zustand with persistence
- **UI Components**: Headless UI + Heroicons
- **HTTP Client**: Axios with interceptors
- **Forms**: React Hook Form with validation

## Core Features

### 1. Authentication System
- User registration and login
- JWT token-based authentication
- Password hashing with bcrypt
- Protected routes and middleware

### 2. Kubernetes Cluster Management
- Add clusters via kubeconfig
- Validate cluster connectivity
- Store cluster information securely
- List and manage user clusters
- Delete clusters with confirmation

### 3. AI Agent Integration
- GPT-powered responses for Kubernetes operations
- Query processing with cluster context
- Stack deployment automation
- Error handling and graceful degradation
- History tracking for queries and deployments

### 4. Dashboard Interface
- Modern, responsive design
- Sidebar navigation with icons
- Real-time cluster status
- Quick action buttons
- AI chat interface with query suggestions

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Kubernetes Clusters Table
```sql
CREATE TABLE kubernetes_clusters (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    kube_config TEXT NOT NULL,
    cluster_url VARCHAR(255),
    version VARCHAR(50),
    status VARCHAR(50) DEFAULT 'pending',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Agent Queries Table
```sql
CREATE TABLE agent_queries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    cluster_id INTEGER REFERENCES kubernetes_clusters(id),
    query TEXT NOT NULL,
    response TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Deployments Table
```sql
CREATE TABLE deployments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    cluster_id INTEGER REFERENCES kubernetes_clusters(id),
    stack_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    manifest TEXT,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/profile` - Get user profile

### Kubernetes
- `POST /api/kubernetes/validate` - Validate kubeconfig
- `POST /api/kubernetes/clusters` - Add new cluster
- `GET /api/kubernetes/clusters` - List user clusters
- `DELETE /api/kubernetes/clusters/:id` - Remove cluster
- `GET /api/kubernetes/clusters/:id/resources` - Get cluster resources

### AI Agent
- `POST /api/agent/query` - Send prompt to AI agent
- `POST /api/agent/deploy` - Deploy stack via AI
- `GET /api/agent/queries` - Get query history
- `GET /api/agent/deployments` - Get deployment history

## Frontend Pages

### 1. Login Page (`/login`)
- Email/password form with validation
- Password visibility toggle
- Link to registration
- Error handling with toast notifications

### 2. Dashboard (`/`)
- Cluster overview cards
- Quick action buttons
- AI chat interface with textarea
- Query suggestions
- Response display with syntax highlighting

### 3. Kubernetes Management (`/kubernetes`)
- Add cluster form with validation
- Kubeconfig validation with real-time feedback
- Cluster cards with status indicators
- Delete confirmation modals
- Empty state with call-to-action

### 4. Layout Component
- Responsive sidebar with navigation
- Header with user menu and notifications
- Authentication state management
- Route protection

## Key Implementation Details

### Backend Structure
```
backend/
├── cmd/main.go                 # Application entry point
├── internal/
│   ├── config/config.go        # Configuration management
│   ├── models/                 # Database models
│   ├── handlers/               # HTTP handlers
│   ├── middleware/             # Auth and CORS middleware
│   └── agent/agent.go         # AI agent service
├── pkg/
│   ├── database/database.go    # Database connection
│   └── kubernetes/client.go    # K8s client wrapper
└── go.mod                     # Dependencies
```

### Frontend Structure
```
frontend/
├── pages/                     # Next.js pages
├── components/                # Reusable components
├── store/                    # Zustand stores
├── utils/                    # API utilities
├── styles/                   # Global styles
└── package.json              # Dependencies
```

### Environment Variables
**Backend (.env)**
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=kubernetes_ai_platform
JWT_SECRET=your-secret-key
OPENROUTER_KEY=your-openrouter-api-key
```

**Frontend (.env.local)**
```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Docker Setup
- Multi-stage Dockerfiles for both backend and frontend
- Docker Compose for development environment
- PostgreSQL and Redis services
- Environment variable management

## Development Commands
```bash
# Backend
cd backend
go mod tidy
go run cmd/main.go

# Frontend
cd frontend
npm install
npm run dev

# Database
docker-compose up -d postgres redis
```

## Security Considerations
- JWT token validation
- Password hashing with bcrypt
- Input validation and sanitization
- CORS configuration
- Secure kubeconfig storage
- Rate limiting (to be implemented)

## Error Handling
- Graceful API error responses
- Frontend error boundaries
- Toast notifications for user feedback
- Logging for debugging
- AI response validation

## Testing Strategy
- Unit tests for backend services
- Integration tests for API endpoints
- Frontend component testing
- E2E testing with Playwright (to be implemented)

## Deployment
- Docker containerization
- Environment-specific configurations
- Database migrations
- Health check endpoints
- Monitoring and logging

## Additional Features to Implement
- Real-time cluster monitoring
- Advanced stack templates
- Multi-cluster management
- Role-based access control
- Audit logging
- Backup and restore functionality
- Integration with Helm charts
- Custom resource definitions support

This prompt provides a comprehensive blueprint for building a production-ready Kubernetes AI Agent Platform. The implementation should follow industry best practices for security, scalability, and maintainability. 