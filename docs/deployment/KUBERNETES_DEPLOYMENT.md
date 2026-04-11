# Kubernetes Deployment Guide

**Document Version:** 1.0  
**Last Updated:** April 6, 2026  
**Applies to:** Catalogizer v2.3.0+  
**Kubernetes Version:** 1.28+  

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Cluster Setup](#2-cluster-setup)
3. [Helm Charts](#3-helm-charts)
4. [Configuration](#4-configuration)
5. [Deployment Procedures](#5-deployment-procedures)
6. [Scaling & Auto-scaling](#6-scaling--auto-scaling)
7. [Monitoring](#7-monitoring)
8. [Troubleshooting](#8-troubleshooting)

---

## 1. Prerequisites

### 1.1 Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| kubectl | 1.28+ | Kubernetes CLI |
| Helm | 3.12+ | Package management |
| kustomize | 5.0+ | Configuration management |
| cert-manager | 1.13+ | TLS certificates |
| ingress-nginx | 1.9+ | Ingress controller |

### 1.2 Cluster Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| Nodes | 3 | 5+ |
| vCPU | 8 cores | 16+ cores |
| Memory | 16 GB | 32+ GB |
| Storage | 100 GB SSD | 500 GB+ SSD |
| Network | 1 Gbps | 10 Gbps |

### 1.3 Namespace Structure

```
catalogizer-prod/          # Production workloads
├── catalogizer-api/       # API deployment
├── catalogizer-web/       # Web frontend
├── database/              # PostgreSQL
├── cache/                 # Redis
└── monitoring/            # Prometheus/Grafana

catalogizer-staging/       # Staging environment
catalogizer-dev/          # Development environment
```

---

## 2. Cluster Setup

### 2.1 Namespace Creation

```bash
#!/bin/bash
# setup-namespaces.sh

NAMESPACES=("catalogizer-prod" "catalogizer-staging" "catalogizer-dev")

for ns in "${NAMESPACES[@]}"; do
    kubectl create namespace "$ns" --dry-run=client -o yaml | kubectl apply -f -
    
    # Label namespaces for monitoring
    kubectl label namespace "$ns" app.kubernetes.io/part-of=catalogizer
    kubectl label namespace "$ns" environment="${ns##*-}"
done
```

### 2.2 Secret Management

```bash
#!/bin/bash
# create-secrets.sh

NAMESPACE="catalogizer-prod"

# Database credentials
kubectl create secret generic db-credentials \
    --namespace="$NAMESPACE" \
    --from-literal=username=catalogizer \
    --from-literal=password="$(openssl rand -base64 32)" \
    --dry-run=client -o yaml | kubectl apply -f -

# JWT secret
kubectl create secret generic jwt-secret \
    --namespace="$NAMESPACE" \
    --from-literal=secret="$(openssl rand -base64 64)" \
    --dry-run=client -o yaml | kubectl apply -f -

# API keys
kubectl create secret generic api-keys \
    --namespace="$NAMESPACE" \
    --from-literal=tmdb="$TMDB_API_KEY" \
    --from-literal=omdb="$OMDB_API_KEY" \
    --dry-run=client -o yaml | kubectl apply -f -
```

### 2.3 ConfigMap for Application Settings

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: catalogizer-config
  namespace: catalogizer-prod
data:
  config.json: |
    {
      "api": {
        "port": 8080,
        "log_level": "info",
        "cors_origins": ["https://catalogizer.example.com"]
      },
      "database": {
        "host": "postgres.database.svc.cluster.local",
        "port": 5432,
        "name": "catalogizer"
      },
      "redis": {
        "host": "redis.cache.svc.cluster.local",
        "port": 6379
      },
      "features": {
        "websocket": true,
        "analytics": true,
        "rate_limiting": true
      }
    }
```

---

## 3. Helm Charts

### 3.1 Chart Structure

```
helm/
├── catalogizer/
│   ├── Chart.yaml
│   ├── values.yaml
│   ├── values-prod.yaml
│   ├── values-staging.yaml
│   └── templates/
│       ├── _helpers.tpl
│       ├── deployment-api.yaml
│       ├── deployment-web.yaml
│       ├── service-api.yaml
│       ├── service-web.yaml
│       ├── ingress.yaml
│       ├── hpa.yaml
│       └── pdb.yaml
└── README.md
```

### 3.2 Main Chart (Chart.yaml)

```yaml
apiVersion: v2
name: catalogizer
description: A Helm chart for Catalogizer media management system
type: application
version: 2.2.0
appVersion: "2.2.0"

dependencies:
  - name: postgresql
    version: 12.12.10
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
  
  - name: redis
    version: 18.12.0
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
  
  - name: ingress-nginx
    version: 4.9.0
    repository: https://kubernetes.github.io/ingress-nginx
    condition: ingress-nginx.enabled
```

### 3.3 API Deployment Template

```yaml
# templates/deployment-api.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "catalogizer.fullname" . }}-api
  labels:
    {{- include "catalogizer.labels" . | nindent 4 }}
    app.kubernetes.io/component: api
spec:
  {{- if not .Values.api.autoscaling.enabled }}
  replicas: {{ .Values.api.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "catalogizer.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: api
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
      labels:
        {{- include "catalogizer.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: api
    spec:
      serviceAccountName: {{ include "catalogizer.serviceAccountName" . }}
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      
      containers:
        - name: api
          image: "{{ .Values.api.image.repository }}:{{ .Values.api.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.api.image.pullPolicy }}
          
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
            - name: websocket
              containerPort: 8081
              protocol: TCP
          
          env:
            - name: PORT
              value: "8080"
            - name: GIN_MODE
              value: "release"
            - name: DB_HOST
              value: {{ .Values.postgresql.host | quote }}
            - name: DB_PORT
              value: {{ .Values.postgresql.port | quote }}
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: username
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: password
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: jwt-secret
                  key: secret
          
          volumeMounts:
            - name: config
              mountPath: /app/config.json
              subPath: config.json
            - name: media-storage
              mountPath: /media
          
          resources:
            {{- toYaml .Values.api.resources | nindent 12 }}
          
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          
          readinessProbe:
            httpGet:
              path: /ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
      
      volumes:
        - name: config
          configMap:
            name: {{ include "catalogizer.fullname" . }}-config
        - name: media-storage
          persistentVolumeClaim:
            claimName: {{ include "catalogizer.fullname" . }}-media
```

### 3.4 Values File (values-prod.yaml)

```yaml
# Production values
api:
  replicaCount: 3
  
  image:
    repository: ghcr.io/catalogizer/catalog-api
    tag: "2.2.0"
    pullPolicy: IfNotPresent
  
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 1000m
      memory: 2Gi
  
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80
  
  podDisruptionBudget:
    enabled: true
    minAvailable: 2

web:
  replicaCount: 2
  
  image:
    repository: ghcr.io/catalogizer/catalog-web
    tag: "2.2.0"
  
  resources:
    limits:
      cpu: 1000m
      memory: 1Gi
    requests:
      cpu: 500m
      memory: 512Mi

postgresql:
  enabled: true
  host: postgres.database.svc.cluster.local
  port: 5432
  
  auth:
    existingSecret: db-credentials
    database: catalogizer
  
  primary:
    persistence:
      size: 100Gi
      storageClass: fast-ssd
    resources:
      limits:
        cpu: 2000m
        memory: 4Gi
      requests:
        cpu: 1000m
        memory: 2Gi
  
  readReplicas:
    replicaCount: 2

redis:
  enabled: true
  architecture: replication
  
  auth:
    enabled: true
    existingSecret: redis-credentials
  
  master:
    persistence:
      size: 10Gi
  
  replica:
    replicaCount: 2

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "100m"
  
  hosts:
    - host: catalogizer.example.com
      paths:
        - path: /api
          pathType: Prefix
          service: api
        - path: /
          pathType: Prefix
          service: web
  
  tls:
    - secretName: catalogizer-tls
      hosts:
        - catalogizer.example.com
```

---

## 4. Configuration

### 4.1 Ingress Configuration

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: catalogizer
  namespace: catalogizer-prod
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    # WebSocket support
    nginx.ingress.kubernetes.io/proxy-http-version: "1.1"
    nginx.ingress.kubernetes.io/proxy-set-headers: |
      Upgrade $http_upgrade
      Connection "upgrade"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - catalogizer.example.com
      secretName: catalogizer-tls
  rules:
    - host: catalogizer.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: catalogizer-api
                port:
                  number: 8080
          - path: /ws
            pathType: Prefix
            backend:
              service:
                name: catalogizer-api
                port:
                  number: 8081
          - path: /
            pathType: Prefix
            backend:
              service:
                name: catalogizer-web
                port:
                  number: 80
```

### 4.2 Network Policies

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: catalogizer-api
  namespace: catalogizer-prod
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/component: api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: postgresql
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: redis
      ports:
        - protocol: TCP
          port: 6379
```

---

## 5. Deployment Procedures

### 5.1 Initial Deployment

```bash
#!/bin/bash
# deploy.sh

NAMESPACE="catalogizer-prod"
ENVIRONMENT="prod"
VERSION="2.2.0"

echo "=== Deploying Catalogizer $VERSION to $NAMESPACE ==="

# 1. Add Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# 2. Create namespace
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# 3. Create secrets
./scripts/create-secrets.sh

# 4. Deploy dependencies
echo "Deploying PostgreSQL..."
helm upgrade --install postgres bitnami/postgresql \
    --namespace "$NAMESPACE" \
    --values helm/values/postgres-$ENVIRONMENT.yaml

echo "Deploying Redis..."
helm upgrade --install redis bitnami/redis \
    --namespace "$NAMESPACE" \
    --values helm/values/redis-$ENVIRONMENT.yaml

# 5. Deploy application
echo "Deploying Catalogizer..."
helm upgrade --install catalogizer ./helm/catalogizer \
    --namespace "$NAMESPACE" \
    --values helm/catalogizer/values-$ENVIRONMENT.yaml \
    --set api.image.tag="$VERSION" \
    --set web.image.tag="$VERSION" \
    --wait \
    --timeout 10m

# 6. Verify deployment
echo "Verifying deployment..."
kubectl rollout status deployment/catalogizer-api -n "$NAMESPACE"
kubectl rollout status deployment/catalogizer-web -n "$NAMESPACE"

# 7. Run smoke tests
./scripts/smoke-tests.sh

echo "=== Deployment Complete ==="
```

### 5.2 Blue-Green Deployment

```bash
#!/bin/bash
# blue-green-deploy.sh

VERSION="2.3.0"
NAMESPACE="catalogizer-prod"

# Deploy new version (green)
helm upgrade --install catalogizer-green ./helm/catalogizer \
    --namespace "$NAMESPACE" \
    --values helm/catalogizer/values-prod.yaml \
    --set api.image.tag="$VERSION" \
    --set fullnameOverride=catalogizer-green \
    --wait

# Verify green deployment
if kubectl rollout status deployment/catalogizer-green-api -n "$NAMESPACE"; then
    echo "Green deployment successful, switching traffic..."
    
    # Switch ingress to green
    kubectl patch ingress catalogizer \
        --namespace "$NAMESPACE" \
        --type=json \
        -p='[{"op": "replace", "path": "/spec/rules/0/http/paths/0/backend/service/name", "value":"catalogizer-green-api"}]'
    
    # Verify traffic switch
    sleep 10
    ./scripts/smoke-tests.sh
    
    # Scale down blue
    kubectl scale deployment catalogizer-api --replicas=0 -n "$NAMESPACE"
    kubectl scale deployment catalogizer-web --replicas=0 -n "$NAMESPACE"
    
    echo "Blue-green deployment complete"
else
    echo "Green deployment failed, rolling back..."
    helm rollback catalogizer-green -n "$NAMESPACE"
    exit 1
fi
```

### 5.3 Rollback Procedure

```bash
#!/bin/bash
# rollback.sh

NAMESPACE="catalogizer-prod"
REVISION=${1:-1}

echo "Rolling back to revision $REVISION..."

# Rollback Helm release
helm rollback catalogizer "$REVISION" --namespace "$NAMESPACE"

# Verify rollback
kubectl rollout status deployment/catalogizer-api -n "$NAMESPACE"

# Run smoke tests
./scripts/smoke-tests.sh

echo "Rollback complete"
```

---

## 6. Scaling & Auto-scaling

### 6.1 Horizontal Pod Autoscaler (HPA)

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: catalogizer-api
  namespace: catalogizer-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: catalogizer-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "100"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 15
        - type: Pods
          value: 4
          periodSeconds: 15
      selectPolicy: Max
```

### 6.2 Vertical Pod Autoscaler (VPA)

```yaml
# vpa.yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: catalogizer-api
  namespace: catalogizer-prod
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: catalogizer-api
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
      - containerName: api
        minAllowed:
          cpu: 500m
          memory: 512Mi
        maxAllowed:
          cpu: 4000m
          memory: 8Gi
        controlledResources: ["cpu", "memory"]
```

### 6.3 Cluster Autoscaler

```yaml
# cluster-autoscaler.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cluster-autoscaler
  namespace: kube-system
spec:
  template:
    spec:
      containers:
        - name: cluster-autoscaler
          image: k8s.gcr.io/autoscaling/cluster-autoscaler:v1.28.0
          command:
            - ./cluster-autoscaler
            - --cloud-provider=aws
            - --namespace=catalogizer-prod
            - --nodes=3:20:worker-nodes
            - --scale-down-delay-after-add=10m
            - --scale-down-unneeded-time=10m
```

---

## 7. Monitoring

### 7.1 ServiceMonitor for Prometheus

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: catalogizer-api
  namespace: monitoring
  labels:
    release: prometheus
spec:
  namespaceSelector:
    matchNames:
      - catalogizer-prod
  selector:
    matchLabels:
      app.kubernetes.io/component: api
  endpoints:
    - port: http
      path: /metrics
      interval: 15s
      scrapeTimeout: 10s
```

### 7.2 Prometheus Rules

```yaml
# prometheus-rules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: catalogizer-alerts
  namespace: monitoring
spec:
  groups:
    - name: catalogizer
      rules:
        - alert: CatalogizerHighErrorRate
          expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "High error rate in Catalogizer API"
            
        - alert: CatalogizerHighLatency
          expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High latency in Catalogizer API"
            
        - alert: CatalogizerPodCrashLooping
          expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "Catalogizer pod is crash looping"
```

---

## 8. Troubleshooting

### 8.1 Common Commands

```bash
# Check pod status
kubectl get pods -n catalogizer-prod

# Check pod logs
kubectl logs -n catalogizer-prod deployment/catalogizer-api --tail=100 -f

# Check previous container logs (if crashed)
kubectl logs -n catalogizer-prod deployment/catalogizer-api --previous

# Exec into container
kubectl exec -it -n catalogizer-prod deployment/catalogizer-api -- /bin/sh

# Check events
kubectl get events -n catalogizer-prod --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n catalogizer-prod

# Check HPA status
kubectl get hpa -n catalogizer-prod

# Port forward for local testing
kubectl port-forward -n catalogizer-prod svc/catalogizer-api 8080:8080
```

### 8.2 Debugging Pod Issues

```bash
#!/bin/bash
# debug-pod.sh

POD_NAME=$1
NAMESPACE="catalogizer-prod"

echo "=== Debugging pod: $POD_NAME ==="

echo "\n1. Pod Status:"
kubectl describe pod "$POD_NAME" -n "$NAMESPACE" | grep -A 20 "Events:"

echo "\n2. Resource Usage:"
kubectl top pod "$POD_NAME" -n "$NAMESPACE"

echo "\n3. Recent Logs:"
kubectl logs "$POD_NAME" -n "$NAMESPACE" --tail=50

echo "\n4. Environment Variables:"
kubectl exec "$POD_NAME" -n "$NAMESPACE" -- env | grep -v PASSWORD | grep -v SECRET

echo "\n5. Network Connectivity:"
kubectl exec "$POD_NAME" -n "$NAMESPACE" -- netstat -tlnp

echo "\n=== Debug Complete ==="
```

### 8.3 Database Connection Issues

```bash
# Test database connectivity
kubectl run -it --rm debug --image=postgres:15 --restart=Never -- \
    psql postgres://username:password@postgres.database.svc.cluster.local:5432/catalogizer \
    -c "SELECT 1"

# Check database logs
kubectl logs -n catalogizer-prod -l app.kubernetes.io/name=postgresql

# Check connection pool status
kubectl exec -it deployment/catalogizer-api -n catalogizer-prod -- \
    wget -qO- http://localhost:8080/debug/pprof/heap
```

---

**Document Control:**
- Version: 1.0
- Approved by: [DevOps Lead]
- Date approved: April 6, 2026
- Next review: July 6, 2026

