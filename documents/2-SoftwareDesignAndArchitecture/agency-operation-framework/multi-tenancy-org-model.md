# Multi-Tenancy & Organization Model Specification

## Overview

This document provides a comprehensive specification for the Multi-Tenancy and Organization Model within CodeValdCortex, defining isolation mechanisms, organizational hierarchy, role-based access control (RBAC), and billing dimensions. It addresses namespace isolation, network policies, compute quotas, noisy-neighbor protections, org/BU/project structures, role matrices, approval chains, and cost visibility.

## 1. Isolation Model

### 1.1 Namespace Architecture

**Namespace Hierarchy**:

```typescript
interface Namespace {
  namespaceId: string;              // e.g., "org_acme_bu_finance_proj_001"
  name: string;                     // Human-readable name
  type: NamespaceType;
  parentNamespaceId?: string;       // For nested namespaces
  
  // Hierarchy
  organizationId: string;
  businessUnitId?: string;
  projectId?: string;
  
  // Isolation Settings
  isolation: IsolationPolicy;
  
  // Resource Limits
  quotas: ResourceQuotas;
  
  // Network
  networkPolicy: NetworkPolicy;
  
  // Metadata
  createdAt: ISODateTime;
  createdBy: string;
  labels: Record<string, string>;
  annotations: Record<string, string>;
}

type NamespaceType = 
  | "organization"     // Top-level org namespace
  | "business_unit"    // Business unit namespace
  | "project"          // Project/team namespace
  | "environment";     // Environment-specific (dev, staging, prod)

interface IsolationPolicy {
  level: "strict" | "standard" | "relaxed";
  
  // Compute Isolation
  computeIsolation: {
    dedicatedNodes: boolean;        // Dedicated compute nodes
    cpuPinning: boolean;            // Pin CPU cores
    memoryReservation: boolean;     // Reserve memory
  };
  
  // Storage Isolation
  storageIsolation: {
    dedicatedVolumes: boolean;      // Separate storage volumes
    encryptionRequired: boolean;    // Encryption at rest
    backupIsolation: boolean;       // Isolated backup storage
  };
  
  // Network Isolation
  networkIsolation: {
    vpcPerNamespace: boolean;       // Dedicated VPC/VNET
    subnetIsolation: boolean;       // Isolated subnets
    egressControl: boolean;         // Control outbound traffic
  };
  
  // Data Isolation
  dataIsolation: {
    databasePerNamespace: boolean;  // Dedicated database
    schemaIsolation: boolean;       // Schema-level isolation
    rowLevelSecurity: boolean;      // Row-level security (RLS)
  };
}
```

**Namespace Naming Convention**:

```
Format: {org}_{bu}_{project}_{environment}
Examples:
  - org_acme                           (Organization)
  - org_acme_bu_finance                (Business Unit)
  - org_acme_bu_finance_proj_risk      (Project)
  - org_acme_bu_finance_proj_risk_prod (Environment)
```

### 1.2 Resource Quotas

**Quota Configuration**:

```typescript
interface ResourceQuotas {
  // Compute Quotas
  compute: {
    maxAgents: number;                // Max concurrent agents
    maxCPUCores: number;              // Total CPU cores
    maxMemoryGB: number;              // Total memory in GB
    maxGPUs: number;                  // GPU count (for ML workloads)
  };
  
  // Storage Quotas
  storage: {
    maxStorageGB: number;             // Total storage in GB
    maxDatabaseSize: number;          // Database size in GB
    maxArtifactStorage: number;       // Artifact storage in GB
    maxBackupRetention: number;       // Backup retention in days
  };
  
  // Network Quotas
  network: {
    maxEgressGB: number;              // Egress bandwidth per month
    maxIngressGB: number;             // Ingress bandwidth per month
    maxConnections: number;           // Max concurrent connections
  };
  
  // Work Item Quotas
  workItems: {
    maxActiveWorkItems: number;       // Max active work items
    maxWorkItemsPerDay: number;       // Daily creation limit
    maxCompensationRetries: number;   // Max compensation attempts
  };
  
  // API & Rate Limits
  api: {
    requestsPerMinute: number;        // API rate limit
    requestsPerDay: number;           // Daily API quota
    webhookCallsPerDay: number;       // Webhook quota
  };
  
  // Audit & Retention
  audit: {
    maxAuditLogRetention: number;     // Audit log retention in days
    maxTraceRetention: number;        // Trace retention in days
    maxMetricsRetention: number;      // Metrics retention in days
  };
}
```

**Quota Enforcement**:

```typescript
interface QuotaEnforcement {
  namespaceId: string;
  quotas: ResourceQuotas;
  
  // Current Usage
  currentUsage: {
    compute: ComputeUsage;
    storage: StorageUsage;
    network: NetworkUsage;
    workItems: WorkItemUsage;
    api: APIUsage;
  };
  
  // Enforcement Actions
  enforcement: {
    softLimit: number;                // Warning threshold (e.g., 0.8 = 80%)
    hardLimit: number;                // Hard limit (e.g., 1.0 = 100%)
    
    onSoftLimitExceeded: "warn" | "throttle";
    onHardLimitExceeded: "block" | "queue" | "spillover";
    
    spilloverNamespace?: string;      // Overflow to another namespace
  };
  
  // Monitoring
  alerts: QuotaAlert[];
}

interface ComputeUsage {
  activeAgents: number;
  cpuCoresUsed: number;
  memoryGBUsed: number;
  gpusUsed: number;
  utilizationPercentage: number;
}

interface QuotaAlert {
  alertId: string;
  quotaType: string;
  threshold: number;                  // Percentage
  currentUsage: number;
  limit: number;
  triggeredAt: ISODateTime;
  notifiedRoles: string[];
}
```

