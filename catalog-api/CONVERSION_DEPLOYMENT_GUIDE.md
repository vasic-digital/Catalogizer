# Conversion API Production Deployment Guide

## Overview
This guide covers deployment of the Catalogizer Conversion API service in production environments.

## Prerequisites

### System Requirements
- **Go 1.24+**: Runtime environment
- **SQLite 3.36+**: Database (bundled with application)
- **FFmpeg 4.4+**: Video/audio conversion engine
- **ImageMagick 7.0+**: Image conversion engine
- **TLS Certificate**: For HTTPS in production

### Hardware Requirements
- **CPU**: 4+ cores recommended for concurrent conversions
- **RAM**: 8GB+ minimum, 16GB+ recommended
- **Storage**: Fast SSD for temporary conversion files
- **Network**: Sufficient bandwidth for source/target file access

## Installation

### 1. Install External Dependencies

#### Ubuntu/Debian
```bash
# Update package manager
sudo apt update

# Install FFmpeg
sudo apt install -y ffmpeg

# Install ImageMagick
sudo apt install -y imagemagick

# Verify installations
ffmpeg -version
convert -version
```

#### CentOS/RHEL
```bash
# Enable EPEL repository
sudo yum install -y epel-release

# Install FFmpeg
sudo yum install -y ffmpeg

# Install ImageMagick
sudo yum install -y ImageMagick

# Verify installations
ffmpeg -version
convert -version
```

#### macOS
```bash
# Using Homebrew
brew install ffmpeg
brew install imagemagick

# Verify installations
ffmpeg -version
convert -version
```

### 2. Deploy Application

#### A. Binary Deployment (Recommended)
```bash
# Build production binary
go build -ldflags="-s -w" -o catalogizer-api .

# Create production directories
sudo mkdir -p /opt/catalogizer/{bin,data,logs,temp}
sudo mkdir -p /var/log/catalogizer

# Copy binary and permissions
sudo cp catalogizer-api /opt/catalogizer/bin/
sudo chmod +x /opt/catalogizer/bin/catalogizer-api

# Create system user
sudo useradd -r -s /bin/false catalogizer
sudo chown -R catalogizer:catalogizer /opt/catalogizer
sudo chown -R catalogizer:catalogizer /var/log/catalogizer
```

#### B. Docker Deployment
```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Build application
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o catalogizer-api .

# Production image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ffmpeg imagemagick sqlite

# Create app user
RUN adduser -D -s /bin/sh catalogizer

# Copy binary and setup
WORKDIR /app
COPY --from=builder /app/catalogizer-api .
RUN chown catalogizer:catalogizer catalogizer-api

USER catalogizer

EXPOSE 8080

CMD ["./catalogizer-api"]
```

```bash
# Build and run
docker build -t catalogizer-api:latest .
docker run -d \
  --name catalogizer-api \
  -p 8080:8080 \
  -v /opt/catalogizer/data:/app/data \
  -v /opt/catalogizer/temp:/tmp \
  catalogizer-api:latest
```

## Configuration

### 1. Environment Variables
```bash
# Create production environment file
sudo tee /opt/catalogizer/.env > /dev/null <<EOF
# Server Configuration
HOST=0.0.0.0
PORT=8080
GIN_MODE=release

# Database Configuration  
DB_PATH=/opt/catalogizer/data/catalog.db

# Authentication Configuration
JWT_SECRET=$(openssl rand -base64 32)
JWT_EXPIRATION_HOURS=24
ENABLE_AUTH=true

# Conversion Configuration
TEMP_DIR=/opt/catalogizer/temp
MAX_CONCURRENT_CONVERSIONS=5
CONVERSION_TIMEOUT=3600

# Logging Configuration
LOG_LEVEL=info
LOG_FORMAT=json
LOG_FILE=/var/log/catalogizer/catalogizer.log
EOF

# Secure environment file
sudo chmod 600 /opt/catalogizer/.env
sudo chown catalogizer:catalogizer /opt/catalogizer/.env
```

