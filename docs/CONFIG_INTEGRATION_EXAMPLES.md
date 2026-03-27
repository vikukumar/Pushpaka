# Configuration Integration Examples

## Using DatabaseConfig in Your Code

### Example 1: Initialize PostgreSQL with SSL in main.go

```go
package main

import (
    "log"
    "github.com/vikukumar/pushpaka/internal/config"
    "github.com/vikukumar/pushpaka/internal/database"
)

func main() {
    // Load application config
    cfg := config.Load()
    
    // If using PostgreSQL with SSL support:
    if cfg.DatabaseDriver == "postgres" {
        // Use the new structured config loader
        db, err := database.NewPostgresWithConfig(cfg.DatabaseConfig)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        defer db.Close()
        
        // Now db is ready with proper SSL and connection pool config
        // ... rest of application setup
    }
}
```

### Example 2: Initialize Redis with Custom Configuration

```go
package main

import (
    "log"
    "github.com/vikukumar/pushpaka/internal/config"
    "github.com/vikukumar/pushpaka/internal/database"
)

func main() {
    cfg := config.Load()
    
    // Connect to Redis with optimized pool settings
    rdb, err := database.NewRedisWithConfig(cfg.RedisConfig)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    defer rdb.Close()
    
    // Use Redis client with configured pool size and timeouts
    // ... queue operations, caching, etc.
}
```

### Example 3: Check Configuration at Startup

```go
package main

import (
    "log"
    "github.com/vikukumar/pushpaka/internal/config"
)

func main() {
    cfg := config.Load()
    
    // Log configuration details for debugging
    log.Printf("Starting Pushpaka (Mode: %s)", cfg.AppMode)
    log.Printf("Database: %s @ %s:%d", 
        cfg.DatabaseConfig.DBName,
        cfg.DatabaseConfig.Host,
        cfg.DatabaseConfig.Port)
    
    if cfg.DatabaseConfig.SSLMode != "disable" {
        log.Printf("Database SSL: %s (cert: %s)",
            cfg.DatabaseConfig.SSLMode,
            cfg.DatabaseConfig.SSLCertFile)
    }
    
    log.Printf("Redis: %s:%d (pool size: %d)",
        cfg.RedisConfig.Host,
        cfg.RedisConfig.Port,
        cfg.RedisConfig.PoolSize)
}
```

## Using DatabaseConfig in Service Layer

### Example 4: Database Repository with SSL Support

```go
package repositories

import (
    "github.com/jmoiron/sqlx"
    "github.com/vikukumar/pushpaka/internal/config"
    "github.com/vikukumar/pushpaka/internal/database"
)

type DeploymentRepository struct {
    db *sqlx.DB
}

// NewWithCustomConfig creates a repository with custom database configuration
func NewDeploymentRepository(dbCfg *config.DatabaseConfig) (*DeploymentRepository, error) {
    db, err := database.NewPostgresWithConfig(dbCfg)
    if err != nil {
        return nil, err
    }
    
    return &DeploymentRepository{db: db}, nil
}

// GetDeployment retrieves a deployment by ID
func (r *DeploymentRepository) GetDeployment(id string) (*Deployment, error) {
    var d Deployment
    err := r.db.Get(&d, "SELECT * FROM deployments WHERE id = $1", id)
    return &d, err
}
```

### Example 5: Service Layer with Connection Management

```go
package services

import (
    "context"
    "github.com/vikukumar/pushpaka/internal/config"
    "github.com/vikukumar/pushpaka/internal/database"
    "github.com/redis/go-redis/v9"
)

type DeploymentService struct {
    repo  *DeploymentRepository
    redis *redis.Client
}

// NewDeploymentService initializes the service with database and Redis
func NewDeploymentService(cfg *config.Config) (*DeploymentService, error) {
    // Initialize database with SSL if configured
    db, err := database.NewPostgresWithConfig(cfg.DatabaseConfig)
    if err != nil {
        return nil, err
    }
    
    // Initialize Redis with custom pool configuration
    redis, err := database.NewRedisWithConfig(cfg.RedisConfig)
    if err != nil {
        return nil, err
    }
    
    repo := NewDeploymentRepository(db, cfg.DatabaseConfig)
    
    return &DeploymentService{
        repo:  repo,
        redis: redis,
    }, nil
}

// StartDeployment starts a new deployment
func (s *DeploymentService) StartDeployment(ctx context.Context, id string) error {
    // Use database connection
    deployment, err := s.repo.GetDeployment(id)
    if err != nil {
        return err
    }
    
    // Queue job in Redis
    jobID := "deploy_" + id
    return s.redis.Set(ctx, jobID, deployment, 0).Err()
}
```

## Testing with Different Configurations

### Example 6: Test with Development Configuration

```go
package tests

import (
    "os"
    "testing"
    "github.com/vikukumar/pushpaka/internal/config"
)

func TestWithDevelopmentConfig(t *testing.T) {
    // Set development environment
    os.Setenv("APP_ENV", "development")
    os.Setenv("DB_DRIVER", "sqlite")
    os.Setenv("DB_PATH", ":memory:") // In-memory SQLite for tests
    
    cfg := config.Load()
    
    // Verify development defaults
    if cfg.AppMode != "development" {
        t.Fatalf("Expected development mode, got %s", cfg.AppMode)
    }
    
    if cfg.DatabaseConfig.SSLMode != "disable" {
        t.Fatalf("Development should have SSL disabled")
    }
}
```

