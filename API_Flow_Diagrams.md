# 🔗 Blockchain SCF API Flow Diagrams với Events Sync

This document contains Mermaid flow diagrams for all API endpoints across the **Multi-Peer Blockchain SCF System** with **gRPC Events Sync Architecture**.

## 📋 System Overview

### Architecture Components
- **Peer Services**: Independent microservices for each business role
  - `peer-anchor` (Port 8084): Contract creation + Events Sync
  - `peer-main-bank` (Port 8082): Bank approvals, token issuance + Events Sync
  - `peer-supplier` (Port 8083): Supplier approvals, token transfers & settlements + Events Sync
- **Orderer Cluster**: PBFT consensus + Events Sync handler
- **gRPC**: Direct communication between peers and orderer (SubmitTx + SubmitEvent)
- **MongoDB**: Dual-database architecture (blockchain_private + blockchain_public)

### Flow Patterns
- 🟢 **Start/End**: Process start or end
- 🔵 **Process**: Main business logic step
- 🟡 **Decision**: Conditional check
- 🟠 **Database**: Database operation (private peer DB)
- 🟣 **Response**: API response
- 🔴 **Error**: Error handling
- 🟡 **gRPC**: Direct communication (SubmitTx/SubmitEvent)
- 🏛️ **Orderer**: PBFT consensus & Events Sync
- 🔵 **Events Sync**: Synchronization to public blockchain

---

## 🔗 **Contract Management APIs**

### **Peer Anchor (Port 8084)**

### POST /contract/create - Anchor Creates Contract
```mermaid
flowchart TD
    A[🟢 Receive Contract Data] --> B[🔵 Validate Anchor Authorization]
    B --> C[🟡 Check Contract ID]
    C -->|No ID provided| D[🔵 Generate UUID Contract ID]
    C -->|ID provided| E[🔵 Use Provided ID]
    D --> F[🟠 Save Contract to blockchain_private]
    E --> F
    F --> G[🔵 Log CONTRACT_CREATED Event to blockchain_private]
    G --> H[🟡 SubmitTx to Orderer via gRPC]
    H --> I[🟡 SubmitEvent to Orderer via gRPC]
    I --> J[🟣 Return Success Response]

    J --> K[🏛️ Orderer PBFT Consensus]
    K --> L[🏛️ Events Sync to blockchain_public]
    L --> M[🏛️ Stream Blocks to All Peers]
```

### **Peer Main Bank (Port 8082)**

### POST /contract/{id}/approve-bank - Bank Approves Contract & Issues Token
```mermaid
flowchart TD
    A[🟢 Receive Bank Approval Request] --> B[🔵 Validate Bank Authorization]
    B --> C[🟠 Get Contract from blockchain_private]
    C --> D[🟡 Contract Exists?]
    D -->|No| E[🔴 Return 404 Not Found]
    D -->|Yes| F[🟡 Bank Has Permission?]
    F -->|No| G[🔴 Return 403 Forbidden]
    F -->|Yes| H[🟠 Update bankApproved = true in blockchain_private]
    H --> I[🔵 Calculate Total Amount from Suppliers]
    I --> J[🟠 Create Token with Issuer = Bank in blockchain_private]
    J --> K[🟠 Create Initial Balance for Anchor in blockchain_private]
    K --> L[🔵 Log CONTRACT_BANK_APPROVED_TOKEN_GENERATED in blockchain_private]
    L --> M[🟡 SubmitTx to Orderer via gRPC]
    M --> N[🟡 SubmitEvent to Orderer via gRPC]
    N --> O[🟣 Return Success Response with Token ID]

    O --> P[🏛️ Orderer PBFT Consensus]
    P --> Q[🏛️ Events Sync to blockchain_public]
    Q --> R[🏛️ Stream Blocks to All Peers]
```

### **Peer Supplier (Port 8083)**

