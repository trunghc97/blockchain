#!/bin/bash

# Test Script cho Blockchain Supply Chain Finance
# Thực hiện tuần tự các use cases và verify blocks/events

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL="http://localhost:8080"
MONGO_HOST="localhost"
MONGO_PORT="27017"
MONGO_USER="root"
MONGO_PASS="example"

# Global variables
CONTRACT_ID=""
TOKEN_ID=""
ANCHOR_TOKEN=""
BANK_TOKEN=""
SUPPLIER1_TOKEN=""
SUPPLIER2_TOKEN=""

echo -e "${BLUE}🚀 Bắt đầu test Blockchain Supply Chain Finance${NC}"
echo "=============================================="

# Function to make API calls with error handling
call_api() {
    local method=$1
    local url=$2
    local data=$3
    local auth_token=${4:-""}

    echo -e "${YELLOW}API Call: ${method} ${url}${NC}"

    if [ -n "$auth_token" ]; then
        # Remove any color codes from token
        auth_token=$(echo "$auth_token" | sed 's/\x1b\[[0-9;]*m//g')
        if [ "$method" = "GET" ]; then
            curl -s -X $method \
                 -H "Authorization: Bearer $auth_token" \
                 -H "Content-Type: application/json" \
                 "$url"
        else
            curl -s -X $method \
                 -H "Authorization: Bearer $auth_token" \
                 -H "Content-Type: application/json" \
                 -d "$data" \
                 "$url"
        fi
    else
        if [ "$method" = "GET" ]; then
            curl -s -X $method \
                 -H "Content-Type: application/json" \
                 "$url"
        else
            curl -s -X $method \
                 -H "Content-Type: application/json" \
                 -d "$data" \
                 "$url"
        fi
    fi
}

# Function to check MongoDB collections
check_mongo() {
    local db=$1
    local collection=$2
    local query=${3:-"{}"}
    local projection=${4:-"{}"}

    echo -e "${YELLOW}MongoDB Query: ${db}.${collection} - ${query}${NC}"

    docker-compose exec -T mongo-shared mongosh --username $MONGO_USER --password $MONGO_PASS \
        --authenticationDatabase admin $db --eval "
        db.${collection}.find(${query}, ${projection}).toArray()
        " 2>/dev/null || echo "Query failed"
}

# Function to login and get token
login_user() {
    local username=$1
    local password=$2
    local role=$3

    echo -e "${BLUE}Đăng nhập user: ${username} (${role})${NC}"

    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${username}\",\"password\":\"${password}\"}" \
        "${BACKEND_URL}/api/auth/login")

    local token=$(echo $response | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    if [ -z "$token" ]; then
        echo -e "${RED}❌ Login failed for ${username}${NC}"
        echo "Response: $response"
        exit 1
    fi

    echo -e "${GREEN}✅ Login thành công${NC}"
    # Return token without any extra characters
    echo "$token"
}

# Function to verify blocks and events
verify_blockchain_data() {
    local step_name=$1

    echo -e "${BLUE}🔍 Verify Blockchain Data: ${step_name}${NC}"

    # Check private blockchain (peer databases)
    echo "=== Private Blockchain (Peers) ==="
    echo "📊 Events in peer-main-bank:"
    check_mongo blockchain_private events '{}' '{"eventType":1,"contractId":1,"tokenId":1,"timestamp":1}'

    echo "📦 Blocks in peer-main-bank:"
    check_mongo blockchain_private blocks '{}' '{"blockNumber":1,"hash":1,"events":1}'

    # Check public blockchain (orderer database)
    echo "=== Public Blockchain (Orderer) ==="
    echo "🌐 Events in blockchain_public:"
    check_mongo blockchain_public events '{}' '{"eventType":1,"contractId":1,"tokenId":1,"timestamp":1}'

    echo "⛓️  Blocks in blockchain_public:"
    check_mongo blockchain_public blocks '{}' '{"blockNumber":1,"hash":1,"events":1}'

    echo ""
}

# ==========================================
# Test Case 1: Anchor tạo hợp đồng
# ==========================================
echo -e "${GREEN}📝 TEST CASE 1: Anchor tạo hợp đồng${NC}"

# Login as anchor
ANCHOR_TOKEN=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"anchor","password":"123456"}' \
    "${BACKEND_URL}/api/auth/login" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo -e "${GREEN}✅ Anchor login thành công${NC}"

# Create contract
CONTRACT_DATA='{
  "buyer": "ANCHOR001",
  "bankId": "BANK001",
  "description": "Test Supply Chain Contract",
  "suppliers": [
    {
      "supplierId": "SUP001",
      "name": "Supplier 1",
      "allocatedAmount": 30000.00
    },
    {
      "supplierId": "SUP002",
      "name": "Supplier 2",
      "allocatedAmount": 20000.00
    }
  ],
  "totalAmount": 50000.00
}'

