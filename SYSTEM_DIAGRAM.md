# Hệ thống Blockchain Supply Chain Finance (SCF) với Blockchain Gateway Architecture

## Tổng quan kiến trúc với Blockchain Gateway & PBFT Consensus

### 🆕 **Kiến trúc mới: Blockchain Gateway Architecture**
Hệ thống blockchain permissioned với kiến trúc hoàn toàn mới:

- **🔗 blockchain-gw**: Blockchain Gateway - Fabric Client chịu trách nhiệm tổng hợp endorsements và gửi transactions sang Orderer cluster trên port 9090
- **📡 gRPC Direct Communication**: Peer services giao tiếp với blockchain-gw qua gRPC APIs để thực hiện endorsement
- **🏗️ Endorsement Layer**: Peers chỉ đóng vai trò Endorser, chạy chaincode logic và trả về endorsements
- **💾 State Persistence**: blockchain-gw quản lý world state trong MongoDB shared
- **🏛️ PBFT Consensus**: Practical Byzantine Fault Tolerance với 3-node orderer cluster (f=1 fault tolerance)
- **⚡ Real-time Processing**: gRPC streaming đảm bảo peers nhận blocks ngay lập tức

## Kiến trúc Blockchain Gateway & PBFT Consensus

### **Core Architecture Principles:**

- **✅ Public-Private Network Separation**: API Gateway là public-facing, blockchain-gw + peers + orderer = private consortium network
- **✅ Blockchain Gateway Architecture**: blockchain-gw đóng vai trò Fabric Client + Endorsement Aggregator trong Private Blockchain Network
- **✅ Endorsement Pattern**: Peers chỉ thực hiện endorsement, không gửi trực tiếp lên Orderer
- **✅ Proposal-Response Flow**: blockchain-gw gửi ProposalRequest đến peers → peers chạy chaincode → trả endorsement
- **✅ PBFT Consensus**: 3-node orderer cluster với f=1 fault tolerance, Pre-Prepare → Prepare → Commit → Finalize phases
- **✅ Centralized State Management**: blockchain-gw quản lý world state trong MongoDB shared
- **✅ Cryptographic Signatures**: Mỗi orderer ký blocks với ECDSA, quorum 2f+1 signatures
- **✅ Real-time Block Streaming**: Orderers stream finalized blocks về peers và blockchain-gw

## Trạng thái triển khai hiện tại ✅

### ✅ **Hoàn thành (10/10 containers chạy ổn định)**

| Service | Port | Status | Implementation |
|---------|------|--------|----------------|
| **blockchain-gw** | 9090 | ✅ Running | Fabric Client + Endorsement Aggregator - trong Private Blockchain Network |
| **peer-main-bank** | 8082 | ✅ Running | Endorser Peer - gRPC EvaluateProposal + REST API, không gửi trực tiếp lên orderer |
| **peer-supplier** | 8083 | ✅ Running | Endorser Peer - gRPC EvaluateProposal + REST API, không gửi trực tiếp lên orderer |
| **peer-anchor** | 8084 | ✅ Running | Endorser Peer - gRPC EvaluateProposal + REST API, không gửi trực tiếp lên orderer |
| **orderer-ord1** | 7050 | ✅ Running | PBFT leader, consensus engine, block ordering |
| **orderer-ord2** | 7060 | ✅ Running | PBFT follower, consensus participant |
| **orderer-ord3** | 7070 | ✅ Running | PBFT follower, consensus participant |
| **mongo-shared** | 27017 | ✅ Running | World state database cho blockchain-gw |
| **backend** | 8080 | ✅ Running | Spring Boot API gateway - public-facing entry point |
| **frontend** | 4200 | ✅ Running | Angular 17 UI |

### ✅ **Blockchain Gateway Architecture End-to-End Test**

