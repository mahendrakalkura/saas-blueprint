# SaaS Blueprint

Full-stack SaaS application with React frontend, Go backend, and PostgreSQL database.

## Tech Stack

**Frontend:**
- React 19 with TypeScript support
- Vite 7 for fast development
- Chakra UI 3 for beautiful components
- React Router v7 for navigation
- TanStack Query (React Query) v5 for server state
- React Hook Form + Zod for form validation
- Vitest for testing
- Hot Module Replacement (HMR)

**Backend:**
- Go 1.23 with Chi Router
- sqlc for type-safe SQL queries
- JWT authentication with refresh tokens
- Structured logging (zerolog)
- Air for hot reloading
- Comprehensive middleware stack

**Infrastructure:**
- PostgreSQL 16 with migrations (goose)
- Redis 7 for caching/sessions
- MinIO for S3-compatible object storage
- Docker Compose for orchestration
- Multi-stage builds for production

## Quick Start

1. **Clone and setup:**
   ```bash
   cp .env.example .env
   # Edit .env if needed
   ```

2. **Start all services:**
   ```bash
   docker-compose up
   ```

3. **Run database migrations:**
   ```bash
   make migrate-up
   ```

4. **Seed demo data (optional):**
   ```bash
   make seed
   ```

5. **Access the application:**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - MinIO Console: http://localhost:9001
   - Database: localhost:5432 (postgres/postgres)

6. **Demo Credentials:**
   ```
   Email: admin@example.com
   Password: password123
   ```

## Features

### ✅ Authentication
- User registration with email
- Login with JWT tokens (access + refresh)
- Password reset flow
- Email verification (tokens ready, email service pending)
- Protected routes on frontend
- Session management in database

### ✅ Database
- PostgreSQL with connection pooling
- Type-safe queries with sqlc
- Database migrations with goose
- Comprehensive schema (users, sessions, organizations, audit logs, etc.)
- Full-text search support

### ✅ Infrastructure
- Docker Compose orchestration
- Multi-stage Docker builds
- Hot reloading (frontend & backend)
- Graceful shutdown handling
- Health check endpoints
- MinIO for file storage (ready)
- Redis for caching (ready)

### ✅ Frontend
- Modern React 19 with hooks
- Dark mode support
- Form validation
- Loading states & error boundaries
- Protected routes
- API client with auto-refresh

### ✅ Backend
- API versioning (v1)
- Request validation
- Security headers (CSP, XSS, CSRF protection)
- Pagination helpers
- Structured logging
- CORS configuration

## Development

### Hot Reloading

Both frontend and backend have hot reloading enabled:
- **Frontend**: Vite HMR reloads instantly on file changes
- **Backend**: Air rebuilds and restarts on Go file changes

### Database Migrations

Create a new migration:
```bash
make migrate-create name=add_new_table
```

Run migrations:
```bash
make migrate-up
```

Rollback last migration:
```bash
make migrate-down
```

### Using sqlc

1. Add database schema to `backend/db/migrations/*.sql`
2. Add queries to `backend/db/queries/*.sql`
3. Generate Go code:
   ```bash
   make sqlc
   ```

### Seeding Data

Populate the database with demo users:
```bash
make seed
```

This creates three demo accounts (all with password: `password123`):
- admin@example.com
- john@example.com
- jane@example.com

### API Development

- API endpoints are in `backend/internal/api/v1/`
- All API routes are proxied from frontend via `/api/*`
- Use the auth middleware for protected routes
- Backend includes CORS middleware for local development

### Project Structure

```
saas-blueprint/
├── frontend/           # React + Vite frontend
│   ├── src/
│   ├── public/
│   ├── Dockerfile
│   └── vite.config.js
├── backend/            # Go backend
│   ├── main.go
│   ├── db/
│   │   ├── migrations/
│   │   ├── queries/
│   │   └── sqlc/
│   ├── Dockerfile
│   └── sqlc.yaml
├── docker-compose.yml
└── .env
```

## Available Commands

### Docker Commands
```bash
# Start services
docker-compose up

# Start in background
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Rebuild services
docker-compose up --build
```

### Backend Commands
```bash
# Access backend container
docker-compose exec backend sh

# Generate sqlc code
docker-compose exec backend sqlc generate

# Run Go tests
docker-compose exec backend go test ./...
```

### Frontend Commands
```bash
# Access frontend container
docker-compose exec frontend sh

# Install new package
docker-compose exec frontend npm install package-name

# Run linter
docker-compose exec frontend npm run lint
```

## Environment Variables

All environment variables are configured in `.env`:

- `DB_HOST`: Database hostname (default: db)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database username (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: postgres)
- `BACKEND_PORT`: Backend API port (default: 8080)
- `FRONTEND_PORT`: Frontend dev server port (default: 3000)

## Health Check

Test if the backend is running:
```bash
curl http://localhost:8080/api/health
```

Expected response:
```json
{"status":"ok","message":"Backend is running"}
```
