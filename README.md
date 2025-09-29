# 🔗 Blockchain Supply Chain Finance (SCF) System với PBFT Consensus

## 📋 Tổng quan

Hệ thống blockchain permissioned thế hệ mới cho Supply Chain Finance, tích hợp **PBFT Consensus Architecture** và **Multi-Peer Architecture** để đảm bảo:

- **✅ Minh bạch tuyệt đối**: Tất cả giao dịch được ghi nhận immutable trên blockchain
- **✅ Phân quyền linh hoạt**: Multi-peer architecture với role-based access
- **✅ High Throughput**: PBFT consensus cho transaction ordering và finality
- **✅ Fault Tolerance**: 3-node PBFT cluster với f=1 fault tolerance (2f+1 signatures)
- **✅ Scalability**: Horizontal scaling cho từng peer type
- **✅ No External Dependencies**: Chạy hoàn toàn local-only với gRPC communication

### 🎯 Các bên tham gia trong hệ thống SCF

| Bên tham gia | Vai trò | Peer Service | Database riêng |
|-------------|---------|-------------|---------------|
| **Anchor/Buyer** | Tạo hợp đồng SCF | `peer-anchor:8084` | `blockchain_anchor` |
| **Main Bank** | Phát hành token, phê duyệt | `peer-main-bank:8082` | `blockchain_main_bank` |
| **Supplier** | Phê duyệt, chuyển token | `peer-supplier:8083` | `blockchain_supplier` |
| **Orderer Cluster** | PBFT Consensus & Ordering | `orderer-ord1/2/3:7050/60/70` | `blockchain` (shared) |

## Mục lục