**Quota Monitoring Dashboard**:

```yaml
quota_dashboard:
  metrics:
    - name: compute_utilization
      description: "CPU/Memory/GPU usage vs quota"
      visualization: gauge
      alert_threshold: 80%
    
    - name: storage_utilization
      description: "Storage usage vs quota"
      visualization: bar_chart
      alert_threshold: 90%
    
    - name: network_egress
      description: "Monthly egress vs quota"
      visualization: line_chart
      alert_threshold: 85%
    
    - name: api_rate_limit
      description: "API requests vs rate limit"
      visualization: time_series
      alert_threshold: 95%
  
  refresh_interval: 60s
  retention: 90d
```

### 1.3 Network Policies

**Network Isolation Rules**:

```typescript
interface NetworkPolicy {
  policyId: string;
  namespaceId: string;
  
  // Ingress Rules
  ingress: IngressRule[];
  
  // Egress Rules
  egress: EgressRule[];
  
  // Default Behavior
  defaultIngressAction: "allow" | "deny";
  defaultEgressAction: "allow" | "deny";
  
  // DNS
  dnsPolicy: DNSPolicy;
}

interface IngressRule {
  ruleId: string;
  description: string;
  
  // Source
  from: {
    namespaces?: string[];            // Allow from specific namespaces
    ipBlocks?: string[];              // Allow from IP ranges (CIDR)
    serviceAccounts?: string[];       // Allow from service accounts
  };
  
  // Ports & Protocols
  ports: {
    port: number;
    protocol: "TCP" | "UDP" | "SCTP";
  }[];
  
  // Action
  action: "allow" | "deny";
  priority: number;                   // Lower = higher priority
}

interface EgressRule {
  ruleId: string;
  description: string;
  
  // Destination
  to: {
    namespaces?: string[];            // Allow to specific namespaces
    ipBlocks?: string[];              // Allow to IP ranges (CIDR)
    domains?: string[];               // Allow to specific domains
    services?: string[];              // Allow to specific services
  };
  
  // Ports & Protocols
  ports: {
    port: number;
    protocol: "TCP" | "UDP" | "SCTP";
  }[];
  
  // Action
  action: "allow" | "deny";
  priority: number;
}

interface DNSPolicy {
  allowedDomains: string[];           // Whitelist of domains
  blockedDomains: string[];           // Blacklist of domains
  customDNSServers?: string[];        // Custom DNS servers
  dnssecRequired: boolean;            // Require DNSSEC
}
```

**Example Network Policies**:

```yaml
# Production namespace - strict egress control
network_policy_prod:
  namespace: org_acme_bu_finance_proj_risk_prod
  
  ingress:
    - rule_id: allow_internal_api
      from:
        namespaces: [org_acme_bu_finance_proj_risk_staging]
      ports:
        - port: 8080
          protocol: TCP
      action: allow
  
  egress:
    - rule_id: allow_database
      to:
        services: [postgresql.internal]
      ports:
        - port: 5432
          protocol: TCP
      action: allow
    
    - rule_id: allow_external_apis
      to:
        domains: [api.stripe.com, api.plaid.com]
      ports:
        - port: 443
          protocol: TCP
      action: allow
    
    - rule_id: deny_all_other
      to:
        ipBlocks: [0.0.0.0/0]
      action: deny
      priority: 1000
  
  dns_policy:
    allowed_domains: [*.internal, api.stripe.com, api.plaid.com]
    dnssec_required: true

# Development namespace - relaxed egress
network_policy_dev:
  namespace: org_acme_bu_finance_proj_risk_dev
  
  ingress:
    - rule_id: allow_all_internal
      from:
        namespaces: [org_acme_*]
      action: allow
  
  egress:
    - rule_id: allow_all
      to:
        ipBlocks: [0.0.0.0/0]
      action: allow
  
  dns_policy:
    allowed_domains: [*]
```

### 1.4 Noisy Neighbor Protections

**Anti-Noisy-Neighbor Strategies**:

