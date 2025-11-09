# Production Deployment Guide

This guide covers deploying the SaaS Blueprint to production using Docker Compose.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Server Setup](#server-setup)
- [SSL/TLS Configuration](#ssltls-configuration)
- [Environment Configuration](#environment-configuration)
- [Database Migration](#database-migration)
- [Deployment](#deployment)
- [Monitoring](#monitoring)
- [Backup Strategy](#backup-strategy)
- [Scaling](#scaling)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **OS**: Ubuntu 22.04 LTS or similar Linux distribution
- **CPU**: 2+ cores recommended
- **RAM**: 4GB minimum, 8GB+ recommended
- **Storage**: 50GB+ SSD
- **Network**: Static IP address and domain name

### Software Requirements

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt install docker-compose-plugin -y

# Add current user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
docker compose version
```

## Server Setup

### 1. Create Application Directory

```bash
# Create app directory
sudo mkdir -p /opt/saas-blueprint
sudo chown $USER:$USER /opt/saas-blueprint
cd /opt/saas-blueprint

# Clone repository
git clone https://github.com/yourusername/saas-blueprint.git .
```

### 2. Configure Firewall

```bash
# Allow SSH, HTTP, HTTPS
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Optional: Prometheus, Grafana (restrict to your IP)
sudo ufw allow from YOUR_IP_ADDRESS to any port 9090
sudo ufw allow from YOUR_IP_ADDRESS to any port 3000

# Enable firewall
sudo ufw --force enable
sudo ufw status
```

## SSL/TLS Configuration

### Option 1: Let's Encrypt (Recommended)

```bash
# Install Certbot
sudo apt install certbot -y

# Stop nginx if running
docker compose down nginx

# Obtain certificate
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d api.yourdomain.com \
  --email your@email.com \
  --agree-tos

# Copy certificates to nginx directory
sudo mkdir -p nginx/ssl
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem nginx/ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem nginx/ssl/key.pem
sudo chown -R $USER:$USER nginx/ssl

# Setup auto-renewal
echo "0 0 * * 0 certbot renew --quiet && docker compose restart nginx" | sudo tee -a /etc/crontab
```

### Option 2: Self-Signed Certificate (Development/Testing)

```bash
# Generate self-signed certificate
mkdir -p nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/key.pem \
  -out nginx/ssl/cert.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
```

## Environment Configuration

### 1. Create Production Environment File

```bash
# Copy example environment file
cp .env.production.example .env.production

# Edit with your values
nano .env.production
```

### 2. Generate Secure Secrets

```bash
# Generate JWT secret (32+ characters)
openssl rand -hex 32

# Generate random passwords
openssl rand -base64 32
```

### 3. Required Environment Variables

Update `.env.production` with:

- **Database credentials**: Strong password for PostgreSQL
- **JWT_SECRET**: Long random string (32+ characters)
- **Redis password**: Strong password for Redis
- **Stripe keys**: Production API keys from Stripe Dashboard
- **Email API key**: Production API key from Resend
- **OAuth credentials**: Production OAuth app credentials
- **Domain**: Your actual domain name

## Database Migration

### 1. Initialize Database

```bash
# Start only database service
docker compose -f docker-compose.prod.yml up -d postgres

# Wait for database to be ready
docker compose -f docker-compose.prod.yml exec postgres \
  pg_isready -U postgres

# Run migrations (when migration tool is implemented)
# docker compose -f docker-compose.prod.yml run --rm backend \
#   ./server migrate up
```

### 2. Create Initial Admin User (if needed)

```bash
# Connect to database
docker compose -f docker-compose.prod.yml exec postgres \
  psql -U postgres -d saas_production

# Create admin user manually or use seed script
```

## Deployment

### 1. Build and Start Services

```bash
# Load environment variables
set -a
source .env.production
set +a

# Build images
docker compose -f docker-compose.prod.yml build

# Start all services
docker compose -f docker-compose.prod.yml up -d

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

### 2. Verify Deployment

```bash
# Check service health
docker compose -f docker-compose.prod.yml ps

# Test API health endpoint
curl https://yourdomain.com/api/v1/health

# Test frontend
curl https://yourdomain.com

# Check metrics endpoint (from allowed IP)
curl http://localhost/metrics
```

### 3. Configure DNS

Point your domain to your server's IP address:

```
A Record: yourdomain.com → YOUR_SERVER_IP
A Record: api.yourdomain.com → YOUR_SERVER_IP
```

## Monitoring

### 1. Access Prometheus

```
URL: http://YOUR_SERVER_IP:9090
```

Secure Prometheus:
```bash
# Use SSH tunnel
ssh -L 9090:localhost:9090 user@YOUR_SERVER_IP
# Access at http://localhost:9090
```

### 2. Access Grafana

```
URL: http://YOUR_SERVER_IP:3000
Username: admin
Password: (from GRAFANA_PASSWORD in .env.production)
```

Add Prometheus data source:
- URL: `http://prometheus:9090`
- Access: Server (default)

### 3. Setup Alerts (Optional)

Create alert rules in `prometheus/rules/` and configure Alertmanager.

## Backup Strategy

### 1. Database Backups

```bash
# Create backup script
cat > /opt/saas-blueprint/backup-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/saas-blueprint/backups"
mkdir -p $BACKUP_DIR
DATE=$(date +%Y%m%d_%H%M%S)

docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U postgres saas_production | \
  gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +30 -delete
EOF

chmod +x /opt/saas-blueprint/backup-db.sh

# Schedule daily backups
echo "0 2 * * * cd /opt/saas-blueprint && ./backup-db.sh" | crontab -
```

### 2. File Storage Backups

```bash
# Backup MinIO data
docker compose -f docker-compose.prod.yml exec minio \
  mc mirror /data /backup

# Or use volume backup
docker run --rm \
  -v saas_minio_data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/minio_$(date +%Y%m%d).tar.gz -C /data .
```

### 3. Configuration Backups

```bash
# Backup environment and configs
tar czf config_backup_$(date +%Y%m%d).tar.gz \
  .env.production \
  nginx/ \
  prometheus/
```

## Scaling

### Horizontal Scaling

Add more backend instances:

```yaml
# In docker-compose.prod.yml
backend:
  deploy:
    replicas: 3
    update_config:
      parallelism: 1
      delay: 10s
```

### Vertical Scaling

Adjust resource limits:

```yaml
backend:
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 2G
      reservations:
        cpus: '1'
        memory: 1G
```

### Database Scaling

For high traffic, consider:
- PostgreSQL read replicas
- Connection pooling (PgBouncer)
- Separate read/write connections

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker compose -f docker-compose.prod.yml logs backend

# Check service status
docker compose -f docker-compose.prod.yml ps

# Restart specific service
docker compose -f docker-compose.prod.yml restart backend
```

### Database Connection Issues

```bash
# Verify database is running
docker compose -f docker-compose.prod.yml exec postgres pg_isready

# Check connection from backend
docker compose -f docker-compose.prod.yml exec backend \
  wget --spider http://postgres:5432
```

### Memory Issues

```bash
# Check container memory usage
docker stats

# Increase container memory limits in docker-compose.prod.yml
```

### SSL Certificate Issues

```bash
# Verify certificate
openssl x509 -in nginx/ssl/cert.pem -text -noout

# Check nginx configuration
docker compose -f docker-compose.prod.yml exec nginx nginx -t

# Reload nginx
docker compose -f docker-compose.prod.yml exec nginx nginx -s reload
```

### High CPU/Memory Usage

```bash
# Identify resource-heavy containers
docker stats

# Check application logs
docker compose -f docker-compose.prod.yml logs --tail=100 backend

# Review Prometheus metrics
```

## Maintenance

### Updating the Application

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d

# Run migrations if needed
# docker compose -f docker-compose.prod.yml run --rm backend ./server migrate up
```

### Log Rotation

Configure Docker log rotation in `/etc/docker/daemon.json`:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

Restart Docker:
```bash
sudo systemctl restart docker
```

### Cleanup

```bash
# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Remove stopped containers
docker container prune
```

## Security Best Practices

1. **Use strong passwords** for all services
2. **Enable firewall** and restrict access
3. **Keep software updated** regularly
4. **Use SSL/TLS** for all connections
5. **Restrict database access** to backend only
6. **Enable audit logging**
7. **Regular security scans**
8. **Backup encryption** for sensitive data
9. **Rate limiting** (already configured)
10. **Monitor for suspicious activity**

## Support

For issues and questions:
- GitHub Issues: https://github.com/yourusername/saas-blueprint/issues
- Documentation: https://docs.yourdomain.com
- Email: support@yourdomain.com

## License

MIT License - see LICENSE file for details