```bash
# Test successful - Blockchain Gateway Architecture ✅
User Request → API GW → blockchain-gw → ProposalRequest to Peers → Endorsements → SubmitTx to Orderer → PBFT Consensus → Block Creation → Events Sync ✅

# Complete test flow:
✅ CONTRACT_CREATED via blockchain-gw + peer endorsements + orderer submission
✅ CONTRACT_BANK_APPROVED_TOKEN_GENERATED via blockchain-gw + peer endorsements + orderer submission
✅ CONTRACT_FULLY_APPROVED via blockchain-gw + peer endorsements + orderer submission
✅ TOKEN_TRANSFERRED via blockchain-gw + peer endorsements + orderer submission
✅ TOKEN_SETTLED via blockchain-gw + peer endorsements + orderer submission

Sample blockchain-gw transaction submission:
{
  "transaction_id": "tx_0845917c79d28d6da74de438031b97a4",
  "sender_id": "blockchain-gw",
  "transaction_type": "CONTRACT_CREATE",
  "transaction_data": "...",
  "endorsements": [
    {
      "peer_id": "peer-anchor",
      "rwset": "...",
      "signature": "ecdsa_sig_1..."
    },
    {
      "peer_id": "peer-main-bank", 
      "rwset": "...",
      "signature": "ecdsa_sig_2..."
    },
    {
      "peer_id": "peer-supplier",
      "rwset": "...",
      "signature": "ecdsa_sig_3..."
    }
  ],
  "timestamp": "2025-10-01T16:00:48Z"
}

Sample PBFT consensus block with signatures:
{
  "_id": "block_1",
  "height": 1,
  "timestamp": "2025-10-01T16:00:48Z",
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

### ✅ **Implemented APIs**

**Peer Anchor (Port 8084) - Endorser Only:**
- `gRPC EvaluateProposal()` - Anchor endorsement → blockchain-gw → ProposalRequest → Endorsement → blockchain-gw SubmitTx
- `GET /contract/{id}` - Contract details từ blockchain-gw
- `GET /contract/list` - Danh sách contracts từ blockchain-gw
- `GET /health` - Health check

**Peer Main Bank (Port 8082) - Endorser Only:**
- `gRPC EvaluateProposal()` - Bank endorsement → blockchain-gw → ProposalRequest → Endorsement → blockchain-gw SubmitTx
- `GET /contract/list` - List tất cả contracts từ blockchain-gw
- `GET /contract/{id}` - Contract details từ blockchain-gw
- `GET /contract/{id}/ledger` - Contract audit trail từ blockchain-gw
- `GET /token/issued/{bankId}` - Tokens issued by bank từ blockchain-gw
- `GET /tokens` - All tokens issued by bank từ blockchain-gw

**Peer Supplier (Port 8083) - Endorser Only:**
- `gRPC EvaluateProposal()` - Supplier endorsement → blockchain-gw → ProposalRequest → Endorsement → blockchain-gw SubmitTx
- `GET /balances/account/{accountId}` - Account balances của supplier từ blockchain-gw
- `GET /health` - Health check

**blockchain-gw (Port 9090) - Fabric Client + Endorsement Aggregator:**
- `ProposalRequest()` - Gửi proposal đến peers để lấy endorsements
- `CollectEndorsements()` - Thu thập endorsements từ peers
- `ValidateEndorsements()` - Kiểm tra endorsement policy
- `SubmitTx()` - Gửi transaction + endorsements sang orderer cluster

**Orderer gRPC APIs (Ports 7050-7070):**
- `SubmitTx(Transaction) → SubmitTxReply` - Submit transaction for PBFT consensus
- `StreamBlocks(StreamBlocksReq) → stream Block` - Stream finalized blocks to peers
- `Consensus(ConsensusMsg) → Ack` - PBFT consensus messages between orderers

```mermaid
graph TB
    subgraph "Public Network - External Access"
        subgraph "Frontend Layer (Angular 17)"
            subgraph "Role-based Components"
                ANC[Anchor Components<br/>Contract Form, Token Mgmt]
                BNK[Bank Components<br/>Token Overview, Ledger]
                SUP[Supplier Components<br/>Approval, Token Transfer]
            end
            subgraph "Shared Components"
                AUTH[Login, Auth Guard]
                NAV[Navbar, Routing]
                LED[Ledger Viewer]
            end
        end

        subgraph "API Gateway Layer (Spring Boot 3.1.5)"
            AR[API Router<br/>Port 8080<br/>Public-facing Entry Point]
            AUTH_S[Authentication Service<br/>JWT Tokens]
            PR[Smart Routing Service<br/>Forward to Private Network]
        end
    end

    subgraph "Private Blockchain Network - Consortium Only"
        subgraph "blockchain-gw (Port 9090) - Fabric Client Service"
            GW[Blockchain Gateway<br/>✅ Running<br/>Internal Orchestrator]
            PROPOSAL[ProposalRequest<br/>Send to Peers]
            COLLECT[CollectEndorsements<br/>Gather from Peers]
            SUBMIT[SubmitTx<br/>Send to Orderer]
            POLICY[Endorsement Policy<br/>Check Quorum]
        end

        subgraph "Peer Network - Endorsers Only"
            subgraph "Peer Main Bank (Port 8082)"
                MB_API[gRPC Endpoint<br/>EvaluateProposal]
                MB_ENDORSE[Endorsement Logic<br/>Chaincode Execution]
                subgraph "Bank Chaincode Logic"
                    MB_CON[Contract Approval<br/>Bank Validation]
                    MB_TOK[Token Issuance<br/>Bank Authority]
                    MB_LED[Ledger Management<br/>Bank Oversight]
                end
                MB_DB[(MongoDB Private<br/>Local State)]
            end

            subgraph "Peer Supplier (Port 8083)"
                SUP_API[gRPC Endpoint<br/>EvaluateProposal]
                SUP_ENDORSE[Endorsement Logic<br/>Chaincode Execution]
                subgraph "Supplier Chaincode Logic"
                    SUP_APP[Contract Approval<br/>Supplier Validation]
                    SUP_TOK[Token Transfer<br/>P2P Circulation]
                    SUP_BAL[Balance Management<br/>Token Holdings]
                end
                SUP_DB[(MongoDB Private<br/>Local State)]
            end

            subgraph "Peer Anchor (Port 8084)"
                ANC_API[gRPC Endpoint<br/>EvaluateProposal]
                ANC_ENDORSE[Endorsement Logic<br/>Chaincode Execution]
                subgraph "Anchor Chaincode Logic"
                    ANC_CON[Contract Creation<br/>Anchor Authority]
                    ANC_TOK[Token Reception<br/>Initial Ownership]
                    ANC_LED[Contract Ledger<br/>Anchor Tracking]
                end
                ANC_DB[(MongoDB Private<br/>Local State)]
            end
        end

        subgraph "Orderer Cluster (PBFT 3f+1 nodes, f=1)"
            ORD1[Orderer Leader<br/>Port 7050<br/>✅ PBFT Leader<br/>Consensus Engine]
            ORD2[Orderer Follower<br/>Port 7060<br/>✅ PBFT Follower<br/>Consensus Participant]
            ORD3[Orderer Follower<br/>Port 7070<br/>✅ PBFT Follower<br/>Consensus Participant]
            subgraph "PBFT Consensus Phases"
                PREP[Pre-Prepare<br/>Leader broadcasts<br/>proposed block]
                PREPARE[Prepare<br/>Followers validate<br/>send prepare msgs]
                COMMIT[Commit<br/>Quorum reached<br/>2f+1 signatures]
                FINAL[Finalize<br/>Block committed<br/>Stream to network]
            end
        end

        subgraph "Database Layer"
            SHARED_DB[(MongoDB Shared<br/>World State Public<br/>Managed by blockchain-gw)]
        end
    end

    %% Public Network Flow: User → API Gateway
    ANC --> AR
    BNK --> AR
    SUP --> AR

    %% Public to Private: API Gateway → blockchain-gw
    AR --> PR
    PR -->|Forward Request<br/>Private Channel| GW

    %% Private Network Flow: blockchain-gw → Peers → Endorsements
    GW -->|ProposalRequest| MB_API
    GW -->|ProposalRequest| SUP_API
    GW -->|ProposalRequest| ANC_API

    %% Peers execute chaincode and return endorsements
    MB_API --> MB_ENDORSE
    SUP_API --> SUP_ENDORSE
    ANC_API --> ANC_ENDORSE

    MB_ENDORSE -->|Endorsement + RWSet| GW
    SUP_ENDORSE -->|Endorsement + RWSet| GW
    ANC_ENDORSE -->|Endorsement + RWSet| GW

    %% blockchain-gw submits to Orderer
    GW -->|SubmitTx + Endorsements| ORD1

    %% PBFT Consensus Process
    ORD1 --> PREP
    PREP --> ORD2
    PREP --> ORD3
    ORD2 --> PREPARE
    ORD3 --> PREPARE
    PREPARE --> COMMIT
    COMMIT --> FINAL

    %% Orderers stream finalized blocks to network
    FINAL -->|StreamBlocks| MB_API
    FINAL -->|StreamBlocks| SUP_API
    FINAL -->|StreamBlocks| ANC_API
    FINAL -->|StreamBlocks| GW

    %% Database connections
    MB_API --> MB_DB
    SUP_API --> SUP_DB
    ANC_API --> ANC_DB
    GW --> SHARED_DB
    FINAL --> SHARED_DB

    %% Response flow: Private → Public
    GW -->|Transaction Result<br/>Private Channel| PR
    PR -->|Response| AR
    AR --> ANC
    AR --> BNK
    AR --> SUP
