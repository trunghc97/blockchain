# Blockchain Transfer System PoC

Hệ thống chuyển tiền có nhiều cấp duyệt sử dụng blockchain, được xây dựng với các công nghệ:
- Frontend: Angular
- Backend: Spring Boot
- Blockchain Service: Golang + MongoDB
- Supplier Mock: Golang

## Kiến trúc hệ thống

```
Frontend (Angular) -> Backend (Spring Boot) -> MS Blockchain (Go) -> Supplier Mock (Go)
                                                    |
                                                MongoDB
```

## Các thành phần

1. **Frontend (port 4200)**
   - Giao diện tạo giao dịch chuyển tiền
   - Giao diện approve giao dịch
   - Hiển thị trạng thái giao dịch

2. **Backend (port 8080)**
   - REST API cho Frontend
   - Chuyển tiếp request tới MS Blockchain

3. **MS Blockchain (port 8081)**
   - Xử lý các giao dịch blockchain
   - Lưu trữ transaction và block trong MongoDB
   - Gọi Supplier API khi đủ số lượng approve

4. **Supplier Mock (port 8082)**
   - Mock API của hệ thống Supplier
   - Luôn trả về kết quả thành công

5. **MongoDB (port 27017)**
   - Lưu trữ transactions, blocks và world state

## Cách chạy

1. Cài đặt Docker và Docker Compose

2. Build và chạy các services:
   ```bash
   docker-compose up --build
   ```

### Cache Docker

Hệ thống đã được cấu hình cache Docker cơ bản:

- **Frontend**: Cache npm dependencies
- **Backend**: Cache Maven dependencies
- **MS Blockchain & Supplier Mock**: Cache Go modules

**Lợi ích**:
- Giảm thời gian build cho lần rebuild
- Tận dụng lại Docker layers đã build trước đó

4. Truy cập các endpoints:
   - Frontend: http://localhost:4200
   - Backend: http://localhost:8080
   - MS Blockchain: http://localhost:8081
   - Supplier Mock: http://localhost:8082

## Luồng hoạt động

1. **Tạo giao dịch**
   - User nhập thông tin chuyển tiền trên Frontend
   - Frontend gọi Backend API `/transfer/create`
   - Backend chuyển tiếp tới MS Blockchain `/tx/create`
   - MS Blockchain tạo transaction và world state

2. **Approve giao dịch**
   - Approver nhập Transaction ID và Approver ID
   - Frontend gọi Backend API `/transfer/approve`
   - Backend chuyển tiếp tới MS Blockchain `/tx/approve`
   - MS Blockchain kiểm tra số lượng approve
   - Nếu đủ số approve, gọi Supplier API
   - Nếu Supplier trả về thành công, tạo transaction EXECUTE

3. **Kiểm tra trạng thái**
   - Frontend gọi Backend API `/transfer/status/{id}`
   - Backend chuyển tiếp tới MS Blockchain `/tx/status/{id}`
   - MS Blockchain trả về world state của giao dịch

## Cấu trúc dữ liệu

1. **Transaction**
   - ID: ObjectID
   - TransactionID: string
   - FromAccount: string
   - ToAccount: string
   - Amount: float64
   - Status: string (PENDING/APPROVED/EXECUTED)
   - Type: string (CREATE/APPROVE/EXECUTE)
   - ApproverID: string
   - Timestamp: datetime

2. **Block**
   - ID: ObjectID
   - BlockNumber: int64
   - Timestamp: datetime
   - PreviousHash: string
   - Hash: string
   - Transactions: []Transaction

3. **World State**
   - ID: ObjectID
   - TransactionID: string
   - FromAccount: string
   - ToAccount: string
   - Amount: float64
   - Status: string
   - ApprovalCount: int
   - LastUpdated: datetime