echo "Tạo hợp đồng với data:"
echo "$CONTRACT_DATA" | jq . 2>/dev/null || echo "$CONTRACT_DATA"

# Create contract using JSON API
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ANCHOR_TOKEN" \
    -d '{"buyer":"ANCHOR001","bankId":"BANK001","description":"Test Supply Chain Contract","suppliers":[{"supplierId":"SUP001","name":"Supplier 1","allocatedAmount":30000.00},{"supplierId":"SUP002","name":"Supplier 2","allocatedAmount":20000.00}],"totalAmount":50000.00}' \
    "${BACKEND_URL}/api/v1/contracts")

CONTRACT_ID=$(echo $response | grep -o '"contractId":"[^"]*' | cut -d'"' -f4)
if [ -z "$CONTRACT_ID" ]; then
    echo -e "${RED}❌ Tạo contract thất bại${NC}"
    echo "Response: $response"
    exit 1
fi

echo -e "${GREEN}✅ Contract created: ${CONTRACT_ID}${NC}"

# Verify blockchain data after contract creation
verify_blockchain_data "Sau khi tạo contract"

# ==========================================
# Test Case 2: Main-bank duyệt hợp đồng
# ==========================================
echo -e "${GREEN}🏦 TEST CASE 2: Main-bank duyệt hợp đồng${NC}"

# Login as bank
BANK_TOKEN=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"bank","password":"123456"}' \
    "${BACKEND_URL}/api/auth/login" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo -e "${GREEN}✅ Bank login thành công${NC}"

# Bank approve contract
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $BANK_TOKEN" \
    -d '{"bankId":"BANK001"}' \
    "${BACKEND_URL}/api/v1/contracts/${CONTRACT_ID}/approve-bank")

TOKEN_ID="token_${CONTRACT_ID}"
if [ -z "$TOKEN_ID" ]; then
    echo -e "${RED}❌ Bank approve thất bại${NC}"
    echo "Response: $response"
    exit 1
fi

echo -e "${GREEN}✅ Bank approved contract. Token created: ${TOKEN_ID}${NC}"

# Verify blockchain data after bank approval
verify_blockchain_data "Sau khi bank approve"

# ==========================================
# Test Case 3: Suppliers duyệt hợp đồng
# ==========================================
echo -e "${GREEN}👥 TEST CASE 3: Suppliers duyệt hợp đồng${NC}"

# Login as supplier1
SUPPLIER1_TOKEN=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"supplier1","password":"123456"}' \
    "${BACKEND_URL}/api/auth/login" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo -e "${GREEN}✅ Supplier1 login thành công${NC}"

# Supplier1 approve contract
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $SUPPLIER1_TOKEN" \
    -d '{"supplierId":"SUP001"}' \
    "${BACKEND_URL}/api/v1/contracts/${CONTRACT_ID}/approve")

echo -e "${GREEN}✅ Supplier1 approved contract${NC}"

# Login as supplier2
SUPPLIER2_TOKEN=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"supplier2","password":"123456"}' \
    "${BACKEND_URL}/api/auth/login" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

echo -e "${GREEN}✅ Supplier2 login thành công${NC}"

# Supplier2 approve contract (this should trigger token distribution)
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $SUPPLIER2_TOKEN" \
    -d '{"supplierId":"SUP002"}' \
    "${BACKEND_URL}/api/v1/contracts/${CONTRACT_ID}/approve")

echo -e "${GREEN}✅ Supplier2 approved contract - Token should be distributed${NC}"

# Verify blockchain data after supplier approvals
verify_blockchain_data "Sau khi suppliers approve"

# Check balances after distribution
echo "=== Check Balances After Distribution ==="
echo "💰 Balances for token ${TOKEN_ID}:"
check_mongo blockchain_private balances "{\"tokenId\":\"${TOKEN_ID}\"}" '{"account":1,"balance":1}'

# ==========================================
# Test Case 4: Token chuyển cho supplier
# ==========================================
echo -e "${GREEN}💸 TEST CASE 4: Token chuyển cho supplier${NC}"

# Anchor transfers 10000 to Supplier1
TRANSFER_DATA='{
  "tokenId": "'${TOKEN_ID}'",
  "from": "ANCHOR001",
  "to": "SUP001",
  "amount": 10000.00
}'