```

## Luồng nghiệp vụ chi tiết

### Quy trình Supply Chain Finance với PBFT Consensus Flow:

```mermaid
sequenceDiagram
    participant User(App/Web)
    participant APIGW(Java)
    participant GW as blockchain-gw (Private)
    participant PeerA as Peer Anchor
    participant PeerB as Peer Bank
    participant PeerS as Peer Supplier
    participant ORD as Orderer PBFT
    participant LED as Ledger (MongoDB)

    User->>APIGW(Java): Submit TX
    APIGW(Java)->>GW: Forward Request (Private)

    GW->>PeerA: ProposalRequest
    GW->>PeerB: ProposalRequest
    GW->>PeerS: ProposalRequest

    PeerA->>PeerA: Execute Chaincode
    PeerA->>GW: Endorsement + RWSet
    PeerB->>PeerB: Execute Chaincode
    PeerB->>GW: Endorsement + RWSet
    PeerS->>PeerS: Execute Chaincode
    PeerS->>GW: Endorsement + RWSet

    GW->>GW: Collect endorsements + Check Policy
    GW->>ORD: Submit TX + endorsements
    ORD->>LED: Commit Block
    LED-->>GW: TX committed
    GW-->>APIGW(Java): TX result
    APIGW(Java)-->>User: Response

## Data Flow Architecture

### Luồng xử lý dữ liệu hoàn chỉnh:

```mermaid
flowchart TD
    subgraph "Contract Creation Flow"
        C1[Anchor submits contract form<br/>with PDF file]
        C2[Validate contract data<br/>+ file format]
        C3[Process file upload<br/>save to storage]
        C4[Generate contract ID<br/>+ metadata]
        C5[Save to contracts collection<br/>with file URL]
        C6[Auto-generate token ID<br/>symbol from contract]
        C7[Save to tokens collection<br/>issuer=Bank, owner=Bank]
        C8[Create initial balance<br/>Bank=totalAmount]
        C9[Create blockchain event<br/>CREATE type]
        C10[Return success response<br/>with contract + token data]
    end

    subgraph "Contract Approval Flow"
        A1[Supplier submits approval<br/>via approval form]
        A2[Validate approver auth<br/>check supplier role]
        A3[Check contract status<br/>not already approved]
        A4[Verify all suppliers approved<br/>business logic check]
        A5[Update contract.approved = true<br/>mark as executed]
        A6[Transfer token ownership<br/>Bank → Supplier]
        A7[Update balances<br/>Bank=0, Supplier=amount]
        A8[Create APPROVE_SUPPLIER event<br/>+ EXECUTE event]
        A9[Return approval response<br/>with updated data]
    end

    subgraph "Token Transfer Flow"
        T1[Supplier initiates transfer<br/>select recipient + amount]
        T2[Validate sender ownership<br/>check token ownership]
        T3[Check sufficient balance<br/>balance >= amount]
        T4[Validate recipient exists<br/>user authentication]
        T5[Debit sender balance<br/>sender.balance -= amount]
        T6[Credit receiver balance<br/>receiver.balance += amount]
        T7[Update token ownership<br/>if full transfer]
        T8[Create TRANSFER event<br/>blockchain record]
        T9[Return transfer response<br/>with transaction details]
    end

    subgraph "Block Building Flow"
        B1[Timer triggers every 10s<br/>scheduled task]
        B2[Query unincluded events<br/>max 10 events]
        B3[Check events available<br/>skip if empty]
        B4[Calculate Merkle root<br/>SHA256 of event IDs]
        B5[Generate block hash<br/>prevHash + merkleRoot + timestamp]
        B6[Create new block record<br/>save to blocks collection]
        B7[Mark events as included<br/>update events collection]
        B8[Log block creation<br/>for monitoring]
    end

    subgraph "Ledger Query Flow"
        L1[User requests ledger view<br/>contract or block data]
        L2[Validate user permissions<br/>role-based access]
        L3[Query blocks/events<br/>by contract ID or block range]
        L4[Derive world state<br/>from event history]
        L5[Format ledger data<br/>human-readable format]
        L6[Return immutable audit trail<br/>blockchain-verified data]
    end

    C1 --> C2 --> C3 --> C4 --> C5 --> C6 --> C7 --> C8 --> C9 --> C10
    A1 --> A2 --> A3 --> A4 --> A5 --> A6 --> A7 --> A8 --> A9
    T1 --> T2 --> T3 --> T4 --> T5 --> T6 --> T7 --> T8 --> T9

    B1 --> B2 --> B3 --> B4 --> B5 --> B6 --> B7 --> B8

    L1 --> L2 --> L3 --> L4 --> L5 --> L6

    C9 --> B2
    A8 --> B2
    T8 --> B2
```

## Database Schema

### Mô hình dữ liệu MongoDB hoàn chỉnh:

```mermaid
erDiagram
    CONTRACT ||--o{ TOKEN : "1 contract issues 1 token"
    TOKEN ||--o{ BALANCE : "1 token has multiple balances"
    CONTRACT ||--o{ EVENT : "1 contract has multiple events"
    EVENT ||--o{ BLOCK : "events grouped into blocks"
    USER ||--o{ BALANCE : "1 user has multiple balances"
    USER ||--o{ TOKEN : "user can own tokens"
    USER ||--o{ EVENT : "user creates events"

    CONTRACT {
        string id PK "Contract unique ID"
        string description "Contract description"
        string buyer "Anchor/Buyer ID"
        object suppliers "Array of supplier objects"
        float totalAmount "Total contract amount"
        string status "PENDING|READY_TO_EXECUTE|EXECUTED"
        string fileUrl "URL to uploaded contract PDF"
        timestamp createdAt "Creation timestamp"
        timestamp updatedAt "Last update timestamp"
        object history "Approval history"
    }

    TOKEN {
        string id PK "Token unique ID"
        string contractId FK "Reference to contract"
        string symbol "Token symbol (e.g., SCF-001)"
        float totalSupply "Total token supply"
        string issuer "Bank ID (token issuer)"
        string owner "Current owner ID"
        timestamp createdAt "Token creation timestamp"
    }

    BALANCE {
        string tokenId FK "Reference to token"
        string account "User ID (account holder)"
        float balance "Account balance for this token"
        timestamp lastUpdated "Last balance update"
    }

    USER {
        string id PK "User unique ID"
        string username "Login username"
        string password "Hashed password"
        string role "ANCHOR|BANK|SUPPLIER"
        string name "Display name"
        timestamp createdAt "User creation timestamp"
    }

    EVENT {
        string eventId PK "Event unique ID"
        string contractId FK "Related contract ID"
        string type "CREATE|APPROVE_SUPPLIER|EXECUTE|TRANSFER"
        string actorId "User who created event"
        object payload "Event-specific data"
        timestamp timestamp "Event timestamp"
        boolean included "Included in block?"
    }

    BLOCK {
        int blockNumber PK "Sequential block number"
        timestamp timestamp "Block creation time"
        array events "Array of included events"
        string prevHash "Hash of previous block"
        string hash "Current block hash"
        string merkleRoot "Merkle root of events"
    }
```

