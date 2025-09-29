FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

# Set Go proxy to avoid GitHub authentication issues
ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app

# Copy source code
COPY . .

# Download and tidy dependencies
RUN go mod download
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orderer main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/orderer .

# Copy genesis config
COPY --from=builder /app/genesis ./genesis

# Copy secrets
COPY --from=builder /app/secrets ./secrets

# Expose port
EXPOSE 7050 7060 7070

# Run the application
CMD ["./orderer"]
