# 🔗 Blockchain Supply Chain Finance (SCF) System

Hệ thống blockchain permissioned cho Supply Chain Finance với PBFT Consensus và Events Sync Architecture.

## 📋 Yêu cầu hệ thống

- Docker & Docker Compose
- Node.js (cho key generation)
- 4GB+ RAM, 10GB+ disk space

## 🚀 Cách chạy hệ thống

### 1. Tạo keys cho Orderer nodes
```bash
node scripts/generate-orderer-keys.js
```

### 2. Chạy tất cả services
```bash
docker-compose up --build
```

### 3. Chạy background (production)
```bash
docker-compose up -d --build
```

## 📊 Kiểm tra trạng thái

### Health checks
```bash
# Kiểm tra tất cả services
docker-compose ps

# Health endpoints
curl http://localhost:8082/health  # Main Bank
curl http://localhost:8083/health  # Supplier
curl http://localhost:8084/health  # Anchor
curl http://localhost:8080/actuator/health  # Backend
```

### Logs
```bash
# Logs tất cả services
docker-compose logs -f

# Logs service cụ thể
docker-compose logs -f peer-anchor
docker-compose logs -f orderer-ord1
docker-compose logs -f backend
```

## 🌐 Truy cập hệ thống

| Service | URL | Mô tả |
|---------|-----|--------|
| Frontend | http://localhost:4200 | UI chính |
| Backend API | http://localhost:8080 | REST APIs |
| API Docs | http://localhost:8080/swagger-ui.html | Documentation |
| MongoDB | mongodb://localhost:27017 | Database |

### Tài khoản test
- **Anchor**: `anchor` / `123456`
- **Bank**: `bank` / `123456`
- **Supplier**: `supplier1` / `123456`

## 🛠️ Commands hữu ích

```bash
# Dừng hệ thống
docker-compose down

# Reset database
docker-compose down -v

# Rebuild service
docker-compose up --build peer-anchor

# Truy cập container
docker-compose exec peer-anchor sh
```

## 🛠️ Development Workflow

### Development Setup
```bash
# 1. Clone repository
git clone <repository-url>
cd blockchain

# 2. Generate orderer keys
node scripts/generate-orderer-keys.js

# 3. Start all services
docker-compose up -d --build

# 4. Monitor startup logs
docker-compose logs -f --tail=50
```

### Code Development
```bash
# For Go services (peer-*)
# Modify code in peer-*/ directory
# Auto-restart enabled in docker-compose.yml

# For Spring Boot (backend)
# Modify code in backend/src/main/java/
# Auto-restart enabled

# For Angular (frontend)
# Access http://localhost:4200
# Hot reload enabled via nginx
```

### Debugging
```bash
# Access container shell
docker-compose exec peer-anchor sh
docker-compose exec orderer-ord1 sh
docker-compose exec backend bash

# View real-time logs
docker-compose logs -f peer-anchor

# Check service health
curl -s http://localhost:8084/health

# Test gRPC endpoints
grpcurl -plaintext localhost:7050 list
```

### Testing
```bash
# Reset databases for testing
docker-compose exec mongo-shared mongosh --username root --password example \
  --authenticationDatabase admin blockchain_private --eval "db.dropDatabase()"

docker-compose exec mongo-shared mongosh --username root --password example \
  --authenticationDatabase admin blockchain_public --eval "db.dropDatabase()"

# Re-initialize databases
docker-compose exec mongo-shared mongosh --username root --password example \
  --authenticationDatabase admin < init-mongo.js
```

## 🚀 Production Deployment

### Prerequisites
- Docker & Docker Compose (latest versions)
- 8GB+ RAM, 20GB+ disk space
- Stable network connection
- SSL certificates (recommended)

### Security Setup
```bash
# Change default passwords in docker-compose.yml
# Use environment variables for sensitive data
# Configure SSL/TLS for all services
# Implement proper firewall rules
```

### Production Configuration
```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  # Add production-specific configurations
  # - Resource limits
  # - Health checks
  # - Logging drivers
  # - Secrets management
  # - Network isolation
```

### Deployment Steps
```bash
# 1. Clone repository
git clone <repository-url>
cd blockchain

# 2. Generate production keys
node scripts/generate-orderer-keys.js

# 3. Configure environment
cp .env.example .env
# Edit .env with production values

# 4. Build and deploy
docker-compose -f docker-compose.prod.yml up -d --build

# 5. Verify deployment
docker-compose -f docker-compose.prod.yml ps
curl -s https://your-domain/health

# 6. Setup monitoring (optional)
# - ELK stack
# - Prometheus + Grafana
# - Log aggregation
```

### Scaling
```bash
# Scale peer services
docker-compose up -d --scale peer-supplier=3

# Add load balancer
# Configure nginx or traefik for load balancing

# Database scaling
# - MongoDB replica sets
# - Sharding for high throughput
```

### Backup & Recovery
```bash
# Backup databases
docker-compose exec mongo-shared mongosh --username root --password example \
  --authenticationDatabase admin --eval "db.fsyncLock()"

# Create volume snapshots
docker run --rm -v blockchain_mongo-shared:/data -v $(pwd):/backup alpine tar czf /backup/backup.tar.gz -C /data .

# Restore procedure
docker-compose down -v
docker-compose up -d mongo-shared
docker-compose exec mongo-shared mongosh --username root --password example \
  --authenticationDatabase admin --eval "db.fsyncUnlock()"
```

### Monitoring & Maintenance
```bash
# Health monitoring
curl -s http://localhost:8082/health
curl -s http://localhost:8083/health
curl -s http://localhost:8084/health

# Resource monitoring
docker stats

# Log rotation
docker-compose logs --tail=1000 > logs_$(date +%Y%m%d).log

# Update deployment
docker-compose pull
docker-compose up -d --build
```

## 📚 Tài liệu chi tiết

- [SYSTEM_DIAGRAM.md](SYSTEM_DIAGRAM.md) - Kiến trúc hệ thống
- [API_Flow_Diagrams.md](API_Flow_Diagrams.md) - API flows
- [System_Design_Document.md](System_Design_Document.md) - Tài liệu thiết kế hệ thống chi tiết
- [k8s/README.md](k8s/README.md) - Hướng dẫn triển khai Kubernetes production
