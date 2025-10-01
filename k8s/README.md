# Kubernetes Deployment Manifests

This directory contains Kubernetes manifests for deploying the Blockchain Supply Chain Finance system to production.

## Directory Structure

```
k8s/
├── base/                    # Base manifests (namespace, config, secrets, RBAC)
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   └── rbac.yaml
├── mongodb/                 # MongoDB StatefulSet with replica set
│   ├── statefulset.yaml
│   ├── service.yaml
│   ├── pvc.yaml
│   └── init-script-configmap.yaml
├── orderer/                 # Orderer cluster with Raft consensus
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   └── hpa.yaml
├── peers/                   # Peer nodes (main-bank, supplier, anchor)
│   ├── peer-main-bank/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── hpa.yaml
│   ├── peer-supplier/
│   └── peer-anchor/
├── backend/                 # Spring Boot API Gateway
│   ├── deployment.yaml
│   ├── service.yaml
│   └── hpa.yaml
├── frontend/                # Angular SPA with ingress
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
├── monitoring/              # Prometheus monitoring stack
│   ├── service-monitors.yaml
│   └── alert-rules.yaml
└── README.md
```

## Prerequisites

- Kubernetes cluster (v1.24+)
- kubectl configured
- Helm (optional)
- NGINX Ingress Controller
- cert-manager (for SSL certificates)

## Quick Start

```bash
# 1. Create namespace and base resources
kubectl apply -f k8s/base/

# 2. Deploy MongoDB
kubectl apply -f k8s/mongodb/

# 3. Deploy orderer cluster
kubectl apply -f k8s/orderer/

# 4. Deploy peer nodes
kubectl apply -f k8s/peers/

# 5. Deploy backend API
kubectl apply -f k8s/backend/

# 6. Deploy frontend
kubectl apply -f k8s/frontend/

# 7. Deploy monitoring (optional)
kubectl apply -f k8s/monitoring/
```

## Configuration

### Environment Variables

Before deploying, update the following:

1. **Domain Name**: Update `your-domain.com` in `k8s/frontend/ingress.yaml`
2. **Container Images**: Update image references to your registry
3. **Secrets**: Update passwords in `k8s/base/secret.yaml`
4. **Storage Class**: Update storage classes for your cluster

### Secrets Management

```bash
# Generate secure passwords
openssl rand -base64 32

# Update secrets
kubectl apply -f k8s/base/secret.yaml
```

### Orderer Keys

```bash
# Generate orderer keys (from project root)
node scripts/generate-orderer-keys.js

# Create configmap
kubectl create configmap orderer-keys-config \
  --from-file=secrets/ord1/ \
  --from-file=secrets/ord2/ \
  --from-file=secrets/ord3/ \
  -n blockchain-production
```

## Scaling

### Horizontal Pod Autoscaling

All services include HPA configurations:

```bash
# Scale backend based on CPU/memory
kubectl apply -f k8s/backend/hpa.yaml

# Scale orderer cluster
kubectl scale deployment orderer-cluster --replicas=5

# Scale peer nodes
kubectl scale deployment peer-supplier --replicas=3
```

### MongoDB Scaling

```bash
# Scale MongoDB replica set
kubectl scale statefulset mongodb-shared --replicas=5
```

## Monitoring

### Prometheus & Grafana

```bash
# Install Prometheus stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/prometheus

# Install Grafana
helm repo add grafana https://grafana.github.io/helm-charts
helm install grafana grafana/grafana

# Access Grafana
kubectl port-forward svc/grafana 3000:80
# Default credentials: admin/admin
```

### Custom Dashboards

Import blockchain-specific dashboards from `k8s/monitoring/dashboards/`

## Backup & Recovery

### Database Backup

```yaml
# CronJob for automated backups
kubectl apply -f k8s/mongodb/backup-cronjob.yaml
```

### Disaster Recovery

```bash
# Restore from backup
kubectl apply -f k8s/mongodb/restore-job.yaml

# Failover procedures
kubectl scale deployment orderer-cluster --replicas=0
kubectl apply -f k8s/orderer/new-cluster.yaml
```

## Security

### Network Policies

Apply network policies to restrict traffic:

```yaml
kubectl apply -f k8s/base/network-policy.yaml
```

### Pod Security Standards

```bash
kubectl label namespace blockchain-production \
  pod-security.kubernetes.io/enforce=restricted
```

## Troubleshooting

### Common Issues

1. **Pods not starting**: Check resource limits and node capacity
2. **Services not accessible**: Verify service selectors and labels
3. **Ingress not working**: Check ingress class and TLS certificates
4. **Database connection**: Verify MongoDB replica set status

### Debug Commands

```bash
# Check pod status
kubectl get pods -n blockchain-production

# View logs
kubectl logs -f deployment/backend -n blockchain-production

# Debug containers
kubectl exec -it deployment/backend -n blockchain-production -- /bin/bash

# Check services
kubectl get services -n blockchain-production

# Verify ingress
kubectl describe ingress blockchain-ingress -n blockchain-production
```

## CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/k8s-deploy.yml
name: Deploy to Kubernetes
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Configure kubectl
      uses: azure/k8s-set-context@v2
      with:
        kubeconfig: ${{ secrets.KUBE_CONFIG }}
    - name: Deploy
      run: |
        kubectl apply -f k8s/
        kubectl rollout status deployment/backend -n blockchain-production
        kubectl rollout status deployment/frontend -n blockchain-production
```

## Performance Tuning

### Resource Optimization

```yaml
# Adjust based on load testing
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi
```

### JVM Tuning

```bash
# For Spring Boot applications
JAVA_OPTS="-Xmx2g -Xms512m -XX:+UseG1GC -XX:+UseContainerSupport"
```

## Maintenance

### Updates

```bash
# Update images
kubectl set image deployment/backend backend=blockchain/backend:v1.1.0

# Rolling updates
kubectl rollout restart deployment/backend

# Check rollout status
kubectl rollout status deployment/backend
```

### Cleanup

```bash
# Remove all resources
kubectl delete namespace blockchain-production

# Remove specific components
kubectl delete -f k8s/monitoring/
kubectl delete -f k8s/frontend/
```

---

For more detailed information, refer to the main [README.md](../README.md) in the project root.