### POST /contract/{id}/approve - Supplier Approves Contract
```mermaid
flowchart TD
    A[🟢 Receive Supplier Approval Request] --> B[🔵 Validate Supplier Authorization]
    B --> C[🟠 Get Contract from blockchain_private]
    C --> D[🟡 Contract Exists?]
    D -->|No| E[🔴 Return 404 Not Found]
    D -->|Yes| F[🟡 Bank Approved Contract?]
    F -->|No| G[🔴 Return 403 Forbidden]
    F -->|Yes| H[🟡 Supplier Authorized for Contract?]
    H -->|No| I[🔴 Return 403 Forbidden]
    H -->|Yes| J[🟠 Update Supplier Status to APPROVED in blockchain_private]
    J --> K[🟠 Refresh Contract Data]
    K --> L[🔵 Check All Suppliers Approved?]
    L -->|No| M[🔵 Log SUPPLIER_APPROVED Event in blockchain_private]
    L -->|Yes| N[🟠 Update Contract Status to approved in blockchain_private]
    N --> O[🟠 Distribute Token Balances to All Suppliers in blockchain_private]
    O --> P[🟠 Remove Anchor's Balance in blockchain_private]
    P --> Q[🔵 Log CONTRACT_FULLY_APPROVED Event in blockchain_private]
    M --> R[🟡 SubmitTx to Orderer via gRPC]
    Q --> R
    R --> S[🟡 SubmitEvent to Orderer via gRPC]
    S --> T[🟣 Return Success Response]

    T --> U[🏛️ Orderer PBFT Consensus]
    U --> V[🏛️ Events Sync to blockchain_public]
    V --> W[🏛️ Stream Blocks to All Peers]
```

---

## 💰 **Token Management APIs**

### POST /token/transfer - Supplier Transfers Token (Peer Supplier)
```mermaid
flowchart TD
    A[🟢 Receive Transfer Request] --> B[🔵 Validate Supplier Authorization]
    B --> C[🟠 Get Token Info from blockchain_private]
    C --> D[🟡 Token Exists?]
    D -->|No| E[🔴 Return 404 Not Found]
    D -->|Yes| F[🟠 Get Sender Balance from blockchain_private]
    F --> G[🟡 Sender Has Balance?]
    G -->|No| H[🔴 Return 400 Insufficient Balance]
    G -->|Yes| I[🟡 Balance Sufficient?]
    I -->|No| J[🔴 Return 400 Insufficient Balance]
    I -->|Yes| K[🟠 Update Sender Balance - amount in supplier DB]
    K --> L[🟠 Get Receiver Balance from supplier DB]
    L --> M[🟡 Receiver Balance Exists?]
    M -->|No| N[🟠 Create New Balance for Receiver in supplier DB]
    M -->|Yes| O[🟠 Update Receiver Balance + amount in supplier DB]
    N --> P[🔵 Log TOKEN_TRANSFERRED Event in supplier DB]
    O --> P
    P --> Q[🟠 Create Block Entry in supplier DB]
    Q --> R[📨 Publish BLOCK_CREATED to Kafka scf-channel-tx]
    R --> S[🟣 Return Success Response]

    S --> T[🏛️ Orderer Consumes Kafka Message]
    T --> U[🏛️ Order Block & Save to Public Ledger]
    U --> V[🏛️ Broadcast Ordered Block to All Peers]
```

### POST /token/settle - Supplier Settles Token with Bank (Peer Supplier)
```mermaid
flowchart TD
    A[🟢 Receive Settlement Request] --> B[🔵 Validate Supplier Authorization]
    B --> C[🟠 Get Supplier Balance for Token from supplier DB]
    C --> D[🟡 Balance Exists?]
    D -->|No| E[🔴 Return 400 No Balance]
    D -->|Yes| F[🟠 Get Token Info from supplier DB]
    F --> G[🟡 Token Exists?]
    G -->|No| H[🔴 Return 404 Not Found]
    G -->|Yes| I[🟠 Get Contract Info from supplier DB]
    I --> J[🟡 Contract Exists?]
    J -->|No| K[🔴 Return 404 Not Found]
    J -->|Yes| L[🟠 Delete Supplier's Balance from supplier DB]
    L --> M[🔵 Log TOKEN_SETTLED Event in supplier DB]
    M --> N[🟠 Create Block Entry in supplier DB]
    N --> O[📨 Publish BLOCK_CREATED to Kafka scf-channel-tx]
    O --> P[🟣 Return Success Response]

    P --> Q[🏛️ Orderer Consumes Kafka Message]
    Q --> R[🏛️ Order Block & Save to Public Ledger]
    R --> S[🏛️ Broadcast Ordered Block to All Peers]
```