### Mối quan hệ dữ liệu chính:
- **Contract → Token**: Mỗi hợp đồng tạo ra một token duy nhất
- **Token → Balances**: Một token có thể có nhiều balance records cho các accounts khác nhau
- **Contract → Events**: Mỗi thay đổi trên contract được ghi thành event
- **Events → Blocks**: Events được nhóm lại thành blocks theo thời gian
- **User → Balances**: User có thể sở hữu nhiều tokens (nhiều balance records)
- **Event Sourcing**: World state được derive từ chuỗi events trong blocks

## Component Architecture

### Kiến trúc components chi tiết theo layer:

```mermaid
graph TB
    subgraph "Angular 17 Frontend"
        subgraph "Core Components"
            LOGIN[Login Component<br/>JWT Authentication]
            NAV[Navbar Component<br/>Role-based Navigation]
            GUARD[Auth Guard<br/>Route Protection]
        end

        subgraph "Anchor Components (/anchor/*)"
            CF[Contract Form<br/>File Upload + Suppliers]
            CTM[Contract Token Mgmt<br/>Created Contracts View]
        end

        subgraph "Bank Components (/bank/*)"
            BTM[Bank Token Mgmt<br/>Issued Tokens Overview]
            LED_BANK[Ledger Viewer<br/>Audit Dashboard]
        end

        subgraph "Supplier Components (/supplier/*)"
            STM[Supplier Token Mgmt<br/>Owned Tokens]
            TF[Transfer Form<br/>Token Transfer UI]
            APP[Contract Approval<br/>Approval Workflow]
        end

        subgraph "Shared Components"
            CS[Contract Status<br/>Status Dashboard]
            LV[Ledger Viewer<br/>Blockchain Explorer]
            SL[Status List<br/>Generic List Component]
        end

        subgraph "Services Layer"
            AS[Auth Service<br/>JWT + Role Mgmt]
            CTS[Contract Service<br/>Contract CRUD]
            TS[Token Service<br/>Token Operations]
            LS[Ledger Service<br/>Blockchain Queries]
        end
    end

    subgraph "Spring Boot 3.1.5 Backend API Gateway"
        subgraph "Controllers"
            UC[User Controller<br/>Auth endpoints]
            CC[Contract Controller<br/>Contract + File APIs]
            TC[Token Controller<br/>Token management]
        end

        subgraph "Services"
            AUS[Auth Service<br/>JWT validation]
            COS[Contract Service<br/>Business logic]
            TOS[Token Service<br/>Transfer logic]
        end

        subgraph "Integration"
            BC[Blockchain Client<br/>Peer Services proxy]
        end

        subgraph "Security"
            SEC[Spring Security<br/>JWT + CORS]
            UP[File Upload<br/>Multipart handling]
        end
    end

    subgraph "gRPC Direct Communication Layer"
        subgraph "SCF Chaincode Service"
            CHAINCODE[Blockchain Gateway<br/>Port 9090<br/>Transaction Aggregator Service]
            SC_METHODS[Contract Management<br/>Token Operations<br/>State Persistence]
        end

        subgraph "gRPC Communication Protocols"
            GRPC_TX[gRPC Transaction APIs<br/>SubmitTx, StreamBlocks<br/>Direct Peer-to-Orderer]
            GRPC_CHAINCODE[gRPC Gateway APIs<br/>AggregateTransactions, SubmitToOrderer<br/>Transaction Management]
            GRPC_CONSENSUS[gRPC Consensus APIs<br/>PBFT Messages<br/>Leader Election]
        end

        subgraph "gRPC Streaming & State Sync"
            STREAMING[gRPC Bidirectional Streams<br/>Real-time Block Delivery<br/>Connection Pooling]
            STATE_SYNC[State Synchronization<br/>Merkle Tree Validation<br/>Block Integrity]
        end
    end

    subgraph "Peer Services Layer (Endorsing Peers)"
        subgraph "Peer Main Bank (✅ IMPLEMENTED)"
            MB_API[REST API<br/>Port 8082<br/>✅ Contract creation]
            MB_GRPC[gRPC Client<br/>✅ Direct to Orderer<br/>SubmitTx + StreamBlocks]
            MB_HANDLER[Contract Handlers<br/>Bank approval, Token issuance]
            MB_CHAINCODE[gRPC to Gateway<br/>Transaction Submission]
            MB_DB[(MongoDB Private<br/>✅ World State<br/>blockchain_private)]
        end

        subgraph "Peer Supplier (✅ IMPLEMENTED)"
            SUP_API[REST API<br/>Port 8083<br/>✅ Token transfer]
            SUP_GRPC[gRPC Client<br/>✅ Direct to Orderer<br/>SubmitTx + StreamBlocks]
            SUP_HANDLER[Token Handlers<br/>Transfer, Balance mgmt]
            SUP_CHAINCODE[gRPC to Gateway<br/>Transaction Submission]
            SUP_DB[(MongoDB Private<br/>✅ World State<br/>blockchain_private)]
        end

        subgraph "Peer Anchor (✅ IMPLEMENTED)"
            ANC_API[REST API<br/>Port 8084<br/>✅ Contract form]
            ANC_GRPC[gRPC Client<br/>✅ Direct to Orderer<br/>SubmitTx + StreamBlocks]
            ANC_HANDLER[Contract Handlers<br/>Creation, Ledger]
            ANC_CHAINCODE[gRPC to Gateway<br/>Transaction Submission]
            ANC_DB[(MongoDB Private<br/>✅ World State<br/>blockchain_private)]
        end

        subgraph "Peer Business Logic"
            PB_VAL[Validation Layer<br/>Input validation<br/>Business rules]
            PB_ES[Event Sourcing<br/>State management<br/>Transaction logic]
            PB_SEC[Security Layer<br/>Digital signatures<br/>Access control]
        end
    end

    subgraph "Orderer Cluster (PBFT Consensus) ✅ IMPLEMENTED"
        subgraph "Orderer Node 1 (Leader)"
            ORD1_API[gRPC API<br/>Port 7050<br/>✅ SubmitTx, StreamBlocks]
            ORD1_PBFT[PBFT Leader<br/>✅ Pre-Prepare, Commit<br/>Consensus coordination]
            ORD1_PROC[Transaction Processor<br/>✅ Block creation<br/>Merkle tree building]
            ORD1_BLOCK[Block Builder<br/>✅ ECDSA signing<br/>Block finalization]
        end

        subgraph "Orderer Node 2 (Follower)"
            ORD2_API[gRPC API<br/>Port 7060<br/>✅ Backup ordering]
            ORD2_PBFT[PBFT Follower<br/>✅ Prepare, Commit<br/>Consensus participation]
            ORD2_PROC[Transaction Processor<br/>Block validation]
        end

        subgraph "Orderer Node 3 (Follower)"
            ORD3_API[gRPC API<br/>Port 7070<br/>✅ Backup ordering]
            ORD3_PBFT[PBFT Follower<br/>✅ Prepare, Commit<br/>Consensus participation]
            ORD3_PROC[Transaction Processor<br/>Block validation]
        end

        subgraph "Orderer Shared Database"
            ORD_DB[(MongoDB Shared<br/>✅ Dual Databases<br/>Private + Public<br/>✅ Events & blocks persisted)]
        end

        subgraph "Orderer Core Components"
            ORD_CONS[PBFT Consensus Engine<br/>3f+1 fault tolerance<br/>Leader election]
            ORD_CRYPTO[Cryptographic Service<br/>ECDSA signatures<br/>SHA256 hashing]
            ORD_LEDGER[Ledger Manager<br/>Block validation<br/>Merkle tree verification]
        end
    end

    subgraph "Data Layer"
        subgraph "MongoDB Cluster"
            MDB_SHARED[MongoDB Shared<br/>✅ Events, Users<br/>Port 27017]
            MDB_MAIN[MongoDB Main Bank<br/>✅ Contracts, Tokens]
            MDB_SUP[MongoDB Supplier<br/>✅ Balances, Transfers]
            MDB_ANC[MongoDB Anchor<br/>✅ Contracts, Ledgers]
        end

        subgraph "Data Access Layer"
            DAL_MODELS[Data Models<br/>Go structs<br/>Schema definitions]
            DAL_DRIVER[MongoDB Driver<br/>Connection pooling<br/>CRUD operations]
            DAL_CACHE[Redis Cache<br/>Session storage<br/>Performance]
        end
    end

    LOGIN --> AS
    NAV --> AS
    GUARD --> AS

    CF --> CTS
    CTM --> CTS
    BTM --> TS
    STM --> TS
    TF --> TS
    APP --> CTS
    CS --> CTS
    LV --> LS
    LED_BANK --> LS

    AS --> UC
    CTS --> CC
    TS --> TC
    LS --> CC

    CC --> COS
    TC --> TOS
    UC --> AUS

    COS --> BC
    TOS --> BC
    AUS --> BC

    %% Backend to Peer Services
    BC --> MB_API
    BC --> SUP_API
    BC --> ANC_API

    %% Peer Services to Chaincode Service
    MB_HANDLER --> MB_CHAINCODE
    SUP_HANDLER --> SUP_CHAINCODE
    ANC_HANDLER --> ANC_CHAINCODE

    MB_CHAINCODE --> CHAINCODE
    SUP_CHAINCODE --> CHAINCODE
    ANC_CHAINCODE --> CHAINCODE

    %% Peer Services to Orderer (gRPC Direct)
    MB_HANDLER --> MB_GRPC
    SUP_HANDLER --> SUP_GRPC
    ANC_HANDLER --> ANC_GRPC

    MB_GRPC --> ORD1_API
    SUP_GRPC --> ORD1_API
    ANC_GRPC --> ORD1_API

    %% PBFT Consensus Process
    ORD1_PBFT --> ORD2_PBFT
    ORD1_PBFT --> ORD3_PBFT
    ORD2_PBFT --> ORD1_PBFT
    ORD3_PBFT --> ORD1_PBFT

    %% Orderer Processing and Database Storage
    ORD1_PROC --> ORD_DB
    ORD2_PROC --> ORD_DB
    ORD3_PROC --> ORD_DB

    %% Orderers stream blocks directly to peers
    ORD1_BLOCK --> MB_GRPC
    ORD1_BLOCK --> SUP_GRPC
    ORD1_BLOCK --> ANC_GRPC

    %% Database connections
    MB_API --> MB_DB
    SUP_API --> SUP_DB
    ANC_API --> ANC_DB

    ORD1_API --> ORD_DB
    ORD2_API --> ORD_DB
    ORD3_API --> ORD_DB

    %% Business logic connections
    CH --> CB
    TH --> TB
    LH --> LB

    CB --> PB_VAL
    TB --> PB_VAL
    CB --> PB_ES
    TB --> PB_ES
    CB --> PB_SEC
    TB --> PB_SEC

    PB_ES --> ORD_CONS
    ORD_CONS --> ORD_CRYPTO
    ORD_CONS --> ORD_LEDGER

    %% Data layer connections
    DAL_MODELS --> MDB_SHARED
    DAL_MODELS --> MDB_MAIN
    DAL_MODELS --> MDB_SUP
    DAL_MODELS --> MDB_ANC

    DAL_DRIVER --> MDB_SHARED
    DAL_DRIVER --> MDB_MAIN
    DAL_DRIVER --> MDB_SUP
    DAL_DRIVER --> MDB_ANC

    DAL_CACHE --> DAL_DRIVER
```

