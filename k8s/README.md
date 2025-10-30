# Kubernetes Deployment Guide for DHCP2P

This guide explains how to deploy the DHCP2P application on Kubernetes.

## Contents

Files in this directory:
- `namespace.yaml` - Create the application namespace
- `configmap.yaml` - Non-sensitive configuration
- `secret.yaml.example` - Template for sensitive values
- `deployment.yaml` - Main application Deployment
- `service.yaml` - Service to expose the app inside the cluster
- `ingress.yaml` - Ingress for external access (optional)
- `hpa.yaml` - Horizontal Pod Autoscaler (optional)
- `postgres-statefulset.yaml` - PostgreSQL StatefulSet (optional)
- `redis-statefulset.yaml` - Redis StatefulSet (optional)

## Requirements

- Kubernetes cluster 1.20+
- Configured kubectl
- DHCP2P Docker image built and pushed to a registry

## Deployment Steps

### 1) Create Namespace

```bash
kubectl apply -f namespace.yaml
```

### 2) Create ConfigMap

```bash
kubectl apply -f configmap.yaml
```

### 3) Create Secret

Important: Create a Secret with your real values before deploying:

```bash
# Copy template
cp secret.yaml.example secret.yaml

# Edit secret.yaml with your actual values
# Then apply it:
kubectl apply -f secret.yaml
```

Or create the Secret from the command line:

```bash
kubectl create secret generic dhcp2p-secrets \
  # In-cluster Postgres (no SSL)
  --from-literal=DATABASE_URL='postgres://user:password@postgres-service:5432/dhcp2p?sslmode=disable' \
  --from-literal=REDIS_URL='redis-service:6379' \
  --from-literal=REDIS_PASSWORD='your-redis-password' \
  -n dhcp2p
```

If you use a managed database requiring TLS (e.g., RDS/Cloud SQL), switch to `sslmode=require` and provide the appropriate certificates if needed.

### 4) (Optional) Deploy PostgreSQL and Redis

If you wish to deploy PostgreSQL and Redis inside the cluster (not recommended for production):

```bash
# Deploy PostgreSQL
kubectl apply -f postgres-statefulset.yaml

# Deploy Redis
kubectl apply -f redis-statefulset.yaml

# Wait for readiness
kubectl wait --for=condition=ready pod -l app=postgres -n dhcp2p --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n dhcp2p --timeout=300s
```

Note: For production, prefer managed services like AWS RDS/Google Cloud SQL (PostgreSQL) and AWS ElastiCache/Google Memorystore (Redis).

### 5) Update Deployment Image

Before deploying, update the image in `deployment.yaml`:

```bash
# Replace image name in deployment.yaml
sed -i 's|dhcp2p:latest|your-registry/dhcp2p:v1.0.0|g' deployment.yaml
```

Or edit `deployment.yaml` directly:

```yaml
containers:
- name: dhcp2p
  image: your-registry/dhcp2p:v1.0.0  # Update here
```

### 6) Deploy the Application

```bash
kubectl apply -f deployment.yaml
```

### 7) Create Service

```bash
kubectl apply -f service.yaml
```

### 8) (Optional) Create Ingress

To expose the app externally:

```bash
# Update ingress.yaml with your domain
sed -i 's/dhcp2p.example.com/your-domain.com/g' ingress.yaml

# Apply ingress
kubectl apply -f ingress.yaml
```

Note: You need an Ingress Controller installed (e.g., nginx-ingress).

### 9) (Optional) Deploy HPA

Enable autoscaling based on CPU/Memory:

```bash
# Ensure metrics-server is installed
kubectl top nodes  # quick check

# Deploy HPA
kubectl apply -f hpa.yaml
```

## One-shot Deploy

To deploy everything (except Secret and optional DB/Redis if using managed services):

```bash
# Namespace and ConfigMap
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml

# Secret (ensure it's configured correctly)
kubectl apply -f secret.yaml

# App
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# Optional
kubectl apply -f ingress.yaml
kubectl apply -f hpa.yaml
```

## Verify Deployment

```bash
# Pods
kubectl get pods -n dhcp2p

# Logs
kubectl logs -f deployment/dhcp2p -n dhcp2p

# Services
kubectl get svc -n dhcp2p

# Health endpoint (via port-forward)
kubectl port-forward svc/dhcp2p-service 8088:8088 -n dhcp2p
curl http://localhost:8088/health
```

## Environment Profiles

### Development

```bash
# Use local image
sed -i 's|dhcp2p:latest|dhcp2p:dev|g' deployment.yaml

# Reduce resources
sed -i 's/replicas: 2/replicas: 1/g' deployment.yaml
```

### Production

1. Use managed database and cache services
2. Update `DATABASE_URL` and `REDIS_URL` in the Secret
3. Enable TLS in Ingress
4. Set appropriate resource requests/limits
5. Enable HPA
6. Use a production-grade image registry

## Scaling

### Manual scaling

```bash
kubectl scale deployment dhcp2p --replicas=5 -n dhcp2p
```

### Auto-scaling with HPA

HPA automatically scales based on CPU and memory. See `hpa.yaml`:

```yaml
minReplicas: 2
maxReplicas: 10
metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        averageUtilization: 80
```

## Troubleshooting

### Pod does not start

```bash
# Events
kubectl describe pod <pod-name> -n dhcp2p

# Logs
kubectl logs <pod-name> -n dhcp2p
```

### Database/Redis connectivity issues

```bash
# Test connectivity from a pod
kubectl exec -it <pod-name> -n dhcp2p -- sh
ping postgres-service
ping redis-service
```

### Health check failures

```bash
# Test health endpoint
kubectl exec -it <pod-name> -n dhcp2p -- curl http://localhost:8088/health
```

## Security Best Practices

1. Use Secrets: never hardcode credentials in ConfigMaps
2. RBAC: create a ServiceAccount with least privilege
3. Network Policies: limit traffic between pods/namespaces as needed
4. Pod Security Standards: apply security contexts as in deployment.yaml
5. Image Security: scan images before deploying
6. TLS: enable TLS for all external communications

## Backup and Recovery

### Database Backup

```bash
# Backup from PostgreSQL pod (if using StatefulSet)
kubectl exec postgres-0 -n dhcp2p -- pg_dump -U dhcp2p dhcp2p > backup.sql

# Restore
kubectl exec -i postgres-0 -n dhcp2p -- psql -U dhcp2p dhcp2p < backup.sql
```

### Redis Backup

```bash
# Redis persistence is handled automatically when using StatefulSet with PVC
# Manual backup:
kubectl exec redis-0 -n dhcp2p -- redis-cli -a $REDIS_PASSWORD BGSAVE
```

## Monitoring

To monitor the application:

1. Use Prometheus and Grafana
2. Expose/enable a metrics endpoint (if available)
3. Optionally use a service mesh (e.g., Istio)
4. Consider cloud monitoring solutions (CloudWatch, Cloud Monitoring, etc.)

## Rollback

```bash
# Show deployment history
kubectl rollout history deployment/dhcp2p -n dhcp2p

# Roll back to previous revision
kubectl rollout undo deployment/dhcp2p -n dhcp2p

# Roll back to a specific revision
kubectl rollout undo deployment/dhcp2p --to-revision=2 -n dhcp2p
```

## Cleanup

To delete everything:

```bash
kubectl delete namespace dhcp2p
```

Note: This deletes all resources in the namespace, including data if using PersistentVolumes bound to the namespace.

