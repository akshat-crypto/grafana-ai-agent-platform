#!/bin/bash

# ELK Stack Deployment Script
# This script deploys Elasticsearch, Logstash, Kibana, and Filebeat to Kubernetes

set -e

echo "🚀 Starting ELK Stack deployment..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if helm is available
if ! command -v helm &> /dev/null; then
    echo "❌ Helm is not installed. Please install Helm first."
    exit 1
fi

# Check if we're connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo "✅ Connected to Kubernetes cluster: $(kubectl config current-context)"

# Create namespace for ELK stack
echo "📦 Creating namespace for ELK stack..."
kubectl create namespace elk-stack --dry-run=client -o yaml | kubectl apply -f -

# Add Elastic Helm repository
echo "📚 Adding Elastic Helm repository..."
helm repo add elastic https://helm.elastic.co
helm repo update

# Deploy Elasticsearch
echo "🔍 Deploying Elasticsearch..."
helm install elasticsearch elastic/elasticsearch \
    --namespace elk-stack \
    --set replicas=1 \
    --set resources.requests.memory="1Gi" \
    --set resources.requests.cpu="500m" \
    --set resources.limits.memory="2Gi" \
    --set resources.limits.cpu="1000m" \
    --set volumeClaimTemplate.resources.requests.storage="10Gi" \
    --set securityContext.runAsUser=1000 \
    --set securityContext.fsGroup=1000 \
    --wait --timeout=10m

# Wait for Elasticsearch to be ready
echo "⏳ Waiting for Elasticsearch to be ready..."
kubectl wait --for=condition=ready pod -l app=elasticsearch-master -n elk-stack --timeout=300s

# Deploy Logstash
echo "📥 Deploying Logstash..."
helm install logstash elastic/logstash \
    --namespace elk-stack \
    --set replicas=1 \
    --set resources.requests.memory="512Mi" \
    --set resources.requests.cpu="250m" \
    --set resources.limits.memory="1Gi" \
    --set resources.limits.cpu="500m" \
    --wait --timeout=10m

# Deploy Kibana
echo "📊 Deploying Kibana..."
helm install kibana elastic/kibana \
    --namespace elk-stack \
    --set replicas=1 \
    --set resources.requests.memory="512Mi" \
    --set resources.requests.cpu="250m" \
    --set resources.limits.memory="1Gi" \
    --set resources.limits.cpu="500m" \
    --set service.type=ClusterIP \
    --wait --timeout=10m

# Deploy Filebeat
echo "📋 Deploying Filebeat..."
helm install filebeat elastic/filebeat \
    --namespace elk-stack \
    --set daemonset.enabled=true \
    --set daemonset.resources.requests.memory="100Mi" \
    --set daemonset.resources.requests.cpu="100m" \
    --set daemonset.resources.limits.memory="200Mi" \
    --set daemonset.resources.limits.cpu="200m" \
    --wait --timeout=10m

# Wait for all pods to be ready
echo "⏳ Waiting for all ELK stack pods to be ready..."
kubectl wait --for=condition=ready pod -l app=logstash -n elk-stack --timeout=300s
kubectl wait --for=condition=ready pod -l app=kibana -n elk-stack --timeout=300s

# Get service information
echo "🔍 Getting service information..."
kubectl get services -n elk-stack

# Port forward services for local access
echo "🌐 Setting up port forwarding for local access..."
echo "📊 Kibana will be available at: http://localhost:5601"
echo "🔍 Elasticsearch will be available at: http://localhost:9200"
echo "📥 Logstash will be available at: http://localhost:5044"

# Create port forwarding in background
kubectl port-forward -n elk-stack service/kibana-kibana 5601:5601 &
KIBANA_PID=$!

kubectl port-forward -n elk-stack service/elasticsearch-master 9200:9200 &
ES_PID=$!

echo "✅ ELK Stack deployment completed successfully!"
echo ""
echo "📋 Deployment Summary:"
echo "   - Namespace: elk-stack"
echo "   - Elasticsearch: Running (port 9200)"
echo "   - Logstash: Running (port 5044)"
echo "   - Kibana: Running (port 5601)"
echo "   - Filebeat: Running (DaemonSet)"
echo ""
echo "🔗 Access URLs:"
echo "   - Kibana: http://localhost:5601"
echo "   - Elasticsearch: http://localhost:9200"
echo ""
echo "🛑 To stop port forwarding, run:"
echo "   kill $KIBANA_PID $ES_PID"
echo ""
echo "🗑️  To uninstall the ELK stack, run:"
echo "   helm uninstall filebeat logstash kibana elasticsearch -n elk-stack"
echo "   kubectl delete namespace elk-stack" 