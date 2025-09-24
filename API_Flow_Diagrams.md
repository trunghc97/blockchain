# Blockchain API Flow Diagrams

This document contains Mermaid flow diagrams for all API endpoints in the Golang blockchain microservice.

## Legend
- 🟢 **Start/End**: Process start or end
- 🔵 **Process**: Main processing step
- 🟡 **Decision**: Conditional check
- 🟠 **Database**: Database operation
- 🟣 **Response**: API response
- 🔴 **Error**: Error handling

---

## Contract Management APIs

### POST /contract/create - Create New Contract
```mermaid
flowchart TD
    A[🟢 Receive Contract Data] --> B[🔵 Validate Input Data]
    B --> C[🟡 Check Contract ID]
    C -->|No ID provided| D[🔵 Generate Random Contract ID]
    C -->|ID provided| E[🔵 Use Provided ID]
    D --> F[🟠 Save Contract to MongoDB]
    E --> F
    F --> G[🔵 Log CONTRACT_CREATED Event]
    G --> H[🟠 Create Block Entry]
    H --> I[🟣 Return Success Response]
```

### POST /contract/{id}/approve-bank - Bank Approve Contract
```mermaid
flowchart TD
    A[🟢 Receive Bank Approval Request] --> B[🔵 Validate Bank ID]
    B --> C[🟠 Get Contract from Database]
    C --> D[🟡 Contract Exists?]
    D -->|No| E[🔴 Return 404 Not Found]
    D -->|Yes| F[🟡 Bank Has Permission?]
    F -->|No| G[🔴 Return 403 Forbidden]
    F -->|Yes| H[🟠 Update bankApproved = true]
    H --> I[🔵 Calculate Total Amount from Suppliers]
    I --> J[🟠 Create Token with Issuer = SYSTEM]
    J --> K[🟠 Create Initial Balance for Anchor]
    K --> L[🔵 Log CONTRACT_BANK_APPROVED_TOKEN_GENERATED Event]
    L --> M[🟠 Create Block Entry]
    M --> N[🟣 Return Success Response with Token ID]
```

### POST /contract/{id}/approve - Supplier Approve Contract
```mermaid
flowchart TD
    A[🟢 Receive Supplier Approval Request] --> B[🔵 Validate Supplier ID]
    B --> C[🟠 Get Contract from Database]
    C --> D[🟡 Contract Exists?]
    D -->|No| E[🔴 Return 404 Not Found]
    D -->|Yes| F[🟡 Bank Approved Contract?]
    F -->|No| G[🔴 Return 403 Forbidden]
    F -->|Yes| H[🟡 Supplier Authorized for Contract?]
    H -->|No| I[🔴 Return 403 Forbidden]
    H -->|Yes| J[🟠 Update Supplier Status to APPROVED]
    J --> K[🟠 Refresh Contract Data]
    K --> L[🔵 Check All Suppliers Approved?]
    L -->|No| M[🔵 Log SUPPLIER_APPROVED Event]
    L -->|Yes| N[🟠 Update Contract Status to approved]
    N --> O[🟠 Distribute Token Balances to All Suppliers]
    O --> P[🟠 Remove Anchor's Balance]
    P --> Q[🔵 Log CONTRACT_FULLY_APPROVED Event]
    M --> R[🟠 Create Block Entry]
    Q --> R
    R --> S[🟣 Return Success Response]
```

---

## Token Management APIs