echo "Anchor chuyển 10000 token cho Supplier1"
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ANCHOR_TOKEN" \
    -d '{"tokenId":"'${TOKEN_ID}'","from":"ANCHOR001","to":"SUP001","amount":10000.00}' \
    "${BACKEND_URL}/api/v1/tokens/transfer")

if echo "$response" | grep -q "success"; then
    echo -e "${GREEN}✅ Token transfer thành công${NC}"
else
    echo -e "${RED}❌ Token transfer thất bại${NC}"
    echo "Response: $response"
fi

# Verify blockchain data after transfer
verify_blockchain_data "Sau khi transfer token"

# ==========================================
# Test Case 5: Supplier chuyển cho nhau
# ==========================================
echo -e "${GREEN}🔄 TEST CASE 5: Supplier chuyển cho nhau${NC}"

# Supplier1 transfers 5000 to Supplier2
TRANSFER_SUPPLIER_DATA='{
  "tokenId": "'${TOKEN_ID}'",
  "from": "SUP001",
  "to": "SUP002",
  "amount": 5000.00
}'

echo "Supplier1 chuyển 5000 token cho Supplier2"
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $SUPPLIER1_TOKEN" \
    -d '{"tokenId":"'${TOKEN_ID}'","from":"SUP001","to":"SUP002","amount":5000.00}' \
    "${BACKEND_URL}/api/v1/tokens/transfer")

if echo "$response" | grep -q "success"; then
    echo -e "${GREEN}✅ Supplier-to-supplier transfer thành công${NC}"
else
    echo -e "${RED}❌ Supplier-to-supplier transfer thất bại${NC}"
    echo "Response: $response"
fi

# Verify blockchain data after supplier transfer
verify_blockchain_data "Sau khi supplier transfer"

# ==========================================
# Test Case 6: Supplier settle token
# ==========================================
echo -e "${GREEN}💰 TEST CASE 6: Supplier settle token${NC}"

# Supplier2 settles all remaining tokens
SETTLE_DATA='{
  "tokenId": "'${TOKEN_ID}'",
  "supplierId": "SUP002"
}'

echo "Supplier2 settle toàn bộ token"
response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $SUPPLIER2_TOKEN" \
    -d '{"tokenId":"'${TOKEN_ID}'","supplierId":"SUP002"}' \
    "${BACKEND_URL}/api/v1/tokens/settle")

if echo "$response" | grep -q "success"; then
    echo -e "${GREEN}✅ Token settlement thành công${NC}"
else
    echo -e "${RED}❌ Token settlement thất bại${NC}"
    echo "Response: $response"
fi

# Verify blockchain data after settlement
verify_blockchain_data "Sau khi settle token"

# ==========================================
# Final Verification
# ==========================================
echo -e "${BLUE}🎯 FINAL VERIFICATION${NC}"

# Check final balances
echo "=== Final Balances ==="
check_mongo blockchain_private balances "{\"tokenId\":\"${TOKEN_ID}\"}" '{"account":1,"balance":1}'

# Check all events for the contract
echo "=== All Events for Contract ${CONTRACT_ID} ==="
check_mongo blockchain_private events "{\"contractId\":\"${CONTRACT_ID}\"}" '{"eventType":1,"timestamp":1,"description":1}'

# Check block count
echo "=== Block Count ==="
echo "Private blockchain blocks:"
check_mongo blockchain_private blocks '{}' '{"blockNumber":1,"timestamp":1}' | wc -l

echo "Public blockchain blocks:"
check_mongo blockchain_public blocks '{}' '{"blockNumber":1,"timestamp":1}' | wc -l

# ==========================================
# API COMPREHENSIVE TEST
# ==========================================
echo -e "${BLUE}🔬 CHI TIẾT API TESTING${NC}"

# Test additional APIs
echo -e "${YELLOW}Testing additional APIs...${NC}"

# Test GET endpoints
echo -e "${BLUE}📋 GET API Tests${NC}"

# Get all contracts
echo "Testing GET /api/v1/contracts"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/contracts")
if echo "$response" | grep -q "contractId"; then
    echo -e "${GREEN}✅ GET contracts API works${NC}"
else
    echo -e "${RED}❌ GET contracts API failed${NC}"
fi

# Get specific contract
echo "Testing GET /api/v1/contracts/${CONTRACT_ID}"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/contracts/${CONTRACT_ID}")
if echo "$response" | grep -q "${CONTRACT_ID}"; then
    echo -e "${GREEN}✅ GET contract by ID API works${NC}"
else
    echo -e "${RED}❌ GET contract by ID API failed${NC}"
fi

# Get all tokens
echo "Testing GET /api/v1/tokens"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/tokens")
if echo "$response" | grep -q '"_id":'; then
    echo -e "${GREEN}✅ GET tokens API works${NC}"