### GET /token/{id} - Get Token Information
```mermaid
flowchart TD
    A[🟢 Receive Token ID] --> B[🟠 Query Token from Database]
    B --> C[🟡 Token Found?]
    C -->|No| D[🔴 Return 404 Not Found]
    C -->|Yes| E[🟣 Return Token Data]
```

---

## 🔍 **Query APIs**

### GET /contract/list - Get All Contracts (Peer Anchor)
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query All Contracts from anchor DB]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Contracts Array]
```

### GET /contract/list - Get All Contracts (Peer Main Bank)
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query All Contracts from main-bank DB]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Contracts Array]
```

### GET /tokens - Get All Tokens (Peer Main Bank)
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query All Tokens from main-bank DB]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Tokens Array]
```

### GET /balances/account/{accountId} - Get Account Balances (Peer Supplier)
```mermaid
flowchart TD
    A[🟢 Receive Account ID] --> B[🟠 Query Balances by Account ID from supplier DB]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Balances Array]
```

### GET /suppliers - Get All Suppliers (Peer Supplier)
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query Users with Role SUPPLIER from shared DB]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Suppliers Array]
```

---

## Ledger & Block APIs

### GET /contract/{id}/ledger - Get Contract Ledger
```mermaid
flowchart TD
    A[🟢 Receive Contract ID] --> B[🟠 Query Events for Contract]
    B --> C[🟠 Query Events for Token (token_contractId)]
    C --> D[🔵 Combine All Events]
    D --> E[🟠 Query Blocks Containing Event IDs]
    E --> F[🟠 Query Token Balances]
    F --> G[🟡 All Queries Successful?]
    G -->|No| H[🔴 Return 500 Internal Error]
    G -->|Yes| I[🟣 Return Ledger Data]
```

### POST /blocks/hash/update - Update Block Hashes
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query All Blocks Sorted by Block Number]
    B --> C[🔵 Initialize Updated Count = 0]
    C --> D{🟡 More Blocks?}
    D -->|Yes| E[🔵 Get Block Data]
    E --> F[🔵 Calculate Previous Hash]
    F --> G[🔵 Calculate New Hash]
    G --> H[🟠 Update Block in Database]
    H --> I[🔵 Increment Updated Count]
    I --> D
    D -->|No| J[🟣 Return Update Summary]
```

---

### **Complete System Flow**

```mermaid
flowchart TD
    subgraph "User Actions"
        A1[Anchor Creates Contract]
        A2[Bank Approves Contract]
        A3[Supplier Approves Contract]
        A4[Supplier Transfers Token]
        A5[Supplier Settles Token]
    end

    subgraph "Peer Services"
        P1[peer-anchor:8084]
        P2[peer-main-bank:8082]
        P3[peer-supplier:8083]
    end

    subgraph "Private Databases"
        DB1[(anchor DB)]
        DB2[(main-bank DB)]
        DB3[(supplier DB)]
    end

    subgraph "Kafka Messaging"
        K1[📨 scf-channel-tx]
        K2[📨 audit-channel-tx]
    end

    subgraph "Orderer Cluster"
        O1[🏛️ Orderer Nodes 7050-7070]
    end

    subgraph "Public Ledger"
        PDB[(Shared MongoDB)]
    end

    A1 --> P1
    P1 --> DB1
    P1 --> K1

    A2 --> P2
    P2 --> DB2
    P2 --> K2

    A3 --> P3
    A3 --> DB3
    A3 --> K1

    A4 --> P3
    A4 --> DB3
    A4 --> K1

    A5 --> P3
    A5 --> DB3
    A5 --> K1

    K1 --> O1
    K2 --> O1
    O1 --> PDB
    PDB --> P1
    PDB --> P2
    PDB --> P3
```

---

## 🏗️ **Multi-Peer Architecture Patterns**

### **Database Collections Per Peer**
| Collection | Peer Anchor | Peer Main Bank | Peer Supplier | Shared DB |
|------------|-------------|----------------|---------------|-----------|
| **contracts** | ✅ Private | ✅ Private | ✅ Private | ❌ |
| **tokens** | ❌ | ✅ Private | ✅ Private | ❌ |
| **balances** | ❌ | ✅ Private | ✅ Private | ❌ |
| **events** | ✅ Private | ✅ Private | ✅ Private | ✅ Public |
| **blocks** | ✅ Private | ✅ Private | ✅ Private | ❌ |
| **users** | ❌ | ❌ | ❌ | ✅ Public |

### **Business Logic Distribution**

#### **Contract Management Flow**
```mermaid
flowchart LR
    A[Anchor Creates Contract] --> B[peer-anchor DB]
    B --> C[Kafka Publish]
    C --> D[Orderer Ordering]
    D --> E[Public Ledger]
    E --> F[Sync to All Peers]
