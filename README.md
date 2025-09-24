# Blockchain Supply Chain Finance (SCF) System

Hệ thống blockchain permissioned cho Supply Chain Finance, cho phép các bên tham gia (Buyer, Bank, Supplier) thực hiện các giao dịch tài trợ chuỗi cung ứng một cách minh bạch và bất biến.

## Mục lục

1. [Giới thiệu](#giới-thiệu)
2. [Kiến trúc tổng quan](#kiến-trúc-tổng-quan)
3. [Mô hình dữ liệu](#mô-hình-dữ-liệu)
4. [Luồng nghiệp vụ](#luồng-nghiệp-vụ)
5. [Sơ đồ Sequence](#sơ-đồ-sequence)
6. [API chính](#api-chính)
7. [Frontend](#frontend)
8. [Thiết kế blockchain](#thiết-kế-blockchain)
9. [Cách chạy hệ thống](#cách-chạy-hệ-thống)

## Giới thiệu

### Mục tiêu hệ thống

Hệ thống blockchain SCF permissioned được thiết kế để:

- **Minh bạch**: Tất cả giao dịch được ghi nhận trên blockchain và có thể truy vết
- **Bất biến**: Dữ liệu đã ghi không thể thay đổi
- **Phân quyền**: Chỉ các bên được ủy quyền mới có thể tham gia
- **Tự động hóa**: Quy trình phê duyệt và thực thi hợp đồng được tự động hóa

### Các bên tham gia

- **Buyer (Người mua)**: Khởi tạo hợp đồng SCF, yêu cầu tài trợ
- **Bank (Ngân hàng)**: Cung cấp tài trợ, phê duyệt hợp đồng
- **Supplier (Người bán)**: Nhận tài trợ, phê duyệt hợp đồng

## Kiến trúc tổng quan

Hệ thống được xây dựng theo kiến trúc microservices với 4 thành phần chính:

### Frontend (Angular 17)
- **Công nghệ**: Angular 17, Angular Material, Bootstrap 5.3, TypeScript 5.2
- **Chức năng**: Giao diện người dùng đa vai trò cho các bên tham gia (Anchor, Bank, Supplier)
- **Components**: Contract forms, approval workflows, token management, ledger viewer
- **Port**: 4200 (served qua nginx proxy)

### Backend (Spring Boot 3.1.5)
- **Công nghệ**: Java 17, Spring Boot 3.1.5, Spring Security (JWT), Spring Data MongoDB, Lombok
- **Chức năng**:
  - Authentication & Authorization với JWT tokens
  - Contract management với file upload support
  - User role management (Anchor, Bank, Supplier)
  - REST API gateway tới blockchain service
  - Error handling và logging
- **Port**: 8080

### ms-blockchain (Go 1.21)
- **Công nghệ**: Go 1.21, Gorilla Mux (HTTP router), MongoDB driver, CORS middleware
- **Chức năng**:
  - Blockchain ledger management với event sourcing
  - Token issuance và transfer logic
  - Automatic block building (mỗi 10 giây)
  - RESTful API cho contract và token operations
  - Immutable audit trail qua blockchain blocks
- **Port**: 8081

### MongoDB (Latest)
- **Chức năng**: Cơ sở dữ liệu lưu trữ tất cả dữ liệu với authentication
- **Collections**:
  - `users`: Thông tin người dùng và roles (Anchor, Bank, Supplier)
  - `contracts`: Metadata hợp đồng SCF với file attachments
  - `tokens`: Thông tin tokens được phát hành từ contracts
  - `balances`: Số dư token của từng account
  - `events`: Sự kiện blockchain chưa được include vào block
  - `blocks`: Blocks của blockchain (immutable ledger)
- **Port**: 27017

### Docker Compose
Tất cả services được orchestrate bởi Docker Compose với network chung `blockchain-network` để đảm bảo kết nối an toàn giữa các services.

## Mô hình dữ liệu

### Contracts Collection
Lưu trữ metadata của hợp đồng SCF:
```json
{
  "contractId": "string",
  "description": "string",
  "buyer": "string",
  "suppliers": [
    {
      "supplierId": "string",
      "name": "string",
      "allocatedAmount": 100000.0,
      "status": "PENDING|READY_TO_EXECUTE|EXECUTED|FAILED"
    }
  ],
  "totalAmount": 500000.0,
  "status": "PENDING|PARTIALLY_APPROVED|READY_TO_EXECUTE|EXECUTED",
  "fileUrl": "string",
  "createdAt": "timestamp",
  "updatedAt": "timestamp",
  "history": [...]
}
```

### Events Collection
Lưu trữ các sự kiện blockchain chưa được include vào block:
```json
{
  "eventId": "string",
  "contractId": "string",
  "type": "CREATE|APPROVE_SUPPLIER|EXECUTE",
  "actorId": "string",
  "payload": {...},
  "timestamp": "timestamp",
  "included": false
}
```

### Blocks Collection
Lưu trữ các block của blockchain (immutable):
```json
{
  "blockNumber": 1,
  "timestamp": "timestamp",
  "events": [...],
  "prevHash": "string",
  "hash": "string",
  "merkleRoot": "string"
}
```

### Tokens Collection
Lưu trữ thông tin về các token được phát hành từ hợp đồng:
```json
{
  "id": "string",
  "contractId": "string",
  "symbol": "string",
  "totalSupply": 500000.0,
  "issuer": "string", // Bank ID
  "owner": "string", // Current owner (initially Bank, then Supplier)
  "createdAt": "timestamp"
}
```

### Balances Collection
Lưu trữ số dư token của từng account:
```json
{
  "tokenId": "string",
  "account": "string", // User ID
  "balance": 500000.0,
  "lastUpdated": "timestamp"
}
```

### Mối quan hệ dữ liệu

- **Contracts → Tokens**: Mỗi contract phát hành một token
- **Tokens → Balances**: Mỗi token có balances cho các accounts
- **Contracts ↔ Events**: Mỗi contract có history events
- **Events → Blocks**: Events được gom nhóm thành blocks
- **Blocks**: Chuỗi liên kết qua hash, tạo thành immutable ledger
- **World State**: Trạng thái hiện tại được derive từ events trong blocks

## Luồng nghiệp vụ

### 1. Anchor tạo hợp đồng
1. Anchor đăng nhập và tạo hợp đồng mới với file đính kèm
2. Hệ thống tự động phát hành token cho Bank (issuer = Bank, owner = Bank)
3. Tạo balance record cho Bank với toàn bộ số lượng token
4. Contract được lưu với status `PENDING`, token được tạo và gán cho Bank

### 2. Suppliers phê duyệt hợp đồng
1. Suppliers xem danh sách hợp đồng cần phê duyệt
2. Mỗi supplier approve → tạo event `APPROVE_SUPPLIER`
3. Khi tất cả suppliers đã approve:
   - Contract status = `READY_TO_EXECUTE`
   - Token ownership chuyển từ Bank sang Supplier
   - Balance cập nhật: Bank = 0, Supplier = amount

### 3. Token management & transfer
1. Supplier có thể chuyển token cho Supplier khác
2. Bank có thể xem tất cả tokens đã phát hành
3. Mọi giao dịch token được ghi lại như events và include vào blocks
4. Balances được cập nhật real-time sau mỗi transaction

### 4. Block builder tạo blocks
1. Block builder chạy định kỳ mỗi 10 giây
2. Gom tối đa 10 events chưa included
3. Tạo block mới với:
   - Merkle root từ SHA256 các event IDs
   - Block hash = SHA256(prevHash + merkleRoot + timestamp)
4. Đánh dấu events đã included

### 5. Immutable ledger
- Blocks được append-only, không thể modify
- Ledger = chuỗi blocks + events bên trong
- World state được tính toán từ events trong tất cả blocks

## Sơ đồ Sequence

Dưới đây là sơ đồ sequence mô tả luồng tương tác đầy đủ của hệ thống SCF:

```mermaid
sequenceDiagram
    participant A as Anchor
    participant S as Supplier
    participant B as Bank
    participant FE as Frontend (Angular)
    participant BE as Backend (Spring Boot)
    participant BC as Blockchain Service (Go)
    participant DB as MongoDB

    %% 1. Tạo hợp đồng & Token
    rect rgb(240, 248, 255)
        Note over A,DB: Tạo hợp đồng + Token tự động
        A->>FE: Submit contract form (with file)
        FE->>BE: POST /api/contracts
        BE->>DB: Save contract metadata
        BE->>BC: POST /contract/create
        BC->>DB: Insert CREATE event
        BC->>DB: Auto-create token (issuer=Bank)
        BC->>DB: Create balance (Bank=totalAmount)
        BC-->>BE: Contract + Token created
        BE-->>FE: Success response
        FE-->>A: Contract created
    end

    %% 2. Phê duyệt & Transfer Token
    rect rgb(255, 248, 240)
        Note over S,DB: Supplier phê duyệt
        S->>FE: Click approve button
        FE->>BE: POST /api/contracts/{id}/approve
        BE->>BC: POST /contract/approve
        BC->>DB: Check contract status
        BC->>DB: Insert APPROVE_SUPPLIER event

        alt All suppliers approved
            BC->>DB: Transfer token (Bank → Supplier)
            BC->>DB: Update balances (Bank=0, Supplier=amount)
            BC->>DB: Update contract.approved = true
        end

        BC-->>BE: Success response
        BE-->>FE: Updated contract
        FE-->>S: Approval confirmed
    end

    %% 3. Token Transfer
    rect rgb(248, 255, 240)
        Note over S,DB: Supplier chuyển token
        S->>FE: Initiate token transfer
        FE->>BE: POST /api/tokens/transfer
        BE->>BC: POST /token/transfer
        BC->>DB: Check sender balance & ownership
        BC->>DB: Debit sender balance
        BC->>DB: Credit receiver balance
        BC->>DB: Update token ownership
        BC->>DB: Insert TRANSFER event
        BC-->>BE: Transfer successful
        BE-->>FE: Success response
        FE-->>S: Transfer confirmed
    end

    %% 4. Block building (chạy định kỳ)
    rect rgb(255, 240, 248)
        Note over BC,DB: Block Builder (mỗi 10s)
        loop Every 10 seconds
            BC->>DB: Find unincluded events
            alt Events found (max 10)
                BC->>BC: Calculate Merkle root
                BC->>BC: Calculate block hash
                BC->>DB: Insert new block
                BC->>DB: Mark events as included
            end
        end
    end

    %% 5. Bank xem tokens
    rect rgb(240, 255, 248)
        Note over B,DB: Bank xem tokens đã phát hành
        B->>FE: View issued tokens
        FE->>BE: GET /api/tokens/issued/{bankId}
        BE->>BC: GET /token/issued/{bankId}
        BC->>DB: Query tokens by issuer
        DB-->>BC: Token list with owners
        BC-->>BE: Formatted response
        BE-->>FE: Token data
        FE-->>B: Display token overview
    end
```

## API chính

### Backend APIs (Spring Boot)

#### Tạo hợp đồng
```
POST /api/contracts
Content-Type: multipart/form-data
- file: MultipartFile (optional)
- contract: JSON string

Response: Contract object
```

#### Lấy danh sách hợp đồng
```
GET /api/contracts
Authorization: Bearer {token}

Response: Array of contracts
```

#### Phê duyệt hợp đồng
```
POST /api/contracts/{id}/approve
Authorization: Bearer {token}

Response: Updated contract
```

#### Lấy ledger của hợp đồng
```
GET /api/contracts/{id}/ledger
Authorization: Bearer {token}

Response: Ledger data với events và blocks
```

#### Lấy thông tin token
```
GET /api/tokens/{id}
Authorization: Bearer {token}

Response: Token information
```

#### Chuyển token
```
POST /api/tokens/transfer
Authorization: Bearer {token}
{
  "tokenId": "string",
  "fromUserId": "string",
  "toUserId": "string",
  "amount": number
}

Response: Transfer result
```

#### Lấy tokens đã phát hành bởi Bank
```
GET /api/tokens/issued/{bankId}
Authorization: Bearer {token}

Response: Array of tokens issued by bank
```

### Blockchain APIs (Go service)

#### Tạo contract event
```
POST /contract/create
{
  "contractId": "string",
  "description": "string",
  "buyer": "string",
  "suppliers": [...],
  "totalAmount": number
}

Response: {"contractId": "string", "status": "success"}
```

#### Phê duyệt contract
```
POST /contract/approve
{
  "contractId": "string",
  "supplierId": "string"
}

Response: {"status": "success"}
```

#### Lấy ledger của contract
```
GET /contract/{id}/ledger

Response: {
  "contractId": "string",
  "events": [...],
  "total": number
}
```

#### Lấy danh sách blocks
```
GET /ledger/blocks

Response: Array of block info
```

#### Lấy thông tin token
```
GET /token/{id}

Response: Token details
```

#### Chuyển token
```
POST /token/transfer
{
  "tokenId": "string",
  "fromUserId": "string",
  "toUserId": "string",
  "amount": number
}

Response: {"status": "success"}
```

#### Lấy tokens đã phát hành bởi Bank
```
GET /token/issued/{bankId}

Response: Array of tokens issued by bank
```

#### Lấy tất cả tokens
```
GET /tokens

Response: Array of all tokens
```

#### Lấy balances của account
```
GET /balances/account/{accountId}

Response: Array of balances for account
```

#### Lấy balances của token
```
GET /balances/token/{tokenId}

Response: Array of balances for token
```

#### Lấy danh sách suppliers
```
GET /suppliers

Response: Array of supplier information
```

#### Cập nhật block hashes
```
POST /blocks/hash/update

Response: Hash update result
```

## Frontend

### Cấu trúc Components
Giao diện được tổ chức theo role-based navigation với các components chính:

#### Core Components
- **Login Component** (`/login`): Xác thực người dùng với JWT
- **Navbar Component**: Navigation menu theo role (Anchor/Bank/Supplier)
- **Auth Guard**: Bảo vệ routes theo quyền truy cập

#### Role-based Components

##### Anchor Components (`/anchor/*`)
- **Contract Form** (`contract-form/`): Tạo hợp đồng mới với file upload
- **Contract Token Management** (`contract-token-management/anchor/`): Quản lý tokens đã tạo

##### Bank Components (`/bank/*`)
- **Contract Token Management** (`contract-token-management/bank/`): Xem và quản lý tokens đã phát hành

##### Supplier Components (`/supplier/*`)
- **Contract Token Management** (`contract-token-management/supplier/`): Quản lý tokens sở hữu
- **Transfer Form** (`transfer-form/`): Chuyển token cho supplier khác

#### Shared Components
- **Contract Approval** (`contract-approval/`): Interface phê duyệt hợp đồng
- **Contract Status** (`contract-status/`): Dashboard trạng thái hợp đồng
- **Ledger Viewer** (`ledger-viewer/`): Hiển thị blockchain ledger
- **Status List** (`status-list/`): Component hiển thị danh sách

### Routing Structure
```
/ → /login (redirect if not authenticated)
/login → LoginComponent
/anchor → ContractFormComponent (Anchor only)
/bank → BankTokenManagementComponent (Bank only)
/supplier → SupplierTokenManagementComponent (Supplier only)
/contracts/approve → ContractApprovalComponent
/contracts → ContractStatusComponent
/ledger → LedgerViewerComponent
```

### Models & Services
- **Auth Models**: User, LoginRequest, AuthResponse
- **Contract Models**: Contract, SupplierAllocation
- **Token Models**: Token, Balance, TransferRequest
- **Ledger Models**: Block, Event, LedgerEntry

- **Auth Service**: JWT authentication, role management
- **Contract Service**: CRUD operations cho contracts
- **Token Service**: Token transfer, balance queries
- **Ledger Service**: Blockchain data retrieval

## Scripts & Utilities

### Password Generation Script
Script tiện ích để tạo password hash cho user authentication:

```bash
# Chạy script generate password
cd scripts
node generate-password.js

# Script sử dụng bcryptjs để hash passwords
# Output: Hashed password để sử dụng trong database seeding
```

### MongoDB Initialization
File `init-mongo.js` được sử dụng để khởi tạo MongoDB với:
- Root user authentication
- Initial database setup
- Collection indexes cho performance tối ưu

## Thiết kế blockchain

### Permissioned Blockchain
- Chỉ các node được ủy quyền mới tham gia
- Mỗi bên (Bank, Buyer, Supplier) giữ bản sao ledger
- Consensus dựa trên sự đồng thuận của các bên

### Event Sourcing Architecture
- Tất cả thay đổi trạng thái được ghi thành events
- Events là nguồn sự thật duy nhất
- World state được rebuild từ events

### Block Structure
- **Block Number**: Số thứ tự block
- **Timestamp**: Thời gian tạo block
- **Events**: Mảng các events được include
- **PrevHash**: Hash của block trước
- **Hash**: SHA256(prevHash + merkleRoot + timestamp)
- **Merkle Root**: Merkle tree root của các event IDs

### Immutable Ledger
- Blocks chỉ có thể append, không modify
- Mỗi block tham chiếu đến block trước qua hash
- Tạo thành chuỗi hash liên kết bất biến

### World State Derivation
- Trạng thái hiện tại của contracts được tính từ events
- Không lưu state riêng biệt, chỉ derive khi cần
- Đảm bảo tính nhất quán và audit-able

## Cách chạy hệ thống

### Yêu cầu hệ thống
- **Docker**: version 20.10+ với Docker Compose V2
- **RAM**: Tối thiểu 4GB, khuyến nghị 8GB
- **CPU**: 2 cores trở lên
- **Disk**: 5GB free space
- **Ports**: 4200, 8080, 8081, 27017 phải khả dụng

### Cấu hình môi trường

#### Environment Variables
Các service sử dụng environment variables được định nghĩa trong `docker-compose.yml`:

```yaml
# Backend service
SPRING_DATA_MONGODB_HOST=mongo
SPRING_DATA_MONGODB_PORT=27017
SPRING_DATA_MONGODB_DATABASE=blockchain

# Blockchain service
MONGO_URI=mongodb://root:example@mongo:27017/blockchain?authSource=admin
```

#### MongoDB Configuration
- **Username**: root
- **Password**: example
- **Database**: blockchain
- **Authentication DB**: admin

### Lệnh triển khai

```bash
# 1. Clone repository
git clone <repository-url>
cd blockchain

# 2. Build và chạy tất cả services (foreground)
docker-compose up --build

# 3. Hoặc chạy background
docker-compose up -d --build

# 4. Kiểm tra trạng thái services
docker-compose ps

# 5. Xem logs real-time
docker-compose logs -f
```

### Truy cập hệ thống

| Service | URL | Mô tả |
|---------|-----|--------|
| **Frontend** | http://localhost:4200 | Giao diện người dùng |
| **Backend API** | http://localhost:8080 | REST API Spring Boot |
| **Blockchain API** | http://localhost:8081 | Blockchain service API |
| **MongoDB** | localhost:27017 | Database (external access) |

### Tài khoản test

| Role | Username | Password | Quyền truy cập |
|------|----------|----------|----------------|
| **Anchor** | `anchor` | `123456` | Tạo hợp đồng |
| **Bank** | `bank` | `123456` | Phát hành token |
| **Supplier 1-10** | `supplier1` - `supplier10` | `123456` | Phê duyệt & chuyển token |

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
- MongoDB có thể scale horizontally với replica sets
- Backend services có thể scale với load balancer
- Blockchain service hiện tại single-node, có thể extend thành multi-node

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
