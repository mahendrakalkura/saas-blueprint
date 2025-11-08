# SaaS Blueprint

Full-stack SaaS application with React frontend, Go backend, and PostgreSQL database.

## Tech Stack

**Frontend:**
- React 19
- Vite 7
- Chakra UI 3
- Hot Module Replacement (HMR)

**Backend:**
- Go 1.23
- Chi Router
- sqlc for type-safe SQL queries
- Air for hot reloading

**Database:**
- PostgreSQL 16

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

3. **Access the application:**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Database: localhost:5432

## Development

### Hot Reloading

Both frontend and backend have hot reloading enabled:
- **Frontend**: Vite HMR reloads instantly on file changes
- **Backend**: Air rebuilds and restarts on Go file changes

### Using sqlc

1. Add database schema to `backend/db/migrations/*.sql`
2. Add queries to `backend/db/queries/*.sql`
3. Generate Go code:
   ```bash
   docker-compose exec backend sqlc generate
   ```

See `backend/db/queries/.gitkeep` for query syntax examples.

### API Development

- API endpoints should be added in `backend/main.go`
- All API routes are proxied from frontend via `/api/*`
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
