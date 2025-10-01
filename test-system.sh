#!/bin/bash

# Simple system test script to verify all services are running

echo "🧪 Testing Blockchain System..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to test HTTP endpoint
test_endpoint() {
    local url=$1
    local expected_status=${2:-200}
    local description=$3

    echo -n "Testing $description... "

    response=$(curl -s -o /dev/null -w "%{http_code}" "$url")

    if [ "$response" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASS${NC}"
        return 0
    else
        echo -e "${RED}❌ FAIL (status: $response, expected: $expected_status)${NC}"
        return 1
    fi
}

# Test backend API
echo "🔍 Testing Backend API (http://localhost:8080)"
test_endpoint "http://localhost:8080/api/health" 200 "Backend Health"

# Test peer services
echo ""
echo "🔍 Testing Peer Services"

test_endpoint "http://localhost:8084/health" 200 "Peer Anchor Health"
test_endpoint "http://localhost:8082/health" 200 "Peer Main Bank Health"
test_endpoint "http://localhost:8083/health" 200 "Peer Supplier Health"

# Test MongoDB connection (basic check)
echo ""
echo "🔍 Testing MongoDB Connection"
if docker-compose exec -T mongo-shared mongosh --username root --password example --authenticationDatabase admin --eval "db.runCommand('ping')" > /dev/null 2>&1; then
    echo -e "MongoDB Connection: ${GREEN}✅ PASS${NC}"
else
    echo -e "MongoDB Connection: ${RED}❌ FAIL${NC}"
fi

# Test frontend
echo ""
echo "🔍 Testing Frontend"
test_endpoint "http://localhost:4200" 200 "Frontend"

echo ""
echo "🎉 System test completed!"
echo "All services should be running successfully."
