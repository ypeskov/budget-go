# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based budget management REST API built with Echo framework. It provides endpoints for managing personal finances including accounts, transactions, categories, currencies, and user settings. The application uses PostgreSQL as the database and implements JWT-based authentication.

## Development Commands

### Building and Running
- **Local development**: `go run ./cmd/api` (runs on port configured in .env, defaults to 8000)
- **Build binary**: `go build -o budget-go ./cmd/api`
- **Docker build**: `./build-and-push.sh` (builds for linux/arm64 by default)
- **Docker build with push**: `./build-and-push.sh push [tag]`

### Environment Setup
- Copy `.env.sample` to `.env` and configure database connection
- Required PostgreSQL database (connection details in config)
- No test framework is currently configured in the project

## Architecture Overview

### Main Components
- **Entry point**: `cmd/api/main.go` - Handles graceful shutdown and server initialization
- **Configuration**: `internal/config/config.go` - Environment-based config using godotenv
- **Database**: `internal/database/db.go` - PostgreSQL connection with sqlx, singleton pattern
- **Server**: `internal/server/server.go` - HTTP server setup
- **Routes**: `internal/routes/routes.go` - Echo router with middleware setup
- **Services**: `internal/services/manager.go` - Business logic layer with service manager pattern

### Layer Architecture
The application follows a clean architecture pattern:
1. **Routes** (`internal/routes/*`) - HTTP handlers and routing
2. **Services** (`internal/services/*`) - Business logic layer
3. **Repositories** (`internal/repositories/*`) - Data access layer
4. **Models** (`internal/models/*`) - Domain entities
5. **DTOs** (`internal/dto/*`) - Data transfer objects

### Key Features
- JWT authentication with custom middleware (`internal/middleware/tokenMiddleware.go`)
- CORS enabled for all origins
- Graceful shutdown handling
- Structured logging with logrus
- Environment-based configuration
- Docker support with multi-stage builds

### API Endpoints
- `/health` - Health check endpoint
- `/auth/*` - Authentication (login/register)
- `/accounts/*` - Account management
- `/categories/*` - Transaction categories
- `/transactions/*` - Financial transactions
- `/currencies/*` - Currency and exchange rates
- `/settings/*` - User preferences
- `/reports/*` - Financial reporting

### Database
- Uses PostgreSQL with sqlx for database operations
- Singleton database connection pattern
- Repository pattern for data access
- Models represent database entities with proper Go tags

### Configuration
Environment variables (with defaults):
- `SERVER_PORT=8000`
- `LOG_LEVEL=info`
- `DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=postgres`, `DB_PASSWORD=password`, `DB_NAME=budget`
- `SECRET_KEY=SECRET_KEY` (for JWT signing)

### Reverse Proxy Setup
Includes Caddyfile for development reverse proxy setup (port 8080) routing requests between different services.

## Important Notes
- No test files exist in the codebase currently
- Services are managed through a central Manager that handles dependency injection
- Authentication is required for all endpoints except `/auth/*` and `/health`
- All protected routes use JWT middleware for authentication
- Database connection is shared across all repositories via singleton pattern