### POST /token/transfer - Transfer Token Between Accounts
```mermaid
flowchart TD
    A[🟢 Receive Transfer Request] --> B[🔵 Validate Transfer Data]
    B --> C[🟠 Get Token Info from Database]
    C --> D[🟡 Token Exists?]
    D -->|No| E[🔴 Return 404 Not Found]
    D -->|Yes| F[🟠 Get Sender Balance]
    F --> G[🟡 Sender Has Balance?]
    G -->|No| H[🔴 Return 400 Insufficient Balance]
    G -->|Yes| I[🟡 Balance Sufficient?]
    I -->|No| J[🔴 Return 400 Insufficient Balance]
    I -->|Yes| K[🟠 Update Sender Balance - amount]
    K --> L[🟠 Get Receiver Balance]
    L --> M[🟡 Receiver Balance Exists?]
    M -->|No| N[🟠 Create New Balance for Receiver]
    M -->|Yes| O[🟠 Update Receiver Balance + amount]
    N --> P[🔵 Log TOKEN_TRANSFERRED Event]
    O --> P
    P --> Q[🟠 Create Block Entry]
    Q --> R[🟡 Anchor Has Zero Balance?]
    R -->|Yes| S[🟠 Auto-approve Contract]
    R -->|No| T[🟣 Return Success Response]
    S --> T
```

### POST /token/settle - Settle Token with Bank
```mermaid
flowchart TD
    A[🟢 Receive Settlement Request] --> B[🔵 Validate TokenId & SupplierId]
    B --> C[🟠 Get Supplier Balance for Token]
    C --> D[🟡 Balance Exists?]
    D -->|No| E[🔴 Return 400 No Balance]
    D -->|Yes| F[🟠 Get Token Info from Database]
    F --> G[🟡 Token Exists?]
    G -->|No| H[🔴 Return 404 Not Found]
    G -->|Yes| I[🟠 Get Contract Info from Token]
    I --> J[🟡 Contract Exists?]
    J -->|No| K[🔴 Return 404 Not Found]
    J -->|Yes| L[🟠 Delete Supplier's Balance]
    L --> M[🔵 Log TOKEN_SETTLED Event]
    M --> N[🟠 Create Block Entry]
    N --> O[🟣 Return Success Response]
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

## Query APIs

### GET /contract/list - Get All Contracts
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query All Contracts from Database]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Contracts Array]
```

### GET /tokens - Get All Tokens
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query All Tokens from Database]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Tokens Array]
```

### GET /balances/account/{accountId} - Get Account Balances
```mermaid
flowchart TD
    A[🟢 Receive Account ID] --> B[🟠 Query Balances by Account ID]
    B --> C[🟡 Query Successful?]
    C -->|No| D[🔴 Return 500 Internal Error]
    C -->|Yes| E[🟣 Return Balances Array]
```

### GET /suppliers - Get All Suppliers
```mermaid
flowchart TD
    A[🟢 Receive Request] --> B[🟠 Query Users with Role SUPPLIER]
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

## Common Block Creation Flow
```mermaid
flowchart TD
    A[🔵 Event Logged] --> B[🟠 Get Next Block Number]
    B --> C[🟠 Get Current Timestamp]
    C --> D[🟠 Get Previous Block Hash]
    D --> E[🔵 Calculate Block Hash]
    E --> F[🟠 Save Block to Database]
    F --> G[🟣 Continue with API Response]
```

---

## Error Handling Patterns
```mermaid
flowchart TD
    A[🟡 Operation Failed?] -->|Yes| B[🔴 Log Error Message]
    B --> C[🔴 Return HTTP Error Response]
    C --> D[🟢 End]
    A -->|No| E[🟣 Return Success Response]
    E --> D
```

## Database Collections Used
- **contracts**: Contract information and approval status
- **tokens**: Token metadata (issuer, owner, total amount)
- **balances**: Token balances per account
- **events**: Audit trail of all blockchain operations
- **blocks**: Block data with hashes and event references
- **users**: User information and roles

## Key Business Logic Flows

### Token Lifecycle
1. **Creation**: Bank approves contract → System creates token → Anchor receives initial balance
2. **Distribution**: All suppliers approve → Token distributed to suppliers → Anchor balance removed
3. **Transfer**: Suppliers transfer tokens between themselves
4. **Settlement**: Supplier settles with bank → Balance removed → Contract closed

### Block Creation
Every significant operation creates:
1. An event record with details
2. A block containing the event
3. Block hash calculated from previous block + current data
4. Immutable audit trail maintained