## Deployment Architecture

### Kiến trúc triển khai với Docker Compose - Multi-layer Architecture:

```mermaid
graph TB
    subgraph "Public Network (public-network)"
        subgraph "Frontend Layer"
            NGINX[nginx:alpine<br/>Angular 17 SPA<br/>Port 4200:80]
            ANG[Angular Build<br/>Static files<br/>TypeScript 5.2]
        end

        subgraph "API Gateway Layer"
            JAVA[Spring Boot 3.1.5<br/>API Gateway<br/>Port 8080:8080]
            SEC[Spring Security<br/>JWT + CORS]
            MUL[File Upload<br/>Multipart handling]
        end

        subgraph "Shared Database Layer"
            MONGO_SHARED[MongoDB Shared<br/>Events & Users<br/>Port 27017:27017]
            INIT_SHARED[init-mongo.js<br/>User setup]
            DATA_SHARED[Persistent volumes<br/>Event storage]
        end
    end

    subgraph "gRPC Communication Layer"
        subgraph "SCF Chaincode Service"
            CHAINCODE[blockchain-gw: golang<br/>Transaction Aggregator<br/>Port 9090:9090]
            CHAINCODE_CONF[Gateway config<br/>gRPC server<br/>Transaction aggregation]
        end

        subgraph "gRPC Direct APIs"
            GRPC_APIS[gRPC APIs<br/>SubmitTx, StreamBlocks<br/>PBFT Consensus]
            GRPC_CONF[gRPC config<br/>Direct communication<br/>No message broker]
        end
    end

    subgraph "Orderer Network (orderer-network)"
        subgraph "Orderer Cluster Layer"
            ORD1[orderer-ord1: golang<br/>Leader Node<br/>Port 7050:7050]
            ORD2[orderer-ord2: golang<br/>Follower Node<br/>Port 7060:7060]
            ORD3[orderer-ord3: golang<br/>Follower Node<br/>Port 7070:7070]
            ORD_CONF[Orderer config<br/>PBFT consensus<br/>MongoDB URI]
        end
    end

    subgraph "Peer Network (peer-network)"
        subgraph "Peer Main Bank Layer"
            MB_PEER[peer-main-bank: golang<br/>Main Bank Peer<br/>Port 8082:8082]
            MB_MONGO[mongo-main-bank<br/>Bank world state]
            MB_CONF[Bank config<br/>gRPC clients<br/>MongoDB URI]
        end

        subgraph "Peer Supplier Layer"
            SUP_PEER[peer-supplier: golang<br/>Supplier Peer<br/>Port 8083:8083]
            SUP_MONGO[mongo-supplier<br/>Supplier world state]
            SUP_CONF[Supplier config<br/>gRPC clients<br/>MongoDB URI]
        end

        subgraph "Peer Anchor Layer"
            ANC_PEER[peer-anchor: golang<br/>Anchor Peer<br/>Port 8084:8084]
            ANC_MONGO[mongo-anchor<br/>Anchor world state]
            ANC_CONF[Anchor config<br/>gRPC clients<br/>MongoDB URI]
        end
    end

    %% Network connections
    NGINX --> JAVA
    JAVA --> MB_PEER
    JAVA --> SUP_PEER
    JAVA --> ANC_PEER

    %% gRPC Direct Communication
    MB_PEER --> ORD1
    SUP_PEER --> ORD1
    ANC_PEER --> ORD1

    MB_PEER --> CHAINCODE
    SUP_PEER --> CHAINCODE
    ANC_PEER --> CHAINCODE

    ORD1 --> MONGO_SHARED
    ORD2 --> MONGO_SHARED
    ORD3 --> MONGO_SHARED

    MB_PEER --> MB_MONGO
    SUP_PEER --> SUP_MONGO
    ANC_PEER --> ANC_MONGO

    %% gRPC Direct Communication
    MB_PEER -->|gRPC Direct| ORD1
    SUP_PEER -->|gRPC Direct| ORD1
    ANC_PEER -->|gRPC Direct| ORD1

    subgraph "External Access Points"
        subgraph "Web Interfaces"
            ANC_UI[Anchor Portal<br/>localhost:4200/anchor<br/>Contract creation]
            BANK_UI[Bank Dashboard<br/>localhost:4200/bank<br/>Token monitoring]
            SUP_UI[Supplier Portal<br/>localhost:4200/supplier<br/>Token management]
        end

        subgraph "Direct API Access"
            MB_API[Main Bank API<br/>localhost:8082<br/>Contract & Token APIs]
            SUP_API[Supplier API<br/>localhost:8083<br/>Transfer APIs]
            ANC_API[Anchor API<br/>localhost:8084<br/>Contract APIs]
        end

        subgraph "System Monitoring"
            DB_EXT[MongoDB Direct<br/>localhost:27017<br/>Development access]
            GRPC_EXT[gRPC Tools<br/>grpcurl, Postman<br/>API inspection]
        end
    end

    ANC_UI --> NGINX
    BANK_UI --> NGINX
    SUP_UI --> NGINX

    MB_API --> MB_PEER
    SUP_API --> SUP_PEER
    ANC_API --> ANC_PEER

    DB_EXT --> MONGO_SHARED
    GRPC_EXT --> GRPC_APIS

    subgraph "Docker Networks"
        PUB_NET[public-network<br/>Frontend, Gateway, Shared DB]
        GRPC_NET[grpc-network<br/>Chaincode, gRPC APIs]
        ORDERER_NET[orderer-network<br/>Orderer cluster]
        PEER_NET[peer-network<br/>Peer services]
    end

    subgraph "Service Dependencies"
        DEP1[depends_on scf-chaincode<br/>All peers]
        DEP2[depends_on mongo-shared<br/>Orderers only]
        DEP3[depends_on mongo-*<br/>Each peer to its DB]
        DEP4[depends_on orderer-ord1<br/>Peers for ordering]
    end

    subgraph "Volume Mounts"
        VOL1[./orderer-cluster/genesis<br/>Orderer genesis config]
        VOL2[./init-mongo.js<br/>Database initialization]
        VOL3[Persistent volumes<br/>Data persistence]
    end

    subgraph "Environment Variables"
        ENV_GRPC[ORDERER_ENDPOINTS=orderer-ord1:7050<br/>Peer gRPC config]
        ENV_MONGO[MONGO_URI=mongodb://...<br/>Service-specific]
        ENV_PEER[PEER_NODE_TYPE=...<br/>Peer identification]
        ENV_ORDERER[ORDERER_NODE_ID=...<br/>Orderer identification]
    end
```