```typescript
interface NoisyNeighborProtection {
  namespaceId: string;
  
  // Detection
  detection: {
    enabled: boolean;
    metrics: NoiseMetric[];
    anomalyThreshold: number;         // Standard deviations from mean
    detectionWindow: number;          // milliseconds
  };
  
  // Throttling
  throttling: {
    enabled: boolean;
    cpuThrottle: CPUThrottle;
    memoryThrottle: MemoryThrottle;
    networkThrottle: NetworkThrottle;
    ioThrottle: IOThrottle;
  };
  
  // Fairness Scheduling
  fairnessPolicy: FairnessPolicy;
  
  // Alerts
  alerts: NoisyNeighborAlert[];
}

interface NoiseMetric {
  metricName: string;
  threshold: number;
  action: "log" | "throttle" | "quarantine";
}

interface CPUThrottle {
  maxCPUPercent: number;              // Max CPU % per namespace
  burstAllowed: boolean;              // Allow short bursts
  burstDuration: number;              // Max burst duration (ms)
}

interface MemoryThrottle {
  maxMemoryPercent: number;           // Max memory % per namespace
  oomKillPriority: number;            // Priority for OOM killer
  swapAllowed: boolean;               // Allow swap usage
}

interface NetworkThrottle {
  maxBandwidthMbps: number;           // Max bandwidth per namespace
  priorityClass: "high" | "medium" | "low";
  burstAllowed: boolean;
}

interface IOThrottle {
  maxIOPS: number;                    // Max I/O operations per second
  maxThroughputMBps: number;          // Max throughput
  priorityClass: "high" | "medium" | "low";
}

interface FairnessPolicy {
  algorithm: "fair_share" | "weighted_fair_queuing" | "deficit_round_robin";
  weights: Record<string, number>;    // Namespace weights
  preemption: boolean;                // Allow preemption
}
```

**Noisy Neighbor Detection Algorithm**:

```
For each namespace:
  1. Collect metrics over detection window (e.g., 5 minutes)
     - CPU usage
     - Memory usage
     - Network bandwidth
     - Disk I/O
  
  2. Calculate percentiles (P50, P95, P99)
  
  3. Compare against cluster averages:
     IF namespace.P95 > cluster.mean + (threshold * stddev):
       Flag as noisy neighbor
  
  4. Determine impact:
     - CPU contention: Check other namespaces' CPU wait times
     - Memory pressure: Check swap usage, OOM events
     - Network: Check packet drops, latency increases
     - Disk: Check I/O wait times
  
  5. Take action based on severity:
     - Mild: Log and alert
     - Moderate: Apply throttling
     - Severe: Quarantine and require manual review
```

**Example Noisy Neighbor Configuration**:

```yaml
noisy_neighbor_protection:
  detection:
    enabled: true
    metrics:
      - metric: cpu_usage_p95
        threshold: 0.80              # 80% CPU
        action: throttle
      
      - metric: memory_usage_p95
        threshold: 0.90              # 90% memory
        action: throttle
      
      - metric: network_egress_mbps
        threshold: 1000              # 1 Gbps
        action: throttle
      
      - metric: disk_iops
        threshold: 10000             # 10k IOPS
        action: throttle
    
    anomaly_threshold: 2.0           # 2 standard deviations
    detection_window: 300000         # 5 minutes
  
  throttling:
    enabled: true
    cpu_throttle:
      max_cpu_percent: 70
      burst_allowed: true
      burst_duration: 60000          # 1 minute
    
    memory_throttle:
      max_memory_percent: 80
      oom_kill_priority: 10
      swap_allowed: false
    
    network_throttle:
      max_bandwidth_mbps: 500
      priority_class: medium
      burst_allowed: true
    
    io_throttle:
      max_iops: 5000
      max_throughput_mbps: 100
      priority_class: medium
  
  fairness_policy:
    algorithm: weighted_fair_queuing
    weights:
      org_acme_bu_finance_proj_risk_prod: 3
      org_acme_bu_finance_proj_risk_staging: 2
      org_acme_bu_finance_proj_risk_dev: 1
    preemption: true
```

## 2. Organization & RBAC Model

### 2.1 Organizational Hierarchy

**Organization Structure**:

```typescript
interface Organization {
  organizationId: string;             // e.g., "org_acme"
  name: string;                       // e.g., "Acme Corporation"
  domain: string;                     // e.g., "acme.com"
  
  // Hierarchy
  businessUnits: BusinessUnit[];
  
  // Settings
  settings: OrgSettings;
  
  // Billing
  billingAccount: BillingAccount;
  
  // Compliance
  complianceFrameworks: string[];     // e.g., ["SOC2", "HIPAA", "GDPR"]
  
  // Metadata
  createdAt: ISODateTime;
  owner: string;                      // User ID
}

interface BusinessUnit {
  businessUnitId: string;             // e.g., "bu_finance"
  name: string;                       // e.g., "Finance & Risk"
  organizationId: string;
  
  // Hierarchy
  projects: Project[];
  
  // Leadership
  lead: string;                       // User ID
  stakeholders: string[];             // User IDs
  
  // Settings
  inheritSettings: boolean;           // Inherit org settings
  overrideSettings?: Partial<OrgSettings>;
  
  // Budget
  budgetAllocation?: BudgetAllocation;
  
  // Metadata
  createdAt: ISODateTime;
}

interface Project {
  projectId: string;                  // e.g., "proj_risk_analysis"
  name: string;                       // e.g., "Financial Risk Analysis"
  businessUnitId: string;
  
  // Environments
  environments: Environment[];
  
  // Team
  owner: string;                      // User ID
  members: ProjectMember[];
  
  // Settings
  inheritSettings: boolean;
  overrideSettings?: Partial<OrgSettings>;
  
  // Budget
  budgetAllocation?: BudgetAllocation;
  
  // Metadata
  createdAt: ISODateTime;
  status: "active" | "archived" | "suspended";
}

interface Environment {
  environmentId: string;              // e.g., "env_prod"
  name: string;                       // e.g., "Production"
  type: "development" | "staging" | "production" | "sandbox";
  projectId: string;
  
  // Namespace
  namespaceId: string;
  
  // Protection
  protected: boolean;                 // Require approvals for changes
  approvalChain?: ApprovalChain;
  
  // Metadata
  createdAt: ISODateTime;
}

interface OrgSettings {
  // Security
  ssoRequired: boolean;
  mfaRequired: boolean;
  ipWhitelist?: string[];
  
  // Data Governance
  dataResidency: string[];            // e.g., ["US", "EU"]
  dataClassification: "PUBLIC" | "INTERNAL" | "CONFIDENTIAL" | "RESTRICTED";
  
  // Audit
  auditLogRetention: number;          // Days
  
  // Defaults
  defaultQuotas: ResourceQuotas;
  defaultNetworkPolicy: NetworkPolicy;
}

interface ProjectMember {
  userId: string;
  role: Role;
  joinedAt: ISODateTime;
}
```

