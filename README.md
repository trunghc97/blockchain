# Blockchain Transfer System PoC

Hệ thống chuyển tiền có nhiều cấp duyệt sử dụng blockchain, được xây dựng với các công nghệ:
- Frontend: Angular + Angular Material
- Backend: Spring Boot + Spring Security + JWT
- Blockchain Service: Golang + MongoDB

## Thông tin đăng nhập

Hệ thống tự động tạo sẵn các tài khoản sau khi khởi động lần đầu:

1. Tài khoản Anchor:
   - Username: anchor
   - Password: 123456
   - Role: ANCHOR

2. Tài khoản Supplier (10 tài khoản):
   - Username: supplier1, supplier2, ..., supplier10
   - Password: 123456
   - Role: SUPPLIER

## Yêu cầu hệ thống

1. **Docker & Docker Compose**
   - Docker version 20.10.0 trở lên
   - Docker Compose version 2.0.0 trở lên

2. **Môi trường phát triển (nếu cần)**
   - Node.js 18.x
   - Java JDK 17
   - Go 1.21
   - MongoDB 6.0

## Cài đặt và chạy

### Sử dụng Docker (Khuyến nghị)

1. Clone repository:
   ```bash
   git clone <repository-url>
   cd blockchain
   ```

2. Build và chạy các services:
   ```bash
   docker-compose up --build
   ```

3. Truy cập các endpoints:
   - Frontend: http://localhost:4200
   - Backend: http://localhost:8080
   - MS Blockchain: http://localhost:8081
   - Supplier Mock: http://localhost:8082

### Phát triển local

1. **Frontend (Angular)**
   ```bash
   cd frontend
   npm install
   npm start
   ```

2. **Backend (Spring Boot)**
   ```bash
   cd backend
   ./mvnw spring-boot:run
   ```

3. **MS Blockchain (Go)**
   ```bash
   cd ms-blockchain
   go mod download
   go run main.go
   ```

4. **Supplier Mock (Go)**
   ```bash
   cd supplier-mock
   go mod download
   go run main.go
   ```

5. **MongoDB**
   ```bash
   mongod --dbpath=./data/mongodb
   ```

## Kiến trúc hệ thống

```
Frontend (Angular) -> Backend (Spring Boot) -> MS Blockchain (Go) -> Supplier Mock (Go)
                                                    |
                                                MongoDB
```

## Các thành phần

1. **Frontend (port 4200)**
   - Giao diện đăng nhập và xác thực
   - Giao diện tạo giao dịch chuyển tiền
   - Giao diện approve giao dịch
   - Hiển thị trạng thái giao dịch
   - Xem lịch sử blockchain

2. **Backend (port 8080)**
   - Xác thực và phân quyền với JWT
   - REST API cho Frontend
   - Chuyển tiếp request tới MS Blockchain
   - Tự động tạo user mẫu khi khởi động

3. **MS Blockchain (port 8081)**
   - Xử lý các giao dịch blockchain
   - Lưu trữ transaction và block trong MongoDB
   - Gọi Supplier API khi đủ số lượng approve

4. **Supplier Mock (port 8082)**
   - Mock API của hệ thống Supplier
   - Luôn trả về kết quả thành công

5. **MongoDB (port 27017)**
   - Lưu trữ users và thông tin xác thực
   - Lưu trữ transactions, blocks và world state

## Cache Docker

Hệ thống đã được cấu hình cache Docker cơ bản:

- **Frontend**: Cache npm dependencies và node_modules
- **Backend**: Cache Maven dependencies
- **MS Blockchain & Supplier Mock**: Cache Go modules

**Lợi ích**:
- Giảm thời gian build cho lần rebuild
- Tận dụng lại Docker layers đã build trước đó

## Luồng hoạt động

1. **Đăng nhập**
   - User nhập username/password
   - Backend xác thực và trả về JWT token
   - Frontend lưu token và thêm vào header

2. **Tạo giao dịch**
   - User nhập thông tin chuyển tiền trên Frontend
   - Frontend gọi Backend API `/transfer/create`
   - Backend chuyển tiếp tới MS Blockchain `/tx/create`
   - MS Blockchain tạo transaction và world state

3. **Approve giao dịch**
   - Approver nhập Transaction ID và Approver ID
   - Frontend gọi Backend API `/transfer/approve`
   - Backend chuyển tiếp tới MS Blockchain `/tx/approve`
   - MS Blockchain kiểm tra số lượng approve
   - Nếu đủ số approve, gọi Supplier API
   - Nếu Supplier trả về thành công, tạo transaction EXECUTE

4. **Kiểm tra trạng thái**
   - Frontend gọi Backend API `/transfer/status/{id}`
   - Backend chuyển tiếp tới MS Blockchain `/tx/status/{id}`
   - MS Blockchain trả về world state của giao dịch

## Cấu trúc dữ liệu

1. **User**
   - ID: string
   - Username: string
   - Password: string (đã mã hóa)
   - Role: string (ANCHOR/SUPPLIER)

2. **Transaction**
   - ID: ObjectID
   - TransactionID: string
   - FromAccount: string
   - ToAccount: string
   - Amount: float64
   - Status: string (PENDING/APPROVED/EXECUTED)
   - Type: string (CREATE/APPROVE/EXECUTE)
   - ApproverID: string
   - Timestamp: datetime

3. **Block**
   - ID: ObjectID
   - BlockNumber: int64
   - Timestamp: datetime
   - PreviousHash: string
   - Hash: string
   - Transactions: []Transaction

4. **World State**
   - ID: ObjectID
   - TransactionID: string
   - FromAccount: string
   - ToAccount: string
   - Amount: float64
   - Status: string
   - ApprovalCount: int
   - LastUpdated: datetime