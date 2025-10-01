# 🚀 Go Build Optimization Guide

## 📦 Dependency Caching

### Docker Layer Caching
Dockerfiles đã được tối ưu để cache Go dependencies:

```dockerfile
# Copy go mod files first for better caching
COPY peer-anchor/go.mod peer-anchor/go.sum ./peer-anchor/

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download

# Copy source code after dependencies are cached
COPY peer-anchor/ .
```

**Lợi ích:**
- Chỉ download dependencies khi `go.mod`/`go.sum` thay đổi
- Build nhanh hơn khi chỉ thay đổi source code

### Docker Volume Caching
Docker Compose sử dụng named volumes để cache Go modules:

```yaml
volumes:
  - go-mod-cache:/go/pkg/mod
  - go-build-cache:/root/.cache/go-build
```

**Lợi ích:**
- Dependencies được lưu trữ persistently qua các lần build
- Không cần download lại trừ khi thay đổi

## 🛠️ Usage

### Build với cache
```bash
# Build lần đầu (sẽ download dependencies)
docker-compose build

# Build lần sau (sử dụng cache)
docker-compose build
```

### Pre-cache dependencies local
```bash
# Chạy script để cache dependencies local
./build-go-cache.sh
```

### Clean cache
```bash
# Clean Docker cache
docker system prune -f

# Clean Go volumes
docker volume rm $(docker volume ls -q | grep go-)
```

## 📊 Performance Comparison

### Trước khi tối ưu:
- Build time: ~3-5 phút mỗi lần
- Download dependencies: Mỗi lần build

### Sau khi tối ưu:
- Build time: ~30-60 giây (chỉ thay đổi source)
- Download dependencies: Chỉ khi go.mod thay đổi
- Cache hit: ~80-90% cho source code changes

## 🔧 Troubleshooting

### Cache không hoạt động
```bash
# Force rebuild without cache
docker-compose build --no-cache

# Check cache volumes
docker volume ls | grep go-
```

### Dependencies corrupted
```bash
# Remove and recreate volumes
docker volume rm go-mod-cache go-build-cache
docker volume create go-mod-cache
docker volume create go-build-cache
```

### Network issues
```bash
# Use different Go proxy
export GOPROXY=https://goproxy.cn,direct

# Or disable proxy
export GOPROXY=direct
```

## 💡 Best Practices

1. **Sử dụng named volumes** cho persistent caching
2. **Chạy `./build-go-cache.sh`** khi update dependencies
3. **Clean cache** định kỳ để tránh disk đầy
4. **Monitor build time** để đánh giá hiệu quả
5. **Use `--no-cache`** khi có vấn đề với cache

## 📈 Monitoring

```bash
# Check cache size
docker system df -v

# Monitor build time
time docker-compose build peer-anchor

# Check volume usage
docker volume inspect go-mod-cache
```