### Network Architecture - Multi-Network Isolation:

```mermaid
graph TB
    subgraph "🔒 Public Network (public-network)"
        subgraph "External Access Layer"
            INTERNET((Internet))
            DEV_TOOLS[Development Tools<br/>localhost:4200, 8080, 27017]
        end

        subgraph "Web Tier"
            NGINX[nginx<br/>Port 4200<br/>Static Files]
            SPRING[Spring Boot<br/>Port 8080<br/>API Gateway]
        end

        subgraph "Shared Services"
            MONGO_SHARED[MongoDB Shared<br/>Port 27017<br/>Events, Users]
        end

        INTERNET --> NGINX
        INTERNET --> SPRING
        INTERNET --> DEV_TOOLS
        NGINX --> SPRING
        SPRING --> MONGO_SHARED
    end

    subgraph "📡 gRPC Communication Network"
        subgraph "Chaincode Infrastructure"
            CHAINCODE[blockchain-gw<br/>Port 9090<br/>Transaction Aggregator<br/>Gateway Service]

            subgraph "gRPC Methods"
                CONTRACT_METHODS[CreateContract<br/>ApproveContract<br/>FinalizeContract]
                TOKEN_METHODS[IssueToken<br/>TransferToken<br/>SettleToken]
            end
        end

        subgraph "Direct Communication"
            GRPC_APIS[gRPC APIs<br/>SubmitTx<br/>StreamBlocks<br/>Consensus]
        end
    end

    subgraph "🏛️ Orderer Network (orderer-network)"
        subgraph "Ordering Service"
            ORDERER1[orderer-ord1<br/>Port 7050<br/>✅ Leader<br/>PBFT Consensus + DB]
            ORDERER2[orderer-ord2<br/>Port 7060<br/>Follower<br/>PBFT Participant]
            ORDERER3[orderer-ord3<br/>Port 7070<br/>Follower<br/>PBFT Participant]
        end

        ORDERER1 -->|PBFT Consensus| ORDERER2
        ORDERER1 -->|PBFT Consensus| ORDERER3
        ORDERER2 -->|PBFT Consensus| ORDERER1
        ORDERER3 -->|PBFT Consensus| ORDERER1

        ORDERER1 -->|Block Storage| MONGO_SHARED
        ORDERER2 -->|Block Storage| MONGO_SHARED
        ORDERER3 -->|Block Storage| MONGO_SHARED
    end

    subgraph "🏢 Peer Network (peer-network)"
        subgraph "Endorsing Peers"
            PEER_MAIN_BANK[peer-main-bank<br/>Port 8082<br/>✅ Main Bank<br/>gRPC Direct]
            PEER_SUPPLIER[peer-supplier<br/>Port 8083<br/>✅ Supplier<br/>gRPC Direct]
            PEER_ANCHOR[peer-anchor<br/>Port 8084<br/>✅ Anchor<br/>gRPC Direct]
        end

        subgraph "Peer Databases"
            MONGO_MAIN_BANK[mongo-main-bank<br/>Bank World State]
            MONGO_SUPPLIER[mongo-supplier<br/>Supplier World State]
            MONGO_ANCHOR[mongo-anchor<br/>Anchor World State]
        end

        PEER_MAIN_BANK -->|gRPC SubmitTx| ORDERER1
        PEER_SUPPLIER -->|gRPC SubmitTx| ORDERER1
        PEER_ANCHOR -->|gRPC SubmitTx| ORDERER1

        PEER_MAIN_BANK -->|state| MONGO_MAIN_BANK
        PEER_SUPPLIER -->|state| MONGO_SUPPLIER
        PEER_ANCHOR -->|state| MONGO_ANCHOR
    end

    %% Cross-network communication via gRPC
    ORDERER1 -.->|stream blocks| PEER_MAIN_BANK
    ORDERER1 -.->|stream blocks| PEER_SUPPLIER
    ORDERER1 -.->|stream blocks| PEER_ANCHOR

    PEER_MAIN_BANK -.->|direct gRPC| ORDERER1
    PEER_SUPPLIER -.->|direct gRPC| ORDERER1
    PEER_ANCHOR -.->|direct gRPC| ORDERER1

    %% API Gateway to Peers
    SPRING -->|proxies| PEER_MAIN_BANK
    SPRING -->|proxies| PEER_SUPPLIER
    SPRING -->|proxies| PEER_ANCHOR

    subgraph "Network Security Zones"
        ZONE_PUBLIC[🌐 Public Zone<br/>External access<br/>Load balancer]
        ZONE_GRPC[📡 Communication Zone<br/>gRPC direct calls<br/>Chaincode service]
        ZONE_ORDERER[🏛️ Ordering Zone<br/>Transaction ordering<br/>Trusted nodes only]
        ZONE_PEER[🏢 Peer Zone<br/>Business logic<br/>Endorsing peers]
    end

    subgraph "Network Flow Summary"
        FLOW1[Public → Peers<br/>API calls via Gateway]
        FLOW2[Peers → Chaincode<br/>Business logic via gRPC]
        FLOW3[Peers → Orderer<br/>SubmitTx via gRPC direct]
        FLOW4[Orderer → Orderer<br/>PBFT consensus replication]
        FLOW5[Orderer → Shared DB<br/>Blocks & events persisted]
        FLOW6[Orderer → Peers<br/>StreamBlocks via gRPC]
        FLOW7[Peers → Peer DBs<br/>World state updates]
    end
```

