# Hệ thống Blockchain Supply Chain Finance (SCF) với Token Management

## Tổng quan kiến trúc

Hệ thống blockchain permissioned cho Supply Chain Finance với đầy đủ chức năng quản lý hợp đồng và token transfer, bao gồm:

- **Permissioned Blockchain**: Chỉ các node được ủy quyền tham gia
- **Event Sourcing**: Tất cả thay đổi được ghi thành events
- **Token Issuance**: Tự động phát hành token khi tạo hợp đồng
- **Token Transfer**: Chuyển giao token giữa các bên
- **Immutable Ledger**: Audit trail không thể thay đổi

```mermaid
graph TB
    subgraph "Frontend (Angular 17)"
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

    subgraph "Backend (Spring Boot 3.1.5)"
        API[REST API Gateway<br/>Port 8080]
        AUTH_S[Authentication Service<br/>JWT Tokens]
        CONT[Contract Controller<br/>File Upload, CRUD]
        TOK_S[Token Service<br/>Transfer Logic]
        BC_C[Blockchain Client<br/>ms-blockchain integration]
    end

    subgraph "ms-blockchain (Go 1.21)"
        REST[REST API<br/>Port 8081]
        subgraph "Business Logic"
            CON_H[Contract Handlers<br/>Create, Approve]
            TOK_H[Token Handlers<br/>Transfer, Query]
            LED_H[Ledger Handlers<br/>Blocks, Events]
        end
        BLDR[Block Builder<br/>Auto-build every 10s]
    end

    subgraph "MongoDB (World State)"
        USR[users collection<br/>Role-based auth]
        CON[contracts collection<br/>SCF contracts + files]
        TOK[tokens collection<br/>Digital assets]
        BAL[balances collection<br/>Account balances]
        EVT[events collection<br/>Blockchain events]
        BLK[blocks collection<br/>Immutable ledger]
    end

    ANC --> API
    BNK --> API
    SUP --> API

    API --> AUTH_S
    API --> CONT
    API --> TOK_S
    CONT --> BC_C
    TOK_S --> BC_C

    BC_C --> REST
    REST --> CON_H
    REST --> TOK_H
    REST --> LED_H

    CON_H --> USR
    CON_H --> CON
    TOK_H --> TOK
    TOK_H --> BAL
    LED_H --> EVT
    LED_H --> BLK

    BLDR --> EVT
    BLDR --> BLK
```

## Luồng nghiệp vụ chi tiết

### Quy trình Supply Chain Finance hoàn chỉnh:

```mermaid
sequenceDiagram
    participant Anchor
    participant Bank
    participant Supplier
    participant Frontend
    participant Backend
    participant MSBlockchain
    participant MongoDB

    %% 1. Anchor tạo hợp đồng & Token tự động phát hành
    rect rgb(240, 248, 255)
        Note over Anchor,MongoDB: Tạo hợp đồng + Token tự động
        Anchor->>Frontend: Submit contract form (with PDF file)
        Frontend->>Backend: POST /api/contracts (multipart/form-data)
        Backend->>MSBlockchain: POST /contract/create
        MSBlockchain->>MongoDB: Save contract metadata + file
        MSBlockchain->>MongoDB: Auto-create token (issuer=Bank, owner=Bank)
        MSBlockchain->>MongoDB: Create balance record (Bank=totalAmount)
        MSBlockchain->>MongoDB: Insert CREATE event
        MSBlockchain-->>Backend: Contract + Token created
        Backend-->>Frontend: Success response
        Frontend-->>Anchor: Contract created notification
    end

    %% 2. Suppliers phê duyệt & Token transfer
    rect rgb(255, 248, 240)
        Note over Supplier,MongoDB: Phê duyệt hợp đồng
        Supplier->>Frontend: Click approve button
        Frontend->>Backend: POST /api/contracts/{id}/approve
        Backend->>MSBlockchain: POST /contract/approve
        MSBlockchain->>MongoDB: Check contract status & approver auth
        MSBlockchain->>MongoDB: Insert APPROVE_SUPPLIER event

        alt All suppliers approved
            MSBlockchain->>MongoDB: Update contract.approved = true
            MSBlockchain->>MongoDB: Transfer token ownership (Bank → Supplier)
            MSBlockchain->>MongoDB: Update balances (Bank=0, Supplier=amount)
            MSBlockchain->>MongoDB: Insert EXECUTE event
        end

        MSBlockchain-->>Backend: Approval successful
        Backend-->>Frontend: Contract updated
        Frontend-->>Supplier: Approval confirmed
    end

    %% 3. Token transfer giữa Suppliers
    rect rgb(248, 255, 240)
        Note over Supplier,MongoDB: Chuyển token
        Supplier->>Frontend: Initiate token transfer
        Frontend->>Backend: POST /api/tokens/transfer
        Backend->>MSBlockchain: POST /token/transfer
        MSBlockchain->>MongoDB: Validate sender ownership & balance
        MSBlockchain->>MongoDB: Debit sender balance
        MSBlockchain->>MongoDB: Credit receiver balance
        MSBlockchain->>MongoDB: Update token ownership
        MSBlockchain->>MongoDB: Insert TRANSFER event
        MSBlockchain-->>Backend: Transfer successful
        Backend-->>Frontend: Success response
        Frontend-->>Supplier: Transfer confirmed
    end

    %% 4. Block builder tự động
    rect rgb(255, 240, 248)
        Note over MSBlockchain,MongoDB: Block building định kỳ
        loop Every 10 seconds
            MSBlockchain->>MongoDB: Find unincluded events (max 10)
            alt Events found
                MSBlockchain->>MSBlockchain: Calculate Merkle root
                MSBlockchain->>MSBlockchain: Generate block hash
                MSBlockchain->>MongoDB: Create new block
                MSBlockchain->>MongoDB: Mark events as included
            end
        end
    end

    %% 5. Bank xem tổng quan tokens
    rect rgb(240, 255, 248)
        Note over Bank,MongoDB: Bank monitoring
        Bank->>Frontend: View issued tokens dashboard
        Frontend->>Backend: GET /api/tokens/issued/{bankId}
        Backend->>MSBlockchain: GET /token/issued/{bankId}
        MSBlockchain->>MongoDB: Query tokens by issuer
        MongoDB-->>MSBlockchain: Token list with current owners
        MSBlockchain-->>Backend: Formatted token data
        Backend-->>Frontend: Token overview
        Frontend-->>Bank: Display token dashboard
    end

    %% 6. Ledger audit trail
    rect rgb(248, 240, 255)
        Note over Bank,MongoDB: Audit & compliance
        Bank->>Frontend: View contract ledger
        Frontend->>Backend: GET /api/contracts/{id}/ledger
        Backend->>MSBlockchain: GET /contract/{id}/ledger
        MSBlockchain->>MongoDB: Query blocks containing contract events
        MongoDB-->>MSBlockchain: Block data with events
        MSBlockchain-->>Backend: Immutable ledger data
        Backend-->>Frontend: Audit trail
        Frontend-->>Bank: Display blockchain ledger
    end
```

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

### Kiến trúc components chi tiết:

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

    subgraph "Spring Boot 3.1.5 Backend"
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
            BC[Blockchain Client<br/>ms-blockchain proxy]
        end

        subgraph "Security"
            SEC[Spring Security<br/>JWT + CORS]
            UP[File Upload<br/>Multipart handling]
        end
    end

    subgraph "Go 1.21 ms-blockchain"
        subgraph "HTTP Handlers"
            CH[Contract Handlers<br/>Create, Approve, Query]
            TH[Token Handlers<br/>Transfer, Balance, Query]
            LH[Ledger Handlers<br/>Blocks, Events]
            UH[Utility Handlers<br/>Health, Metrics]
        end

        subgraph "Business Logic"
            CB[Contract Business<br/>Validation + State]
            TB[Token Business<br/>Transfer + Ownership]
            LB[Ledger Business<br/>Event sourcing]
        end

        subgraph "Blockchain Core"
            BB[Block Builder<br/>Auto block creation]
            BC_VAL[Block Validation<br/>Hash + Merkle]
            ES[Event Sourcing<br/>State derivation]
        end

        subgraph "Data Access"
            MD[Models<br/>Data structures]
            DB[MongoDB Driver<br/>Database operations]
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

    BC --> CH
    BC --> TH
    BC --> LH

    CH --> CB
    TH --> TB
    LH --> LB

    CB --> ES
    TB --> ES
    LB --> ES

    ES --> BB
    BB --> BC_VAL

    CB --> MD
    TB --> MD
    LB --> MD

    MD --> DB
    BC_VAL --> DB