1. [Giới thiệu](#giới-thiệu)
2. [Cách chạy hệ thống](#cách-chạy-hệ-thống)
3. [Tài liệu chi tiết](#tài-liệu-chi-tiết)

## Giới thiệu

### Mục tiêu hệ thống

Hệ thống blockchain SCF permissioned với PBFT consensus được thiết kế để:

- **Minh bạch**: Tất cả giao dịch được ghi nhận trên blockchain và có thể truy vết
- **Bất biến**: Dữ liệu đã ghi không thể thay đổi nhờ PBFT cryptographic signatures
- **Phân quyền**: Chỉ các bên được ủy quyền mới có thể tham gia
- **Tự động hóa**: Quy trình phê duyệt và thực thi hợp đồng được tự động hóa
- **Consensus Guarantee**: PBFT đảm bảo transaction ordering và finality với 2f+1 signatures






## Tài liệu chi tiết

### 📊 **Kiến trúc và thiết kế hệ thống**
Chi tiết kiến trúc Multi-Peer với PBFT Consensus:
- **[SYSTEM_DIAGRAM.md](SYSTEM_DIAGRAM.md)** - Kiến trúc tổng quan, PBFT consensus flow, component diagrams, data flow, database schema, business logic

### 🔗 **API Documentation**
Luồng xử lý chi tiết cho tất cả APIs:
- **[API_Flow_Diagrams.md](API_Flow_Diagrams.md)** - Flow diagrams, business logic patterns, database collections, architecture patterns

## 🚀 Cách chạy hệ thống

### 🔐 Thiết lập Keys cho Orderer Nodes

**Quan trọng**: Trước khi chạy hệ thống, bạn cần tạo private keys cho PBFT orderer nodes.

#### **Tạo Keys tự động**

```bash
# Chạy script tạo key ECDSA PKCS8 cho 3 orderer nodes
node scripts/generate-orderer-keys.js
```

Script sẽ tạo:
- `secrets/ord1/private.pem` & `secrets/ord1/public.pem`
- `secrets/ord2/private.pem` & `secrets/ord2/public.pem`
- `secrets/ord3/private.pem` & `secrets/ord3/public.pem`

#### **Tạo Keys thủ công (tùy chọn)**

Nếu muốn tạo keys thủ công bằng OpenSSL:

```bash
# Tạo thư mục secrets
mkdir -p secrets/ord1 secrets/ord2 secrets/ord3

# Tạo ECDSA key pair cho từng orderer
for ord in ord1 ord2 ord3; do
    # Tạo private key (PKCS8 format)
    openssl ecparam -genkey -name prime256v1 -out secrets/${ord}/private.pem

    # Tạo public key
    openssl ec -in secrets/${ord}/private.pem -pubout -out secrets/${ord}/public.pem
done
```

### 📋 Yêu cầu hệ thống

| Tài nguyên | Tối thiểu | Khuyến nghị | Mục đích |
|-----------|-----------|-------------|----------|
| **Docker** | 20.10+ với Compose V2 | Latest | Container orchestration |
| **RAM** | 4GB | 8GB+ | Multi-service chạy đồng thời |
| **CPU** | 2 cores | 4 cores+ | PBFT consensus, MongoDB processing |
| **Disk** | 5GB | 10GB+ | Logs, databases, containers |
| **Network** | Stable internet | High-speed | gRPC communication |

### ⚙️ Ports cần khả dụng

| Port | Service | Mục đích |
|------|---------|----------|
| **4200** | Frontend (nginx) | Angular UI |
| **8080** | Backend (Spring Boot) | API Gateway |
| **8082** | peer-main-bank | Bank operations |
| **8083** | peer-supplier | Supplier operations |
| **8084** | peer-anchor | Anchor operations |
| **7050** | orderer-ord1 | PBFT Primary Node |
| **7060** | orderer-ord2 | PBFT Replica Node |
| **7070** | orderer-ord3 | PBFT Replica Node |
| **27017** | MongoDB | Database access |

### 🛠️ Lệnh triển khai nhanh

```bash
# 1. Clone và vào thư mục
git clone <repository-url>
cd blockchain

# 2. Tạo private keys cho orderer nodes (bắt buộc)
node scripts/generate-orderer-keys.js

# 3. Chạy tất cả services (recommended cho lần đầu)
docker-compose up --build

# 4. Hoặc chạy background (production-like)
docker-compose up -d --build

# 5. Kiểm tra health của tất cả services
docker-compose ps

# 6. Xem logs để monitor khởi động
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
ORDERER_ADDR=orderer-ord1:7050
MONGO_URI=mongodb://root:example@mongo-shared:27017/{database}

# Orderer Cluster (PBFT)
PBFT_NODE_ID=ord1
ORDERER_PORT=7050
PBFT_F=1
MONGO_URI=mongodb://root:example@mongo-shared:27017/blockchain?authSource=admin
```

#### **MongoDB Multi-Database Architecture**

Hệ thống sử dụng **data segregation** với databases riêng biệt:

```
mongo-shared:blockchain (Public Ledger - PBFT Signed)
├── blocks: PBFT signed blocks with ECDSA signatures
├── users: User authentication
└── events: Legacy event log (if needed)

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

#### **PBFT Consensus Configuration**

```yaml
# PBFT Parameters
PBFT_F=1                              # Max faulty nodes (3f+1 = 4 nodes total)
PBFT_NODE_ID=ord1|ord2|ord3         # Unique node identifier
PBFT_QUORUM=3                        # 2f+1 = 3 signatures required

# gRPC Communication
ORDERER_ADDR=orderer-ord1:7050       # Primary orderer for peers
ORDERER_PORT=7050|7060|7070         # Individual orderer ports

# Cryptographic Keys
secrets/ord*/private.pem             # ECDSA private keys (auto-generated)
secrets/ord*/public.pem              # ECDSA public keys (auto-generated)
```

### Monitoring & Debugging

#### Logs Management
```bash
# Xem logs tất cả services
docker-compose logs -f

# Xem logs service cụ thể
docker-compose logs -f backend
docker-compose logs -f orderer-ord1
docker-compose logs -f peer-anchor
docker-compose logs -f peer-main-bank
docker-compose logs -f peer-supplier
docker-compose logs -f mongo-shared

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
docker-compose exec orderer-ord1 sh
docker-compose exec peer-anchor sh
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
docker-compose logs mongo-shared

# Kiểm tra MongoDB connectivity
docker-compose exec mongo-shared mongo --username root --password example --authenticationDatabase admin
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
      - ORDERER_ADDR=orderer-ord1:7050
      - MONGO_URI=mongodb://root:example@mongo-main-bank-2:27017/blockchain_main_bank_2?authSource=admin
    depends_on:
      - mongo-shared
      - orderer-ord1
    networks:
      - peer-network
      - public-network

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

### ✅ **Hoàn thành (11/11 containers chạy ổn định)**

| Service | Port | Status | Implementation |
|---------|------|--------|----------------|
| **peer-main-bank** | 8082 | ✅ Running | Contract creation, bank approval, token issuance, gRPC client |
| **peer-supplier** | 8083 | ✅ Running | Contract approval, token transfer, balance management, gRPC client |
| **peer-anchor** | 8084 | ✅ Running | Contract creation, token reception, ledger tracking, gRPC client |
| **orderer-ord1** | 7050 | ✅ Running | PBFT primary, consensus engine, block signing, MongoDB storage |
| **orderer-ord2** | 7060 | ✅ Running | PBFT replica, consensus participant, block validation |
| **orderer-ord3** | 7070 | ✅ Running | PBFT replica, consensus participant, block validation |
| **mongo-shared** | 27017 | ✅ Running | Block storage with PBFT signatures |
| **mongo-main-bank** | - | ✅ Running | Main bank world state |
| **mongo-supplier** | - | ✅ Running | Supplier world state |
| **mongo-anchor** | - | ✅ Running | Anchor world state |
| **backend** | 8080 | ✅ Running | Spring Boot API gateway |
| **frontend** | 4200 | ✅ Running | Angular 17 UI |

### ✅ **PBFT Consensus End-to-End Test**

```bash
# Test successful - Peer gửi tx qua gRPC → Orderer PBFT consensus → Block signed → Stream về peers ✅
Transaction submitted via gRPC → PBFT consensus (Pre-Prepare→Prepare→Commit) → Block with signatures stored ✅

Sample stored block with PBFT signatures:
{
  "_id": "block_1",
  "height": 1,
  "timestamp": "2025-09-29T10:30:00Z",
  "transactions": [...],
  "previous_hash": "genesis",
  "hash": "block_hash_123...",
  "merkle_root": "merkle_456...",
  "signatures": [
    {
      "orderer_id": "ord1",
      "signature": "ecdsa_sig_1...",
      "public_key": "-----BEGIN PUBLIC KEY-----\n..."
    },
    {
      "orderer_id": "ord2",
      "signature": "ecdsa_sig_2...",
      "public_key": "-----BEGIN PUBLIC KEY-----\n..."
    },
    {
      "orderer_id": "ord3",
      "signature": "ecdsa_sig_3...",
      "public_key": "-----BEGIN PUBLIC KEY-----\n..."
    }
  ]
}
```

### 🚀 **Test Commands để verify:**

```bash
# Health checks
curl -s http://localhost:8082/health
curl -s http://localhost:8083/health
curl -s http://localhost:8084/health

# Check orderer gRPC endpoints
grpcurl -plaintext localhost:7050 list
grpcurl -plaintext localhost:7060 list
grpcurl -plaintext localhost:7070 list

# Database verification - check PBFT signed blocks
docker-compose exec mongo-shared mongosh --username root --password example --authenticationDatabase admin \
blockchain --eval "db.blocks.find().sort({height: -1}).limit(3).toArray()"

# Verify ECDSA signatures in blocks
docker-compose exec mongo-shared mongosh --username root --password example --authenticationDatabase admin \
blockchain --eval "db.blocks.find({}, {height: 1, signatures: 1}).sort({height: -1}).limit(1)"
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

## 🔐 **PBFT Consensus Overview**

### Practical Byzantine Fault Tolerance (PBFT)
PBFT là thuật toán consensus cho phép hệ thống chịu được **f faulty nodes** trong tổng số **3f+1 nodes**. Với f=1, chúng ta cần **4 nodes** để chịu được 1 node bị lỗi.

### PBFT Phases
1. **Pre-Prepare**: Primary node broadcast proposed block
2. **Prepare**: Replica nodes validate và send prepare messages
3. **Commit**: Quorum reached (2f+1), block finalized with signatures

### Cryptographic Security
- **ECDSA Signatures**: Mỗi orderer ký blocks với private key
- **Quorum Validation**: Cần 2f+1 signatures để block valid
- **Public Key Verification**: Peers verify signatures với public keys

### Advantages over Traditional Consensus
- **No External Dependencies**: Chạy hoàn toàn local-only
- **Low Latency**: Direct gRPC communication
- **Deterministic Finality**: Blocks finalized trong vòng vài giây
- **Cryptographic Proofs**: Mỗi block có cryptographic evidence

---

## 📚 **Tài liệu tham khảo**

- **[SYSTEM_DIAGRAM.md](SYSTEM_DIAGRAM.md)** - Kiến trúc chi tiết, PBFT consensus flow, diagrams, database schema
- **[API_Flow_Diagrams.md](API_Flow_Diagrams.md)** - API flow diagrams, business logic patterns

---

**🎉 Hệ thống Blockchain Supply Chain Finance đã chuyển đổi thành công sang kiến trúc PBFT-only với consensus ordering!**