### Docker Compose Services - Detailed Breakdown:

#### **Peer Services (Endorsing Peers)**
| Service | Network | Ports | Database | gRPC Role | Status |
|---------|---------|-------|----------|-----------|--------|
| **peer-main-bank** | peer-network | 8082:8082 | mongo-private | Client (Orderer + Chaincode) | ✅ Running |
| **peer-supplier** | peer-network | 8083:8083 | mongo-private | Client (Orderer + Chaincode) | ✅ Running |
| **peer-anchor** | peer-network | 8084:8084 | mongo-private | Client (Orderer + Chaincode) | ✅ Running |

#### **Orderer Services (PBFT Consensus)**
| Service | Network | Ports | PBFT Role | Database | Status |
|---------|---------|-------|-----------|----------|--------|
| **orderer-ord1** | orderer-network | 7050:7050 | Leader + Storage | mongo-shared | ✅ Running |
| **orderer-ord2** | orderer-network | 7060:7060 | Follower | - | ✅ Running |
| **orderer-ord3** | orderer-network | 7070:7070 | Follower | - | ✅ Running |

#### **Chaincode Services**
| Service | Network | Ports | Purpose | Status |
|---------|---------|-------|---------|--------|
| **scf-chaincode** | peer-network | 9090:9090 | Smart contracts engine | ✅ Running |
| **mongo-shared** | public-network | 27017:27017 | Event storage | ✅ Running |
| **mongo-main-bank** | peer-network | - | Bank world state | ✅ Running |
| **mongo-supplier** | peer-network | - | Supplier world state | ✅ Running |
| **mongo-anchor** | peer-network | - | Anchor world state | ✅ Running |

#### **Application Services**
| Service | Network | Ports | Purpose | Status |
|---------|---------|-------|---------|--------|
| **frontend** | public-network | 4200:80 | Angular UI | ✅ Running |
| **backend** | public-network | 8080:8080 | Spring API Gateway | ✅ Running |

### Network Architecture - Security & Isolation:

#### **4-Layer Network Design** 🔒

| Network | Purpose | Services | Security Level | Access Pattern |
|---------|---------|----------|----------------|----------------|
| **public-network** | External access | Frontend, Backend, Mongo-Shared | 🌐 Public | Internet → Services |
| **grpc-network** | Direct communication | Chaincode, gRPC APIs | 📡 Internal | gRPC direct calls |
| **orderer-network** | PBFT consensus | Orderer cluster | 🏛️ Trusted | PBFT replication + Shared DB |
| **peer-network** | Business logic | Peer services + DBs | 🏢 Restricted | API Gateway + gRPC direct |

#### **Cross-Network Communication Flow** 🔄

```mermaid
flowchart LR
    subgraph "User Request"
        UI[Frontend UI<br/>Port 4200]
        API[API Gateway<br/>Port 8080]
    end

    subgraph "Business Logic Layer"
        PEER1[Peer Main Bank<br/>Port 8082]
        PEER2[Peer Supplier<br/>Port 8083]
        PEER3[Peer Anchor<br/>Port 8084]
    end

    subgraph "Chaincode Layer"
        CHAINCODE[Blockchain Gateway<br/>Port 9090]
        METHODS[Gateway Methods<br/>AggregateTransactions, SubmitToOrderer]
    end

    subgraph "Ordering Layer"
        ORDERER[Orderer Cluster<br/>Ports 7050-7070]
    end

    subgraph "Storage Layer"
        SHARED_DB[MongoDB Shared<br/>Port 27017]
        PEER_DB[Peer Databases<br/>Isolated per peer]
    end

    UI --> API
    API --> PEER1
    API --> PEER2
    API --> PEER3

    PEER1 --> CHAINCODE
    PEER2 --> CHAINCODE
    PEER3 --> CHAINCODE

    CHAINCODE --> ORDERER
    ORDERER --> SHARED_DB

    ORDERER --> PEER1
    ORDERER --> PEER2
    ORDERER --> PEER3

    PEER1 --> PEER_DB
    PEER2 --> PEER_DB
    PEER3 --> PEER_DB
```

#### **Key Architecture Principles** 🏗️

