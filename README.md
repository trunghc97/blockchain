# 🔗 Blockchain Supply Chain Finance (SCF) System

## 📋 Tổng quan

Hệ thống blockchain permissioned thế hệ mới cho Supply Chain Finance, tích hợp **Kafka Event-Driven Architecture** và **Multi-Peer Architecture** để đảm bảo:

- **✅ Minh bạch tuyệt đối**: Tất cả giao dịch được ghi nhận immutable trên blockchain
- **✅ Phân quyền linh hoạt**: Multi-peer architecture với role-based access
- **✅ High Throughput**: Kafka messaging cho event streaming và async processing
- **✅ Fault Tolerance**: Orderer cluster với consensus mechanism
- **✅ Scalability**: Horizontal scaling cho từng peer type

### 🎯 Các bên tham gia trong hệ thống SCF

| Bên tham gia | Vai trò | Peer Service | Database riêng |
|-------------|---------|-------------|---------------|
| **Anchor/Buyer** | Tạo hợp đồng SCF | `peer-anchor:8084` | `blockchain_anchor` |
| **Main Bank** | Phát hành token, phê duyệt | `peer-main-bank:8082` | `blockchain_main_bank` |
| **Supplier** | Phê duyệt, chuyển token | `peer-supplier:8083` | `blockchain_supplier` |
| **Orderer Cluster** | Ordering & Consensus | `orderer-ord1/2/3` | `blockchain` (shared) |

## Mục lục

