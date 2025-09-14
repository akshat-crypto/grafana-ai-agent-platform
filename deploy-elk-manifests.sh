#!/bin/bash

# Simple ELK Stack Deployment using Kubernetes Manifests
# This script deploys the ELK stack using the YAML manifests

set -e

echo "🚀 Starting ELK Stack deployment using Kubernetes manifests..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if we're connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo "✅ Connected to Kubernetes cluster: $(kubectl config current-context)"

# Apply the ELK stack manifests
echo "📦 Applying ELK stack manifests..."
kubectl apply -f elk-stack-manifests.yaml

# Wait for the namespace to be created
echo "⏳ Waiting for namespace to be created..."
kubectl wait --for=condition=active namespace/elk-stack --timeout=60s

# Wait for Elasticsearch to be ready
echo "⏳ Waiting for Elasticsearch to be ready..."
kubectl wait --for=condition=ready pod -l app=elasticsearch -n elk-stack --timeout=300s

# Wait for Logstash to be ready
echo "⏳ Waiting for Logstash to be ready..."
kubectl wait --for=condition=ready pod -l app=logstash -n elk-stack --timeout=300s

# Wait for Kibana to be ready
echo "⏳ Waiting for Kibana to be ready..."
kubectl wait --for=condition=ready pod -l app=kibana -n elk-stack --timeout=300s

# Wait for Filebeat to be ready
echo "⏳ Waiting for Filebeat to be ready..."
kubectl wait --for=condition=ready pod -l app=filebeat -n elk-stack --timeout=300s

# Get service information
echo "🔍 Getting service information..."
kubectl get services -n elk-stack

# Get pod status
echo "📊 Getting pod status..."
kubectl get pods -n elk-stack

echo "✅ ELK Stack deployment completed successfully!"
echo ""
echo "📋 Deployment Summary:"
echo "   - Namespace: elk-stack"
echo "   - Elasticsearch: Running (port 9200)"
echo "   - Logstash: Running (port 5044)"
echo "   - Kibana: Running (port 5601)"
echo "   - Filebeat: Running (DaemonSet)"
echo ""
echo "🌐 To access the services, you can use port forwarding:"
echo "   kubectl port-forward -n elk-stack service/kibana 5601:5601"
echo "   kubectl port-forward -n elk-stack service/elasticsearch 9200:9200"
echo ""
echo "🔗 Access URLs (after port forwarding):"
echo "   - Kibana: http://localhost:5601"
echo "   - Elasticsearch: http://localhost:9200"
echo ""
echo "🗑️  To uninstall the ELK stack, run:"
echo "   kubectl delete -f elk-stack-manifests.yaml" 