### 2. Production Configuration File
```bash
sudo tee /opt/catalogizer/config.json > /dev/null <<EOF
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "read_timeout": 30,
    "write_timeout": 30,
    "idle_timeout": 120,
    "enable_cors": false,
    "enable_https": true,
    "tls_cert_file": "/opt/catalogizer/certs/server.crt",
    "tls_key_file": "/opt/catalogizer/certs/server.key"
  },
  "database": {
    "path": "/opt/catalogizer/data/catalog.db",
    "max_open_connections": 25,
    "max_idle_connections": 5,
    "enable_wal": true,
    "cache_size": -2000
  },
  "auth": {
    "jwt_secret": "\${JWT_SECRET}",
    "jwt_expiration_hours": 24,
    "enable_auth": true
  },
  "conversion": {
    "temp_dir": "/opt/catalogizer/temp",
    "max_concurrent_conversions": 5,
    "conversion_timeout_seconds": 3600,
    "ffmpeg_path": "/usr/bin/ffmpeg",
    "imagemagick_path": "/usr/bin/convert"
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "file",
    "file_path": "/var/log/catalogizer/catalogizer.log",
    "max_size": 100,
    "max_backups": 7,
    "compress": true
  }
}
EOF
```

## Service Management

### 1. Systemd Service
```bash
# Create systemd service file
sudo tee /etc/systemd/system/catalogizer-api.service > /dev/null <<EOF
[Unit]
Description=Catalogizer Conversion API
After=network.target

[Service]
Type=simple
User=catalogizer
Group=catalogizer
WorkingDirectory=/opt/catalogizer
ExecStart=/opt/catalogizer/bin/catalogizer-api
Environment=CONFIG_FILE=/opt/catalogizer/config.json
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/catalogizer/data /opt/catalogizer/temp /var/log/catalogizer

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable catalogizer-api
sudo systemctl start catalogizer-api

# Check status
sudo systemctl status catalogizer-api
```

### 2. Service Health Checks
```bash
# Check if service is running
sudo systemctl is-active catalogizer-api

# Check service logs
sudo journalctl -u catalogizer-api -f

# Check API health
curl -k https://localhost:8080/health

# Check conversion formats endpoint
curl -k -H "Authorization: Bearer <JWT_TOKEN>" \
  https://localhost:8080/api/v1/conversion/formats
```

## TLS/SSL Configuration

### 1. Generate Self-Signed Certificate (Testing)
```bash
# Create certificates directory
sudo mkdir -p /opt/catalogizer/certs
sudo chown catalogizer:catalogizer /opt/catalogizer/certs

# Generate self-signed certificate
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /opt/catalogizer/certs/server.key \
  -out /opt/catalogizer/certs/server.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Set permissions
sudo chmod 600 /opt/catalogizer/certs/server.key
sudo chmod 644 /opt/catalogizer/certs/server.crt
sudo chown catalogizer:catalogizer /opt/catalogizer/certs/*
```

### 2. Let's Encrypt Certificate (Production)
```bash
# Install certbot
sudo apt install -y certbot

# Generate certificate
sudo certbot certonly --standalone -d your-domain.com

# Copy certificates to application
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem \
  /opt/catalogizer/certs/server.crt
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem \
  /opt/catalogizer/certs/server.key

# Set permissions
sudo chown catalogizer:catalogizer /opt/catalogizer/certs/*
sudo chmod 600 /opt/catalogizer/certs/server.key
```

## Database Migration

### 1. Initial Setup
```bash
# As catalogizer user, run initial migration
sudo -u catalogizer /opt/catalogizer/bin/catalogizer-api -migrate-only

# Verify database creation
sudo -u catalogizer sqlite3 /opt/catalogizer/data/catalog.db \
  "SELECT name FROM sqlite_master WHERE type='table';"

# Check conversion_jobs table
sudo -u catalogizer sqlite3 /opt/catalogizer/data/catalog.db \
  "PRAGMA table_info(conversion_jobs);"
```

### 2. Backup Strategy
```bash
# Create backup script
sudo tee /opt/catalogizer/backup.sh > /dev/null <<'EOF'
#!/bin/bash
BACKUP_DIR="/opt/catalogizer/backups"
DB_PATH="/opt/catalogizer/data/catalog.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"
sqlite3 "$DB_PATH" ".backup $BACKUP_DIR/catalogizer_$DATE.db"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "catalogizer_*.db" -mtime +7 -delete
EOF

sudo chmod +x /opt/catalogizer/backup.sh
sudo chown catalogizer:catalogizer /opt/catalogizer/backup.sh

# Add to cron (daily at 2 AM)
echo "0 2 * * * /opt/catalogizer/backup.sh" | sudo crontab -u catalogizer -
```

## Monitoring and Logging

### 1. Log Rotation
```bash
# Create logrotate configuration
sudo tee /etc/logrotate.d/catalogizer > /dev/null <<EOF
/var/log/catalogizer/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 catalogizer catalogizer
    postrotate
        systemctl reload catalogizer-api
    endscript
}
EOF
```