### Example 7: Test with Production Configuration

```go
package tests

import (
    "os"
    "testing"
    "github.com/vikukumar/pushpaka/internal/config"
)

func TestWithProductionConfig(t *testing.T) {
    // Set production environment
    os.Setenv("APP_ENV", "production")
    os.Setenv("DB_HOST", "postgres.example.com")
    os.Setenv("DB_SSL_MODE", "require")
    os.Setenv("DB_SSL_CERT_FILE", "/path/to/cert.pem")
    os.Setenv("REDIS_HOST", "redis.example.com")
    os.Setenv("REDIS_PASSWORD", "secret")
    
    cfg := config.Load()
    
    // Verify production configuration
    if cfg.AppMode != "production" {
        t.Fatalf("Expected production mode, got %s", cfg.AppMode)
    }
    
    if cfg.DatabaseConfig.SSLMode != "require" {
        t.Fatalf("Production should require SSL")
    }
    
    if cfg.DatabaseConfig.MaxOpenConns < 20 {
        t.Fatalf("Production should have adequate connection pool size")
    }
}
```

## Middleware for Configuration-Based Behavior

### Example 8: Log Middleware with Configuration

```go
package middleware

import (
    "github.com/rs/zerolog/log"
    "github.com/vikukumar/pushpaka/internal/config"
)

func LoggingMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Only log verbose details in development
            if cfg.AppMode == "development" {
                log.Debug().
                    Str("method", r.Method).
                    Str("path", r.RequestURI).
                    Msg("HTTP request")
            } else {
                log.Info().
                    Str("method", r.Method).
                    Str("path", r.RequestURI).
                    Msg("HTTP request")
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Example 9: Health Check Endpoint

```go
package handlers

import (
    "net/http"
    "github.com/vikukumar/pushpaka/internal/config"
    "github.com/vikukumar/pushpaka/internal/database"
)

// HealthCheckHandler returns application health status
func HealthCheckHandler(cfg *config.Config, db *sqlx.DB, redis *redis.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        health := map[string]interface{}{
            "status": "ok",
            "mode":   cfg.AppMode,
        }
        
        // Check database
        if err := db.Ping(); err != nil {
            health["database"] = "error: " + err.Error()
            w.WriteHeader(http.StatusServiceUnavailable)
        } else {
            health["database"] = "ok"
        }
        
        // Check Redis
        ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
        defer cancel()
        
        if err := redis.Ping(ctx).Err(); err != nil {
            health["redis"] = "error: " + err.Error()
            if health["database"] == "ok" {
                w.WriteHeader(http.StatusServiceUnavailable)
            }
        } else {
            health["redis"] = "ok"
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(health)
    }
}
```

## Migration from Legacy Configuration

### Example 10: Backward Compatibility Layer

```go
package config

import (
    "os"
    "strings"
)

// MigrateLegacyConfig helps migrate from old DATABASE_URL to new DB_* variables
func MigrateLegacyConfig() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        return // New config already in use
    }
    
    // Parse legacy DATABASE_URL if new variables not set
    if os.Getenv("DB_HOST") == "" {
        // Extract host from PostgreSQL URL
        // Format: postgres://user:password@host:port/dbname?options
        
        // This is handled automatically by LoadDatabaseConfig()
        // which builds URL from DB_* variables or uses DATABASE_URL
    }
    
    log.Warn().Msg("Using legacy DATABASE_URL - please migrate to DB_* variables")
}
```

## Environment Variable Documentation

### PostgreSQL Variables
```
DB_HOST=localhost          # Server hostname
DB_PORT=5432               # Server port
DB_USER=postgres           # Username
DB_PASSWORD=               # Password (should be in secrets)
DB_NAME=pushpaka           # Database name
DB_SSL_MODE=disable        # SSL mode (disable|prefer|require|verify-ca|verify-full)
DB_SSL_CERT_FILE=          # Path to CA certificate
DB_MAX_OPEN_CONNS=10       # Max open connections
DB_MAX_IDLE_CONNS=3        # Max idle connections
DB_CONN_MAX_LIFETIME=15m   # Connection lifetime
```

### Redis Variables
```
REDIS_HOST=localhost       # Server hostname
REDIS_PORT=6379            # Server port
REDIS_PASSWORD=            # Password (if needed)
REDIS_DB=0                 # Database number
REDIS_MAX_RETRIES=3        # Retry attempts
REDIS_POOL_SIZE=5          # Connection pool size
REDIS_MIN_IDLE_CONNS=1     # Min idle connections
REDIS_POOL_TIMEOUT=4s      # Pool timeout
REDIS_READ_TIMEOUT=3s      # Read timeout
REDIS_WRITE_TIMEOUT=3s     # Write timeout
```

## Best Practices

1. **Always use `NewPostgresWithConfig()` in production** for SSL certificate support
2. **Set `APP_ENV` correctly** to get appropriate defaults for your deployment
3. **Load configuration early** in your application startup
4. **Test SSL connections locally** before deploying to production
5. **Document custom pool settings** if you override the defaults
6. **Use environment variables** for sensitive data (passwords, API keys)
7. **Never commit actual secrets** to version control - use examples files instead
8. **Monitor connection pool metrics** in production to adjust pool sizes
9. **Use structured logging** to record configuration at startup for debugging
10. **Validate configuration at startup** to fail fast if required variables are missing