```

## Deployment Architecture

### Kiến trúc triển khai với Docker Compose:

```mermaid
graph TB
    subgraph "Docker Compose Network (blockchain-network)"
        subgraph "Frontend Service"
            NGINX[nginx:80<br/>Angular 17 SPA<br/>Port 4200]
            ANG[Angular Build<br/>Static files<br/>TypeScript 5.2]
        end

        subgraph "Backend Service"
            JAVA[Java 17<br/>Spring Boot 3.1.5<br/>Port 8080]
            SEC[Spring Security<br/>JWT Authentication]
            MUL[Multipart Upload<br/>File handling]
        end

        subgraph "ms-blockchain Service"
            GO[Go 1.21<br/>Gorilla Mux<br/>Port 8081]
            BB[Block Builder<br/>Auto every 10s]
            CORS[CORS enabled<br/>Cross-origin]
        end

        subgraph "MongoDB Service"
            MONGO[MongoDB latest<br/>Auth enabled<br/>Port 27017]
            INIT[init-mongo.js<br/>Database setup]
            DATA[Persistent volumes<br/>Data storage]
        end
    end

    NGINX --> JAVA
    JAVA --> GO
    JAVA --> MONGO
    GO --> MONGO
    GO --> BB

    BB -.-> MONGO

    subgraph "External Access Points"
        subgraph "User Interfaces"
            ANC_UI[Anchor UI<br/>localhost:4200/anchor<br/>Contract creation]
            BANK_UI[Bank UI<br/>localhost:4200/bank<br/>Token monitoring]
            SUP_UI[Supplier UI<br/>localhost:4200/supplier<br/>Token management]
        end

        subgraph "API Endpoints"
            REST_API[REST APIs<br/>localhost:8080<br/>JSON responses]
            BC_API[Blockchain APIs<br/>localhost:8081<br/>Direct blockchain]
        end

        subgraph "Database Access"
            DB_EXT[MongoDB External<br/>localhost:27017<br/>For development]
        end
    end

    ANC_UI --> NGINX
    BANK_UI --> NGINX
    SUP_UI --> NGINX

    REST_API --> JAVA
    BC_API --> GO
    DB_EXT --> MONGO

    subgraph "Environment Configuration"
        ENV1[Frontend<br/>Dockerfile<br/>nginx.conf]
        ENV2[Backend<br/>application.yml<br/>JVM settings]
        ENV3[Blockchain<br/>config.go<br/>MongoDB URI]
        ENV4[MongoDB<br/>docker-compose.yml<br/>Auth credentials]
    end

    subgraph "Development Tools"
        LOGS[docker-compose logs<br/>Real-time monitoring]
        SHELL[docker-compose exec<br/>Container access]
        DEBUG[Hot reload<br/>Development mode]
    end

    subgraph "Production Considerations"
        SECURE[Security hardening<br/>Secrets management]
        SCALE[Horizontal scaling<br/>Load balancing]
        MONITOR[Health checks<br/>Metrics collection]
        BACKUP[Data backup<br/>Disaster recovery]
    end
```

### Docker Compose Services:

| Service | Image | Ports | Dependencies | Environment |
|---------|-------|-------|--------------|-------------|
| **frontend** | nginx:alpine | 4200:80 | backend | Static files |
| **backend** | openjdk:17 | 8080:8080 | mongo, ms-blockchain | Spring profiles |
| **ms-blockchain** | golang:1.21 | 8081:8081 | mongo | MongoDB URI |
| **mongo** | mongo:latest | 27017:27017 | - | Root credentials |

### Network Architecture:
- **Internal Network**: `blockchain-network` (bridge driver)
- **Service Discovery**: DNS resolution giữa containers
- **Security**: Isolated network, controlled access
- **Scalability**: Services có thể scale independently

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

### Blockchain APIs (Go ms-blockchain):

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/contract/create` | Tạo contract + auto token issuance |
| POST | `/contract/approve` | Phê duyệt contract + token transfer |
| GET | `/contract/{id}` | Chi tiết contract |
| GET | `/contract/list` | Danh sách contracts |
| GET | `/contract/{id}/ledger` | Contract audit trail |
| GET | `/token/{id}` | Token information |
| POST | `/token/transfer` | Direct token transfer |
| GET | `/token/issued/{bankId}` | Tokens issued by bank |
| GET | `/tokens` | All tokens |
| GET | `/balances/account/{accountId}` | Account balances |
| GET | `/balances/token/{tokenId}` | Token balances |
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
