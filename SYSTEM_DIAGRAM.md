# Hệ thống Blockchain + Contract + Token Issuance

## Tổng quan kiến trúc

```mermaid
graph TB
    subgraph "Frontend (Angular)"
        A[Anchor UI<br/>/anchor]
        B[Bank UI<br/>/bank]
        S[Supplier UI<br/>/supplier]
    end

    subgraph "Backend (Spring Boot)"
        API[API Gateway<br/>Port 8080]
        CT[ContractTokenController]
        CTS[ContractTokenService]
    end

    subgraph "ms-blockchain (Go)"
        REST[REST API<br/>Port 8081]
        CC[Chaincode<br/>contract.go + token.go]
        H[Handlers<br/>MongoDB integration]
    end

    subgraph "MongoDB World State"
        CON[contracts collection]
        TOK[tokens collection]
        BAL[balances collection]
        EVT[events collection]
    end

    A --> API
    B --> API
    S --> API

    API --> CT --> CTS --> REST
    REST --> H --> CC
    H --> CON
    H --> TOK
    H --> BAL
    H --> EVT
```

## Luồng nghiệp vụ chi tiết

```mermaid
sequenceDiagram
    participant Anchor
    participant Bank
    participant Supplier
    participant Backend
    participant MSBlockchain
    participant MongoDB

    %% 1. Anchor tạo hợp đồng
    Anchor->>Backend: POST /api/contracts
    Backend->>MSBlockchain: POST /v1/contracts
    MSBlockchain->>MongoDB: Save contract (approved=false)
    MSBlockchain->>MongoDB: Auto-issue token (owner=Bank)
    MSBlockchain->>MongoDB: Create balance (Bank=amount)
    MSBlockchain-->>Backend: Contract + Token created
    Backend-->>Anchor: Success response

    %% 2. Approvers duyệt hợp đồng
    Supplier->>Backend: POST /api/contracts/{id}/approve
    Backend->>MSBlockchain: POST /v1/contracts/{id}/approve
    MSBlockchain->>MongoDB: Update contract (approved=true)
    MSBlockchain->>MongoDB: Transfer token (Bank -> Supplier)
    MSBlockchain->>MongoDB: Update balance (Bank=0, Supplier=amount)
    MSBlockchain-->>Backend: Approval successful
    Backend-->>Supplier: Success response

    %% 3. Supplier chuyển token
    Supplier->>Backend: POST /api/tokens/transfer
    Backend->>MSBlockchain: POST /v1/tokens/transfer
    MSBlockchain->>MongoDB: Check balance & ownership
    MSBlockchain->>MongoDB: Update balances (from- , to+)
    MSBlockchain->>MongoDB: Update token owner
    MSBlockchain-->>Backend: Transfer successful
    Backend-->>Supplier: Success response

    %% 4. Bank xem tokens đã phát hành
    Bank->>Backend: GET /api/tokens/issued/{bankId}
    Backend->>MSBlockchain: GET /v1/tokens/issued/{bankId}
    MSBlockchain->>MongoDB: Query tokens by issuer
    MongoDB-->>MSBlockchain: Token list with current owners
    MSBlockchain-->>Backend: Token data
    Backend-->>Bank: Display token list
```

## Data Flow Architecture

```mermaid
flowchart TD
    subgraph "Contract Creation Flow"
        C1[Anchor submits contract form]
        C2[Validate contract data]
        C3[Generate contract ID]
        C4[Save to contracts collection]
        C5[Auto-generate token ID]
        C6[Save to tokens collection]
        C7[Create initial balance for Bank]
        C8[Return success response]
    end

    subgraph "Contract Approval Flow"
        A1[Approver submits approval]
        A2[Validate approver authorization]
        A3[Check contract not already approved]
        A4[Update contract.approved = true]
        A5[Transfer token ownership: Bank → Supplier]
        A6[Update balances: Bank=0, Supplier=amount]
        A7[Return approval response]
    end

    subgraph "Token Transfer Flow"
        T1[Supplier initiates transfer]
        T2[Validate sender ownership]
        T3[Check sufficient balance]
        T4[Debit sender balance]
        T5[Credit receiver balance]
        T6[Update token ownership]
        T7[Return transfer response]
    end

    C1 --> C2 --> C3 --> C4 --> C5 --> C6 --> C7 --> C8
    A1 --> A2 --> A3 --> A4 --> A5 --> A6 --> A7
    T1 --> T2 --> T3 --> T4 --> T5 --> T6 --> T7
```