**Hierarchy Visualization**:

```
Organization (Acme Corporation)
├── Business Unit (Finance & Risk)
│   ├── Project (Financial Risk Analysis)
│   │   ├── Environment (Production)
│   │   ├── Environment (Staging)
│   │   └── Environment (Development)
│   └── Project (Fraud Detection)
│       ├── Environment (Production)
│       └── Environment (Development)
├── Business Unit (Operations)
│   └── Project (Logistics Optimization)
│       ├── Environment (Production)
│       └── Environment (Development)
└── Business Unit (Marketing)
    └── Project (Customer Analytics)
        └── Environment (Production)
```

### 2.2 Role Matrix

**Role Definitions**:

```typescript
interface Role {
  roleId: string;
  name: string;
  description: string;
  permissions: Permission[];
  scope: RoleScope;
}

type RoleScope = 
  | "organization"    // Org-wide permissions
  | "business_unit"   // BU-level permissions
  | "project"         // Project-level permissions
  | "environment";    // Environment-level permissions

interface Permission {
  resource: ResourceType;
  actions: Action[];
  conditions?: Condition[];
}

type ResourceType = 
  | "agents"
  | "work_items"
  | "goals"
  | "namespaces"
  | "projects"
  | "business_units"
  | "organizations"
  | "users"
  | "roles"
  | "billing"
  | "audit_logs"
  | "network_policies"
  | "quotas";

type Action = 
  | "create"
  | "read"
  | "update"
  | "delete"
  | "execute"
  | "approve"
  | "audit";

interface Condition {
  field: string;
  operator: "equals" | "not_equals" | "in" | "not_in" | "matches";
  value: any;
}
```

**Standard Roles**:

| Role | Scope | Description | Key Permissions |
|------|-------|-------------|-----------------|
| **Organization Owner** | Organization | Full control over org | All permissions on all resources |
| **Business Unit Lead** | Business Unit | Manage BU and projects | Create/update/delete projects, manage BU members, view BU billing |
| **Project Owner** | Project | Full control over project | Create/update/delete environments, manage project members, configure quotas |
| **Developer** | Project/Environment | Build and deploy agents | Create/update agents, work items, goals; execute work items |
| **Operator** | Project/Environment | Operate and monitor | Read agents, work items, goals; execute operations; view metrics |
| **Auditor** | Organization/BU/Project | Review and audit | Read-only access to all resources, audit logs, compliance reports |
| **Risk Manager** | Organization/BU | Manage risk policies | Create/update risk policies, view all risk assessments, audit access |
| **Viewer** | Organization/BU/Project | Read-only access | Read all resources within scope |

**Detailed Role Permissions Matrix**:

```typescript
const rolePermissions: Record<string, Role> = {
  organization_owner: {
    roleId: "org_owner",
    name: "Organization Owner",
    description: "Full administrative control over the organization",
    scope: "organization",
    permissions: [
      { resource: "*", actions: ["*"] }  // All permissions
    ]
  },
  
  business_unit_lead: {
    roleId: "bu_lead",
    name: "Business Unit Lead",
    description: "Manage business unit and projects",
    scope: "business_unit",
    permissions: [
      { resource: "business_units", actions: ["read", "update"] },
      { resource: "projects", actions: ["create", "read", "update", "delete"] },
      { resource: "users", actions: ["read", "create", "update"] },
      { resource: "billing", actions: ["read"] },
      { resource: "audit_logs", actions: ["read"] },
      { resource: "quotas", actions: ["read", "update"] }
    ]
  },
  
  project_owner: {
    roleId: "proj_owner",
    name: "Project Owner",
    description: "Full control over project and environments",
    scope: "project",
    permissions: [
      { resource: "projects", actions: ["read", "update"] },
      { resource: "environments", actions: ["create", "read", "update", "delete"] },
      { resource: "agents", actions: ["create", "read", "update", "delete", "execute"] },
      { resource: "work_items", actions: ["create", "read", "update", "delete", "approve"] },
      { resource: "goals", actions: ["create", "read", "update", "delete", "approve"] },
      { resource: "users", actions: ["read", "create", "update"] },
      { resource: "network_policies", actions: ["create", "read", "update", "delete"] },
      { resource: "quotas", actions: ["read", "update"] },
      { resource: "billing", actions: ["read"] }
    ]
  },
  
  developer: {
    roleId: "developer",
    name: "Developer",
    description: "Build and deploy agents and work items",
    scope: "project",
    permissions: [
      { resource: "agents", actions: ["create", "read", "update", "execute"] },
      { resource: "work_items", actions: ["create", "read", "update", "execute"] },
      { resource: "goals", actions: ["create", "read", "update"] },
      { resource: "environments", actions: ["read"] },
      { resource: "audit_logs", actions: ["read"], 
        conditions: [{ field: "userId", operator: "equals", value: "$current_user" }] }
    ]
  },
  
  operator: {
    roleId: "operator",
    name: "Operator",
    description: "Operate and monitor agents",
    scope: "environment",
    permissions: [
      { resource: "agents", actions: ["read", "execute"] },
      { resource: "work_items", actions: ["read", "execute"] },
      { resource: "goals", actions: ["read"] },
      { resource: "environments", actions: ["read"] },
      { resource: "audit_logs", actions: ["read"] }
    ]
  },
  
  auditor: {
    roleId: "auditor",
    name: "Auditor",
    description: "Review and audit all activities",
    scope: "organization",
    permissions: [
      { resource: "*", actions: ["read", "audit"] }
    ]
  },
  
  risk_manager: {
    roleId: "risk_manager",
    name: "Risk Manager",
    description: "Manage risk policies and assessments",
    scope: "organization",
    permissions: [
      { resource: "goals", actions: ["read", "update", "approve"] },
      { resource: "work_items", actions: ["read", "approve"] },
      { resource: "agents", actions: ["read"] },
      { resource: "audit_logs", actions: ["read", "audit"] },
      { resource: "network_policies", actions: ["read", "update"] },
      { resource: "quotas", actions: ["read"] }
    ]
  },
  
  viewer: {
    roleId: "viewer",
    name: "Viewer",
    description: "Read-only access to resources",
    scope: "project",
    permissions: [
      { resource: "*", actions: ["read"] }
    ]
  }
};
```

### 2.3 Approval Chains

**Approval Chain Configuration**:

```typescript
interface ApprovalChain {
  chainId: string;
  name: string;
  description: string;
  
  // Trigger Conditions
  triggers: ApprovalTrigger[];
  
  // Approval Steps
  steps: ApprovalStep[];
  
  // Policy
  policy: ApprovalPolicy;
  
  // Timeouts & Escalation
  timeouts: ApprovalTimeout[];
}

interface ApprovalTrigger {
  resourceType: ResourceType;
  action: Action;
  conditions?: Condition[];
}

interface ApprovalStep {
  stepId: string;
  name: string;
  order: number;                      // Execution order
  
  // Approvers
  approvers: ApproverConfig;
  
  // Requirements
  requirements: {
    minimumApprovals: number;         // e.g., 1 of 3 approvers
    unanimousRequired: boolean;       // All must approve
    vetoRights: string[];             // Roles with veto power
  };
  
  // Actions
  onApprove: string[];                // Actions to execute on approval
  onReject: string[];                 // Actions to execute on rejection
  
  // Timeout
  timeout: number;                    // milliseconds
  onTimeout: "auto_approve" | "auto_reject" | "escalate";
}

interface ApproverConfig {
  type: "roles" | "users" | "dynamic";
  
  // Static approvers
  roles?: string[];
  users?: string[];
  
  // Dynamic approvers (computed at runtime)
  dynamicRule?: string;               // e.g., "resource.owner" or "project.lead"
}

interface ApprovalPolicy {
  parallelApprovals: boolean;         // Allow parallel approvals
  allowSelfApproval: boolean;         // Requester can approve
  requireComments: boolean;           // Comments required
  auditRequired: boolean;             // Full audit trail
}

interface ApprovalTimeout {
  stepId: string;
  duration: number;                   // milliseconds
  action: "escalate" | "auto_approve" | "auto_reject";
  escalateTo?: string[];              // Roles to escalate to
}
```

**Standard Approval Chains**:

```yaml
# Production deployment approval
approval_chain_prod_deploy:
  name: "Production Deployment Approval"
  triggers:
    - resource_type: agents
      action: create
      conditions:
        - field: environment.type
          operator: equals
          value: production
  
  steps:
    - step_id: tech_lead_approval
      name: "Technical Lead Review"
      order: 1
      approvers:
        type: roles
        roles: [project_owner, technical_lead]
      requirements:
        minimum_approvals: 1
        unanimous_required: false
      timeout: 3600000              # 1 hour
      on_timeout: escalate
    
    - step_id: security_review
      name: "Security Review"
      order: 2
      approvers:
        type: roles
        roles: [security_engineer, risk_manager]
      requirements:
        minimum_approvals: 1
        veto_rights: [risk_manager]
      timeout: 7200000              # 2 hours
      on_timeout: escalate
    
    - step_id: final_approval
      name: "Final Approval"
      order: 3
      approvers:
        type: roles
        roles: [bu_lead, org_owner]
      requirements:
        minimum_approvals: 1
      timeout: 14400000             # 4 hours
      on_timeout: auto_reject
  
  policy:
    parallel_approvals: false
    allow_self_approval: false
    require_comments: true
    audit_required: true

# High-risk work item approval
approval_chain_high_risk_work:
  name: "High-Risk Work Item Approval"
  triggers:
    - resource_type: work_items
      action: execute
      conditions:
        - field: riskLevel
          operator: in
          value: [HIGH, CRITICAL]
  
  steps:
    - step_id: risk_assessment
      name: "Risk Manager Assessment"
      order: 1
      approvers:
        type: roles
        roles: [risk_manager]
      requirements:
        minimum_approvals: 1
        veto_rights: [risk_manager]
      timeout: 1800000              # 30 minutes
      on_timeout: auto_reject
    
    - step_id: executive_approval
      name: "Executive Approval"
      order: 2
      approvers:
        type: roles
        roles: [bu_lead, org_owner]
      requirements:
        minimum_approvals: 1
      timeout: 3600000              # 1 hour
      on_timeout: escalate
  
  policy:
    parallel_approvals: false
    allow_self_approval: false
    require_comments: true
    audit_required: true

# Quota increase approval
approval_chain_quota_increase:
  name: "Quota Increase Approval"
  triggers:
    - resource_type: quotas
      action: update
      conditions:
        - field: increase_percentage
          operator: ">="
          value: 0.2               # 20% increase
  
  steps:
    - step_id: project_owner_approval
      name: "Project Owner Justification"
      order: 1
      approvers:
        type: dynamic
        dynamic_rule: "project.owner"
      requirements:
        minimum_approvals: 1
        require_justification: true
      timeout: 3600000
      on_timeout: auto_reject
    
    - step_id: bu_lead_approval
      name: "Business Unit Lead Approval"
      order: 2
      approvers:
        type: roles
        roles: [bu_lead]
      requirements:
        minimum_approvals: 1
      timeout: 7200000
      on_timeout: escalate
  
  policy:
    parallel_approvals: false
    allow_self_approval: false
    require_comments: true
    audit_required: true
```

## 3. Billing Dimensions

### 3.1 Metering & Cost Tracking

**Billing Dimensions**:

```typescript
interface BillingDimensions {
  organizationId: string;
  billingPeriod: BillingPeriod;
  
  // Compute Costs
  compute: {
    agentHours: AgentHourMetrics;
    cpuHours: ResourceHourMetrics;
    memoryGBHours: ResourceHourMetrics;
    gpuHours: ResourceHourMetrics;
  };
  
  // Storage Costs
  storage: {
    databaseGB: StorageMetrics;
    artifactStorageGB: StorageMetrics;
    backupStorageGB: StorageMetrics;
    archiveStorageGB: StorageMetrics;
  };
  
  // Network Costs
  network: {
    egressGB: NetworkMetrics;
    ingressGB: NetworkMetrics;
    dataTransferGB: NetworkMetrics;
  };
  
  // Audit & Retention
  audit: {
    auditLogGB: StorageMetrics;
    traceStorageGB: StorageMetrics;
    metricsStorageGB: StorageMetrics;
    retentionDays: number;
  };
  
  // Support & Services
  support: {
    supportPlan: "community" | "standard" | "premium" | "enterprise";
    incidentCount: number;
    prioritySupportHours: number;
  };
  
  // Total Costs
  totalCost: Cost;
}

interface BillingPeriod {
  startDate: ISODateTime;
  endDate: ISODateTime;
  period: "daily" | "weekly" | "monthly" | "quarterly" | "annual";
}

interface AgentHourMetrics {
  totalAgentHours: number;
  breakdown: {
    agentType: string;
    hours: number;
    cost: number;
    costPerHour: number;
  }[];
}

interface ResourceHourMetrics {
  totalHours: number;
  peakUsage: number;
  averageUsage: number;
  cost: number;
  costPerUnit: number;
}

interface StorageMetrics {
  totalGB: number;
  peakGB: number;
  averageGB: number;
  cost: number;
  costPerGB: number;
}

interface NetworkMetrics {
  totalGB: number;
  peakMbps: number;
  averageGBPerDay: number;
  cost: number;
  costPerGB: number;
}

interface Cost {
  amount: number;
  currency: string;
  breakdown: CostBreakdown[];
}

interface CostBreakdown {
  category: string;
  amount: number;
  percentage: number;
}
```

**Cost Calculation Formulas**:

```typescript
// Agent Hours Cost
const calculateAgentHoursCost = (metrics: AgentHourMetrics): number => {
  return metrics.breakdown.reduce((total, item) => {
    return total + (item.hours * item.costPerHour);
  }, 0);
};

// Storage Cost (tiered pricing)
const calculateStorageCost = (gb: number, tierPricing: TierPricing[]): number => {
  let cost = 0;
  let remainingGB = gb;
  
  for (const tier of tierPricing) {
    if (remainingGB <= 0) break;
    
    const gbInTier = Math.min(remainingGB, tier.maxGB - tier.minGB);
    cost += gbInTier * tier.pricePerGB;
    remainingGB -= gbInTier;
  }
  
  return cost;
};

// Network Egress Cost (distance-based)
const calculateEgressCost = (
  egressGB: number, 
  sourceRegion: string, 
  destRegion: string,
  pricingTable: PricingTable
): number => {
  const route = `${sourceRegion}->${destRegion}`;
  const pricePerGB = pricingTable[route] || pricingTable.default;
  return egressGB * pricePerGB;
};
```

### 3.2 Pricing Models (OSS & Commercial)

**Open Source Pricing** (Cost Visibility):

