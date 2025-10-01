FROM golang:1.23-alpine AS builder

# Install git for go mod operations
RUN apk add --no-cache git

# Set Go proxy to avoid GitHub authentication issues
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app

# Copy go mod files and share directory first for better caching
COPY peer-main-bank/go.mod peer-main-bank/go.sum ./peer-main-bank/
COPY share/ ./share/

# Change to peer-main-bank directory for Go operations
WORKDIR /app/peer-main-bank

# Disable go work mode for this build
ENV GOWORK=off

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download

# Copy source code after dependencies are cached
COPY peer-main-bank/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o peer .

FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/peer-main-bank/peer .

EXPOSE 8082

CMD ["./peer"]