1. [Giới thiệu](#giới-thiệu)
2. [Cách chạy hệ thống](#cách-chạy-hệ-thống)
3. [Tài liệu chi tiết](#tài-liệu-chi-tiết)

## Giới thiệu

### Mục tiêu hệ thống

Hệ thống blockchain SCF permissioned được thiết kế để:

- **Minh bạch**: Tất cả giao dịch được ghi nhận trên blockchain và có thể truy vết
- **Bất biến**: Dữ liệu đã ghi không thể thay đổi
- **Phân quyền**: Chỉ các bên được ủy quyền mới có thể tham gia
- **Tự động hóa**: Quy trình phê duyệt và thực thi hợp đồng được tự động hóa






## Tài liệu chi tiết

### 📊 **Kiến trúc và thiết kế hệ thống**
Chi tiết kiến trúc Multi-Peer với Kafka Event-Driven:
- **[SYSTEM_DIAGRAM.md](SYSTEM_DIAGRAM.md)** - Kiến trúc tổng quan, component diagrams, data flow, database schema, business logic

### 🔗 **API Documentation**
Luồng xử lý chi tiết cho tất cả APIs:
- **[API_Flow_Diagrams.md](API_Flow_Diagrams.md)** - Flow diagrams, business logic patterns, database collections, architecture patterns

## 🚀 Cách chạy hệ thống

### 📋 Yêu cầu hệ thống

| Tài nguyên | Tối thiểu | Khuyến nghị | Mục đích |
|-----------|-----------|-------------|----------|
| **Docker** | 20.10+ với Compose V2 | Latest | Container orchestration |
| **RAM** | 4GB | 8GB+ | Multi-service chạy đồng thời |
| **CPU** | 2 cores | 4 cores+ | Kafka, MongoDB processing |
| **Disk** | 5GB | 10GB+ | Logs, databases, containers |
| **Network** | Stable internet | High-speed | Kafka messaging |

### ⚙️ Ports cần khả dụng

| Port | Service | Mục đích |
|------|---------|----------|
| **4200** | Frontend (nginx) | Angular UI |
| **8080** | Backend (Spring Boot) | API Gateway |
| **8082** | peer-main-bank | Bank operations |
| **8083** | peer-supplier | Supplier operations |
| **8084** | peer-anchor | Anchor operations |
| **27017** | MongoDB | Database access |
| **9092** | Kafka | External monitoring |
| **2181** | Zookeeper | Kafka coordination |

### 🛠️ Lệnh triển khai nhanh

```bash
# 1. Clone và vào thư mục
git clone <repository-url>
cd blockchain

# 2. Chạy tất cả services (recommended cho lần đầu)
docker-compose up --build

# 3. Hoặc chạy background (production-like)
docker-compose up -d --build

# 4. Kiểm tra health của tất cả services
docker-compose ps

# 5. Xem logs để monitor khởi động
docker-compose logs -f
```

### 🌐 Truy cập hệ thống

#### **Web Interfaces**
| Interface | URL | Credentials | Chức năng |
|-----------|-----|-------------|-----------|
| **Frontend** | http://localhost:4200 | Xem bảng dưới | Role-based UI |
| **API Docs** | http://localhost:8080/swagger-ui.html | - | REST API documentation |

#### **Direct Service Access**
| Service | URL | Mục đích |
|---------|-----|----------|
| **Backend API** | http://localhost:8080/api/v1/** | REST endpoints |
| **Peer Main Bank** | http://localhost:8082/** | Bank operations |
| **Peer Supplier** | http://localhost:8083/** | Supplier operations |
| **Peer Anchor** | http://localhost:8084/** | Anchor operations |
| **MongoDB** | localhost:27017 | Database queries |

### 👥 Tài khoản test

| Role | Username | Password | Quyền truy cập | UI Path |
|------|----------|----------|----------------|---------|
| **Anchor** | `anchor` | `123456` | Tạo hợp đồng SCF | `/anchor` |
| **Bank** | `bank` | `123456` | Phát hành token, phê duyệt | `/bank` |
| **Supplier 1-10** | `supplier1` | `123456` | Phê duyệt, chuyển token | `/supplier` |

### 🔧 Cấu hình nâng cao

#### **Environment Variables**

Services sử dụng environment variables được định nghĩa trong `docker-compose.yml`:

```yaml
# Backend (Spring Boot)
PEER_ANCHOR_URL=http://peer-anchor:8084
PEER_MAIN_BANK_URL=http://peer-main-bank:8082
PEER_SUPPLIER_URL=http://peer-supplier:8083

# Peer Services (Go)
KAFKA_BROKERS=kafka:29092
MONGO_URI=mongodb://root:example@mongo-shared:27017/{database}

# Orderer Cluster
ORDERER_NODE_ID=ord1
SCF_CHANNEL_TOPIC=scf-channel-tx
```

#### **MongoDB Multi-Database Architecture**

Hệ thống sử dụng **data segregation** với databases riêng biệt:

```
mongo-shared:blockchain (Public Ledger)
├── events: Global event log
├── blocks: Ordered blocks
└── users: User authentication

mongo-main-bank:blockchain_main_bank (Private)
├── contracts: Bank-approved contracts
├── tokens: Bank-issued tokens
└── balances: Bank token balances

mongo-supplier:blockchain_supplier (Private)
├── contracts: Supplier contracts
├── tokens: Supplier tokens
└── balances: Supplier balances

mongo-anchor:blockchain_anchor (Private)
├── contracts: Anchor-created contracts
└── ledger: Contract history
```

#### **Kafka Topics Configuration**

```yaml
# Transaction Topics (Events)
SCF_CHANNEL_TX=scf-channel-tx        # SCF events (contracts, transfers)
AUDIT_CHANNEL_TX=audit-channel-tx    # Bank approval events

# Block Topics (Ordered Results)
SCF_CHANNEL_BLOCKS=scf-channel-blocks      # Ordered SCF blocks
AUDIT_CHANNEL_BLOCKS=audit-channel-blocks  # Ordered audit blocks
```

### Monitoring & Debugging

#### Logs Management
```bash
# Xem logs tất cả services
docker-compose logs -f

# Xem logs service cụ thể
docker-compose logs -f backend
docker-compose logs -f ms-blockchain
docker-compose logs -f mongo

# Xem logs với timestamp
docker-compose logs -f --timestamps

# Xuất logs ra file
docker-compose logs > logs.txt
```

#### Container Management
```bash
# Restart service cụ thể
docker-compose restart backend

# Rebuild và restart service
docker-compose up --build --force-recreate backend

# Truy cập container shell
docker-compose exec backend bash
docker-compose exec ms-blockchain sh
```

### Troubleshooting

#### Common Issues

**1. Port conflicts**
```bash
# Kiểm tra port đang sử dụng
lsof -i :4200
lsof -i :8080
lsof -i :8081
lsof -i :27017

# Thay đổi port trong docker-compose.yml nếu cần
```

**2. MongoDB connection failed**
```bash
# Kiểm tra MongoDB container
docker-compose logs mongo

# Kiểm tra MongoDB connectivity
docker-compose exec mongo mongo --username root --password example --authenticationDatabase admin
```

**3. Services không start được**
```bash
# Kiểm tra dependencies
docker-compose ps

# Restart với clean state
docker-compose down -v
docker-compose up --build
```

**4. Frontend build failed**
```bash
# Clear npm cache
docker-compose exec frontend npm cache clean --force

# Rebuild frontend
docker-compose up --build frontend
```

#### Performance Tuning

**Memory Issues:**
```bash
# Tăng memory limit cho Docker
docker system info

# Cấu hình Docker Desktop với 8GB+ RAM
```

**Storage Issues:**
```bash
# Kiểm tra disk usage
docker system df

# Clean up unused resources
docker system prune -a --volumes
```

### Production Deployment

#### Security Considerations
- Thay đổi default passwords trong `docker-compose.yml`
- Sử dụng secrets management cho sensitive data
- Configure HTTPS với reverse proxy
- Implement rate limiting và CORS policies

#### Scaling

##### **MongoDB Scaling**
- MongoDB có thể scale horizontally với replica sets
- Mỗi peer database có thể có replica riêng

##### **Peer Services Scaling: Thêm Node Main-Bank mới**

Hệ thống hỗ trợ **horizontal scaling** cho peer services. Dưới đây là cách thêm **peer-main-bank-2**:

###### **Bước 1: Thêm services vào docker-compose.yml**

```yaml
# Thêm vào cuối file docker-compose.yml
  peer-main-bank-2:
    build: ./peer-main-bank
    ports:
      - "8085:8082"  # External port 8085, internal port 8082
    environment:
      - PEER_NODE_TYPE=main-bank-2
      - PEER_NODE_ID=main-bank-peer-2
      - PEER_PORT=8082
      - PEER_GRPC_PORT=9095
      - KAFKA_BROKERS=kafka:29092
      - SCF_CHANNEL_TOPIC=scf-channel-tx
      - AUDIT_CHANNEL_TOPIC=audit-channel-tx
      - ORDERER_ADDR=orderer-ord1:7050
      - MONGO_URI=mongodb://root:example@mongo-main-bank-2:27017/blockchain_main_bank_2?authSource=admin
    depends_on:
      - mongo-shared
      - kafka
      - orderer-ord1
    networks:
      - peer-network
      - kafka-network

  mongo-main-bank-2:
    image: mongo:latest
    container_name: blockchain-mongo-main-bank-2
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: example
    volumes:
      - ./init-mongo-bank-2.js:/docker-entrypoint-initdb.d/init-mongo-bank-2.js:ro
      - ./data/mongodb-main-bank-2:/data/db
    networks:
      - peer-network
```

###### **Bước 2: Tạo init script**

Tạo file `init-mongo-bank-2.js`:

```javascript
db = db.getSiblingDB('blockchain_main_bank_2');
db.createCollection('contracts');
db.createCollection('tokens');
db.createCollection('balances');
db.createCollection('events');

db.contracts.createIndex({ "_id": 1 });
db.tokens.createIndex({ "_id": 1 });
db.balances.createIndex({ "tokenId": 1, "account": 1 });

print("Main Bank 2 database initialized");
```

###### **Bước 3: Cập nhật Backend Load Balancing**

Cập nhật `BlockchainService.java` để support multiple bank nodes:

```java
@Value("${peer.main-bank.urls:http://peer-main-bank:8082,http://peer-main-bank-2:8082}")
private String[] mainBankUrls;

private int currentBankIndex = 0;

public Map<String, Object> approveContractByBank(...) {
    String bankUrl = mainBankUrls[currentBankIndex++ % mainBankUrls.length];
    // Use selected bank URL
}
```

###### **Bước 4: Deploy**

```bash
# Build và start new bank node
docker-compose up --build peer-main-bank-2 mongo-main-bank-2

# Test connectivity
curl http://localhost:8085/health
```

##### **Lợi ích Multi-Bank Architecture**
- ✅ **High Availability**: Fault tolerance
- ✅ **Load Balancing**: Round-robin distribution
- ✅ **Regulatory Compliance**: Data isolation
- ✅ **Geographic Scaling**: Multi-region deployment

## 🎯 **Trạng thái triển khai hiện tại**

### ✅ **Hoàn thành (13/13 containers chạy ổn định)**

| Service | Port | Status | Implementation |
|---------|------|--------|----------------|
| **peer-main-bank** | 8082 | ✅ Running | Contract creation, bank approval, token issuance, Kafka producer |
| **peer-supplier** | 8083 | ✅ Running | Contract approval, token transfer, balance management, Kafka producer |
| **peer-anchor** | 8084 | ✅ Running | Contract creation, token reception, ledger tracking, Kafka producer |
| **orderer-ord1** | 7050 | ✅ Running | Kafka consumer, event processing, MongoDB storage |
| **orderer-ord2** | 7060 | ✅ Running | Kafka consumer, block ordering |
| **orderer-ord3** | 7070 | ✅ Running | Kafka consumer, block ordering |
| **kafka** | 9092 | ✅ Running | Message broker với topics `scf-channel-tx`, `audit-channel-tx` |
| **zookeeper** | 2181 | ✅ Running | Kafka coordination |
| **mongo-shared** | 27017 | ✅ Running | Event storage database |
| **mongo-main-bank** | - | ✅ Running | Main bank world state |
| **mongo-supplier** | - | ✅ Running | Supplier world state |
| **mongo-anchor** | - | ✅ Running | Anchor world state |
| **backend** | 8080 | ✅ Running | Spring Boot API gateway |
| **frontend** | 4200 | ✅ Running | Angular 17 UI |

### ✅ **Kafka Messaging End-to-End Test**

```bash
# Test successful - Orderer nhận và lưu events vào database
Event published to Kafka → Orderer consumed → Database stored ✅

Sample stored event:
{
  "_id": "68d603246d3df5837a38bf1a",
  "eventType": "TEST_DATABASE_STORAGE",
  "eventId": "bde11db426d42bc237841bd6145a1283",
  "data": {"contractId": "TEST003", "amount": 3000},
  "timestamp": "1758855971",
  "processed": true,
  "ordererId": "ord1"
}
```

### 🚀 **Test Commands để verify:**

```bash
# Health checks
curl -s http://localhost:8082/health
curl -s http://localhost:8083/health
curl -s http://localhost:8084/health

# Kafka messaging test
echo '{"eventType": "TEST", "timestamp": "'$(date +%s)'", "data": {"message": "test"}}' | \
docker-compose exec -T kafka kafka-console-producer --bootstrap-server localhost:9092 --topic scf-channel-tx

# Database verification
docker-compose exec mongo-shared mongosh --username root --password example --authenticationDatabase admin \
blockchain --eval "db.events.find().toArray()"
```

### Cleanup Commands

```bash
# Dừng tất cả services
docker-compose down

# Dừng và xóa tất cả data (reset hoàn toàn)
docker-compose down -v

# Xóa images (free disk space)
docker-compose down --rmi all

# Deep cleanup
docker system prune -a --volumes --force
```

---

## 📚 **Tài liệu tham khảo**

- **[SYSTEM_DIAGRAM.md](SYSTEM_DIAGRAM.md)** - Kiến trúc chi tiết, diagrams, database schema
- **[API_Flow_Diagrams.md](API_Flow_Diagrams.md)** - API flow diagrams, business logic patterns

---

**🎉 Hệ thống Blockchain Supply Chain Finance đã sẵn sàng với đầy đủ functionality!**