```yaml
# OSS users see costs in "CodeVald Credits" (virtual currency)
oss_pricing_model:
  purpose: "Make resource costs visible without actual billing"
  
  compute:
    agent_hour:
      base: 10 credits/hour
      by_type:
        stateless_tool_caller: 10 credits/hour
        planner_coordinator: 20 credits/hour
        data_access_agent: 15 credits/hour
        long_running_service: 30 credits/hour
    
    cpu_hour: 5 credits/core-hour
    memory_gb_hour: 2 credits/GB-hour
    gpu_hour: 100 credits/GPU-hour
  
  storage:
    database_gb_month: 1 credit/GB-month
    artifact_storage_gb_month: 0.5 credit/GB-month
    backup_storage_gb_month: 0.25 credit/GB-month
  
  network:
    egress_gb: 1 credit/GB
    ingress_gb: 0 credits/GB      # Free
  
  audit:
    audit_log_gb_month: 2 credits/GB-month
    retention_per_day: 0.1 credit/GB-day
  
  monthly_quota:
    free_tier: 10000 credits/month
    alert_threshold: 8000 credits  # 80% of quota

commercial_pricing_model:
  purpose: "Actual billing for managed/cloud deployments"
  
  compute:
    agent_hour:
      base: $0.10/hour
      by_type:
        stateless_tool_caller: $0.10/hour
        planner_coordinator: $0.20/hour
        data_access_agent: $0.15/hour
        long_running_service: $0.30/hour
    
    cpu_hour: $0.05/core-hour
    memory_gb_hour: $0.02/GB-hour
    gpu_hour: $1.00/GPU-hour
  
  storage:
    database_gb_month: $0.10/GB-month
    artifact_storage_gb_month: $0.05/GB-month
    backup_storage_gb_month: $0.025/GB-month
    
    tiered_pricing:
      - tier: 1-100GB
        price: $0.10/GB-month
      - tier: 101-1000GB
        price: $0.08/GB-month
      - tier: 1001GB+
        price: $0.05/GB-month
  
  network:
    egress_gb:
      same_region: $0.01/GB
      cross_region: $0.05/GB
      internet: $0.09/GB
    ingress_gb: $0.00/GB
  
  audit:
    audit_log_gb_month: $0.20/GB-month
    retention_per_day: $0.01/GB-day
  
  support:
    community: $0/month
    standard: $100/month
    premium: $500/month
    enterprise: custom
```

### 3.3 Cost Allocation & Chargebacks

**Cost Allocation Tags**:

```typescript
interface CostAllocation {
  organizationId: string;
  billingPeriod: BillingPeriod;
  
  // Allocation by hierarchy
  byBusinessUnit: BusinessUnitCosts[];
  byProject: ProjectCosts[];
  byEnvironment: EnvironmentCosts[];
  
  // Allocation by tags
  byTags: TagCosts[];
  
  // Shared costs
  sharedCosts: SharedCostAllocation;
}

interface BusinessUnitCosts {
  businessUnitId: string;
  name: string;
  totalCost: number;
  breakdown: CostBreakdown[];
  percentage: number;                 // % of org total
}

interface ProjectCosts {
  projectId: string;
  name: string;
  businessUnitId: string;
  totalCost: number;
  breakdown: CostBreakdown[];
  budgetAllocation?: number;
  budgetRemaining?: number;
  onTrack: boolean;
}

interface EnvironmentCosts {
  environmentId: string;
  name: string;
  type: "development" | "staging" | "production";
  projectId: string;
  totalCost: number;
  breakdown: CostBreakdown[];
}

interface TagCosts {
  tagKey: string;
  tagValue: string;
  totalCost: number;
  breakdown: CostBreakdown[];
}

interface SharedCostAllocation {
  totalSharedCosts: number;
  allocationMethod: "equal" | "proportional" | "weighted";
  
  // Proportional allocation (based on usage)
  proportionalAllocation?: {
    basedOn: "compute" | "storage" | "combined";
    allocations: {
      targetId: string;
      percentage: number;
      amount: number;
    }[];
  };
  
  // Weighted allocation (custom weights)
  weightedAllocation?: {
    weights: Record<string, number>;
    allocations: {
      targetId: string;
      weight: number;
      amount: number;
    }[];
  };
}
```

**Chargeback Reports**:

```typescript
interface ChargebackReport {
  reportId: string;
  organizationId: string;
  billingPeriod: BillingPeriod;
  generatedAt: ISODateTime;
  
  // Executive Summary
  summary: {
    totalCost: number;
    costByBU: BusinessUnitCosts[];
    topCostDrivers: CostDriver[];
    trends: CostTrend[];
  };
  
  // Detailed Breakdown
  details: {
    byBusinessUnit: BusinessUnitCosts[];
    byProject: ProjectCosts[];
    byEnvironment: EnvironmentCosts[];
    byResourceType: ResourceTypeCosts[];
  };
  
  // Recommendations
  recommendations: CostOptimizationRecommendation[];
  
  // Export
  exports: {
    csv: string;                      // CSV export URL
    pdf: string;                      // PDF export URL
    api: string;                      // API endpoint
  };
}

interface CostDriver {
  category: string;
  amount: number;
  percentage: number;
  trend: "increasing" | "stable" | "decreasing";
}

interface CostTrend {
  period: string;
  cost: number;
  changePercentage: number;
}

interface ResourceTypeCosts {
  resourceType: string;
  totalCost: number;
  breakdown: CostBreakdown[];
}

interface CostOptimizationRecommendation {
  recommendationId: string;
  category: "compute" | "storage" | "network" | "audit";
  description: string;
  estimatedSavings: number;
  effort: "low" | "medium" | "high";
  priority: "low" | "medium" | "high";
  actions: string[];
}
```

