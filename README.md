# 🚀 SaaS Blueprint

> **A production-ready, full-stack SaaS starter kit** with authentication, payments, multi-tenancy, real-time features, and comprehensive monitoring.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-19-blue.svg)](https://reactjs.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)](https://www.postgresql.org)

---

## ✨ Features

### 🔐 Authentication & Security
- ✅ **JWT Authentication** - Access & refresh tokens with automatic rotation
- ✅ **Multi-Factor Authentication (MFA)** - TOTP with QR codes and backup codes
- ✅ **OAuth2 Integration** - Google & GitHub login
- ✅ **Password Reset Flow** - Email-based secure password reset
- ✅ **Email Verification** - Token-based email confirmation
- ✅ **Rate Limiting** - Redis-based adaptive rate limiting (global, auth, authenticated)
- ✅ **Session Management** - Database-backed sessions with revocation

### 💳 Payments & Subscriptions (Stripe)
- ✅ **Subscription Management** - Create, update, cancel subscriptions
- ✅ **Checkout Sessions** - Hosted Stripe Checkout integration
- ✅ **Customer Portal** - Self-service billing portal
- ✅ **Webhook Handling** - Secure webhook processing with signature verification
- ✅ **Multiple Plans** - Support for tiers and pricing models

### 🏢 Multi-Tenancy & Organizations
- ✅ **Organizations** - Team workspaces with hierarchical structure
- ✅ **Role-Based Access Control (RBAC)** - Owner, Admin, Member roles
- ✅ **Team Invitations** - Email invitations with expiring tokens
- ✅ **Member Management** - Add, remove, and update team members
- ✅ **Organization Settings** - Customizable organization profiles

### 📁 File Storage
- ✅ **S3-Compatible Storage** - MinIO for object storage
- ✅ **Avatar Upload** - User profile images
- ✅ **File Management** - Upload, download, delete files
- ✅ **Presigned URLs** - Secure file access
- ✅ **Content Type Detection** - Automatic MIME type handling

### 🔔 Real-Time Features
- ✅ **WebSocket Server** - Real-time bidirectional communication
- ✅ **Live Notifications** - Push notifications to connected clients
- ✅ **Presence System** - Track online/offline users
- ✅ **Connection Management** - Automatic reconnection & heartbeat

### 📧 Background Jobs & Email
- ✅ **Asynq Workers** - Distributed task queue with Redis
- ✅ **Email Service** - Resend integration for transactional emails
- ✅ **Job Scheduling** - Periodic and delayed job execution
- ✅ **Job Monitoring** - Queue statistics and worker health

### 🔍 Full-Text Search
- ✅ **PostgreSQL FTS** - Native full-text search with GIN indexes
- ✅ **Weighted Ranking** - Relevance scoring (A for titles, B for content)
- ✅ **Multi-Entity Search** - Search across users, organizations, notifications
- ✅ **Highlighting** - Context-aware search result snippets
- ✅ **Fuzzy Matching** - Unaccent support for international characters
- ✅ **Boolean Operators** - AND, OR, NOT queries

### 🌐 Internationalization (i18n)
- ✅ **Multi-Language Support** - English, Spanish, French
- ✅ **i18next Integration** - Client-side translations
- ✅ **Language Detection** - Browser language detection
- ✅ **Dynamic Language Switching** - Change language without reload

### 📱 Progressive Web App (PWA)
- ✅ **Installable** - Add to home screen on all platforms
- ✅ **Offline Support** - Service worker caching strategies
- ✅ **Push Notifications** - Web push notification ready
- ✅ **Update Prompts** - Automatic app update detection
- ✅ **Install Prompts** - Custom install UI
- ✅ **Offline Indicator** - Real-time connection status

### 📊 Monitoring & Observability
- ✅ **Prometheus Metrics** - 40+ application metrics
- ✅ **Health Checks** - Liveness, readiness, and detailed health endpoints
- ✅ **Structured Logging** - JSON logs with zerolog
- ✅ **Grafana Dashboards** - Pre-configured monitoring dashboards
- ✅ **Error Tracking** - Comprehensive error boundaries
- ✅ **Performance Monitoring** - Request duration, database queries

### 📖 API Documentation
- ✅ **OpenAPI/Swagger** - Interactive API documentation
- ✅ **Swagger UI** - Built-in documentation viewer at `/api/v1/docs`
- ✅ **Auto-Generated Schemas** - Type-safe API contracts

### 🚀 Performance Optimizations
- ✅ **Database Indexes** - Comprehensive indexing for common queries
- ✅ **Redis Caching** - Response caching with configurable TTL
- ✅ **HTTP Caching** - Cache-Control, ETags, compression
- ✅ **Connection Pooling** - Optimized database connections
- ✅ **Code Splitting** - Vite chunk optimization
- ✅ **Lazy Loading** - On-demand route loading

### 🧪 Testing & Quality
- ✅ **Unit Tests** - Backend test coverage with fixtures
- ✅ **Integration Tests** - API endpoint testing
- ✅ **CI/CD Pipeline** - GitHub Actions workflow
- ✅ **Test Utilities** - Comprehensive test helpers
- ✅ **Code Coverage** - Automated coverage reporting

### 🐳 Production Deployment
- ✅ **Docker Compose** - Full production stack configuration
- ✅ **Nginx Reverse Proxy** - SSL/TLS termination, load balancing
- ✅ **Multi-Stage Builds** - Optimized Docker images
- ✅ **Health Checks** - All services monitored
- ✅ **Deployment Guide** - Comprehensive documentation
- ✅ **Backup Strategies** - Database, files, and config backups

---

## 🏗️ Tech Stack

### Frontend
- **React 19** - Latest React with concurrent features
- **Vite** - Lightning-fast build tool
- **Chakra UI 3** - Accessible component library
- **TanStack Query v5** - Powerful async state management
- **React Router v7** - Client-side routing
- **React Hook Form + Zod** - Type-safe form validation
- **i18next** - Internationalization framework
- **Lucide React** - Beautiful icon library

### Backend
- **Go 1.24** - High-performance backend
- **Chi Router v5** - Lightweight, idiomatic router
- **sqlc** - Type-safe SQL code generation
- **JWT (golang-jwt)** - Secure authentication
- **Zerolog** - Fast structured logging
- **Validator v10** - Request validation
- **Asynq** - Background job processing
- **gorilla/websocket** - WebSocket support

### Infrastructure
- **PostgreSQL 16** - Robust relational database
- **Redis 7** - Caching and job queue
- **MinIO** - S3-compatible object storage
- **Prometheus** - Metrics collection
- **Grafana** - Visualization platform
- **Nginx** - Reverse proxy and load balancer
- **Docker & Docker Compose** - Containerization

### Integrations
- **Stripe** - Payment processing
- **Resend** - Transactional email
- **Google OAuth** - Social authentication
- **GitHub OAuth** - Developer authentication

---

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Git
- (Optional) Go 1.24+ for local development
- (Optional) Node.js 20+ for local development

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/saas-blueprint.git
   cd saas-blueprint
   ```

2. **Copy environment file:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start all services:**
   ```bash
   docker-compose up -d
   ```

4. **Run database migrations:**
   ```bash
   docker-compose exec backend ./server migrate up
   ```

5. **Access the application:**
   - **Frontend:** http://localhost:3000
   - **Backend API:** http://localhost:8080
   - **API Docs:** http://localhost:8080/api/v1/docs
   - **MinIO Console:** http://localhost:9001
   - **Prometheus:** http://localhost:9090
   - **Grafana:** http://localhost:3001

6. **Demo credentials:**
   ```
   Email: admin@example.com
   Password: password123
   ```

---

## 📁 Project Structure

```
saas-blueprint/
├── frontend/                   # React frontend application
│   ├── src/
│   │   ├── components/        # Reusable UI components
│   │   ├── pages/             # Route pages
│   │   ├── contexts/          # React contexts (Auth, etc.)
│   │   ├── hooks/             # Custom React hooks
│   │   ├── lib/               # Utilities (API, i18n, etc.)
│   │   └── utils/             # Helper functions
│   ├── public/                # Static assets
│   │   ├── manifest.json      # PWA manifest
│   │   ├── sw.js              # Service worker
│   │   └── icons/             # PWA icons
│   ├── Dockerfile             # Dev Dockerfile
│   ├── Dockerfile.prod        # Production Dockerfile
│   └── PWA.md                 # PWA documentation
│
├── backend/                    # Go backend application
│   ├── main.go                # Application entry point
│   ├── db/
│   │   ├── migrations/        # Database migrations
│   │   └── queries/           # sqlc queries
│   ├── internal/
│   │   ├── api/v1/           # API handlers
│   │   ├── auth/             # Authentication
│   │   ├── middleware/       # HTTP middleware
│   │   ├── models/           # Data models
│   │   ├── repository/       # Data access layer
│   │   ├── search/           # Full-text search
│   │   ├── metrics/          # Prometheus metrics
│   │   ├── mfa/              # MFA service
│   │   ├── oauth/            # OAuth providers
│   │   ├── payment/          # Stripe integration
│   │   ├── storage/          # File storage
│   │   ├── websocket/        # WebSocket server
│   │   └── worker/           # Background jobs
│   ├── docs/                  # API documentation
│   ├── Dockerfile             # Dev Dockerfile
│   ├── Dockerfile.prod        # Production Dockerfile
│   ├── TESTING.md             # Testing guide
│   └── Makefile               # Development commands
│
├── nginx/                      # Nginx configuration
│   └── nginx.prod.conf        # Production nginx config
│
├── prometheus/                 # Prometheus configuration
│   └── prometheus.yml         # Scrape configs
│
├── docker-compose.yml          # Development stack
├── docker-compose.prod.yml     # Production stack
├── DEPLOYMENT.md               # Deployment guide
├── PERFORMANCE.md              # Performance guide
└── README.md                   # This file
```

---

## 🔧 Development

### Hot Reloading

Both frontend and backend support hot reloading:

- **Frontend**: Vite HMR - instant updates on file changes
- **Backend**: Air - automatic rebuild and restart

### Database Migrations

```bash
# Create migration
make migrate-create name=add_feature

# Run migrations
make migrate-up

# Rollback last migration
make migrate-down

# Migration status
make migrate-status
```

### Testing

```bash
# Backend tests
cd backend
make test                    # Run all tests
make test-unit              # Unit tests only
make test-integration       # Integration tests only
make test-coverage          # With coverage report

# Frontend tests
cd frontend
npm run test                # Run tests
npm run test:ui             # Interactive UI
npm run test:coverage       # With coverage
```

### Code Generation

```bash
# Generate sqlc code
cd backend
make sqlc

# Generate Swagger docs
cd backend
make swagger
```

---

## 📚 Documentation

- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Production deployment guide
- **[PERFORMANCE.md](./PERFORMANCE.md)** - Performance optimization guide
- **[backend/TESTING.md](./backend/TESTING.md)** - Testing infrastructure guide
- **[frontend/PWA.md](./frontend/PWA.md)** - Progressive Web App guide
- **[API Docs](http://localhost:8080/api/v1/docs)** - Interactive API documentation

---

## 🌟 Key Features Explained

### Authentication Flow

1. **Registration** → Email verification → Account activated
2. **Login** → Password check → MFA (if enabled) → JWT tokens issued
3. **Token Refresh** → Automatic token rotation before expiry
4. **OAuth** → Google/GitHub → Account creation/linking → JWT tokens

### Multi-Tenancy

- Users can create and join multiple organizations
- Role-based permissions (Owner, Admin, Member)
- Data isolation per organization
- Invitation system with expiring tokens

### Real-Time Notifications

1. User connects via WebSocket
2. Server authenticates connection
3. Notifications pushed to user's channel
4. Automatic reconnection on disconnect

### Full-Text Search

- PostgreSQL GIN indexes for fast search
- Weighted ranking (titles > content > descriptions)
- Automatic index updates via triggers
- Support for multi-word queries and operators

### Background Jobs

```go
// Schedule a job
client.Enqueue(
    asynq.NewTask("email:send", payload),
    asynq.Queue("emails"),
    asynq.MaxRetry(3),
)
```

### Caching Strategy

- **Response Caching**: Redis cache for GET requests (5min TTL)
- **HTTP Caching**: Cache-Control headers for static assets
- **ETags**: Conditional requests to save bandwidth
- **PWA Caching**: Service worker for offline support

---

## 🎯 Performance Targets

| Metric | Target | Status |
|--------|--------|--------|
| First Contentful Paint | < 1s | ✅ |
| API Response (p95) | < 100ms | ✅ |
| Database Query (p95) | < 50ms | ✅ |
| Time to Interactive | < 3s | ✅ |
| Cache Hit Rate | > 80% | ✅ |
| Lighthouse Score | > 90 | ✅ |

---

## 🚀 Deployment

### Production Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for comprehensive deployment guide.

**Quick deployment:**

```bash
# 1. Configure environment
cp .env.production.example .env.production
vim .env.production

# 2. Setup SSL certificates
sudo certbot certonly --standalone -d yourdomain.com

# 3. Build and start
docker-compose -f docker-compose.prod.yml up -d

# 4. Run migrations
docker-compose -f docker-compose.prod.yml exec backend ./server migrate up
```

### Docker Production Stack

The production stack includes:
- **PostgreSQL** with persistent volume
- **Redis** with persistence
- **MinIO** for file storage
- **Backend** with health checks
- **Frontend** served by Nginx
- **Nginx** reverse proxy with SSL
- **Prometheus** for metrics
- **Grafana** for dashboards

---

## 🔒 Security

- JWT with RS256 signing (configurable)
- Password hashing with bcrypt
- CSRF protection
- XSS protection headers
- SQL injection prevention (prepared statements)
- Rate limiting on all endpoints
- MFA for enhanced security
- Secure session management
- OAuth2 with state validation
- Webhook signature verification

---

## 📊 Monitoring

### Prometheus Metrics

40+ metrics including:
- HTTP request duration, count, size
- Database query performance
- Cache hit/miss rates
- WebSocket connections
- Background job stats
- Rate limit hits

### Health Checks

- `/api/v1/health` - Overall health status
- `/api/v1/health/live` - Liveness probe
- `/api/v1/health/ready` - Readiness probe

### Grafana Dashboards

Pre-configured dashboards for:
- API performance
- Database statistics
- Cache efficiency
- System resources

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Built with amazing open-source technologies:
- Go, React, PostgreSQL, Redis, Docker
- Stripe, Resend, and many others

---

## 📬 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/saas-blueprint/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/saas-blueprint/discussions)
- **Email**: support@yourdomain.com

---

**Made with ❤️ by the SaaS Blueprint team**
