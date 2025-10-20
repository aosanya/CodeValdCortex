# CodeValdCortex - Kubernetes Infrastructure

## Overview

CodeValdCortex's Kubernetes infrastructure provides the foundational container orchestration platform for enterprise-grade multi-agent AI systems. The infrastructure is designed for high availability, scalability, and enterprise security requirements.

## 1. Cluster Architecture

### Production Cluster Configuration

```yaml
# Kubernetes Cluster Specification
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-config
data:
  cluster.yaml: |
    cluster:
      name: pweza-core-production
      version: "1.28+"
      nodes:
        control-plane:
          count: 3
          instance-type: c5.xlarge
          availability-zones: ["us-east-1a", "us-east-1b", "us-east-1c"]
        worker-nodes:
          count: 6
          instance-type: c5.2xlarge
          auto-scaling:
            min: 3
            max: 20
            target-cpu: 70%
        storage:
          type: gp3
          size: 100Gi
          iops: 3000
      networking:
        cni: calico
        service-mesh: istio
        ingress: nginx-ingress
      addons:
        - metrics-server
        - cluster-autoscaler
        - aws-load-balancer-controller
        - external-dns
```

### Namespace Organization

```yaml
# Namespace Configuration
---
apiVersion: v1
kind: Namespace
metadata:
  name: pweza-core-system
  labels:
    name: pweza-core-system
    istio-injection: enabled
---
apiVersion: v1
kind: Namespace
metadata:
  name: pweza-core-agents
  labels:
    name: pweza-core-agents
    istio-injection: enabled
---
apiVersion: v1
kind: Namespace
metadata:
  name: pweza-core-data
  labels:
    name: pweza-core-data
    istio-injection: enabled
---
apiVersion: v1
kind: Namespace
metadata:
  name: pweza-core-monitoring
  labels:
    name: pweza-core-monitoring
    istio-injection: enabled
```

## 2. Helm Chart Structure

### Chart Organization

```
pweza-core-helm/
├── Chart.yaml
├── values.yaml
├── values-production.yaml
├── values-staging.yaml
├── templates/
│   ├── core-manager/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap.yaml
│   │   └── hpa.yaml
│   ├── agent-pools/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── scaling-policy.yaml
│   ├── database/
│   │   ├── arangodb-cluster.yaml
│   │   ├── storage.yaml
│   │   └── backup-policy.yaml
│   ├── monitoring/
│   │   ├── prometheus.yaml
│   │   ├── grafana.yaml
│   │   └── alertmanager.yaml
│   ├── networking/
│   │   ├── ingress.yaml
│   │   ├── virtual-service.yaml
│   │   └── destination-rule.yaml
│   └── security/
│       ├── rbac.yaml
│       ├── network-policy.yaml
│       └── pod-security-policy.yaml
├── crds/
│   ├── agent-pool-crd.yaml
│   ├── workflow-crd.yaml
│   └── agent-template-crd.yaml
└── tests/
    ├── agent-test.yaml
    ├── workflow-test.yaml
    └── integration-test.yaml
```

### Core Manager Deployment

```yaml
# Core Manager Deployment Template
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "pweza-core.fullname" . }}-manager
  namespace: {{ .Values.namespace.system }}
spec:
  replicas: {{ .Values.manager.replicas }}
  selector:
    matchLabels:
      app: pweza-core-manager
  template:
    metadata:
      labels:
        app: pweza-core-manager
        version: {{ .Chart.AppVersion }}
    spec:
      serviceAccountName: pweza-core-manager
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 2000
      containers:
      - name: manager
        image: "{{ .Values.manager.image.repository }}:{{ .Values.manager.image.tag }}"
        imagePullPolicy: {{ .Values.manager.image.pullPolicy }}
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        - name: grpc
          containerPort: 9090
          protocol: TCP
        - name: metrics
          containerPort: 2112
          protocol: TCP
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: pweza-core-secrets
              key: database-url
        - name: LOG_LEVEL
          value: {{ .Values.manager.logLevel }}
        - name: METRICS_ENABLED
          value: "true"
        resources:
          requests:
            memory: {{ .Values.manager.resources.requests.memory }}
            cpu: {{ .Values.manager.resources.requests.cpu }}
          limits:
            memory: {{ .Values.manager.resources.limits.memory }}
            cpu: {{ .Values.manager.resources.limits.cpu }}
        livenessProbe:
          httpGet:
            path: /health/live
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /etc/pweza-core
          readOnly: true
        - name: tls-certs
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: pweza-core-config
      - name: tls-certs
        secret:
          secretName: pweza-core-tls
```

