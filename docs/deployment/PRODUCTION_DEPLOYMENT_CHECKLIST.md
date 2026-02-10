# Production Deployment Checklist

**Version:** 1.0.0
**Date:** 2026-02-10
**Deployment Type:** [ ] New Deployment [ ] Update/Upgrade

---

## Pre-Deployment Phase

### Infrastructure Preparation

- [ ] **Server Provisioned**
  - OS: ________________
  - CPU Cores: ___ RAM: ___GB Disk: ___GB
  - IP Address: ________________
  - Hostname: ________________

- [ ] **DNS Configuration**
  - [ ] A record created: catalogizer.yourdomain.com → Server IP
  - [ ] AAAA record (IPv6) if applicable
  - [ ] DNS propagation verified (`nslookup catalogizer.yourdomain.com`)

- [ ] **Firewall Rules Configured**
  - [ ] Port 80 (HTTP) open
  - [ ] Port 443 (HTTPS) open
  - [ ] Port 22 (SSH) secured (key-based auth only)
  - [ ] Database port restricted to localhost/internal network
  - [ ] Redis port restricted to localhost/internal network

- [ ] **SSL/TLS Certificates Obtained**
  - [ ] Certificate authority: ________________
  - [ ] Certificate files in place
  - [ ] Certificate expiry date: ________________
  - [ ] Auto-renewal configured (if Let's Encrypt)

### Security Preparation

- [ ] **Passwords Generated**
  - [ ] Database password (32+ characters): ✓
  - [ ] Redis password (32+ characters): ✓
  - [ ] JWT secret (64+ characters, base64): ✓
  - [ ] Admin password (16+ characters): ✓
  - [ ] All passwords stored in password manager: ✓

- [ ] **API Keys Obtained**
  - [ ] TMDB API key: ________________
  - [ ] OMDB API key: ________________
  - [ ] IMDB API key (optional): ________________

- [ ] **Security Scanning Complete**
  - [ ] No HIGH severity vulnerabilities
  - [ ] Gosec scan passed
  - [ ] npm audit passed
  - [ ] Security remediation documented

### Application Preparation

- [ ] **Build Artifacts Ready**
  - [ ] Backend binary built (version: ______)
  - [ ] Frontend built (version: ______)
  - [ ] Build checksums verified
  - [ ] Build notes documented

- [ ] **Configuration Files Prepared**
  - [ ] Environment variables file (.env) ready
  - [ ] Nginx configuration ready
  - [ ] Systemd service file ready (if applicable)
  - [ ] Docker Compose file ready (if applicable)

- [ ] **Database Setup Planned**
  - [ ] Database type chosen: [ ] PostgreSQL [ ] SQLite
  - [ ] Migration scripts reviewed
  - [ ] Seed data prepared (if any)

### Testing & Validation

- [ ] **Pre-Deployment Tests Complete**
  - [ ] All unit tests passing (721+ tests): ✓
  - [ ] Integration tests passing (50+ tests): ✓
  - [ ] Stress tests passing (70+ tests): ✓
  - [ ] E2E tests passing (8+ tests): ✓
  - [ ] No race conditions detected: ✓

- [ ] **Staging Environment Validated**
  - [ ] Staging deployment successful
  - [ ] Staging tests passed
  - [ ] Performance validated
  - [ ] Configuration validated

---

## Deployment Phase

### Step 1: System Setup

- [ ] **System Updated**
  ```bash
  sudo apt update && sudo apt upgrade -y
  # OR
  sudo dnf update -y
  ```

- [ ] **Dependencies Installed**
  - [ ] PostgreSQL installed (if using)
  - [ ] Redis installed (if using)
  - [ ] Nginx installed
  - [ ] Certbot installed (for Let's Encrypt)
  - [ ] Git installed
  - [ ] Required build tools installed

- [ ] **Application User Created**
  ```bash
  sudo useradd -r -s /bin/bash -d /opt/catalogizer -m catalogizer
  ```

- [ ] **Directories Created**
  ```bash
  sudo mkdir -p /opt/catalogizer/{api,web,logs,data,backups}
  sudo chown -R catalogizer:catalogizer /opt/catalogizer
  ```

### Step 2: Database Setup

- [ ] **Database Installed & Running**
  ```bash
  sudo systemctl status postgresql
  # OR verify SQLite support
  sqlite3 --version
  ```

- [ ] **Database Created**
  ```bash
  # PostgreSQL
  sudo -u postgres psql -c "CREATE DATABASE catalogizer;"
  sudo -u postgres psql -c "CREATE USER catalogizer WITH ENCRYPTED PASSWORD 'xxx';"
  sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE catalogizer TO catalogizer;"
  ```

- [ ] **Migrations Run**
  ```bash
  cd /opt/catalogizer/api
  ./catalog-api migrate up
  ```

- [ ] **Database Verified**
  ```bash
  # PostgreSQL
  sudo -u postgres psql -d catalogizer -c "\dt"
  # OR SQLite
  sqlite3 /opt/catalogizer/data/catalogizer.db ".tables"
  ```

### Step 3: Application Deployment

- [ ] **Backend Deployed**
  ```bash
  sudo cp catalog-api /opt/catalogizer/api/
  sudo chown catalogizer:catalogizer /opt/catalogizer/api/catalog-api
  sudo chmod +x /opt/catalogizer/api/catalog-api
  ```

- [ ] **Frontend Deployed**
  ```bash
  sudo cp -r dist/* /opt/catalogizer/web/
  sudo chown -R catalogizer:www-data /opt/catalogizer/web
  sudo chmod -R 755 /opt/catalogizer/web
  ```

- [ ] **Configuration Deployed**
  ```bash
  sudo cp .env /opt/catalogizer/api/
  sudo chown catalogizer:catalogizer /opt/catalogizer/api/.env
  sudo chmod 600 /opt/catalogizer/api/.env
  ```

- [ ] **Configuration Validated**
  ```bash
  /opt/catalogizer/api/catalog-api --validate-config
  ```

### Step 4: Service Configuration

- [ ] **Systemd Service Installed** (if using systemd)
  ```bash
  sudo cp catalogizer-api.service /etc/systemd/system/
  sudo systemctl daemon-reload
  ```

- [ ] **Docker Compose Configured** (if using Docker)
  ```bash
  docker-compose -f docker-compose.prod.yml config
  ```

- [ ] **Service Enabled**
  ```bash
  sudo systemctl enable catalogizer-api
  # OR
  docker-compose -f docker-compose.prod.yml up -d
  ```

### Step 5: Reverse Proxy Setup

- [ ] **Nginx Configuration Deployed**
  ```bash
  sudo cp nginx-catalogizer.conf /etc/nginx/sites-available/catalogizer
  sudo ln -s /etc/nginx/sites-available/catalogizer /etc/nginx/sites-enabled/
  ```

- [ ] **Nginx Configuration Tested**
  ```bash
  sudo nginx -t
  ```

- [ ] **SSL Certificate Configured**
  ```bash
  sudo certbot --nginx -d catalogizer.yourdomain.com
  # OR copy commercial certificate
  sudo cp catalogizer.{crt,key} /etc/nginx/ssl/
  ```

- [ ] **Nginx Reloaded**
  ```bash
  sudo systemctl reload nginx
  ```

### Step 6: Service Startup

- [ ] **Backend Started**
  ```bash
  sudo systemctl start catalogizer-api
  # OR
  docker-compose -f docker-compose.prod.yml up -d api
  ```

- [ ] **Service Status Verified**
  ```bash
  sudo systemctl status catalogizer-api
  # OR
  docker-compose -f docker-compose.prod.yml ps
  ```

- [ ] **Logs Checked for Errors**
  ```bash
  sudo journalctl -u catalogizer-api -n 50
  # OR
  docker-compose -f docker-compose.prod.yml logs --tail=50 api
  ```

---

## Post-Deployment Phase

### Validation & Testing

- [ ] **Health Check Passed**
  ```bash
  curl https://catalogizer.yourdomain.com/health
  # Expected: {"status":"ok", ...}
  ```

- [ ] **Frontend Loads**
  ```bash
  curl -I https://catalogizer.yourdomain.com
  # Expected: HTTP/2 200
  ```

- [ ] **API Responds**
  ```bash
  curl https://catalogizer.yourdomain.com/api/v1/health
  # Expected: {"status":"ok", ...}
  ```

- [ ] **Authentication Works**
  ```bash
  curl -X POST https://catalogizer.yourdomain.com/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"xxx"}'
  # Expected: JWT token returned
  ```

- [ ] **Database Connectivity Verified**
  ```bash
  sudo -u postgres psql -d catalogizer -c "SELECT COUNT(*) FROM users;"
  # Expected: Numeric result
  ```

- [ ] **WebSocket Connection Tested** (if applicable)
  - [ ] Browser developer tools show successful WebSocket connection
  - [ ] Real-time updates working

- [ ] **SSL/TLS Verified**
  ```bash
  echo | openssl s_client -connect catalogizer.yourdomain.com:443 | grep "Verify return code"
  # Expected: Verify return code: 0 (ok)
  ```

- [ ] **HTTPS Redirect Works**
  ```bash
  curl -I http://catalogizer.yourdomain.com
  # Expected: 301 redirect to https://
  ```

### Performance Validation

- [ ] **Response Time Acceptable**
  ```bash
  curl -w "@curl-format.txt" -o /dev/null -s https://catalogizer.yourdomain.com/api/v1/health
  # Expected: < 200ms
  ```

- [ ] **Load Test Passed**
  ```bash
  ab -n 1000 -c 10 https://catalogizer.yourdomain.com/api/v1/health
  # Expected: >100 req/sec, <200ms p99, 0% errors
  ```

- [ ] **Memory Usage Normal**
  ```bash
  ps aux | grep catalog-api
  # Expected: RSS < 2GB
  ```

- [ ] **CPU Usage Acceptable**
  ```bash
  top -b -n 1 | grep catalog-api
  # Expected: CPU < 50% idle
  ```

### Security Validation

- [ ] **Security Headers Present**
  ```bash
  curl -I https://catalogizer.yourdomain.com | grep -i "strict-transport"
  # Expected: HSTS, X-Frame-Options, X-Content-Type-Options headers
  ```

- [ ] **CORS Configured Correctly**
  ```bash
  curl -H "Origin: https://evil.com" https://catalogizer.yourdomain.com/api/v1/health -I
  # Expected: No CORS headers (blocked)
  ```

- [ ] **Rate Limiting Active**
  ```bash
  # Make 200 rapid requests
  for i in {1..200}; do curl https://catalogizer.yourdomain.com/api/v1/health; done
  # Expected: Some 429 Too Many Requests responses
  ```

- [ ] **Default Passwords Changed**
  - [ ] Admin password changed
  - [ ] Database password changed (not default)
  - [ ] JWT secret is unique

- [ ] **Firewall Active**
  ```bash
  sudo ufw status
  # OR
  sudo firewall-cmd --list-all
  # Expected: Only ports 80, 443, 22 open
  ```

### Monitoring & Logging

- [ ] **Logs Configured**
  - [ ] Application logs: `/opt/catalogizer/logs/api.log`
  - [ ] Error logs: `/opt/catalogizer/logs/api-error.log`
  - [ ] Nginx access logs: `/var/log/nginx/catalogizer-access.log`
  - [ ] Nginx error logs: `/var/log/nginx/catalogizer-error.log`

- [ ] **Log Rotation Configured**
  ```bash
  cat /etc/logrotate.d/catalogizer
  # Expected: Log rotation configured for all logs
  ```

- [ ] **Monitoring Alerts Configured** (if using monitoring)
  - [ ] Health check monitoring active
  - [ ] Disk space alerts configured
  - [ ] Memory alerts configured
  - [ ] CPU alerts configured
  - [ ] Error rate alerts configured

- [ ] **Health Check Script Deployed**
  ```bash
  /opt/catalogizer/scripts/health-check.sh
  # Expected: Exit code 0
  ```

- [ ] **Health Check Cron Job Added**
  ```bash
  crontab -l | grep health-check
  # Expected: Cron job present
  ```

### Backup Configuration

- [ ] **Backup Script Deployed**
  ```bash
  /opt/catalogizer/scripts/backup-db.sh
  # Expected: Backup file created
  ```

- [ ] **Backup Cron Job Configured**
  ```bash
  crontab -l | grep backup
  # Expected: Daily backup scheduled
  ```

- [ ] **Backup Destination Verified**
  - [ ] Backup directory exists and is writable
  - [ ] Backup retention policy configured (30 days default)
  - [ ] Off-site backup configured (optional)

- [ ] **Test Backup & Restore**
  - [ ] Manual backup created successfully
  - [ ] Restore tested on non-production system
  - [ ] Restore procedure documented

---

## Documentation Phase

### Update Documentation

- [ ] **Deployment Details Documented**
  - [ ] Server IP and hostname recorded
  - [ ] DNS configuration documented
  - [ ] SSL certificate details recorded
  - [ ] Deployment date and version recorded

- [ ] **Credentials Secured**
  - [ ] All passwords stored in password manager
  - [ ] SSH keys backed up securely
  - [ ] API keys documented
  - [ ] Access credentials shared with team (securely)

- [ ] **Runbook Created**
  - [ ] Service start/stop procedures
  - [ ] Backup/restore procedures
  - [ ] Rollback procedures
  - [ ] Troubleshooting guide

- [ ] **Monitoring Documented**
  - [ ] Monitoring endpoints documented
  - [ ] Alert thresholds documented
  - [ ] On-call procedures documented

### Team Handoff

- [ ] **Team Notification**
  - [ ] Development team notified
  - [ ] Operations team notified
  - [ ] Support team notified
  - [ ] Management notified

- [ ] **Access Provided**
  - [ ] SSH access provided to ops team
  - [ ] Database access provided (if needed)
  - [ ] Monitoring access provided
  - [ ] Log access provided

- [ ] **Training Completed** (if needed)
  - [ ] Team trained on deployment procedures
  - [ ] Team trained on troubleshooting
  - [ ] Team trained on rollback procedures

---

## Post-Launch Monitoring (First 24 Hours)

### Immediate Monitoring (First Hour)

- [ ] **Hour 0: Deployment Complete**
  - Time: ________________
  - All services running: ✓
  - Health checks passing: ✓

- [ ] **Hour 0+15min: First Check**
  - Health check: ✓
  - Error logs: ✓
  - Performance metrics: ✓
  - Memory usage: ✓

- [ ] **Hour 0+30min: Second Check**
  - Health check: ✓
  - Error logs: ✓
  - Performance metrics: ✓
  - Memory usage: ✓

- [ ] **Hour 0+60min: Third Check**
  - Health check: ✓
  - Error logs: ✓
  - Performance metrics: ✓
  - Memory usage: ✓

### Extended Monitoring (First 24 Hours)

- [ ] **Hour 2: Check**
  - Services stable: ✓
  - No critical errors: ✓
  - Performance acceptable: ✓

- [ ] **Hour 4: Check**
  - Services stable: ✓
  - No critical errors: ✓
  - Performance acceptable: ✓

- [ ] **Hour 8: Check**
  - Services stable: ✓
  - No critical errors: ✓
  - Performance acceptable: ✓

- [ ] **Hour 24: Final Check**
  - Services stable: ✓
  - No critical errors: ✓
  - Performance acceptable: ✓
  - Backup completed: ✓

---

## Rollback Plan (If Needed)

### Rollback Decision Criteria

Rollback should be initiated if:
- [ ] Critical security vulnerability discovered
- [ ] Data corruption detected
- [ ] Service unavailable for > 5 minutes
- [ ] Error rate > 5%
- [ ] Performance degradation > 50%

### Rollback Procedure

- [ ] **Stop Current Version**
  ```bash
  sudo systemctl stop catalogizer-api
  # OR
  docker-compose -f docker-compose.prod.yml down
  ```

- [ ] **Restore Database** (if needed)
  ```bash
  gunzip < /opt/catalogizer/backups/pre-deployment.sql.gz | \
    psql -U catalogizer catalogizer
  ```

- [ ] **Deploy Previous Version**
  ```bash
  sudo cp /opt/catalogizer/backups/catalog-api.previous /opt/catalogizer/api/catalog-api
  ```

- [ ] **Start Service**
  ```bash
  sudo systemctl start catalogizer-api
  ```

- [ ] **Verify Rollback**
  ```bash
  curl https://catalogizer.yourdomain.com/health
  ```

- [ ] **Document Rollback**
  - [ ] Rollback reason documented
  - [ ] Post-mortem scheduled
  - [ ] Lessons learned captured

---

## Sign-Off

### Deployment Team

- **Deployed By:** ________________
- **Deployment Date:** ________________
- **Deployment Time:** ________________
- **Deployment Version:** ________________
- **Signature:** ________________

### Verification Team

- **Verified By:** ________________
- **Verification Date:** ________________
- **Verification Time:** ________________
- **All Tests Passed:** [ ] Yes [ ] No
- **Signature:** ________________

### Approval

- **Approved By:** ________________
- **Approval Date:** ________________
- **Production Ready:** [ ] Yes [ ] No
- **Signature:** ________________

---

## Notes & Issues

### Deployment Notes

```
[Add any notes about the deployment process, deviations from plan, or special considerations]
```

### Issues Encountered

```
[Document any issues encountered during deployment and their resolutions]
```

### Follow-Up Actions

```
[List any follow-up actions needed after deployment]
```

---

**Checklist Version:** 1.0.0
**Last Updated:** 2026-02-10
**Document Owner:** Catalogizer Operations Team