### 3.4 Budgets & Alerts

**Budget Configuration**:

```typescript
interface Budget {
  budgetId: string;
  name: string;
  scope: BudgetScope;
  
  // Budget Amount
  amount: number;
  currency: string;
  period: "daily" | "weekly" | "monthly" | "quarterly" | "annual";
  
  // Alerts
  alerts: BudgetAlert[];
  
  // Actions
  actions: BudgetAction[];
  
  // Tracking
  currentSpend: number;
  forecastedSpend: number;
  remainingBudget: number;
  burnRate: number;                   // Per day
  
  // Status
  status: "on_track" | "at_risk" | "exceeded";
  
  // Metadata
  owner: string;
  createdAt: ISODateTime;
  updatedAt: ISODateTime;
}

interface BudgetScope {
  type: "organization" | "business_unit" | "project" | "environment";
  targetId: string;
}

interface BudgetAlert {
  alertId: string;
  threshold: number;                  // Percentage (e.g., 0.8 = 80%)
  notifyRoles: string[];
  notifyUsers: string[];
  channels: ("email" | "slack" | "webhook")[];
  triggered: boolean;
  triggeredAt?: ISODateTime;
}

interface BudgetAction {
  actionId: string;
  trigger: number;                    // Percentage threshold
  action: "notify" | "throttle" | "block" | "shutdown";
  description: string;
  executed: boolean;
  executedAt?: ISODateTime;
}
```

**Example Budget Configuration**:

```yaml
budgets:
  - budget_id: budget_finance_bu_monthly
    name: "Finance BU Monthly Budget"
    scope:
      type: business_unit
      target_id: bu_finance
    
    amount: 10000
    currency: USD
    period: monthly
    
    alerts:
      - threshold: 0.5              # 50%
        notify_roles: [bu_lead]
        channels: [email]
      
      - threshold: 0.8              # 80%
        notify_roles: [bu_lead, project_owner]
        channels: [email, slack]
      
      - threshold: 0.95             # 95%
        notify_roles: [bu_lead, org_owner]
        channels: [email, slack, webhook]
    
    actions:
      - trigger: 1.0                # 100%
        action: throttle
        description: "Throttle non-critical workloads"
      
      - trigger: 1.2                # 120%
        action: block
        description: "Block new work item creation"

  - budget_id: budget_risk_proj_monthly
    name: "Risk Analysis Project Monthly Budget"
    scope:
      type: project
      target_id: proj_risk_analysis
    
    amount: 3000
    currency: USD
    period: monthly
    
    alerts:
      - threshold: 0.75
        notify_roles: [project_owner]
        channels: [email]
      
      - threshold: 0.90
        notify_roles: [project_owner, bu_lead]
        channels: [email, slack]
    
    actions:
      - trigger: 1.0
        action: notify
        description: "Budget exceeded - review required"
```

## 4. Implementation Roadmap

### Phase 1: Namespace Isolation (MVP-038)
- [ ] Implement namespace creation and hierarchy
- [ ] Build resource quota enforcement system
- [ ] Create network policy engine
- [ ] Implement noisy neighbor detection and throttling

### Phase 2: Organization & RBAC (MVP-039)
- [ ] Build org/BU/project hierarchy management
- [ ] Implement role matrix and permission system
- [ ] Create approval chain engine
- [ ] Build user management and SSO integration

### Phase 3: Billing & Metering (MVP-040)
- [ ] Implement metering for all billing dimensions
- [ ] Build cost allocation and chargeback system
- [ ] Create budget tracking and alerting
- [ ] Build cost optimization recommendations engine

### Phase 4: Multi-tenancy Hardening (MVP-041)
- [ ] Add advanced isolation (dedicated nodes, encryption)
- [ ] Implement data residency controls
- [ ] Build tenant migration tools
- [ ] Create compliance reporting dashboards

## 5. Success Metrics

### Technical Metrics
- **Namespace Isolation**: 100% of resources isolated by namespace
- **Quota Enforcement**: <1% quota violations
- **Network Policy Violations**: 0 unauthorized cross-namespace access
- **Noisy Neighbor Incidents**: <0.1% of namespaces flagged per month

### Operational Metrics
- **RBAC Coverage**: 100% of resources protected by RBAC
- **Approval Chain Latency**: <2 hours for P95 approvals
- **Cost Visibility**: 100% of costs allocated to owners
- **Budget Adherence**: >90% of projects within budget

### Business Metrics
- **Multi-tenancy Efficiency**: >80% resource utilization across tenants
- **Cost Predictability**: <10% variance from forecasted spend
- **Tenant Satisfaction**: >4.5/5 satisfaction score
- **Billing Accuracy**: >99.9% accuracy in cost allocation

---

**Document Version**: 1.0.0  
**Last Updated**: 2024-10-30  
**Owner**: CodeValdCortex Architecture Team  
**Status**: Draft - Pending Review
