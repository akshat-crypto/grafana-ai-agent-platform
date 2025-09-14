#!/bin/bash

# Test kubeconfig validation API
echo "Testing kubeconfig validation API..."

# Test with empty kubeconfig
echo "Test 1: Empty kubeconfig"
curl -X POST http://localhost:8080/api/kubernetes/validate \
  -H "Content-Type: application/json" \
  -d '{"kube_config": ""}'

echo -e "\n\n"

# Test with invalid kubeconfig
echo "Test 2: Invalid kubeconfig"
curl -X POST http://localhost:8080/api/kubernetes/validate \
  -H "Content-Type: application/json" \
  -d '{"kube_config": "invalid: yaml"}'

echo -e "\n\n"

# Test with valid kubeconfig format
echo "Test 3: Valid kubeconfig format (but invalid content)"
curl -X POST http://localhost:8080/api/kubernetes/validate \
  -H "Content-Type: application/json" \
  -d '{"kube_config": "apiVersion: v1\nkind: Config\nclusters:\n- name: test\n  cluster:\n    server: https://test:6443\ncontexts:\n- name: test\n  context:\n    cluster: test\n    user: test\ncurrent-context: test\nusers:\n- name: test\n  user:\n    token: test"}'

echo -e "\n\n" 