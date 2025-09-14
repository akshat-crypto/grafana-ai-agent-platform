# ELK Stack Setup Guide

This guide explains how to deploy and use the ELK (Elasticsearch, Logstash, Kibana) stack in your Kubernetes cluster for centralized logging and monitoring.

## What is the ELK Stack?

The ELK stack consists of four main components:

1. **Elasticsearch**: A distributed search and analytics engine that stores and indexes logs
2. **Logstash**: A data processing pipeline that ingests, transforms, and sends data to Elasticsearch
3. **Kibana**: A web interface for exploring and visualizing data stored in Elasticsearch
4. **Filebeat**: A lightweight log shipper that collects logs from various sources and sends them to Logstash

## Prerequisites

- Kubernetes cluster with kubectl configured
- Helm (optional, for Helm-based deployment)
- At least 4GB of available memory
- At least 10GB of available storage

## Deployment Options

### Option 1: Using Helm Charts (Recommended)

1. **Install Helm** (if not already installed):
   ```bash
   curl https://get.helm.sh/helm-v3.15.0-linux-amd64.tar.gz | tar xz
   sudo mv linux-amd64/helm /usr/local/bin/helm
   ```

2. **Deploy using the script**:
   ```bash
   ./deploy-elk-stack.sh
   ```

### Option 2: Using Kubernetes Manifests

1. **Deploy using the script**:
   ```bash
   ./deploy-elk-manifests.sh
   ```

2. **Or manually apply the manifests**:
   ```bash
   kubectl apply -f elk-stack-manifests.yaml
   ```

## Accessing the Services

### Port Forwarding

After deployment, you can access the services using port forwarding:

```bash
# Access Kibana
kubectl port-forward -n elk-stack service/kibana 5601:5601

# Access Elasticsearch
kubectl port-forward -n elk-stack service/elasticsearch 9200:9200

# Access Logstash
kubectl port-forward -n elk-stack service/logstash 5044:5044
```

### Access URLs

- **Kibana**: http://localhost:5601
- **Elasticsearch**: http://localhost:9200
- **Logstash**: http://localhost:5044

## Configuration

### Elasticsearch

- **Memory**: 1GB minimum, 2GB recommended
- **Storage**: 10GB minimum
- **Security**: Disabled for development (enable in production)

### Logstash

- **Memory**: 512MB minimum, 1GB recommended
- **Inputs**: Filebeat (port 5044) and TCP (port 5000)
- **Outputs**: Elasticsearch with daily index rotation

### Kibana

- **Memory**: 512MB minimum, 1GB recommended
- **Features**: Basic visualization and dashboard capabilities

### Filebeat

- **Collection**: Container logs, system logs, and application logs
- **Processing**: Kubernetes metadata enrichment
- **Output**: Logstash for further processing

## Usage Examples

### Viewing Logs in Kibana

1. Open Kibana at http://localhost:5601
2. Go to **Discover** in the left sidebar
3. Create an index pattern: `logstash-*`
4. Set the time field to `@timestamp`
5. Start exploring your logs

### Creating Dashboards

1. In Kibana, go to **Dashboard**
2. Click **Create dashboard**
3. Add visualizations for:
   - Log volume over time
   - Top error messages
   - Container performance metrics
   - Kubernetes events

### Monitoring Specific Applications

1. **Filter by namespace**:
   ```
   kubernetes.namespace: "your-app-namespace"
   ```

2. **Filter by pod**:
   ```
   kubernetes.pod.name: "your-pod-name"
   ```

3. **Filter by log level**:
   ```
   log.level: "ERROR"
   ```

## Troubleshooting

### Common Issues

1. **Elasticsearch not starting**:
   - Check available memory (minimum 1GB required)
   - Verify storage class availability
   - Check pod logs: `kubectl logs -n elk-stack elasticsearch-0`

2. **Logstash connection issues**:
   - Verify Elasticsearch is running
   - Check Logstash logs: `kubectl logs -n elk-stack deployment/logstash`

3. **Filebeat not collecting logs**:
   - Check RBAC permissions
   - Verify host path mounts
   - Check Filebeat logs: `kubectl logs -n elk-stack daemonset/filebeat`

### Useful Commands

```bash
# Check pod status
kubectl get pods -n elk-stack

# Check service status
kubectl get services -n elk-stack

# Check pod logs
kubectl logs -n elk-stack <pod-name>

# Check events
kubectl get events -n elk-stack

# Check resource usage
kubectl top pods -n elk-stack
```

## Production Considerations

### Security

1. **Enable X-Pack security**:
   - Set `xpack.security.enabled: true`
   - Configure authentication and authorization
   - Enable TLS encryption

2. **Network policies**:
   - Restrict access to ELK services
   - Use ingress controllers with proper authentication

3. **RBAC**:
   - Limit Filebeat permissions to minimum required
   - Use dedicated service accounts

### Scaling

1. **Elasticsearch**:
   - Increase replicas for high availability
   - Use dedicated storage classes
   - Configure proper resource limits

2. **Logstash**:
   - Scale horizontally based on log volume
   - Use persistent queues for reliability

3. **Kibana**:
   - Scale for concurrent users
   - Use load balancers for high availability

### Monitoring

1. **Health checks**:
   - Monitor Elasticsearch cluster health
   - Check Logstash pipeline status
   - Monitor Filebeat agent health

2. **Metrics**:
   - Track log ingestion rates
   - Monitor storage usage
   - Alert on service failures

## Uninstalling

### Using Helm
```bash
helm uninstall filebeat logstash kibana elasticsearch -n elk-stack
kubectl delete namespace elk-stack
```

### Using Manifests
```bash
kubectl delete -f elk-stack-manifests.yaml
```

## Additional Resources

- [Elasticsearch Documentation](https://www.elastic.co/guide/index.html)
- [Kubernetes Logging](https://kubernetes.io/docs/concepts/cluster-administration/logging/)
- [ELK Stack Best Practices](https://www.elastic.co/blog/elk-stack-best-practices)

## Support

For issues related to:
- **ELK Stack**: Check the [Elastic forums](https://discuss.elastic.co/)
- **Kubernetes**: Check the [Kubernetes documentation](https://kubernetes.io/docs/)
- **This Setup**: Check the project repository or create an issue 