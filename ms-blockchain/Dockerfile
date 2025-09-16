# Build stage
FROM golang:1.21-alpine AS build

WORKDIR /app

# Copy source code first
COPY . .

# Initialize go modules and download dependencies
RUN go mod init ms-blockchain || true && \
    go mod tidy && \
    go mod download

# Build application with cached dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o main .

# Production stage
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
EXPOSE 8081
CMD ["./main"]