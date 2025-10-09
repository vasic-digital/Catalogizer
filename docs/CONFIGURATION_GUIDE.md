# Catalogizer v3.0 - Configuration Management Guide

## Table of Contents
1. [Overview](#overview)
2. [Configuration Wizard](#configuration-wizard)
3. [Manual Configuration](#manual-configuration)
4. [Configuration Profiles](#configuration-profiles)
5. [System Configuration](#system-configuration)
6. [Environment Variables](#environment-variables)
7. [Database Configuration](#database-configuration)
8. [Storage Configuration](#storage-configuration)
9. [Security Configuration](#security-configuration)
10. [Performance Tuning](#performance-tuning)
11. [Monitoring and Logging](#monitoring-and-logging)
12. [Backup and Recovery](#backup-and-recovery)
13. [Troubleshooting](#troubleshooting)

## Overview

Catalogizer v3.0 provides multiple ways to configure your media management system:

- **Configuration Wizard**: Interactive setup for new installations
- **Configuration Profiles**: Save and manage multiple configurations
- **Manual Configuration**: Direct editing of configuration files
- **Environment Variables**: Override settings via environment
- **API Configuration**: Dynamic configuration through REST API

## Configuration Wizard

The Configuration Wizard provides an interactive way to set up Catalogizer for the first time or create new configurations.

### Starting the Wizard

#### Basic Installation
```bash
# Start basic installation wizard
curl -X POST http://localhost:8080/api/config/wizard/start \
  -H "Content-Type: application/json" \
  -d '{
    "config_type": "basic",
    "quick_install": false
  }'
```

#### Enterprise Installation
```bash
# Start enterprise installation wizard
curl -X POST http://localhost:8080/api/config/wizard/start \
  -H "Content-Type: application/json" \
  -d '{
    "config_type": "enterprise",
    "quick_install": false
  }'
```

#### Development Environment
```bash
# Start development environment wizard
curl -X POST http://localhost:8080/api/config/wizard/start \
  -H "Content-Type: application/json" \
  -d '{
    "config_type": "development",
    "quick_install": true
  }'
```

### Wizard Steps

The wizard guides you through these configuration steps:

1. **System Requirements Check**
   - Verify minimum system requirements
   - Check available tools and dependencies
   - Auto-fix common issues

2. **Database Configuration**
   - Choose database type (SQLite, MySQL, PostgreSQL)
   - Configure connection parameters
   - Test database connectivity

3. **Media Storage Configuration**
   - Select storage type (Local, S3, WebDAV, FTP)
   - Configure storage paths and credentials
   - Set file size and access limits

4. **Security Configuration**
   - Generate JWT secrets
   - Configure session timeouts
   - Set up two-factor authentication
   - Define password policies

5. **Administrator Account**
   - Create initial admin user
   - Set admin credentials
   - Configure admin permissions

6. **Service Configuration**
   - Set server port and host
   - Configure logging levels
   - Enable CORS and security features
   - Set up backup schedules

7. **Final Testing**
   - Test all configurations
   - Verify service connectivity
   - Start services automatically

### Getting Current Step
```bash
curl -X GET "http://localhost:8080/api/config/wizard/{session_id}/step"
```

### Submitting Step Data
```bash
curl -X POST "http://localhost:8080/api/config/wizard/{session_id}/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "server_port": 8080,
    "log_level": "info",
    "enable_cors": true
  }'
```

### Checking Progress
```bash
curl -X GET "http://localhost:8080/api/config/wizard/{session_id}/progress"
```

## Manual Configuration

### Configuration File Structure

The main configuration file is located at `config/config.json`:

```json
{
  "version": "3.0.0",
  "generated_at": "2024-01-15T10:30:00Z",
  "generated_by": "configuration_wizard",
  "configuration": {
    "server": {
      "port": 8080,
      "host": "0.0.0.0",
      "ssl_enabled": false,
      "ssl_cert_path": "",
      "ssl_key_path": "",
      "cors_enabled": true,
      "cors_origins": ["*"],
      "request_timeout": "30s",
      "max_request_size": "100MB"
    },
    "database": {
      "type": "sqlite",
      "connection_string": "./catalogizer.db",
      "max_connections": 10,
      "max_idle_connections": 5,
      "connection_lifetime": "1h",
      "auto_migrate": true
    },
    "storage": {
      "type": "local",
      "path": "./media",
      "max_file_size": "100MB",
      "allowed_extensions": [".jpg", ".png", ".mp4", ".pdf"],
      "thumbnail_enabled": true,
      "thumbnail_sizes": [200, 400, 800]
    },
    "security": {
      "jwt_secret": "your-secret-key-here",
      "jwt_expiration": "24h",
      "password_min_length": 8,
      "password_require_uppercase": true,
      "password_require_lowercase": true,
      "password_require_numbers": true,
      "password_require_symbols": true,
      "two_factor_enabled": false,
      "session_timeout": "24h",
      "max_login_attempts": 5,
      "lockout_duration": "15m"
    },
    "logging": {
      "level": "info",
      "format": "json",
      "output": "file",
      "file_path": "./logs/catalogizer.log",
      "max_size": "100MB",
      "max_age": "30d",
      "max_backups": 10,
      "compress": true
    },
    "features": {
      "analytics_enabled": true,
      "favorites_enabled": true,
      "conversion_enabled": true,
      "sync_enabled": true,
      "stress_testing_enabled": false,
      "error_reporting_enabled": true,
      "log_management_enabled": true
    },
    "performance": {
      "worker_count": 4,
      "queue_size": 1000,
      "cache_enabled": true,
      "cache_size": "512MB",
      "cache_ttl": "1h"
    }
  }
}
```

### Validating Configuration

```bash
# Validate configuration file
curl -X POST http://localhost:8080/api/config/validate \
  -H "Content-Type: application/json" \
  -d @config/config.json
```

### Reloading Configuration

```bash
# Reload configuration without restart
curl -X POST http://localhost:8080/api/config/reload
```

## Configuration Profiles

Configuration profiles allow you to save and manage multiple configurations for different environments or use cases.

### Creating a Profile

```bash
curl -X POST http://localhost:8080/api/config/profiles \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "profile_id": "production-config",
    "name": "Production Configuration",
    "description": "Optimized configuration for production environment",
    "configuration": {
      "server": {
        "port": 443,
        "ssl_enabled": true
      },
      "database": {
        "type": "postgresql",
        "connection_string": "postgresql://user:pass@localhost/catalogizer"
      }
    },
    "tags": ["production", "ssl", "postgresql"]
  }'
```

### Listing Profiles

```bash
curl -X GET http://localhost:8080/api/config/profiles \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Loading a Profile

```bash
curl -X GET "http://localhost:8080/api/config/profiles/production-config" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Activating a Profile

```bash
curl -X POST "http://localhost:8080/api/config/profiles/production-config/activate" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Deleting a Profile

```bash
curl -X DELETE "http://localhost:8080/api/config/profiles/production-config" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## System Configuration

Individual system settings can be managed through the API:

### Setting Configuration Values

```bash
# Set a configuration value
curl -X POST http://localhost:8080/api/config/system \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "server.port",
    "value": "8080",
    "type": "number",
    "description": "Server port number"
  }'
```

### Getting Configuration Values

```bash
# Get a specific configuration value
curl -X GET "http://localhost:8080/api/config/system/server.port" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get all configuration values
curl -X GET http://localhost:8080/api/config/system \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Deleting Configuration Values

```bash
curl -X DELETE "http://localhost:8080/api/config/system/custom.setting" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Environment Variables

Environment variables can override configuration file settings:

### Core Settings

```bash
# Server configuration
export CATALOGIZER_PORT=8080
export CATALOGIZER_HOST=0.0.0.0
export CATALOGIZER_SSL_ENABLED=false

# Database configuration
export CATALOGIZER_DB_TYPE=sqlite
export CATALOGIZER_DB_CONNECTION="./catalogizer.db"

# Storage configuration
export CATALOGIZER_STORAGE_TYPE=local
export CATALOGIZER_STORAGE_PATH="./media"

# Security configuration
export CATALOGIZER_JWT_SECRET="your-secret-key-here"
export CATALOGIZER_SESSION_TIMEOUT=24h

# Logging configuration
export CATALOGIZER_LOG_LEVEL=info
export CATALOGIZER_LOG_PATH="./logs"

# Feature toggles
export CATALOGIZER_ANALYTICS_ENABLED=true
export CATALOGIZER_CONVERSION_ENABLED=true
export CATALOGIZER_SYNC_ENABLED=true
```

### External Service Configuration

```bash
# AWS S3 configuration
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
export S3_BUCKET_NAME="catalogizer-media"

# WebDAV configuration
export WEBDAV_URL="https://webdav.example.com"
export WEBDAV_USERNAME="username"
export WEBDAV_PASSWORD="password"

# SMTP configuration
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT=587
export SMTP_USERNAME="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"

# Slack notifications
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."

# Firebase Crashlytics
export FIREBASE_CRASHLYTICS_URL="https://crashlytics.googleapis.com/..."
export FIREBASE_TOKEN="your-firebase-token"
```

### Development Settings

```bash
# Development mode
export CATALOGIZER_ENV=development
export CATALOGIZER_DEBUG=true
export CATALOGIZER_HOT_RELOAD=true

# Testing configuration
export CATALOGIZER_TEST_MODE=true
export CATALOGIZER_TEST_DB=":memory:"
```

## Database Configuration

### SQLite Configuration

```json
{
  "database": {
    "type": "sqlite",
    "connection_string": "./data/catalogizer.db",
    "pragma": {
      "foreign_keys": "ON",
      "journal_mode": "WAL",
      "synchronous": "NORMAL",
      "cache_size": "10000",
      "temp_store": "MEMORY"
    }
  }
}
```

### PostgreSQL Configuration

```json
{
  "database": {
    "type": "postgresql",
    "host": "localhost",
    "port": 5432,
    "database": "catalogizer",
    "username": "catalogizer_user",
    "password": "secure_password",
    "ssl_mode": "require",
    "max_connections": 20,
    "max_idle_connections": 10,
    "connection_lifetime": "1h",
    "connection_timeout": "30s"
  }
}
```

### MySQL Configuration

```json
{
  "database": {
    "type": "mysql",
    "host": "localhost",
    "port": 3306,
    "database": "catalogizer",
    "username": "catalogizer_user",
    "password": "secure_password",
    "charset": "utf8mb4",
    "collation": "utf8mb4_unicode_ci",
    "max_connections": 20,
    "max_idle_connections": 10,
    "connection_lifetime": "1h"
  }
}
```

## Storage Configuration

### Local Storage

```json
{
  "storage": {
    "type": "local",
    "path": "/var/lib/catalogizer/media",
    "permissions": "0755",
    "max_file_size": "500MB",
    "allowed_extensions": [".jpg", ".jpeg", ".png", ".gif", ".mp4", ".avi", ".pdf", ".epub"],
    "scan_interval": "1h",
    "thumbnail_generation": true,
    "thumbnail_sizes": [150, 300, 600, 1200]
  }
}
```

### Amazon S3 Storage

```json
{
  "storage": {
    "type": "s3",
    "bucket": "catalogizer-media",
    "region": "us-east-1",
    "access_key_id": "your-access-key",
    "secret_access_key": "your-secret-key",
    "endpoint": "",
    "force_path_style": false,
    "encryption": "AES256",
    "storage_class": "STANDARD",
    "cache_control": "max-age=86400"
  }
}
```

### WebDAV Storage

```json
{
  "storage": {
    "type": "webdav",
    "url": "https://webdav.example.com/catalogizer/",
    "username": "your-username",
    "password": "your-password",
    "timeout": "30s",
    "retry_attempts": 3,
    "chunk_size": "10MB"
  }
}
```

## Security Configuration

### Authentication

```json
{
  "security": {
    "authentication": {
      "jwt_secret": "your-256-bit-secret-key-here",
      "jwt_algorithm": "HS256",
      "jwt_expiration": "24h",
      "refresh_token_enabled": true,
      "refresh_token_expiration": "30d",
      "remember_me_enabled": true,
      "remember_me_duration": "30d"
    }
  }
}
```

### Password Policies

```json
{
  "security": {
    "password_policy": {
      "min_length": 8,
      "max_length": 128,
      "require_uppercase": true,
      "require_lowercase": true,
      "require_numbers": true,
      "require_symbols": true,
      "forbidden_words": ["password", "admin", "catalogizer"],
      "history_count": 5,
      "expiration_days": 90
    }
  }
}
```

### Two-Factor Authentication

```json
{
  "security": {
    "two_factor": {
      "enabled": true,
      "issuer": "Catalogizer",
      "qr_code_size": 256,
      "backup_codes_count": 10,
      "grace_period": "30d"
    }
  }
}
```

### Rate Limiting

```json
{
  "security": {
    "rate_limiting": {
      "enabled": true,
      "login_attempts": {
        "max_attempts": 5,
        "window": "15m",
        "lockout_duration": "15m"
      },
      "api_requests": {
        "max_requests": 1000,
        "window": "1h",
        "burst_size": 100
      }
    }
  }
}
```

## Performance Tuning

### Worker Configuration

```json
{
  "performance": {
    "workers": {
      "count": 8,
      "queue_size": 1000,
      "timeout": "30s",
      "retry_attempts": 3,
      "retry_backoff": "exponential"
    }
  }
}
```

### Caching

```json
{
  "performance": {
    "cache": {
      "enabled": true,
      "type": "memory",
      "size": "512MB",
      "ttl": "1h",
      "cleanup_interval": "10m",
      "compression": true
    }
  }
}
```

### Database Optimization

```json
{
  "performance": {
    "database": {
      "connection_pool_size": 20,
      "max_idle_connections": 10,
      "connection_lifetime": "1h",
      "query_timeout": "30s",
      "prepared_statements": true,
      "batch_size": 1000
    }
  }
}
```

## Monitoring and Logging

### Logging Configuration

```json
{
  "logging": {
    "level": "info",
    "format": "json",
    "structured": true,
    "outputs": [
      {
        "type": "file",
        "path": "./logs/catalogizer.log",
        "max_size": "100MB",
        "max_age": "30d",
        "max_backups": 10,
        "compress": true
      },
      {
        "type": "console",
        "level": "warn"
      }
    ],
    "fields": {
      "service": "catalogizer",
      "version": "3.0.0",
      "instance_id": "auto"
    }
  }
}
```

### Metrics Collection

```json
{
  "monitoring": {
    "metrics": {
      "enabled": true,
      "endpoint": "/metrics",
      "interval": "15s",
      "retention": "7d",
      "custom_metrics": true
    },
    "health_check": {
      "enabled": true,
      "endpoint": "/health",
      "interval": "30s",
      "timeout": "5s"
    }
  }
}
```

### Error Reporting

```json
{
  "monitoring": {
    "error_reporting": {
      "enabled": true,
      "sentry_dsn": "https://your-sentry-dsn",
      "slack_webhook": "https://hooks.slack.com/services/...",
      "email_notifications": true,
      "rate_limit": 100
    }
  }
}
```

## Backup and Recovery

### Automatic Backups

```json
{
  "backup": {
    "enabled": true,
    "schedule": "0 2 * * *",
    "retention": {
      "daily": 7,
      "weekly": 4,
      "monthly": 12
    },
    "compression": true,
    "encryption": true,
    "destinations": [
      {
        "type": "local",
        "path": "./backups"
      },
      {
        "type": "s3",
        "bucket": "catalogizer-backups",
        "prefix": "database/"
      }
    ]
  }
}
```

### Creating Manual Backup

```bash
# Create configuration backup
curl -X POST http://localhost:8080/api/config/backup \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "backup_name": "pre-upgrade-backup",
    "backup_type": "manual",
    "compress": true,
    "encrypt": true
  }'
```

### Restoring from Backup

```bash
# List available backups
curl -X GET http://localhost:8080/api/config/backups \
  -H "Authorization: Bearer $JWT_TOKEN"

# Restore from backup
curl -X POST "http://localhost:8080/api/config/backups/{backup_id}/restore" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Troubleshooting

### Common Configuration Issues

#### Database Connection Errors

```bash
# Test database connection
curl -X POST http://localhost:8080/api/config/test/database \
  -H "Content-Type: application/json" \
  -d '{
    "type": "postgresql",
    "host": "localhost",
    "port": 5432,
    "database": "catalogizer",
    "username": "user",
    "password": "pass"
  }'
```

#### Storage Access Errors

```bash
# Test storage configuration
curl -X POST http://localhost:8080/api/config/test/storage \
  -H "Content-Type: application/json" \
  -d '{
    "type": "local",
    "path": "/var/lib/catalogizer/media"
  }'
```

#### Permission Issues

```bash
# Check file permissions
ls -la /var/lib/catalogizer/
chmod 755 /var/lib/catalogizer/media
chown -R catalogizer:catalogizer /var/lib/catalogizer/
```

### Configuration Validation

```bash
# Validate current configuration
curl -X GET http://localhost:8080/api/config/validate

# Check configuration schema
curl -X GET http://localhost:8080/api/config/schema
```

### Debugging Configuration

```bash
# Enable debug logging
export CATALOGIZER_LOG_LEVEL=debug

# View current configuration
curl -X GET http://localhost:8080/api/config/current \
  -H "Authorization: Bearer $JWT_TOKEN"

# Check configuration sources
curl -X GET http://localhost:8080/api/config/sources \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Configuration Reset

```bash
# Reset to default configuration
curl -X POST http://localhost:8080/api/config/reset \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "confirm": true,
    "backup_current": true
  }'
```

## Best Practices

### Security Best Practices

1. **Use Strong Secrets**: Generate random JWT secrets with at least 256 bits
2. **Enable SSL**: Use HTTPS in production environments
3. **Limit Permissions**: Run services with minimal required permissions
4. **Regular Updates**: Keep dependencies and system packages updated
5. **Monitor Access**: Enable audit logging and monitor for suspicious activity

### Performance Best Practices

1. **Database Tuning**: Optimize connection pools and query performance
2. **Caching Strategy**: Implement appropriate caching for frequently accessed data
3. **Resource Limits**: Set appropriate memory and CPU limits
4. **Background Processing**: Use workers for CPU-intensive tasks
5. **Monitoring**: Implement comprehensive monitoring and alerting

### Operational Best Practices

1. **Configuration Management**: Use version control for configuration files
2. **Environment Separation**: Maintain separate configurations for dev/staging/prod
3. **Backup Strategy**: Implement regular automated backups
4. **Documentation**: Keep configuration documentation up to date
5. **Testing**: Test configuration changes in non-production environments first

For more advanced configuration topics, see the [Deployment Guide](DEPLOYMENT_GUIDE.md) and [Troubleshooting Guide](TROUBLESHOOTING_GUIDE.md).