### 2. Monitoring Setup
```bash
# Create health check script
sudo tee /opt/catalogizer/health_check.sh > /dev/null <<'EOF'
#!/bin/bash
API_URL="https://localhost:8080/health"
RESPONSE=$(curl -k -s -o /dev/null -w "%{http_code}" "$API_URL")

if [ "$RESPONSE" -eq 200 ]; then
    echo "OK - API is healthy"
    exit 0
else
    echo "CRITICAL - API is not responding (HTTP $RESPONSE)"
    exit 2
fi
EOF

sudo chmod +x /opt/catalogizer/health_check.sh
sudo chown catalogizer:catalogizer /opt/catalogizer/health_check.sh
```

## Performance Tuning

### 1. Database Optimization
```sql
-- Create additional indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_conversion_jobs_user_status 
ON conversion_jobs(user_id, status);

CREATE INDEX IF NOT EXISTS idx_conversion_jobs_created_status 
ON conversion_jobs(created_at, status);
```

### 2. Conversion Performance
```bash
# Update configuration for high-throughput systems
sudo tee /opt/catalogizer/config.json > /dev/null <<EOF
{
  "conversion": {
    "temp_dir": "/opt/catalogizer/temp",
    "max_concurrent_conversions": 10,
    "conversion_timeout_seconds": 1800,
    "ffmpeg_threads": 4,
    "enable_gpu_acceleration": true
  }
}
EOF
```

## Security Considerations

### 1. File System Security
```bash
# Secure temporary directory
sudo chmod 750 /opt/catalogizer/temp
sudo chown catalogizer:catalogizer /opt/catalogizer/temp

# Secure database directory
sudo chmod 750 /opt/catalogizer/data
sudo chown catalogizer:catalogizer /opt/catalogizer/data

# Configure AppArmor/SELinux if enabled
# (Distribution-specific commands)
```

### 2. Network Security
```bash
# Firewall configuration (UFW example)
sudo ufw allow 8080/tcp
sudo ufw enable

# Rate limiting with nginx (reverse proxy)
sudo tee /etc/nginx/sites-available/catalogizer-api > /dev/null <<EOF
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    
    location /api/v1/conversion/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # SSL configuration
    ssl_certificate /opt/catalogizer/certs/server.crt;
    ssl_certificate_key /opt/catalogizer/certs/server.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
}
EOF
```

## Troubleshooting

### Common Issues

#### 1. FFmpeg Not Found
```bash
# Error: "ffmpeg: command not found"
# Solution:
which ffmpeg
sudo apt install -y ffmpeg  # or appropriate package manager
```

#### 2. Permission Denied on Database
```bash
# Error: "permission denied"
# Solution:
sudo chown -R catalogizer:catalogizer /opt/catalogizer/data
sudo chmod 750 /opt/catalogizer/data
```

#### 3. JWT Secret Not Set
```bash
# Error: "JWT secret not configured"
# Solution:
echo "JWT_SECRET=$(openssl rand -base64 32)" | sudo tee -a /opt/catalogizer/.env
sudo systemctl restart catalogizer-api
```

#### 4. High Memory Usage
```bash
# Solution: Adjust concurrent conversions
# In config.json, reduce max_concurrent_conversions
sudo systemctl restart catalogizer-api
```

## Scaling Considerations

### Horizontal Scaling
- Use load balancer (nginx/HAProxy)
- Shared storage for temp files (NFS/GlusterFS)
- External database (PostgreSQL/MySQL)
- Distributed job queue (Redis/RabbitMQ)

### Vertical Scaling
- Increase CPU cores for concurrent conversions
- Add SSD storage for I/O performance  
- Increase RAM for larger file handling
- GPU acceleration for video conversions

## Maintenance

### Regular Tasks
1. **Daily**: Monitor logs, check disk space
2. **Weekly**: Review conversion performance metrics
3. **Monthly**: Update dependencies, security patches
4. **Quarterly**: Performance tuning, capacity planning

### Updates and Patches
```bash
# Stop service
sudo systemctl stop catalogizer-api

# Backup database
/opt/catalogizer/backup.sh

# Update binary
sudo cp catalogizer-api-new /opt/catalogizer/bin/catalogizer-api
sudo chmod +x /opt/catalogizer/bin/catalogizer-api

# Run migrations if needed
sudo -u catalogizer /opt/catalogizer/bin/catalogizer-api -migrate-only

# Start service
sudo systemctl start catalogizer-api
```

---

**Deployment Guide Version**: 1.0  
**Last Updated**: November 27, 2025  
**API Version**: v1  
**Compatibility**: Catalogizer v1.0+