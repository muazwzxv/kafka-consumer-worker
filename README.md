# kafka-consumer-worker

A kafka-consumer-worker service

## Features

- Clean Architecture (Entity, DTO, Repository, Service, Handler layers)
- Multi-source configuration (TOML file + environment variables)
- Docker & Docker Compose setup
- MySQL database with sqlx
- Type-safe SQL queries with sqlc
- Fiber v2 web framework
- Health check endpoints
- Middleware (logging, error handling, CORS)
- Sample CRUD API for User

## Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose
- Make (optional, for convenience commands)
- sqlc (for SQL code generation) - `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Quick Start

### Using Docker Compose

```bash
make up                 # Start all services
make migrate-up         # Run database migrations
make sqlc-generate      # Generate type-safe SQL code
```

Access the application:
- API: http://localhost:8080
- Health Check: http://localhost:8080/health

- Kafka UI: http://localhost:8090


### Local Development

```bash
docker-compose up -d mysql redis kafka   # Start database services only
make run                                                        # Run application locally
```

## Configuration

The application uses **Viper** for simple, powerful configuration management with automatic environment variable override.

### Configuration Priority

**Environment Variables → TOML File**

(Environment variables have the highest priority)

### Environment Variables

Environment variables automatically override config file values. Nested keys are flattened with underscores:

```bash
# Server configuration
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080
export SERVER_READ_TIMEOUT=5s
export SERVER_WRITE_TIMEOUT=10s

# Database configuration
export DATABASE_HOST=localhost
export DATABASE_PORT=3306
export DATABASE_USER=appuser
export DATABASE_PASSWORD=apppassword
export DATABASE_DATABASE=kafka-consumer-worker
export DATABASE_MAX_OPEN_CONNS=50
export DATABASE_MAX_IDLE_CONNS=20
```

**Environment variable naming convention:**
- Uppercase with underscores
- Nested keys are flattened: `server.host` → `SERVER_HOST`
- Examples:
  - `server.port` → `SERVER_PORT`
  - `database.max_open_conns` → `DATABASE_MAX_OPEN_CONNS`

See `.env.example` for all available options.

### TOML Configuration File

Edit `config.toml` for structured configuration:

```toml
[server]
host = "0.0.0.0"
port = 8080
read_timeout = "5s"
write_timeout = "10s"