## 3. Resource Management

### Horizontal Pod Autoscaler (HPA)

```yaml
# HPA Configuration for Core Manager
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: pweza-core-manager-hpa
  namespace: pweza-core-system
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: pweza-core-manager
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
        name: agent_coordination_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

### Vertical Pod Autoscaler (VPA)

```yaml
# VPA Configuration for Resource Optimization
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: pweza-core-manager-vpa
  namespace: pweza-core-system
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: pweza-core-manager
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: manager
      minAllowed:
        cpu: 100m
        memory: 256Mi
      maxAllowed:
        cpu: 2000m
        memory: 4Gi
      controlledResources: ["cpu", "memory"]
```

## 4. Security Configuration

### RBAC Configuration

```yaml
# Role-Based Access Control
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pweza-core-manager
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["pweza.ai"]
  resources: ["agents", "workflows", "agenttemplates"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pweza-core-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pweza-core-manager
subjects:
- kind: ServiceAccount
  name: pweza-core-manager
  namespace: pweza-core-system
```

### Network Policies

```yaml
# Network Policy for Core System
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: pweza-core-system-policy
  namespace: pweza-core-system
spec:
  podSelector:
    matchLabels:
      app: pweza-core-manager
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: pweza-core-agents
    - namespaceSelector:
        matchLabels:
          name: istio-system
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 9090
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: pweza-core-data
    ports:
    - protocol: TCP
      port: 8529
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

## 5. Custom Resource Definitions

### Agent Pool CRD

```yaml
# Custom Resource Definition for Agent Pools
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: agentpools.pweza.ai
spec:
  group: pweza.ai
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              template:
                type: object
                properties:
                  image:
                    type: string
                  resources:
                    type: object
                  environment:
                    type: array
                    items:
                      type: object
              scaling:
                type: object
                properties:
                  minReplicas:
                    type: integer
                    minimum: 1
                  maxReplicas:
                    type: integer
                    maximum: 100
                  targetUtilization:
                    type: integer
                    minimum: 1
                    maximum: 100
          status:
            type: object
            properties:
              replicas:
                type: integer
              readyReplicas:
                type: integer
              conditions:
                type: array
                items:
                  type: object
  scope: Namespaced
  names:
    plural: agentpools
    singular: agentpool
    kind: AgentPool
    shortNames:
    - ap
```

## 6. Monitoring Integration

### ServiceMonitor Configuration

```yaml
# Prometheus ServiceMonitor for Metrics Collection
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pweza-core-manager
  namespace: pweza-core-system
  labels:
    app: pweza-core-manager
spec:
  selector:
    matchLabels:
      app: pweza-core-manager
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    honorLabels: true
```

## 7. Deployment Commands

### Installation

```bash
# Add Helm repository
helm repo add pweza-core https://charts.pweza.ai
helm repo update

# Install with production values
helm install pweza-core pweza-core/pweza-core \
  --namespace pweza-core-system \
  --create-namespace \
  --values values-production.yaml \
  --wait

# Verify deployment
kubectl get pods -n pweza-core-system
kubectl get services -n pweza-core-system
```

### Upgrade

```bash
# Upgrade to new version
helm upgrade pweza-core pweza-core/pweza-core \
  --namespace pweza-core-system \
  --values values-production.yaml \
  --wait

# Rollback if needed
helm rollback pweza-core 1 --namespace pweza-core-system
```

This Kubernetes infrastructure provides a robust, scalable foundation for CodeValdCortex's multi-agent orchestration platform with enterprise-grade security and monitoring capabilities.