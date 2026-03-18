```mermaid
graph TB
    subgraph "Environment Sources"
        ENV[".env / .env.production.example"]
        YAML["config/config.yaml"]
        SYS["System Environment Variables"]
    end
    
    subgraph "Configuration Layer"
        LOAD["config.Load()"]
        DBCFG["LoadDatabaseConfig(mode)"]
        REDISCFG["LoadRedisConfig(mode)"]
    end
    
    subgraph "Config Objects"
        DBOBJ["DatabaseConfig{
            Host, Port, User, Password
            DBName, SSLMode, SSLCertFile
            MaxOpenConns, MaxIdleConns
            ConnMaxLifetime
        }"]
        
        REDISOBJ["RedisConfig{
            Host, Port, Password, DB
            MaxRetries, PoolSize
            MinIdleConns, PoolTimeout
            ReadTimeout, WriteTimeout
        }"]
        
        APPOBJ["Application Config{
            Database/Redis Configs
            JWT Secret, JWT Expiry
            Base URL, CORS Origins
            Log Level, App Mode
        }"]
    end
    
    subgraph "Connection Builders"
        DBURL["BuildPostgresURL()
        postgres://user:pass@host:port/db
        ?sslmode=require
        &sslrootcert=/path/to/cert"]
        
        REDISURL["BuildRedisURL()
        redis://default:pass@host:port/db"]
        
        SSLCERT["LoadSSLCertificate()
        Loads PEM certificate
        Creates tls.Config"]
    end
    
    subgraph "Database/Redis Initialization"
        PGFUNC["NewPostgresWithConfig(cfg)
        • Loads SSL certificate
        • Creates DB connection
        • Configures pool
        • Tests connection"]
        
        REDISFUNC["NewRedisWithConfig(cfg)
        • Creates Redis options
        • Sets pool parameters
        • Tests connection"]
    end
    
    subgraph "Connection Pools"
        PGPOOL["PostgreSQL Connection Pool
        Max Open: 10-30
        Max Idle: 3-10
        Lifetime: 5-15 min"]
        
        REDISPOOL["Redis Connection Pool
        Pool Size: 5-20
        Min Idle: 1-5
        Read/Write Timeout: 3-5s"]
    end
    
    subgraph "Application"
        APP["Pushpaka API Server
        • Handlers with DB connections
        • Queue workers with Redis
        • Deployment management
        • Git sync services"]
    end
    
    ENV --> LOAD
    YAML --> LOAD
    SYS --> LOAD
    
    LOAD --> DBCFG
    LOAD --> REDISCFG
    
    DBCFG --> DBOBJ
    REDISCFG --> REDISOBJ
    LOAD --> APPOBJ
    
    DBOBJ --> DBURL
    DBOBJ --> SSLCERT
    REDISOBJ --> REDISURL
    
    DBURL --> PGFUNC
    SSLCERT --> PGFUNC
    REDISURL --> REDISFUNC
    
    PGFUNC --> PGPOOL
    REDISFUNC --> REDISPOOL
    
    PGPOOL --> APP
    REDISPOOL --> APP
    APPOBJ --> APP
```

## Configuration Flow Diagram

This diagram shows how environment variables flow through Pushpaka's configuration system:

### Key Steps:

1. **Input Sources**
   - `.env` files for environment variables
   - `config.yaml` for centralized configuration
   - System environment variables (highest priority)

2. **Configuration Loading**
   - `config.Load()` reads from all sources
   - Creates `DatabaseConfig` and `RedisConfig` structures
   - Applies mode-specific defaults (dev/staging/prod)

3. **URL Building**
   - `BuildPostgresURL()` constructs PostgreSQL DSN with SSL parameters
   - `BuildRedisURL()` constructs Redis connection URL
   - SSL certificates loaded if configured

4. **Connection Initialization**
   - `NewPostgresWithConfig()` creates database connection pool
   - `NewRedisWithConfig()` creates Redis connection pool with custom settings
   - Both test connection before returning

5. **Running Application**
   - App uses initialized pools for all operations
   - Deployment management queries database
   - Queue workers use Redis for job distribution

## Environment Variable Priority (Highest to Lowest)

```
System Environment Variables (cli/shell exports)
    ↓
.env file (local override)
    ↓
.env.production.example defaults (documentation)
    ↓
Hardcoded defaults in config.go (fallback)
```

## Mode-Based Defaults

### Development Mode
- Database: SQLite (no network)
- Redis: Optional (in-process queue used)
- SSL: Disabled
- Pool Size: Small (3-5 idle, 10-5 connections)
- Log Level: Debug

### Staging Mode
- Database: PostgreSQL
- Redis: Enabled
- SSL: Prefer (SSL if available)
- Pool Size: Medium (5 idle, 20 connections)
- Log Level: Info

### Production Mode
- Database: PostgreSQL with SSL required
- Redis: Enabled with authentication
- SSL: Enforce (require or verify-full)
- Pool Size: Large (10 idle, 30 connections)
- Log Level: Warning only
