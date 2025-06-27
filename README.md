# ChirpStack

A Twitter-like social media backend built with Go, demonstrating modern web development practices including RESTful APIs, JWT authentication, database management, and comprehensive testing.

## What This Project Does

ChirpStack is a learning-focused backend implementation that provides:

- **User Management**: Registration, authentication, and profile updates
- **Social Media Features**: Create, read, and delete short messages (chirps)
- **JWT Authentication**: Secure token-based authentication with refresh tokens
- **Content Moderation**: Automatic profanity filtering
- **Admin Interface**: Metrics tracking and system management
- **Webhook Integration**: External service integration for user upgrades

## How to Install and Run This Project

### Prerequisites

- **Go 1.24.4+**: [Download Go](https://golang.org/dl/)
- **PostgreSQL 12+**: [Install PostgreSQL](https://www.postgresql.org/download/)
- **Git**: For cloning the repository

### Quick Start

1. **Clone the Repository**
   ```bash
   git clone https://github.com/your-username/chirpstack.git
   cd chirpstack
   ```

2. **Set Up the Database**
   ```bash
   # Create database and user
   createdb chirpstack
   psql chirpstack -c "CREATE USER chirpstack WITH PASSWORD 'password';"
   psql chirpstack -c "GRANT ALL PRIVILEGES ON DATABASE chirpstack TO chirpstack;"

   # Run migrations
   psql "postgres://chirpstack:password@localhost/chirpstack?sslmode=disable" -f sql/schema/001_users.sql
   psql "postgres://chirpstack:password@localhost/chirpstack?sslmode=disable" -f sql/schema/002_chirps.sql
   psql "postgres://chirpstack:password@localhost/chirpstack?sslmode=disable" -f sql/schema/003_add_hashed_password_to_users.sql
   psql "postgres://chirpstack:password@localhost/chirpstack?sslmode=disable" -f sql/schema/004_refresh_tokens.sql
   psql "postgres://chirpstack:password@localhost/chirpstack?sslmode=disable" -f sql/schema/005_add_is_chirpy_red_to_users.sql
   ```

3. **Configure Environment**
   ```bash
   # Create .env file
   cat > .env << EOF
   DB_URL=postgres://chirpstack:password@localhost/chirpstack?sslmode=disable
   PLATFORM=dev
   JWT_TOKEN_SECRET=$(openssl rand -base64 32)
   POLKA_KEY=your-webhook-api-key-here
   EOF
   ```

4. **Install Dependencies and Run**
   ```bash
   go mod tidy
   go run main.go
   ```

5. **Verify Installation**
   ```bash
   curl http://localhost:8080/api/health
   # Should return: OK
   ```

### Testing the API

1. **Create a User**
   ```bash
   curl -X POST http://localhost:8080/api/users \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123"}'
   ```

2. **Login**
   ```bash
   curl -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123"}'
   ```

3. **Create a Chirp** (use token from login response)
   ```bash
   curl -X POST http://localhost:8080/api/chirps \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -d '{"body":"Hello, ChirpStack!"}'
   ```

4. **View All Chirps**
   ```bash
   curl http://localhost:8080/api/chirps
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run HTTP integration tests (server must be running)
# Use your preferred HTTP client with files in http_tests/
# I use https://github.com/rest-nvim/rest.nvim as a plugin with Neovim!
```

## Project Structure

```
chirpstack/
├── main.go                 # Application entry point
├── internal/
│   ├── api/               # HTTP handlers and middleware
│   │   ├── api.go         # Core API configuration
│   │   ├── auth.go        # Authentication endpoints
│   │   ├── chirps.go      # Chirp management endpoints
│   │   ├── users.go       # User management endpoints
│   │   └── admin.go       # Admin endpoints
│   ├── auth/              # Authentication utilities
│   │   ├── auth.go        # JWT and password handling
│   │   └── auth_test.go   # Authentication tests
│   ├── database/          # SQLC generated database code
│   └── utils/             # Shared utilities
│       ├── utils.go       # Helper functions
│       └── utils_test.go  # Utility tests
├── sql/
│   ├── schema/            # Database migrations
│   └── queries/           # SQLC query definitions
├── http_tests/            # HTTP integration tests
├── docs/                  # Project documentation
├── assets/                # Static assets
├── index.html            # Frontend entry point
├── go.mod                # Go module definition
├── sqlc.yaml             # SQLC configuration
```

## Development Commands

```bash
# Build the application
go build

# Run tests
go test ./...

# Generate database code (requires sqlc)
sqlc generate

# Format code
go fmt ./...

# Clean dependencies
go mod tidy
```

## API Documentation

Full API documentation is available in [`docs/API.md`](docs/API.md).

### Quick API Reference

- **Health**: `GET /api/health`
- **Users**: `POST /api/users`, `PUT /api/users`
- **Auth**: `POST /api/login`, `POST /api/refresh`, `POST /api/revoke`
- **Chirps**: `GET|POST /api/chirps`, `GET|DELETE /api/chirps/{id}`
- **Admin**: `GET /admin/metrics`, `POST /admin/reset`
- **Static**: `GET /app/*`

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Follow Go naming conventions
- Update documentation for API changes
- Run `go fmt` and `go vet` before committing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


---

**Thanks [boot.dev](https://www.boot.dev/) for this guided course**
