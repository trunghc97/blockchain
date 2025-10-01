# SCF Chaincode

Smart Contract Engine cho Supply Chain Finance (SCF) system, implement theo phong cách Hyperledger Fabric nhưng sử dụng Go thuần với gRPC.

## 🏗️ Cấu trúc Project

```
scf-chaincode/
├── contracts/
│   ├── contract_management.go  # Contract smart contract
│   └── token_contract.go       # Token smart contract
├── engine/
│   └── server.go               # gRPC server implementation
├── share/
│   └── smartcontract.proto     # gRPC service definitions
├── main.go                     # Entry point
└── go.mod                      # Go module
```

## 🚀 Chạy Project

### 1. Khởi động Smart Contract Engine

```bash
cd scf-chaincode
go run main.go
```

Server sẽ chạy trên port `:9090`

```
Smart Contract Engine running on :9090
```

### 2. Test với gRPC Client

Các peer services sẽ sử dụng gRPC clients để gọi chaincode methods:
- `CreateContract` - Tạo hợp đồng
- `ApproveContract` - Phê duyệt hợp đồng
- `FinalizeContract` - Hoàn tất hợp đồng
- `IssueToken` - Phát hành token
- `TransferToken` - Chuyển nhượng token
- `SettleToken` - Tất toán token

## 📋 Smart Contracts

### Contract Management

**Struct:**
```go
type Contract struct {
    ID          string
    AnchorID    string
    Suppliers   []string
    TotalAmount float64
    Status      string // PENDING | APPROVED | EXECUTED
    FileHash    string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Functions:**
- `CreateContract(anchorID, suppliers[], totalAmount, fileHash)` - Tạo contract mới
- `ApproveContract(contractID, supplierID)` - Supplier phê duyệt
- `FinalizeContract(contractID)` - Hoàn tất contract

### Token Contract

**Struct:**
```go
type Token struct {
    ID          string
    ContractID  string
    Symbol      string
    TotalSupply float64
    Issuer      string
    Owner       string
    Balances    map[string]float64
    CreatedAt   time.Time
}
```

**Functions:**
- `IssueToken(contractID, issuer, totalSupply)` - Phát hành token
- `TransferToken(tokenID, from, to, amount)` - Chuyển token
- `SettleToken(tokenID, supplierID, bankID)` - Tất toán với bank

## 🔗 gRPC API

### Contract Operations
- `CreateContract(CreateContractRequest) returns (ContractResponse)`
- `ApproveContract(ApproveContractRequest) returns (ContractResponse)`
- `FinalizeContract(FinalizeContractRequest) returns (ContractResponse)`

### Token Operations
- `IssueToken(IssueTokenRequest) returns (TokenResponse)`
- `TransferToken(TransferTokenRequest) returns (TokenResponse)`
- `SettleToken(SettleTokenRequest) returns (TokenResponse)`

## 🧪 Test Example

```go
// Create contract
resp, err := client.CreateContract(context.Background(), &share.CreateContractRequest{
    AnchorId:    "ANCHOR001",
    Suppliers:   []string{"SUP001", "SUP002"},
    TotalAmount: 50000.0,
    FileHash:    "file_hash_123",
})

// Response
{
    "success": true,
    "message": "Contract created successfully",
    "contract_id": "contract_1234567890",
    "status": "PENDING"
}
```

## 🔧 Development

### Generate Protobuf (nếu cần)

```bash
protoc --go_out=. --go-grpc_out=. share/smartcontract.proto
```

### Run Tests

```bash
go test ./...
```

## 📚 Dependencies

- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol buffers

## 🎯 Use Case Flow

1. **Anchor tạo contract** → `CreateContract()`
2. **Suppliers duyệt** → `ApproveContract()` (multiple calls)
3. **Contract hoàn tất** → `FinalizeContract()`
4. **Phát hành token** → `IssueToken()`
5. **Chuyển token** → `TransferToken()` (multiple transfers)
6. **Tất toán** → `SettleToken()` (remove balances)

## 🔐 Security Notes

- Hiện tại chưa có authentication (development mode)
- Sử dụng in-memory storage (production nên dùng database)
- Không có validation business logic phức tạp

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Add tests
4. Submit PR

---

*SCF Chaincode - Smart Contracts for Supply Chain Finance*