- **🔒 Network Isolation**: Each layer runs in separate Docker networks
- **📡 Direct Communication**: gRPC enables real-time peer-to-orderer communication
- **🏗️ Decoupled Architecture**: Chaincode service separates business logic from peers
- **🏛️ Consensus Separation**: Orderers handle PBFT consensus, peers handle business logic
- **💾 Data Segregation**: Shared events vs. private world state databases
- **🔄 Fault Tolerance**: PBFT consensus with f=1 fault tolerance across orderer cluster
- **📊 Scalability**: Horizontal scaling possible for each layer independently

## Business Process Summary

### Quy trình Supply Chain Finance hoàn chỉnh:

1. **Anchor tạo hợp đồng** → Upload file PDF + phân bổ suppliers → Bank tự động phát hành token
2. **Suppliers phê duyệt** → Khi tất cả approve → Token transfer từ Bank → Supplier + Contract executed
3. **Token circulation** → Suppliers có thể transfer tokens cho nhau (P2P transfers)
4. **Bank monitoring** → Xem tất cả tokens đã phát hành + audit trail qua blockchain ledger
5. **Immutable audit** → Tất cả giao dịch được ghi vào blockchain với block building tự động

### Các tính năng chính:
- **Permissioned Blockchain**: Chỉ authorized nodes tham gia
- **Event Sourcing**: State derive từ event history
- **Token Issuance**: Auto token creation per contract
- **Token Transfer**: Secure P2P token transfers
- **Audit Trail**: Immutable ledger với Merkle trees

## API Endpoints Summary

### Backend REST APIs (Spring Boot):

| Method | Endpoint | Description | Authentication |
|--------|----------|-------------|----------------|
| POST | `/api/auth/login` | User authentication | - |
| POST | `/api/contracts` | Anchor tạo hợp đồng với file | JWT Bearer |
| GET | `/api/contracts` | Lấy danh sách contracts | JWT Bearer |
| POST | `/api/contracts/{id}/approve` | Supplier phê duyệt contract | JWT Bearer |
| GET | `/api/contracts/{id}/ledger` | Xem contract ledger | JWT Bearer |
| GET | `/api/tokens/{id}` | Thông tin chi tiết token | JWT Bearer |
| POST | `/api/tokens/transfer` | Transfer token giữa users | JWT Bearer |
| GET | `/api/tokens/issued/{bankId}` | Bank xem tokens đã phát hành | JWT Bearer |

### Peer Service APIs (Go Peer Services):

#### **Peer Anchor (Port 8084) - Contract Creation:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/create` | Anchor tạo contract + ledger entry + gRPC SubmitTx |
| GET | `/contract/{id}` | Chi tiết contract từ anchor perspective |
| GET | `/contract/list` | Danh sách contracts đã tạo |

#### **Peer Supplier (Port 8083) - Token Operations:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/{id}/approve` | Supplier phê duyệt contract + token transfer |
| POST | `/token/transfer` | Token transfer giữa suppliers |
| POST | `/token/settle` | Supplier settle token với bank + gRPC SubmitTx |
| GET | `/balances/account/{accountId}` | Account balances của supplier |
| GET | `/health` | Health check |

#### **Peer Main Bank (Port 8082) - Bank Operations:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/{id}/approve-bank` | Bank phê duyệt + token issuance |
| GET | `/contract/list` | Bank xem tất cả contracts |
| GET | `/contract/{id}/ledger` | Contract audit trail |
| GET | `/token/issued/{bankId}` | Tokens issued by bank |
| GET | `/tokens` | All tokens issued by bank |

#### **Common Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/token/{id}` | Token information |
| GET | `/balances/token/{tokenId}` | Token balances across all accounts |
| GET | `/suppliers` | Supplier list |
| POST | `/blocks/hash/update` | Update block hashes |

### Database Collections:
- `users`: User accounts với roles
- `contracts`: SCF contracts với file URLs
- `tokens`: Digital assets từ contracts
- `balances`: Account balances per token
- `events`: Blockchain events (event sourcing)
- `blocks`: Immutable ledger blocks

### Key Features:
- **JWT Authentication**: Role-based access control
- **File Upload**: Contract PDF storage
- **Event Sourcing**: State từ event history
- **Auto Block Building**: Every 10 seconds
- **Merkle Trees**: Block integrity
- **Audit Trail**: Immutable transaction history

---

## 🎯 **Tóm tắt trạng thái hiện tại**

### ✅ **Đã hoàn thành:**
- **9/9 containers** chạy ổn định với gRPC Direct Communication
- **PBFT consensus** end-to-end hoạt động với 3 orderer nodes
- **gRPC communication** giữa Peers ↔ Orderer cluster
- **ECDSA signatures** cho block validation (2f+1 quorum)
- **Real-time block streaming** từ Orderer → Peers
- **Peer services** với gRPC clients thay vì Kafka producers
- **REST APIs** hoàn chỉnh cho SCF workflow

### 🔄 **Đang hoạt động:**
- **PBFT consensus flow**: Pre-Prepare → Prepare → Commit → Finalized blocks
- **Transaction ordering**: Peers → Blockchain Gateway → Orderer mempool → Consensus
- **Block streaming**: Orderers stream finalized blocks với ECDSA signatures
- **Health checks**: Tất cả services có health endpoints
- **gRPC Direct Communication**: Real-time peer-to-orderer communication
- **PBFT Consensus**: 3-node cluster với fault tolerance f=1
- **Blockchain Gateway**: Decoupled business logic trong microservice riêng
- **State Management**: Blockchain gateway quản lý tất cả contract và token state

### 📋 **Next Steps có thể mở rộng:**
1. **Frontend Integration**: Kết nối Angular UI với peer APIs
2. **Authentication**: Implement JWT cho API security
3. **File Upload**: Contract PDF upload functionality
4. **View Changes**: PBFT view change protocol cho fault tolerance
5. **Monitoring**: Metrics và logging nâng cao cho gRPC & PBFT
6. **Performance**: Optimize consensus latency và gRPC throughput

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

# Events Sync verification - check synchronized events
docker-compose exec mongo-shared mongosh --username root --password example --authenticationDatabase admin \
blockchain_public --eval "db.events.find({}, {eventType: 1, contractId: 1, tokenId: 1, amount: 1, supplierId: 1, bankId: 1, from: 1, to: 1}).sort({_id: 1})"

# Private blockchain verification
docker-compose exec mongo-shared mongosh --username root --password example --authenticationDatabase admin \
blockchain_private --eval "db.events.find({}, {eventType: 1, contractId: 1}).sort({_id: 1})"

# Database verification - check PBFT signed blocks
docker-compose exec mongo-shared mongosh --username root --password example --authenticationDatabase admin \
blockchain_public --eval "db.blocks.find().sort({height: -1}).limit(3).toArray()"
```

**🎉 Hệ thống blockchain Supply Chain Finance đã hoàn thành với gRPC Direct Communication & PBFT Consensus Architecture!**
