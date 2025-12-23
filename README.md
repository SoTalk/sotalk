# Soldef - Solana Secure Messaging Platform

A secure, privacy-focused messaging application using Solana wallet addresses as user identities, featuring end-to-end encryption, real-time messaging, and integrated wallet functionality.

## Tech Stack

- **Backend**: Go + Gin framework
- **Database**: PostgreSQL with GORM (ORM)
- **Cache**: Redis
- **Real-time**: WebSocket
- **Blockchain**: Solana
- **Architecture**: Clean Architecture (Uncle Bob's Pattern)
- **Security**: End-to-end encryption, message signing with Solana keys

## Project Structure

```
soldef/
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/           # Business entities (User, Message, Wallet, etc.)
│   ├── usecase/          # Business logic layer
│   ├── delivery/         # HTTP/WebSocket handlers
│   └── repository/       # Database implementations (GORM)
├── pkg/                  # Shared packages (config, logger, etc.)
├── migrations/           # Database migrations
└── scripts/              # Build & deployment scripts
```

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 7+

## Getting Started

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd soldef
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment

```bash
cp .env.example .env
# Edit .env with your local database credentials
```

### 4. Setup local databases

Make sure PostgreSQL and Redis are running locally.

**PostgreSQL:**
```bash
createdb soldef
```

### 5. Run the application

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

### 6. Test health check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "uptime": "1m0s",
  "service": "soldef-api",
  "version": "1.0.0"
}
```

## Development

### Run with hot reload (using air)

```bash
go install github.com/air-verse/air@latest
air
```

### Run tests

```bash
go test ./...
```

### Build for production

```bash
go build -o bin/soldef cmd/api/main.go
```

## API Documentation

- Health Check: `GET /health`
- Readiness Check: `GET /ready`
- API v1: `/api/v1/*` (to be implemented)

## Day 1 Status ✅

- [x] Project structure initialized (Clean Architecture)
- [x] Dependencies installed (Gin, GORM, Redis, Solana, etc.)
- [x] Configuration management
- [x] Logger implementation (Zap)
- [x] GORM database connection
- [x] Health check endpoint
- [x] Basic HTTP server with graceful shutdown

## Next Steps (Day 2)

- [ ] Domain layer - User entity
- [ ] Solana wallet authentication
- [ ] JWT middleware
- [ ] User repository with GORM
- [ ] Auth endpoints

## License

MIT
