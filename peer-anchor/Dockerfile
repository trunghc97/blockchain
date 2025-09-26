FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o peer main.go

FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/peer .

EXPOSE 8084

CMD ["./peer"]