[database]
host = "localhost"
port = 3306
user = "appuser"
password = "apppassword"
database = "kafka-consumer-worker"
max_open_conns = 25
max_idle_conns = 10
conn_max_lifetime = "5m"
```

The configuration file is **optional**. If not present, the application will use environment variables only.

## API Endpoints

### Health Check
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness check (includes database status)

### User API
- `GET /api/v1/users` - List all users (paginated)
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Development

### Available Make Commands

```bash
make help           # Show all available commands
make up             # Start all services
make down           # Stop all services
make logs           # Show logs from all services
make build          # Build the application
make run            # Run the application locally
make migrate-up     # Run database migrations
make migrate-down   # Rollback last migration
make migrate-status # Check migration status
make migrate-create # Create new migration (use NAME=migration_name)
make sqlc-generate  # Generate Go code from SQL queries
make test           # Run tests
make clean          # Stop services and remove volumes
```

### Database Operations

#### Migrations

Create a new migration:
```bash
make migrate-create NAME=add_users_table
```

Run migrations:
```bash
make migrate-up
```

Rollback:
```bash
make migrate-down
```

Check migration status:
```bash
make migrate-status
```

#### Type-Safe SQL with sqlc

This project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries. SQL queries are written in `internal/database/query/` and sqlc generates corresponding Go functions.

**Generate Code:**
```bash
make sqlc-generate
```

**Workflow:**

1. Write your SQL queries in `internal/database/query/*.sql`:
```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: CreateUser :execresult
INSERT INTO users (name, description, status)
VALUES (?, ?, ?);
```

2. Run code generation:
```bash
make sqlc-generate
```

3. Use the generated type-safe functions in your repository:
```go
import "github.com/muazwzxv/kafka-consumer-worker/internal/database/store"

queries := store.New()
user, err := queries.GetUserByID(ctx, db, id)
```

**Benefits:**
- ✅ Compile-time type safety for SQL queries
- ✅ Auto-generated models from database schema
- ✅ No manual struct mapping
- ✅ Catches SQL errors at build time

The generated code is located in `internal/database/store/` and is automatically created from:
- Schema: `internal/database/migrations/*.sql`
- Queries: `internal/database/query/*.sql`

## Project Structure

```
kafka-consumer-worker/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Viper-based configuration
│   ├── database/                # Database setup & migrations
│   │   ├── migrations/         # SQL migration files
│   │   ├── query/              # SQL queries for sqlc
│   │   ├── store/              # Generated sqlc code
│   │   ├── database.go         # Database connection
│   │   └── sqlc.yaml           # sqlc configuration
│   ├── entity/                  # Domain models
│   ├── dto/                     # Data Transfer Objects
│   │   ├── request/            # API request structures
│   │   └── response/           # API response structures
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic layer
│   ├── handler/                 # HTTP handlers
│   └── application.go           # Application wiring
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── config.toml
```

## Technology Stack

- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Dependency Injection**: [samber/do](https://do.samber.dev/) - Type-safe DI with Go generics
- **Database**: MySQL 8.0 with [sqlx](https://github.com/jmoiron/sqlx)
- **Type-Safe SQL**: [sqlc](https://sqlc.dev/) for compile-time SQL validation
- **Configuration**: [Viper](https://github.com/spf13/viper) with TOML + Environment Variables

- **Cache**: Redis 7-alpine


- **Message Queue**: Apache Kafka 7.6.0

- **Migrations**: [goose](https://github.com/pressly/goose)

## Dependency Injection

This project uses [samber/do](https://do.samber.dev/) for dependency injection, providing:

- ✅ **Type-safe** dependency management with Go 1.18+ generics
- ✅ **No reflection** or code generation required
- ✅ **Automatic** dependency resolution
- ✅ **Built-in lifecycle management** (health checks, graceful shutdown)
- ✅ **Easy testing** with service overrides

### How It Works

Services are registered in `internal/application.go` and automatically resolved:

```go
// Register services
injector := do.New()
do.Provide(injector, repository.NewUserRepository)
do.Provide(injector, service.NewUserService)
do.Provide(injector, handler.NewUserHandler)

// Dependencies are automatically resolved
handler := do.MustInvoke[*handler.UserHandler](injector)
```

### Testing with DI

Override services for testing:

```go
func TestUserHandler(t *testing.T) {
    injector := do.New()
    
    // Provide mock service
    do.Provide(injector, func(i do.Injector) (service.UserService, error) {
        return &MockUserService{}, nil
    })
    
    handler := do.MustInvoke[*handler.UserHandler](injector)
    // Test handler with mock service
}
```

**Learn more**: [samber/do Documentation](https://do.samber.dev/docs/getting-started)

## Logging

This project uses Fiber's built-in structured logger for all application logging.

### Log Levels

Ordered from most to least verbose:
- `trace` - Detailed trace information
- `debug` - Debug messages (connection attempts, internal operations)
- `info` - Informational messages (startup, shutdown, operations) **[DEFAULT]**
- `warn` - Warning messages (non-fatal issues, retries)
- `error` - Error messages (failures that don't crash the app)
- `fatal` - Fatal errors (application terminates)
- `panic` - Panic-level errors

### Configuration

Set log level in `config.toml`:
```toml
[server]
log_level = "info"  # Change to "debug" for verbose logs
```

Or via environment variable (overrides config file):
```bash
export SERVER_LOG_LEVEL=debug
./app
```

### Usage Examples

**Simple logging:**
```go
import "github.com/gofiber/fiber/v2/log"

log.Info("server started successfully")
log.Debug("connecting to database")
log.Warn("retry attempt failed")
log.Error("failed to process request")
```

**Structured logging with key-value pairs:**
```go
log.Infow("user created",
    "user_id", 123,
    "email", "user@example.com",
    "ip", "192.168.1.1")
```

**Context-aware logging (in HTTP handlers):**
```go
func (h *Handler) CreateUser(c *fiber.Ctx) error {
    logger := log.WithContext(c.UserContext())
    logger.Infow("processing request",
        "method", c.Method(),
        "path", c.Path(),
        "ip", c.IP())
    // ... handler logic
}
```

### Production Tips

- **Default level**: Use `info` in production (balanced detail/performance)
- **Debug mode**: Set `debug` for troubleshooting issues
- **Log aggregation**: Logs output to `stdout` (Docker/K8s compatible)
- **Compatible with**: Datadog, Grafana Loki, CloudWatch, ELK stack

## Deployment

### Docker

Build the Docker image:
```bash
make docker-build
```

Run with Docker:
```bash
docker run -p 8080:8080 --env-file .env.docker kafka-consumer-worker:latest
```

### Production Checklist

- [ ] Update database credentials
- [ ] Configure log level (set to `info` or `warn` for production)
- [ ] Set up log aggregation (Datadog, Grafana, CloudWatch)
- [ ] Set up monitoring and alerting
- [ ] Enable HTTPS/TLS
- [ ] Configure rate limiting
- [ ] Set up backup strategy
- [ ] Review security headers

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Author

Your Name

---

**Generated with [ready-go-cli](https://github.com/muazwzxv/ready-go-cli)**
