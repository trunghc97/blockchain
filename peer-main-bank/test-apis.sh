#!/bin/bash

# Test individual APIs to find which ones are failing

echo "🧪 Testing Individual APIs..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

test_api() {
    local url=$1
    local description=$2

    echo -n "Testing: $description"
    echo -n " ($url) ... "

    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$url")
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        body=$(echo $response | sed -e 's/HTTPSTATUS.*//g')
        echo -e "${RED}❌ FAIL (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
}

echo "Backend APIs:"
echo "=============="

# 1. GET contract by ID API
test_api "http://localhost:8080/api/v1/contract/123" "GET contract by ID"

# 2. GET tokens API
test_api "http://localhost:8080/api/v1/tokens" "GET tokens API"

# 3. GET token by ID API
test_api "http://localhost:8080/api/v1/tokens/123" "GET token by ID API"

# 4. GET suppliers API
test_api "http://localhost:8080/api/v1/suppliers" "GET suppliers API"

# 5. GET all tokens via contracts API
test_api "http://localhost:8080/api/v1/contract/tokens/all" "GET all tokens via contracts"

# 6. GET balances by token via contracts API
test_api "http://localhost:8080/api/v1/contract/balances/token/123" "GET balances by token via contracts"

# 7. GET tokens issued by system API
test_api "http://localhost:8080/api/v1/tokens/issued/SYSTEM" "GET tokens issued by system"

echo ""
echo "Peer APIs:"
echo "=========="

# Test peer endpoints directly
test_api "http://localhost:8084/contract/123" "Peer Anchor - GET contract"
test_api "http://localhost:8084/tokens" "Peer Anchor - GET tokens"
test_api "http://localhost:8083/token/123" "Peer Supplier - GET token"
test_api "http://localhost:8083/suppliers" "Peer Supplier - GET suppliers"
test_api "http://localhost:8082/token/issued/SYSTEM" "Peer Main Bank - GET tokens issued by system"

echo ""
echo "🎉 API testing completed!"
