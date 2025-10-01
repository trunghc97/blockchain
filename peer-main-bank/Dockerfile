FROM golang:1.23-alpine AS builder

# Install git for go mod operations
RUN apk add --no-cache git

# Set Go proxy to avoid GitHub authentication issues
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app
COPY . .

# Resolve dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o peer .

FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/peer .

EXPOSE 8082

CMD ["./peer"]