```

#### **Token Issuance Flow**
```mermaid
flowchart LR
    A[Bank Approves Contract] --> B[peer-main-bank DB]
    B --> C[Create Token + Initial Balance]
    C --> D[Kafka Publish to audit-channel]
    D --> E[Orderer Ordering]
    E --> F[Public Ledger]
    F --> G[Sync to All Peers]
```

#### **Token Circulation Flow**
```mermaid
flowchart LR
    A[Supplier Transfers Token] --> B[peer-supplier DB]
    B --> C[Update Balances]
    C --> D[Kafka Publish to scf-channel]
    D --> E[Orderer Ordering]
    E --> F[Public Ledger]
    F --> G[Sync to All Peers]
```

#### **Token Settlement Flow**
```mermaid
flowchart LR
    A[Supplier Settles Token] --> B[peer-supplier DB]
    B --> C[Delete Supplier Balance]
    C --> D[Kafka Publish to scf-channel]
    D --> E[Orderer Ordering]
    E --> F[Public Ledger]
    F --> G[Contract Closed]
```

### **Event Streaming Patterns**

#### **Kafka Topic Usage**
- **`scf-channel-tx`**: Contract operations, token transfers, settlements
- **`audit-channel-tx`**: Bank approvals, compliance events
- **`scf-channel-blocks`**: Ordered SCF blocks broadcast
- **`audit-channel-blocks`**: Ordered audit blocks broadcast

#### **Orderer Processing**
```mermaid
flowchart TD
    A[📨 Receive Kafka Events] --> B[🏛️ Validate Event Sequence]
    B --> C[🏛️ Order Events by Timestamp]
    C --> D[🏛️ Create Ordered Block]
    D --> E[🏛️ Calculate Block Hash]
    E --> F[🏛️ Save to Public Ledger]
    F --> G[🏛️ Broadcast to All Peers]
```

### **Data Consistency Guarantees**

#### **Eventual Consistency**
- Private peer databases maintain local state
- Public shared database maintains global ledger
- Kafka ensures reliable message delivery
- Orderer ensures global ordering

#### **Audit Trail**
- Every operation creates immutable event
- Events grouped into ordered blocks
- Block hash chain ensures integrity
- Public ledger provides universal truth

---

## 📋 **API Summary by Peer Service**

### **Peer Anchor (Port 8084)**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/create` | Create contract + Kafka publish |
| GET | `/contract/list` | List contracts |
| GET | `/contract/{id}` | Get contract details |
| GET | `/health` | Health check |

### **Peer Main Bank (Port 8082)**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/{id}/approve-bank` | Bank approve + token issuance |
| GET | `/contract/list` | List all contracts |
| GET | `/contract/{id}` | Get contract details |
| GET | `/contract/{id}/ledger` | Contract audit trail |
| GET | `/token/issued/{bankId}` | Tokens issued by bank |
| GET | `/tokens` | All tokens issued by bank |

### **Peer Supplier (Port 8083)**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/{id}/approve` | Supplier approve contract |
| POST | `/token/transfer` | Transfer tokens between suppliers |
| POST | `/token/settle` | Settle token with bank |
| GET | `/balances/account/{accountId}` | Account balances |
| GET | `/suppliers` | List all suppliers |
| GET | `/health` | Health check |

### **Common Query Endpoints**
| Method | Endpoint | Available On |
|--------|----------|--------------|
| GET | `/token/{id}` | All peers |
| GET | `/balances/token/{tokenId}` | All peers |

---

## 🔄 **Complete API Flow Reference**

**This document provides comprehensive flow diagrams for all APIs in the Multi-Peer Blockchain SCF System with Events Sync, showing the complete journey from user action through peer processing, gRPC communication, PBFT consensus, to public blockchain synchronization.**