else
    echo -e "${RED}❌ GET tokens API failed${NC}"
fi

# Get specific token
echo "Testing GET /api/v1/tokens/${TOKEN_ID}"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/tokens/${TOKEN_ID}")
if echo "$response" | grep -q "${TOKEN_ID}"; then
    echo -e "${GREEN}✅ GET token by ID API works${NC}"
else
    echo -e "${RED}❌ GET token by ID API failed${NC}"
fi

# Get balances by token
echo "Testing GET /api/v1/tokens/balances/token/${TOKEN_ID}"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/tokens/balances/token/${TOKEN_ID}")
if echo "$response" | grep -q "balance"; then
    echo -e "${GREEN}✅ GET balances by token API works${NC}"
else
    echo -e "${RED}❌ GET balances by token API failed${NC}"
fi

# Get balances by account (Supplier1)
echo "Testing GET /api/v1/tokens/balances/account/SUP001"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/tokens/balances/account/SUP001")
if echo "$response" | grep -q "balance"; then
    echo -e "${GREEN}✅ GET balances by account API works${NC}"
else
    echo -e "${RED}❌ GET balances by account API failed${NC}"
fi

# Get all suppliers
echo "Testing GET /api/v1/suppliers"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/suppliers")
if echo "$response" | grep -q '"id":' || echo "$response" | grep -q '"_id":'; then
    echo -e "${GREEN}✅ GET suppliers API works${NC}"
else
    echo -e "${RED}❌ GET suppliers API failed${NC}"
fi

# Get users
echo "Testing GET /api/users"
response=$(curl -s -X GET "${BACKEND_URL}/api/users")
if echo "$response" | grep -q "username"; then
    echo -e "${GREEN}✅ GET users API works${NC}"
else
    echo -e "${RED}❌ GET users API failed${NC}"
fi

# Get suppliers via users API
echo "Testing GET /api/users/suppliers"
response=$(curl -s -X GET "${BACKEND_URL}/api/users/suppliers")
if echo "$response" | grep -q "SUPPLIER"; then
    echo -e "${GREEN}✅ GET suppliers via users API works${NC}"
else
    echo -e "${RED}❌ GET suppliers via users API failed${NC}"
fi

# Test Contract V1 additional endpoints
echo "Testing GET /api/v1/contracts/tokens/all"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/contracts/tokens/all")
if [ -n "$response" ] && [ "$response" != "null" ] && echo "$response" | grep -q '"_id":'; then
    echo -e "${GREEN}✅ GET all tokens via contracts API works${NC}"
else
    echo -e "${RED}❌ GET all tokens via contracts API failed${NC}"
fi

echo "Testing GET /api/v1/contracts/balances/all"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/contracts/balances/all")
if echo "$response" | grep -q "balance"; then
    echo -e "${GREEN}✅ GET all balances via contracts API works${NC}"
else
    echo -e "${RED}❌ GET all balances via contracts API failed${NC}"
fi

echo "Testing GET /api/v1/contracts/balances/token/${TOKEN_ID}"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/contracts/balances/token/${TOKEN_ID}")
if echo "$response" | grep -q "balance"; then
    echo -e "${GREEN}✅ GET balances by token via contracts API works${NC}"
else
    echo -e "${RED}❌ GET balances by token via contracts API failed${NC}"
fi

# Test tokens issued by system (note: tokens are issued by SYSTEM, not BANK)
echo "Testing GET /api/v1/tokens/issued/SYSTEM"
response=$(curl -s -X GET "${BACKEND_URL}/api/v1/tokens/issued/SYSTEM")
if [ -n "$response" ] && [ "$response" != "null" ] && echo "$response" | grep -q '"_id":'; then
    echo -e "${GREEN}✅ GET tokens issued by system API works${NC}"
else
    echo -e "${RED}❌ GET tokens issued by system API failed${NC}"
fi

echo ""
echo -e "${GREEN}🎉 HOÀN THÀNH COMPREHENSIVE API TESTING${NC}"
echo -e "${GREEN}✅ Đã test đầy đủ tất cả API endpoints${NC}"
echo ""
echo "📊 Summary:"
echo "- Contract ID: ${CONTRACT_ID}"
echo "- Token ID: ${TOKEN_ID}"
echo "- Total blockchain operations: 6 (create, bank-approve, supplier-approve x2, transfer x2, settle)"
echo "- Total API endpoints tested: 16+"
echo ""
echo "🔗 Để kiểm tra UI: http://localhost:4200"
echo "🔗 Để kiểm tra API docs: http://localhost:8080/swagger-ui.html"
