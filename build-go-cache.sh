#!/bin/bash

# Script to build Go module cache for faster Docker builds
# This downloads all Go dependencies locally to speed up container builds

set -e

echo "🚀 Building Go module cache for all services..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create cache directories if they don't exist
echo "📁 Creating cache directories..."
mkdir -p .go-cache/pkg
mkdir -p .go-cache/build

# Function to cache dependencies for a service
cache_service() {
    local service_name=$1
    local service_dir=$2

    echo "📦 Caching dependencies for $service_name..."

    if [ ! -f "$service_dir/go.mod" ]; then
        echo -e "${RED}❌ go.mod not found in $service_dir${NC}"
        return 1
    fi

    cd "$service_dir"

    # Download dependencies
    echo "⬇️  Downloading Go modules for $service_name..."
    GOPROXY=https://proxy.golang.org,direct go mod download

    # Optional: Pre-compile some packages to speed up builds
    echo "🔨 Pre-building common packages for $service_name..."
    GOPROXY=https://proxy.golang.org,direct go build -v ./... 2>/dev/null || true

    cd - > /dev/null

    echo -e "${GREEN}✅ $service_name dependencies cached${NC}"
}

# Cache dependencies for each service
cache_service "peer-anchor" "peer-anchor"
cache_service "peer-main-bank" "peer-main-bank"
cache_service "peer-supplier" "peer-supplier"
cache_service "scf-chaincode" "scf-chaincode"

echo -e "${GREEN}🎉 All Go dependencies cached successfully!${NC}"
echo ""
echo "💡 Tips:"
echo "   - Docker builds will now be faster as dependencies are cached"
echo "   - Run this script whenever you update go.mod files"
echo "   - Use 'docker system prune' occasionally to clean old cache"
