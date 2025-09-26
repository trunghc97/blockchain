# Tài liệu Thiết kế Hệ thống Blockchain Token Trading

## Mục lục
1. [Tổng quan hệ thống](#tổng-quan-hệ-thống)
2. [Kiến trúc hệ thống](#kiến-trúc-hệ-thống)
3. [Luồng Use Case](#luồng-use-case)
4. [Luồng Sequence Diagram](#luồng-sequence-diagram)
5. [Thiết kế API](#thiết-kế-api)
6. [Thiết kế Database](#thiết-kế-database)
7. [Luồng Xử lý Nghiệp vụ](#luồng-xử-lý-nghiệp-vụ)

## Tổng quan hệ thống

### Mục đích
Hệ thống Blockchain Permissioned Network với Kafka Messaging cho Supply Chain Finance là nền tảng phân quyền sử dụng event-driven architecture. Hệ thống cho phép:
- **Kafka-based Messaging**: Event-driven communication giữa Peer nodes và Orderer cluster
- **Decoupled Architecture**: Peers publish events asynchronously, Orderers consume và order globally
- **Channel-based Topics**: SCF transactions và Audit events trên các Kafka topics riêng biệt
- **High Throughput**: Async processing với Kafka's scalable messaging infrastructure
- **Fault Tolerance**: Kafka's replication và persistence đảm bảo message durability
- **Event Sourcing**: Tất cả blockchain events được publish và consume qua Kafka

### Phạm vi
- **Event-Driven Architecture**: Kafka làm messaging backbone cho tất cả blockchain events
- **Channel Segregation**: `scf-channel-tx` cho SCF transactions, `audit-channel-tx` cho bank approvals
- **Async Processing**: Peers publish events và continue, Orderers process asynchronously
- **Message Persistence**: Kafka đảm bảo không mất events với replication và persistence
- **Consumer Groups**: Orderer nodes chia sẻ load qua Kafka consumer groups
- **Network Isolation**: Private network cho Peer+Kafka, Orderer access qua bridge

## Kiến trúc hệ thống

### Mô hình kết nối với Kafka Event-Driven Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        FE[Angular Frontend<br/>Port: 4200<br/>Single Page Application]
    end

    subgraph "API Gateway Layer"
        JAVA[Java Spring Boot<br/>Port: 8080<br/>Smart Routing & Auth]
    end

    subgraph "Kafka Messaging Layer"
        subgraph "Zookeeper"
            ZK[Zookeeper<br/>Port: 2181<br/>Coordination]
        end

        subgraph "Kafka Broker"
            KB[Kafka Broker<br/>Port: 9092<br/>Message Broker]
            subgraph "Topics"
                SCF_TX[scf-channel-tx<br/>SCF Transactions]
                AUDIT_TX[audit-channel-tx<br/>Bank Approvals]
                SCF_BLOCKS[scf-channel-blocks<br/>Ordered Blocks]
                AUDIT_BLOCKS[audit-channel-blocks<br/>Ordered Blocks]
            end
        end
    end

    subgraph "Permissioned Peer Network"
        subgraph "Peer Main Bank"
            MB_API[REST API<br/>Port: 8082]
            MB_KAFKA[Kafka Producer<br/>Audit Events]
            MB_DB[(MongoDB<br/>blockchain_main_bank)]
        end

        subgraph "Peer Supplier"
            SUP_API[REST API<br/>Port: 8083]
            SUP_KAFKA[Kafka Producer<br/>SCF Events]
            SUP_DB[(MongoDB<br/>blockchain_supplier)]
        end

        subgraph "Peer Anchor"
            ANC_API[REST API<br/>Port: 8084]
            ANC_KAFKA[Kafka Producer<br/>SCF Events]
            ANC_DB[(MongoDB<br/>blockchain_anchor)]
        end
    end

    subgraph "Orderer Cluster"
        ORD1[Orderer Node 1<br/>Port: 7050<br/>Kafka Consumer<br/>Block Ordering]
        ORD2[Orderer Node 2<br/>Port: 7060<br/>Kafka Consumer<br/>Block Ordering]
        ORD3[Orderer Node 3<br/>Port: 7070<br/>Kafka Consumer<br/>Block Ordering]
    end

    subgraph "Shared Services"
        SHARED_DB[(MongoDB Shared<br/>User Management<br/>blockchain_shared)]
    end

    subgraph "Network Topology"
        PUBLIC_NW[Public Network<br/>External Access]
        PRIVATE_NW[Private Network<br/>Peer + Kafka<br/>Isolated]
        KAFKA_BRIDGE[Kafka Bridge<br/>Orderer Access]
    end

    %% Event Flow
    FE -->|HTTP| JAVA
    JAVA -->|Smart Routing| MB_API
    JAVA -->|Smart Routing| SUP_API
    JAVA -->|Smart Routing| ANC_API

    MB_API -->|Bank Approval Events| MB_KAFKA
    SUP_API -->|SCF Events| SUP_KAFKA
    ANC_API -->|SCF Events| ANC_KAFKA

    MB_KAFKA -->|Publish| AUDIT_TX
    SUP_KAFKA -->|Publish| SCF_TX
    ANC_KAFKA -->|Publish| SCF_TX

    SCF_TX -->|Consume| ORD1
    SCF_TX -->|Consume| ORD2
    SCF_TX -->|Consume| ORD3
    AUDIT_TX -->|Consume| ORD1
    AUDIT_TX -->|Consume| ORD2
    AUDIT_TX -->|Consume| ORD3

    ORD1 -->|Ordered Blocks| SCF_BLOCKS
    ORD2 -->|Ordered Blocks| SCF_BLOCKS
    ORD3 -->|Ordered Blocks| SCF_BLOCKS

    SCF_BLOCKS -->|Broadcast| MB_API
    SCF_BLOCKS -->|Broadcast| SUP_API
    SCF_BLOCKS -->|Broadcast| ANC_API

    MB_API --> MB_DB
    SUP_API --> SUP_DB
    ANC_API --> ANC_DB
    JAVA --> SHARED_DB

    ZK -.->|Coordination| KB
    KB -.->|Topics| SCF_TX
    KB -.->|Topics| AUDIT_TX
    KB -.->|Topics| SCF_BLOCKS
    KB -.->|Topics| AUDIT_BLOCKS

    PUBLIC_NW -.->|External Access| FE
    PUBLIC_NW -.->|External Access| JAVA
    PRIVATE_NW -.->|Internal| MB_KAFKA
    PRIVATE_NW -.->|Internal| SUP_KAFKA
    PRIVATE_NW -.->|Internal| ANC_KAFKA
    KAFKA_BRIDGE -.->|Orderer Access| ORD1
    KAFKA_BRIDGE -.->|Orderer Access| ORD2
    KAFKA_BRIDGE -.->|Orderer Access| ORD3
```

### Kiến trúc Microservices với Multi-peer Network

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Client Layer                                │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                 Angular Frontend                           │    │
│  │  - Components: Contract, Token, Supplier, Bank            │    │
│  │  - Services: API calls, Auth, Data management             │    │
│  │  - UI: Material Design, Responsive                         │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                     │
                        HTTP/REST API (Port 8080)
                                     │
┌─────────────────────────────────────────────────────────────────────┐
│                     Smart API Gateway Layer                         │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │            Java Spring Boot API Gateway                    │    │
│  │  - Controllers: Contract, Token, Supplier                  │    │
│  │  - PeerRoutingService: Business logic routing              │    │
│  │  - Models: DTOs, Entities                                  │    │
│  │  - Auth: JWT validation, Role-based access                 │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                                     │
                  Smart HTTP Routing (Business Logic Based)
                                     │
┌─────────────────────────────────────────────────────────────────────┐
│                   Permissioned Peer Network                          │
│  ┌─────┬─────┬─────┐    ┌─────┬─────┬─────┐    ┌─────┬─────┐         │
│  │MB   │SUP  │ANC  │    │MB   │SUP  │ANC  │    │MB   │SUP  │ANC  │  │
│  │REST │REST │REST │    │gRPC │gRPC │gRPC │    │DB   │DB   │DB   │  │
│  │API  │API  │API  │    │Srv  │Srv  │Srv  │    │Inst │Inst │Inst │  │
│  └─────┴─────┴─────┘    └─────┴─────┴─────┘    └─────┴─────┴─────┘  │
│     Port: 8082-8084         Port: 9092-9094        Isolated DBs     │
└─────────────────────────────────────────────────────────────────────┘
                                     │
                          gRPC Cross-peer Communication
                                     │
┌─────────────────────────────────────────────────────────────────────┐
│                     Orderer Cluster (Public)                        │
│  ┌─────┬─────┬─────┐    ┌─────────────────────────────┐             │
│  │ORD1 │ORD2 │ORD3 │    │      Raft Consensus         │             │
│  │7050 │7060 │7070 │    │  - Global Block Ordering    │             │
│  └─────┴─────┴─────┘    │  - Transaction Sequencing   │             │
│                         │  - Consensus Protocol       │             │
│                         └─────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────────┘
                                     │
                           Globally Ordered Blocks
                                     │
┌─────────────────────────────────────────────────────────────────────┐
│                    Distributed World State                           │
│  ┌─────┬─────┬─────┬─────┐    ┌─────┬─────┬─────┬─────┐             │
│  │CTR  │TOK  │BAL  │EVT  │    │CTR  │TOK  │BAL  │EVT  │   ...       │
│  │Main │Main│Main │Main │    │Sup  │Sup  │Sup  │Sup  │   ...       │
│  │Bank │Bank│Bank │Bank │    │Bank │Bank │Bank │Bank │             │
│  └─────┴─────┴─────┴─────┘    └─────┴─────┴─────┴─────┘             │
│     Isolated Collections             Isolated Collections            │
└─────────────────────────────────────────────────────────────────────┘
```

### Công nghệ sử dụng với kiến trúc mới

| Layer | Technology | Version | Purpose |
|-------|------------|---------|---------|
| **Frontend** | Angular | 17 | Single Page Application với role-based components |
| **API Gateway** | Java Spring Boot | 3.1.5 | Smart routing, authentication, business logic |
| **Peer Nodes** | Golang | 1.21 | Permissioned blockchain operations, gRPC services |
| **Orderer Cluster** | Golang | 1.21 | Raft consensus, global block ordering |
| **Databases** | MongoDB | 7.x | Isolated World State databases per peer |
| **Communication** | gRPC + REST | - | Cross-peer validation và API access |
| **Container** | Docker | 24.x | Multi-service containerization |
| **Orchestration** | Docker Compose | 2.20 | Complex network topology management |
| **Networks** | Docker Networks | - | Isolated public/private/orderer networks |

## Luồng Use Case

### UC-001: Tạo hợp đồng thương mại
**Actor**: Anchor (Người mua)
**Mô tả**: Anchor tạo hợp đồng mới với danh sách suppliers và thông tin chi tiết
**Điều kiện tiên quyết**: Anchor đã đăng nhập
**Luồng chính**:
1. Anchor nhập thông tin hợp đồng
2. Hệ thống tạo ID hợp đồng
3. Lưu hợp đồng vào database
4. Ghi event và tạo block
5. Thông báo thành công

### UC-002: Ngân hàng phê duyệt hợp đồng
**Actor**: Bank (Ngân hàng)
**Mô tả**: Bank phê duyệt hợp đồng và hệ thống tự động phát hành token
**Điều kiện tiên quyết**: Hợp đồng tồn tại, chưa được bank phê duyệt
**Luồng chính**:
1. Bank chọn hợp đồng cần phê duyệt
2. Hệ thống kiểm tra quyền hạn
3. Cập nhật trạng thái bankApproved = true
4. Tính tổng giá trị hợp đồng
5. Tạo token với Issuer = "SYSTEM"
6. Tạo balance ban đầu cho Anchor
7. Ghi event và tạo block
8. Thông báo thành công

### UC-003: Supplier phê duyệt hợp đồng
**Actor**: Supplier (Người bán)
**Mô tả**: Supplier phê duyệt phần của mình trong hợp đồng
**Điều kiện tiên quyết**: Hợp đồng đã được bank phê duyệt
**Luồng chính**:
1. Supplier chọn hợp đồng cần phê duyệt
2. Hệ thống kiểm tra quyền hạn
3. Cập nhật trạng thái supplier
4. Kiểm tra tất cả suppliers đã phê duyệt
5. Nếu tất cả đã phê duyệt:
   - Phân phối token cho tất cả suppliers
   - Xóa balance của Anchor
   - Cập nhật trạng thái hợp đồng
6. Ghi event và tạo block
7. Thông báo thành công

### UC-004: Chuyển nhượng token
**Actor**: Supplier
**Mô tả**: Supplier chuyển token cho supplier khác
**Điều kiện tiên quyết**: Supplier có balance token > 0
**Luồng chính**:
1. Supplier chọn token và số lượng
2. Chọn người nhận
3. Hệ thống kiểm tra balance đủ
4. Cập nhật balance người gửi (-)
5. Cập nhật balance người nhận (+)
6. Ghi event và tạo block
7. Kiểm tra nếu Anchor hết token thì tự động hoàn tất hợp đồng

### UC-005: Tất toán token với ngân hàng
**Actor**: Supplier
**Mô tả**: Supplier tất toán toàn bộ token với ngân hàng
**Điều kiện tiên quyết**: Supplier có balance token > 0
**Luồng chính**:
1. Supplier chọn token cần tất toán
2. Hệ thống kiểm tra balance
3. Xóa toàn bộ balance của supplier
4. Ghi event tất toán
5. Tạo block
6. Thông báo thành công

### UC-006: Xem lịch sử giao dịch
**Actor**: Tất cả users
**Mô tả**: Xem tất cả events và blocks liên quan đến một hợp đồng
**Luồng chính**:
1. Chọn hợp đồng
2. Query tất cả events liên quan
3. Query blocks chứa events
4. Query balances hiện tại
5. Hiển thị ledger hoàn chỉnh

## Luồng Sequence Diagram

### SD-001: Tạo và phê duyệt hợp đồng hoàn chỉnh

```mermaid
sequenceDiagram
    participant A as Anchor
    participant FE as Frontend
    participant BE as Backend (Java)
    participant BC as Blockchain (Go)
    participant DB as MongoDB

    %% Tạo hợp đồng
    A->>FE: Nhập thông tin hợp đồng
    FE->>BE: POST /api/contracts (contract data)
    BE->>BC: POST /contract/create
    BC->>DB: Insert contract
    BC->>DB: Insert event CONTRACT_CREATED
    BC->>DB: Insert block
    BC-->>BE: Success response
    BE-->>FE: Success response
    FE-->>A: Hiển thị hợp đồng đã tạo

    %% Bank phê duyệt
    A->>FE: Yêu cầu bank phê duyệt
    FE->>BE: POST /api/contracts/{id}/approve-bank
    BE->>BC: POST /contract/{id}/approve-bank
    BC->>DB: Update contract bankApproved=true
    BC->>DB: Insert token (issuer=SYSTEM)
    BC->>DB: Insert balance (anchor)
    BC->>DB: Insert event BANK_APPROVED
    BC->>DB: Insert block
    BC-->>BE: Success + tokenId
    BE-->>FE: Success + tokenId
    FE-->>A: Hiển thị token đã tạo

    %% Suppliers phê duyệt
    A->>FE: Thông báo suppliers phê duyệt
    loop Mỗi supplier
        FE->>BE: POST /api/contracts/{id}/approve (supplierId)
        BE->>BC: POST /contract/{id}/approve
        BC->>DB: Update supplier status
        BC->>DB: Check all suppliers approved
        alt Tất cả đã approve
            BC->>DB: Update contract approved=true
            BC->>DB: Insert balances cho tất cả suppliers
            BC->>DB: Delete anchor balance
            BC->>DB: Insert event CONTRACT_FULLY_APPROVED
        else
            BC->>DB: Insert event SUPPLIER_APPROVED
        end
        BC->>DB: Insert block
        BC-->>BE: Success
        BE-->>FE: Success
    end
    FE-->>A: Hợp đồng hoàn tất
```

### SD-002: Chuyển nhượng token

```mermaid
sequenceDiagram
    participant S1 as Supplier 1
    participant FE as Frontend
    participant BE as Backend (Java)
    participant BC as Blockchain (Go)
    participant DB as MongoDB

    S1->>FE: Chọn token & số lượng, chọn người nhận
    FE->>BE: POST /api/v1/tokens/transfer
    BE->>BC: POST /token/transfer
    BC->>DB: Get sender balance
    BC->>DB: Validate balance >= amount
    BC->>DB: Update sender balance (-amount)
    BC->>DB: Get/Update receiver balance (+amount)
    BC->>DB: Insert event TOKEN_TRANSFERRED
    BC->>DB: Insert block
    BC->>DB: Check if anchor balance == 0
    alt Anchor hết token
        BC->>DB: Update contract status APPROVED
    end
    BC-->>BE: Success
    BE-->>FE: Success
    FE-->>S1: Transfer thành công
```

### SD-003: Tất toán token

```mermaid
sequenceDiagram
    participant S as Supplier
    participant FE as Frontend
    participant BE as Backend (Java)
    participant BC as Blockchain (Go)
    participant DB as MongoDB

    S->>FE: Click "Settle with Bank" cho token
    FE->>BE: POST /api/v1/tokens/settle
    BE->>BC: POST /token/settle
    BC->>DB: Get supplier balance
    BC->>DB: Validate balance exists
    BC->>DB: Delete supplier balance
    BC->>DB: Insert event TOKEN_SETTLED
    BC->>DB: Insert block
    BC-->>BE: Success
    BE-->>FE: Success
    FE-->>S: Settlement thành công
```

## Thiết kế API

### API Gateway (Java Spring Boot - Port 8080)

#### 1. Contract APIs

##### POST /api/contracts - Create Contract
**Mô tả**: Tạo hợp đồng mới

**Input**:
```json
{
  "buyer": "ANCHOR001",
  "bankId": "BANK001",
  "description": "Supply Chain Contract",
  "suppliers": [
    {
      "supplierId": "SUP001",
      "name": "Supplier 1",
      "allocatedAmount": 50000.00
    }
  ],
  "totalAmount": 50000.00
}
```

**Output**:
```json
{
  "id": "contract_1234567890",
  "status": "success"
}
```

**Mã lỗi**:
- 400: Invalid contract data
- 500: Internal server error

**Flowchart**:
```mermaid
flowchart TD
    A[Receive request] --> B[Validate input]
    B --> C[Call Go service /contract/create]
    C --> D[Return response]
```

##### POST /api/contracts/{id}/approve-bank - Bank Approve Contract
**Mô tả**: Ngân hàng phê duyệt hợp đồng

**Input**:
```json
{
  "bankId": "BANK001"
}
```

**Output**:
```json
{
  "status": "success",
  "message": "Contract approved by bank successfully",
  "tokenId": "token_contract_123"
}
```

**Mã lỗi**:
- 404: Contract not found
- 403: Bank does not have permission
- 500: Internal server error

#### 2. Token APIs

##### POST /api/v1/tokens/transfer - Transfer Token
**Mô tả**: Chuyển token giữa các tài khoản

**Input**:
```json
{
  "tokenId": "token_contract_123",
  "from": "SUP001",
  "to": "SUP002",
  "amount": 10000.00
}
```

**Output**:
```json
{
  "status": "transferred",
  "message": "Token transferred successfully"
}
```

**Mã lỗi**:
- 400: Invalid transfer data / Insufficient balance
- 404: Token not found
- 500: Internal server error

##### POST /api/v1/tokens/settle - Settle Token
**Mô tả**: Tất toán token với ngân hàng

**Input**:
```json
{
  "tokenId": "token_contract_123",
  "supplierId": "SUP001"
}
```

**Output**:
```json
{
  "status": "settled",
  "message": "Token settled successfully with bank",
  "settledAmount": 25000.00
}
```

**Mã lỗi**:
- 400: Supplier has no balance
- 404: Token not found
- 500: Internal server error

### Blockchain Service APIs (Golang - Port 8081)

#### 1. Contract APIs

##### POST /contract/create
**Input/Output**: Same as Java API

##### POST /contract/{id}/approve-bank
**Input/Output**: Same as Java API

##### POST /contract/{id}/approve
**Input**:
```json
{
  "supplierId": "SUP001"
}
```

**Output**:
```json
{
  "status": "success",
  "message": "Contract approved successfully"
}
```

#### 2. Token APIs

##### POST /token/transfer
**Input/Output**: Same as Java API

##### POST /token/settle
**Input/Output**: Same as Java API

##### GET /token/{id}
**Output**:
```json
{
  "id": "token_contract_123",
  "contractId": "contract_123",
  "symbol": "TK123",
  "total": 50000.00,
  "issuer": "SYSTEM",
  "owner": "ANCHOR001",
  "createdAt": "2024-01-01T10:00:00Z"
}
```

##### GET /tokens
**Output**:
```json
[
  {
    "id": "token_contract_123",
    "contractId": "contract_123",
    "symbol": "TK123",
    "total": 50000.00,
    "issuer": "SYSTEM",
    "owner": "ANCHOR001",
    "createdAt": "2024-01-01T10:00:00Z"
  }
]
```

#### 3. Query APIs

##### GET /contract/list
**Output**:
```json
[
  {
    "_id": "contract_123",
    "description": "Supply Chain Contract",
    "anchorId": "ANCHOR001",
    "bankId": "BANK001",
    "bankApproved": true,
    "suppliers": [...],
    "approved": true,
    "createdAt": "2024-01-01T09:00:00Z"
  }
]
```

##### GET /balances/account/{accountId}
**Output**:
```json
[
  {
    "tokenId": "token_contract_123",
    "account": "SUP001",
    "balance": 25000.00
  }
]
```

## Blockchain Technical Details

### Thuật toán Hash và Cơ chế Blockchain

#### 1. Thuật toán Hash
Hệ thống sử dụng **SHA256** làm thuật toán hash chính cho blockchain:

```go
func calculateBlockHash(blockNumber int64, timestamp, previousHash string, events []string) string {
    hashData := map[string]interface{}{
        "blockNumber":  blockNumber,
        "timestamp":    timestamp,
        "previousHash": previousHash,
        "events":       events,
    }

    jsonData, err := json.Marshal(hashData)
    if err != nil {
        return ""
    }

    hash := sha256.Sum256(jsonData)
    return hex.EncodeToString(hash[:])
}
```

**Đặc điểm của SHA256:**
- **Cryptographic Hash Function**: Một chiều, không thể reverse
- **Deterministic**: Input giống nhau luôn tạo ra output giống nhau
- **Avalanche Effect**: Thay đổi nhỏ trong input tạo ra thay đổi lớn trong output
- **Collision Resistant**: Rất khó tìm 2 inputs khác nhau tạo ra cùng hash

#### 2. Merkle Tree Implementation

Hệ thống sử dụng **Merkle Tree** để tối ưu hóa việc verify các events trong block:

```go
func (b *BlockBuilder) calculateMerkleRoot(eventIds []string) string {
    if len(eventIds) == 0 {
        return b.calculateSHA256("")
    }

    // Calculate SHA256 for each event ID
    hashes := make([]string, len(eventIds))
    for i, eventId := range eventIds {
        hashes[i] = b.calculateSHA256(eventId)
    }

    // Build merkle tree
    for len(hashes) > 1 {
        var newHashes []string
        for i := 0; i < len(hashes); i += 2 {
            left := hashes[i]
            right := ""
            if i+1 < len(hashes) {
                right = hashes[i+1]
            } else {
                right = left // Duplicate last hash if odd number
            }
            newHashes = append(newHashes, b.calculateSHA256(left+right))
        }
        hashes = newHashes
    }

    return hashes[0]
}
```

**Lợi ích của Merkle Tree:**
- **Efficient Verification**: Có thể verify một event mà không cần toàn bộ events
- **Data Integrity**: Phát hiện được thay đổi trong bất kỳ event nào
- **Compact Representation**: Merkle root đại diện cho tất cả events

#### 3. Cách nối các Chain với nhau

Mỗi block được nối với block trước thông qua **Previous Hash**:

```mermaid
graph LR
    B0[Genesis Block<br/>hash: H0] --> B1[Block 1<br/>prevHash: H0<br/>hash: H1]
    B1 --> B2[Block 2<br/>prevHash: H1<br/>hash: H2]
    B2 --> B3[Block 3<br/>prevHash: H2<br/>hash: H3]

    B0 --> H0((H0))
    B1 --> H1((H1))
    B2 --> H2((H2))
    B3 --> H3((H3))
```

**Công thức hash của block:**
```
Block_Hash = SHA256(prevHash + merkleRoot + timestamp)
```

**Genesis Block:**
- `blockNumber = 1`
- `prevHash = "genesis"`
- `hash = SHA256("genesis" + merkleRoot + timestamp)`

#### 4. Block Verification Algorithm

Để verify một block, hệ thống kiểm tra:

```go
func verifyBlock(block Block) bool {
    // 1. Verify block hash
    expectedHash := calculateBlockHash(
        block.BlockNumber,
        block.Timestamp.Format(time.RFC3339),
        block.PrevHash,
        extractEventIds(block.Events)
    )

    if expectedHash != block.Hash {
        return false // Block hash is invalid
    }

    // 2. Verify merkle root
    eventIds := extractEventIds(block.Events)
    expectedMerkleRoot := calculateMerkleRoot(eventIds)

    if expectedMerkleRoot != block.MerkleRoot {
        return false // Merkle root is invalid
    }

    // 3. Verify events exist and are valid
    for _, event := range block.Events {
        if !verifyEvent(event) {
            return false // Event is invalid
        }
    }

    return true // Block is valid
}
```

**Các bước verification:**

1. **Hash Verification**:
   ```
   calculated_hash = SHA256(blockNumber + timestamp + prevHash + events[])
   if calculated_hash != block.hash → INVALID
   ```

2. **Merkle Root Verification**:
   ```
   calculated_merkle = buildMerkleTree(eventIds[])
   if calculated_merkle != block.merkleRoot → INVALID
   ```

3. **Chain Continuity Verification**:
   ```
   if block.prevHash != previousBlock.hash → INVALID
   ```

4. **Event Verification**:
   - Kiểm tra event tồn tại trong database
   - Verify timestamp hợp lệ
   - Kiểm tra business logic rules

#### 5. Blockchain Integrity Verification

Để verify toàn bộ blockchain:

```mermaid
flowchart TD
    A[Start from Genesis Block] --> B[Verify Block 1]
    B --> C{Block Valid?}
    C -->|No| D[INVALID Blockchain]
    C -->|Yes| E[Verify Next Block]
    E --> F{More Blocks?}
    F -->|Yes| B
    F -->|No| G[VALID Blockchain]

    E --> H[Check Chain Continuity]
    H --> I{prevHash matches previous block hash?}
    I -->|No| D
    I -->|Yes| F
```

**Verification Rules:**
- **Genesis Block**: prevHash = "genesis"
- **Chain Continuity**: block[N].prevHash == block[N-1].hash
- **No Double Spending**: Token balances không âm, tổng balance = token.total
- **Business Rules**: Contract states hợp lệ, approvals đúng quyền

#### 6. Security Features

**Cryptographic Security:**
- SHA256 collision resistance
- Timestamp prevents replay attacks
- Event ordering đảm bảo causality

**Data Integrity:**
- Merkle tree cho efficient verification
- Block hash chaining
- Immutable audit trail

**Tamper Detection:**
- Bất kỳ thay đổi nào trong events sẽ làm thay đổi merkle root
- Thay đổi merkle root làm thay đổi block hash
- Thay đổi block hash làm đứt chain continuity

## Thiết kế Database

### MongoDB Collections

#### 1. contracts
```javascript
{
  _id: "contract_1234567890",
  description: "Supply Chain Contract Q1 2024",
  anchorId: "ANCHOR001",
  supplierId: "SUP001", // Primary supplier
  bankId: "BANK001",
  bankApproved: true,
  amount: 50000.00,
  suppliers: [
    {
      supplierId: "SUP001",
      name: "ABC Corporation",
      amount: 30000.00,
      status: "APPROVED"
    },
    {
      supplierId: "SUP002",
      name: "XYZ Ltd",
      amount: 20000.00,
      status: "APPROVED"
    }
  ],
  approvers: ["SUP001", "SUP002"],
  approved: true,
  createdAt: "2024-01-01T09:00:00Z",
  status: "APPROVED"
}
```

**Indexes**:
- `{bankId: 1}`
- `{approved: 1}`
- `{bankApproved: 1}`

#### 2. tokens
```javascript
{
  _id: "token_contract_1234567890",
  contractId: "contract_1234567890",
  symbol: "TK567890",
  total: 50000.00,
  issuer: "SYSTEM",
  owner: "ANCHOR001",
  createdAt: "2024-01-01T10:00:00Z"
}
```

**Indexes**:
- `{contractId: 1}`
- `{issuer: 1}`
- `{owner: 1}`

#### 3. balances
```javascript
{
  tokenId: "token_contract_1234567890",
  account: "SUP001",
  balance: 25000.00,
  transferredFrom: "ANCHOR001"
}
```

**Indexes**:
- `{tokenId: 1, account: 1}` (unique compound index)
- `{account: 1}`

#### 4. events
```javascript
{
  eventId: "evt_1234567890",
  eventType: "CONTRACT_CREATED", // or CONTRACT_BANK_APPROVED, SUPPLIER_APPROVED, etc.
  contractId: "contract_1234567890",
  tokenId: "token_contract_1234567890", // optional
  supplierId: "SUP001", // optional
  bankId: "BANK001", // optional
  anchorId: "ANCHOR001", // optional
  totalAmount: 50000.00, // optional
  settledAmount: 25000.00, // optional
  description: "Bank approved contract and system auto-generated token for anchor",
  timestamp: "2024-01-01T10:00:00Z"
}
```

**Indexes**:
- `{contractId: 1}`
- `{tokenId: 1}`
- `{eventType: 1}`
- `{timestamp: 1}`

#### 5. blocks
```javascript
{
  blockNumber: 1,
  timestamp: "2024-01-01T10:00:00Z",
  events: ["evt_1234567890"],
  previousHash: "genesis",
  hash: "a1b2c3d4e5f6...",
  merkleRoot: "m1n2o3p4q5r6..."
}
```

**Indexes**:
- `{blockNumber: 1}` (unique)
- `{timestamp: 1}`

#### 6. users
```javascript
{
  id: "SUP001",
  username: "supplier1",
  password: "$2a$10$encrypted_password",
  role: "SUPPLIER" // ANCHOR, BANK, SUPPLIER
}
```

**Indexes**:
- `{id: 1}` (unique)
- `{username: 1}` (unique)
- `{role: 1}`

### Database Relationships

```
contracts (1) ──── (1) tokens
    │                    │
    │                    │
    └─── suppliers[] ────┼─── (many) balances
                         │
                         └─── (many) events
                              │
                              └─── (many) blocks
```

### Data Flow Patterns

1. **Contract Creation**: `contracts` → `events` → `blocks`
2. **Token Issuance**: `contracts` → `tokens` → `balances` → `events` → `blocks`
3. **Token Transfer**: `balances` → `events` → `blocks`
4. **Token Settlement**: `balances` → `events` → `blocks`

## Luồng Xử lý Nghiệp vụ

### 1. Contract Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Draft: Contract Created
    Draft --> BankApproval: Bank Approves
    BankApproval --> SupplierApproval: Suppliers Approve
    SupplierApproval --> Active: All Approved
    Active --> Completed: All Tokens Settled
    Completed --> [*]

    Draft --> Cancelled: Cancelled
    BankApproval --> Cancelled: Cancelled
    SupplierApproval --> Cancelled: Cancelled
    Cancelled --> [*]
```

### 2. Token Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Created: Token Created<br/>(by System)
    Created --> Distributed: Distributed to<br/>Suppliers
    Distributed --> Trading: Peer-to-peer<br/>Trading
    Trading --> Settled: Settled with Bank
    Settled --> [*]

    Trading --> PartiallySettled: Some Suppliers<br/>Settled
    PartiallySettled --> Settled: All Settled
```

### 3. Business Rules

#### Contract Rules
- Chỉ Anchor có thể tạo contract
- Bank phải phê duyệt trước khi suppliers có thể phê duyệt
- Tất cả suppliers phải phê duyệt để contract active
- Contract chỉ có thể bị hủy khi ở trạng thái Draft

#### Token Rules
- Token được tạo tự động bởi hệ thống khi bank phê duyệt
- Token ban đầu thuộc về Anchor
- Token được phân phối cho suppliers khi tất cả đã phê duyệt
- Suppliers có thể chuyển token cho nhau
- Suppliers có thể tất toán token với bank bất cứ lúc nào

#### Balance Rules
- Balance không được âm
- Tổng balance của tất cả accounts cho một token luôn bằng token.total
- Balance được cập nhật atomically trong transfer operations

#### Blockchain Rules
- Mọi operation quan trọng đều tạo event
- Mọi event được ghi vào block
- Block hash được tính toán từ previous block + current data
- Blockchain đảm bảo immutability và audit trail

### 4. Security Considerations

#### Authentication & Authorization
- JWT tokens cho user authentication
- Role-based access control (ANCHOR, BANK, SUPPLIER)
- API-level authorization checks

#### Data Integrity
- Blockchain hashing đảm bảo data integrity
- Database transactions cho multi-document operations
- Audit trail hoàn chỉnh

#### Network Security
- HTTPS cho tất cả communications
- CORS configuration
- Docker network isolation

### 5. Performance Considerations

#### Database Optimization
- Compound indexes cho frequent queries
- Pagination cho large result sets
- Connection pooling

#### Caching Strategy
- Redis cache cho frequently accessed data (future enhancement)
- In-memory caching cho user sessions

#### Scalability
- Horizontal scaling với multiple instances
- Database sharding strategy
- Load balancing configuration

---

*Document Version: 1.0*
*Last Updated: January 2024*
*Author: System Design Team*