## Database Schema

```mermaid
erDiagram
    CONTRACT ||--o{ TOKEN : issues
    TOKEN ||--o{ BALANCE : has
    CONTRACT {
        string id PK
        string anchorId
        string supplierId
        string bankId
        float amount
        string_array approvers
        boolean approved
        string createdAt
    }
    TOKEN {
        string id PK
        string contractId FK
        string symbol
        float total
        string issuer
        string owner
        string createdAt
    }
    BALANCE {
        string tokenId FK
        string account
        float balance
    }
```

## Component Architecture

```mermaid
graph TB
    subgraph "Angular Frontend"
        subgraph "Components"
            ANC[AnchorComponent<br/>Contract creation form]
            BNK[BankComponent<br/>Token overview]
            SUP[SupplierComponent<br/>Token management]
        end
        subgraph "Services"
            CTS[ContractTokenService<br/>HTTP client]
        end
        subgraph "Routing"
            RT[AppRoutingModule<br/>/anchor, /bank, /supplier]
        end
    end

    subgraph "Spring Boot Backend"
        subgraph "Controllers"
            CTC[ContractTokenController<br/>REST endpoints]
        end
        subgraph "Services"
            CTSB[ContractTokenService<br/>Blockchain client]
        end
        subgraph "Config"
            CFG[application.yml<br/>blockchain.client.baseUrl]
        end
    end

    subgraph "Go ms-blockchain"
        subgraph "Handlers"
            HND[Contract & Token handlers<br/>MongoDB operations]
        end
        subgraph "Chaincode"
            CHC[ContractChaincode<br/>Business logic]
            TKN[TokenChaincode<br/>Token operations]
        end
    end

    ANC --> CTS
    BNK --> CTS
    SUP --> CTS
    CTS --> CTC
    CTC --> CTSB
    CTSB --> HND
    HND --> CHC
    HND --> TKN
```

## Deployment Architecture

```mermaid
graph TB
    subgraph "Docker Compose Network"
        FE[Frontend Container<br/>nginx:80 → 4200]
        BE[Backend Container<br/>java:8080]
        BC[ms-blockchain Container<br/>go:8081]
        DB[MongoDB Container<br/>mongo:27017]
    end

    FE --> BE
    BE --> BC
    BE --> DB
    BC --> DB

    subgraph "External Access"
        U1[Anchor User → localhost:4200/anchor]
        U2[Bank User → localhost:4200/bank]
        U3[Supplier User → localhost:4200/supplier]
    end

    U1 --> FE
    U2 --> FE
    U3 --> FE
```

## Business Process Summary

1. **Anchor** tạo hợp đồng → **Bank** tự động phát hành token (owner = Bank)
2. **Approvers** duyệt hợp đồng → Token chuyển từ **Bank** → **Supplier**
3. **Supplier** có thể:
   - Chuyển token cho **Supplier** khác
   - Trả lại token cho **Bank** (early payment)
4. **Bank** query toàn bộ token đã phát hành + trạng thái hiện tại

## API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/contracts` | Anchor tạo hợp đồng |
| POST | `/api/contracts/{id}/approve` | Approver duyệt hợp đồng |
| GET | `/api/contracts/{id}` | Chi tiết hợp đồng |
| GET | `/api/tokens/{id}` | Thông tin token |
| POST | `/api/tokens/transfer` | Supplier chuyển token |
| GET | `/api/tokens/issued/{bankId}` | Bank xem tokens đã phát hành |
