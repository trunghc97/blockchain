#!/bin/bash

# Kubernetes Deployment Script for Blockchain System
# Usage: ./deploy.sh [namespace]

set -e

NAMESPACE=${1:-blockchain-production}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Deploying Blockchain System to Kubernetes"
echo "📁 Namespace: $NAMESPACE"
echo "📂 Script Directory: $SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl."
        exit 1
    fi

    # Check cluster access
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot access Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi

    # Check namespace exists or create it
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_info "Creating namespace $NAMESPACE..."
        kubectl create namespace "$NAMESPACE"
    fi

    log_success "Prerequisites check passed"
}

deploy_component() {
    local component=$1
    local path=$2

    log_info "Deploying $component..."

    if kubectl apply -f "$path" -n "$NAMESPACE"; then
        log_success "$component deployed successfully"
    else
        log_error "Failed to deploy $component"
        exit 1
    fi
}

wait_for_pods() {
    local label=$1
    local timeout=${2:-300}

    log_info "Waiting for pods with label '$label' to be ready..."

    if kubectl wait --for=condition=ready pod -l "$label" -n "$NAMESPACE" --timeout="${timeout}s"; then
        log_success "All pods are ready"
    else
        log_warning "Some pods are not ready. Continuing deployment..."
    fi
}

main() {
    echo "========================================"
    echo "🔗 Blockchain System K8s Deployment"
    echo "========================================"

    check_prerequisites

    # Deploy base resources
    log_info "📦 Phase 1: Deploying base resources..."
    deploy_component "namespace" "$SCRIPT_DIR/base/namespace.yaml"
    deploy_component "configmap" "$SCRIPT_DIR/base/configmap.yaml"
    deploy_component "secrets" "$SCRIPT_DIR/base/secret.yaml"
    deploy_component "RBAC" "$SCRIPT_DIR/base/rbac.yaml"

    # Deploy MongoDB
    log_info "🗄️  Phase 2: Deploying MongoDB..."
    deploy_component "MongoDB init script" "$SCRIPT_DIR/mongodb/init-script-configmap.yaml"
    deploy_component "MongoDB service" "$SCRIPT_DIR/mongodb/service.yaml"
    deploy_component "MongoDB PVC" "$SCRIPT_DIR/mongodb/pvc.yaml"
    deploy_component "MongoDB StatefulSet" "$SCRIPT_DIR/mongodb/statefulset.yaml"

    # Wait for MongoDB
    wait_for_pods "app.kubernetes.io/name=mongodb"
    log_info "Initializing MongoDB replica set..."
    kubectl exec -it statefulset/mongodb-shared -n "$NAMESPACE" -- mongosh --eval "rs.status()" || true

    # Deploy orderer cluster
    log_info "🏛️  Phase 3: Deploying Orderer cluster..."
    deploy_component "Orderer config" "$SCRIPT_DIR/orderer/configmap.yaml"
    deploy_component "Orderer service" "$SCRIPT_DIR/orderer/service.yaml"
    deploy_component "Orderer deployment" "$SCRIPT_DIR/orderer/deployment.yaml"
    deploy_component "Orderer HPA" "$SCRIPT_DIR/orderer/hpa.yaml"

    # Wait for orderer
    wait_for_pods "app.kubernetes.io/name=orderer"

    # Deploy peer nodes
    log_info "👥 Phase 4: Deploying Peer nodes..."
    for peer in peer-main-bank peer-supplier peer-anchor; do
        log_info "Deploying $peer..."
        deploy_component "$peer deployment" "$SCRIPT_DIR/peers/$peer/deployment.yaml"
        deploy_component "$peer service" "$SCRIPT_DIR/peers/$peer/service.yaml"
        deploy_component "$peer HPA" "$SCRIPT_DIR/peers/$peer/hpa.yaml"
    done

    # Wait for peers
    wait_for_pods "app.kubernetes.io/name=blockchain-peer"

    # Deploy backend
    log_info "🔧 Phase 5: Deploying Backend API..."
    deploy_component "Backend deployment" "$SCRIPT_DIR/backend/deployment.yaml"
    deploy_component "Backend service" "$SCRIPT_DIR/backend/service.yaml"
    deploy_component "Backend HPA" "$SCRIPT_DIR/backend/hpa.yaml"

    # Wait for backend
    wait_for_pods "app.kubernetes.io/name=backend"

    # Deploy frontend
    log_info "🌐 Phase 6: Deploying Frontend..."
    deploy_component "Frontend deployment" "$SCRIPT_DIR/frontend/deployment.yaml"
    deploy_component "Frontend service" "$SCRIPT_DIR/frontend/service.yaml"
    deploy_component "Frontend ingress" "$SCRIPT_DIR/frontend/ingress.yaml"

    # Wait for frontend
    wait_for_pods "app.kubernetes.io/name=frontend"

    # Deploy monitoring (optional)
    log_info "📊 Phase 7: Deploying Monitoring (optional)..."
    if kubectl apply -f "$SCRIPT_DIR/monitoring/" -n "$NAMESPACE" 2>/dev/null; then
        log_success "Monitoring components deployed"
    else
        log_warning "Monitoring deployment skipped (components may not be installed)"
    fi

    # Final status check
    log_info "📋 Deployment Summary:"
    echo ""
    kubectl get pods -n "$NAMESPACE" --show-labels
    echo ""
    kubectl get services -n "$NAMESPACE"
    echo ""
    kubectl get ingress -n "$NAMESPACE"

    log_success "🎉 Deployment completed successfully!"
    log_info "🌐 Access your application:"
    echo "  Frontend: https://your-domain.com"
    echo "  API Docs: https://your-domain.com/api/swagger-ui.html"
    echo "  Health Check: https://your-domain.com/api/actuator/health"

    log_warning "⚠️  Remember to:"
    echo "  1. Update domain name in ingress.yaml"
    echo "  2. Configure SSL certificates"
    echo "  3. Update container image references"
    echo "  4. Review and update secrets"
}

# Run main function
main "$@"
