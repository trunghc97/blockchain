#!/bin/bash

# Kubernetes Deployment Script for Blockchain System
# Usage: ./deploy-k8s.sh [namespace]

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
    deploy_component "namespace" "$SCRIPT_DIR/k8s-base/namespace.yaml"
    deploy_component "configmap" "$SCRIPT_DIR/k8s-base/configmap.yaml"
    deploy_component "secrets" "$SCRIPT_DIR/k8s-base/secret.yaml"
    deploy_component "RBAC" "$SCRIPT_DIR/k8s-base/rbac.yaml"

    # Deploy MongoDB
    log_info "🗄️  Phase 2: Deploying MongoDB..."
    deploy_component "MongoDB manifests" "$SCRIPT_DIR/mongodb/k8s/"

    # Wait for MongoDB
    wait_for_pods "app.kubernetes.io/name=mongodb"
    log_info "MongoDB cluster is ready"

    # Deploy microservices
    log_info "🏛️  Phase 3: Deploying Orderer cluster..."
    deploy_component "Orderer cluster" "$SCRIPT_DIR/orderer/k8s/"
    wait_for_pods "app.kubernetes.io/name=orderer"

    log_info "👥 Phase 4: Deploying Peer nodes..."
    deploy_component "Peer Main Bank" "$SCRIPT_DIR/peer-main-bank/k8s/"
    deploy_component "Peer Supplier" "$SCRIPT_DIR/peer-supplier/k8s/"
    deploy_component "Peer Anchor" "$SCRIPT_DIR/peer-anchor/k8s/"
    wait_for_pods "app.kubernetes.io/name=blockchain-peer"

    log_info "🔧 Phase 5: Deploying Backend API..."
    deploy_component "Backend API" "$SCRIPT_DIR/backend/k8s/"
    wait_for_pods "app.kubernetes.io/name=backend"

    log_info "🌐 Phase 6: Deploying Frontend..."
    deploy_component "Frontend" "$SCRIPT_DIR/frontend/k8s/"
    wait_for_pods "app.kubernetes.io/name=frontend"

    # Deploy monitoring (optional)
    log_info "📊 Phase 7: Deploying Monitoring (optional)..."
    if kubectl apply -f "$SCRIPT_DIR/monitoring/k8s/" -n "$NAMESPACE" 2>/dev/null; then
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
    echo "  1. Update domain name in frontend/k8s/deployment.yaml"
    echo "  2. Configure SSL certificates"
    echo "  3. Update container image references"
    echo "  4. Review and update secrets in k8s-base/"
}

# Run main function
main "$@"
