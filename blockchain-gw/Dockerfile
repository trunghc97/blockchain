# Build stage
FROM golang:1.23-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set Go proxy to avoid GitHub authentication issues
ENV GOPROXY=https://proxy.golang.org,direct

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY blockchain-gw/go.mod blockchain-gw/go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download

# Copy source code after dependencies are cached
COPY blockchain-gw/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 9090 8080

# Run the binary
CMD ["./main"]
