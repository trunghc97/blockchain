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
curl http://localhost:9090/health  # SCF Chaincode (Smart Contract Engine)
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

## ☸️ Production Deployment với Kubernetes

### Microservice Architecture

Hệ thống được tổ chức theo kiến trúc microservice, mỗi service có thể deploy độc lập:

```
k8s-base/                   # Base manifests (namespace, config, secrets, RBAC)
├── namespace.yaml
├── configmap.yaml
├── secret.yaml
└── rbac.yaml

mongodb/k8s/               # MongoDB StatefulSet với replica set
├── statefulset.yaml
├── service.yaml
├── pvc.yaml
└── init-script-configmap.yaml

# Từng microservice có deployment.yaml riêng
orderer/                   # Orderer cluster (Raft consensus)
└── deployment.yaml        # Service + Deployment + HPA (gộp)

peer-main-bank/            # Peer Main Bank microservice
└── deployment.yaml        # Service + Deployment + HPA (gộp)

peer-supplier/             # Peer Supplier microservice
└── deployment.yaml        # Service + Deployment + HPA (gộp)

peer-anchor/               # Peer Anchor microservice
└── deployment.yaml        # Service + Deployment + HPA (gộp)

backend/                   # Backend API Gateway (Spring Boot)
└── deployment.yaml        # Service + Deployment + HPA (gộp)

frontend/                  # Frontend Angular SPA
└── deployment.yaml        # Service + Deployment + Ingress (gộp)

# Monitoring files (service-monitors.yaml, alert-rules.yaml) có thể được deploy riêng
```

### Prerequisites

- **Kubernetes Cluster**: v1.24+ (EKS, GKE, AKS, hoặc self-hosted)
- **kubectl**: v1.24+ configured cho cluster
- **Helm**: v3.8+ (optional, cho templating)
- **Storage**: Persistent volumes cho MongoDB
- **Ingress Controller**: NGINX hoặc Traefik
- **Cert-Manager**: Cho SSL certificates (optional)

```bash
# Verify cluster access
kubectl cluster-info
kubectl get nodes

# Install prerequisites
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml
```

### Quick Deploy từng Microservice

```bash
# 1. Deploy base infrastructure
kubectl apply -f k8s-base/

# 2. Deploy MongoDB
kubectl apply -f mongodb/

# 3. Deploy từng microservice độc lập
kubectl apply -f orderer/deployment.yaml           # Orderer cluster
kubectl apply -f peer-main-bank/deployment.yaml   # Main Bank peer
kubectl apply -f peer-supplier/deployment.yaml    # Supplier peer
kubectl apply -f peer-anchor/deployment.yaml      # Anchor peer
kubectl apply -f backend/deployment.yaml          # API Gateway
kubectl apply -f frontend/deployment.yaml         # Web UI

# 4. Deploy monitoring (optional)
kubectl apply -f monitoring/service-monitors.yaml
kubectl apply -f monitoring/alert-rules.yaml
```

### Independent Microservice Deployment

Mỗi microservice có thể deploy/update độc lập:

```bash
# Deploy chỉ một service
kubectl apply -f backend/deployment.yaml          # Chỉ deploy backend
kubectl apply -f peer-supplier/deployment.yaml    # Chỉ deploy supplier peer

# Update image cho một service
kubectl set image deployment/backend backend=my-registry/backend:v1.1.0

# Scale chỉ một service
kubectl scale deployment peer-main-bank --replicas=3

# Restart chỉ một service
kubectl rollout restart deployment frontend
```

# Check deployment status
kubectl get pods -n blockchain-production
kubectl get deployments -n blockchain-production
```

## 🏗️ Kiến trúc hệ thống với Chaincode Service

### SCF Chaincode Service (Smart Contract Engine)
Hệ thống đã được cập nhật để tách biệt business logic vào **SCF Chaincode Service** - một microservice riêng biệt chạy trên port 9090.

#### Các thay đổi kiến trúc:
- **SCF Chaincode Service**: Chứa toàn bộ business logic cho contracts và tokens
- **Peer Services**: Chỉ còn là REST API gateways, gọi gRPC đến chaincode service
- **Decoupled Architecture**: Business logic được tách riêng, dễ maintain và scale

#### Smart Contracts:
- **Contract Management**: Create, Approve, Finalize contracts
- **Token Management**: Issue, Transfer, Settle tokens
- **State Persistence**: Lưu trữ state trong MongoDB blockchain_private

#### gRPC Communication:
- Peer services sử dụng gRPC client để gọi chaincode methods
- Protocol buffer definitions trong `share/` directory
- High-performance internal communication

### Service Ports:
- **SCF Chaincode**: `:9090` (gRPC)
- **Peer Main Bank**: `:8082` (REST)
- **Peer Supplier**: `:8083` (REST)
- **Peer Anchor**: `:8084` (REST)
- **Backend API**: `:8080` (REST)
- **Frontend**: `:4200` (HTTP)

## 📚 Tài liệu chi tiết

- [SYSTEM_DIAGRAM.md](SYSTEM_DIAGRAM.md) - Kiến trúc hệ thống
- [API_Flow_Diagrams.md](API_Flow_Diagrams.md) - API flows
- [System_Design_Document.md](System_Design_Document.md) - Tài liệu thiết kế hệ thống chi tiết
- [deploy-k8s.sh](deploy-k8s.sh) - Script triển khai Kubernetes tự động
- [k8s-base/](k8s-base/) - Base manifests cho Kubernetes
- [mongodb/](mongodb/) - MongoDB manifests
