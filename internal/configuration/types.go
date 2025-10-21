package configuration

import (
	"encoding/json"
	"time"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// AgentConfiguration represents a complete agent configuration
type AgentConfiguration struct {
	// ID is the unique identifier for this configuration
	ID string `json:"id" yaml:"id"`

	// Version is the configuration version for tracking changes
	Version string `json:"version" yaml:"version"`

	// Name is a human-readable name for this configuration
	Name string `json:"name" yaml:"name"`

	// Description provides context about this configuration
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// AgentType specifies the type of agent this config is for
	AgentType string `json:"agent_type" yaml:"agent_type"`

	// BaseConfig contains the core agent configuration
	BaseConfig agent.Config `json:"base_config" yaml:"base_config"`

	// RuntimeConfig contains runtime-specific settings
	RuntimeConfig RuntimeConfiguration `json:"runtime_config" yaml:"runtime_config"`

	// DeploymentConfig contains deployment-specific settings
	DeploymentConfig DeploymentConfiguration `json:"deployment_config" yaml:"deployment_config"`

	// EnvironmentVariables for the agent runtime
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty" yaml:"environment_variables,omitempty"`

	// Labels for categorization and selection
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Annotations for additional metadata
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// CreatedAt timestamp
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`

	// UpdatedAt timestamp
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`

	// CreatedBy tracks who created this configuration
	CreatedBy string `json:"created_by,omitempty" yaml:"created_by,omitempty"`
}

// RuntimeConfiguration defines runtime behavior settings
type RuntimeConfiguration struct {
	// AutoRestart determines if agent should restart on failure
	AutoRestart bool `json:"auto_restart" yaml:"auto_restart"`

	// RestartPolicy defines restart behavior
	RestartPolicy RestartPolicy `json:"restart_policy" yaml:"restart_policy"`

	// HealthCheck configuration
	HealthCheck HealthCheckConfig `json:"health_check" yaml:"health_check"`

	// Logging configuration
	Logging LoggingConfig `json:"logging" yaml:"logging"`

	// Metrics configuration
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`

	// Communication settings
	Communication CommunicationConfig `json:"communication" yaml:"communication"`

	// Memory management settings
	Memory MemoryConfig `json:"memory" yaml:"memory"`

	// Security settings
	Security SecurityConfig `json:"security" yaml:"security"`
}

// DeploymentConfiguration defines deployment-specific settings
type DeploymentConfiguration struct {
	// Strategy defines how the agent should be deployed
	Strategy DeploymentStrategy `json:"strategy" yaml:"strategy"`

	// Replicas defines the number of agent instances
	Replicas int `json:"replicas" yaml:"replicas"`

	// Rolling update configuration
	RollingUpdate RollingUpdateConfig `json:"rolling_update,omitempty" yaml:"rolling_update,omitempty"`

	// Resource requirements
	Resources ResourceRequirements `json:"resources" yaml:"resources"`

	// Node selector for placement
	NodeSelector map[string]string `json:"node_selector,omitempty" yaml:"node_selector,omitempty"`

	// Tolerations for node taints
	Tolerations []Toleration `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`

	// Affinity rules
	Affinity *Affinity `json:"affinity,omitempty" yaml:"affinity,omitempty"`
}

// RestartPolicy defines how agents should restart
type RestartPolicy struct {
	// Policy type: Always, OnFailure, Never
	Policy string `json:"policy" yaml:"policy"`

	// MaxRetries before giving up (-1 for unlimited)
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// BackoffMultiplier for exponential backoff
	BackoffMultiplier float64 `json:"backoff_multiplier" yaml:"backoff_multiplier"`

	// InitialDelay before first restart
	InitialDelay time.Duration `json:"initial_delay" yaml:"initial_delay"`

	// MaxDelay caps the backoff delay
	MaxDelay time.Duration `json:"max_delay" yaml:"max_delay"`
}

// HealthCheckConfig defines health checking behavior
type HealthCheckConfig struct {
	// Enabled indicates if health checks are active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Path for HTTP health checks
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Port for health check endpoint
	Port int `json:"port,omitempty" yaml:"port,omitempty"`

	// Interval between health checks
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Timeout for health check requests
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// FailureThreshold before marking unhealthy
	FailureThreshold int `json:"failure_threshold" yaml:"failure_threshold"`

	// SuccessThreshold before marking healthy
	SuccessThreshold int `json:"success_threshold" yaml:"success_threshold"`

	// InitialDelay before starting health checks
	InitialDelay time.Duration `json:"initial_delay" yaml:"initial_delay"`
}

// LoggingConfig defines logging behavior
type LoggingConfig struct {
	// Level defines log level (debug, info, warn, error)
	Level string `json:"level" yaml:"level"`

	// Format defines log format (json, text)
	Format string `json:"format" yaml:"format"`

	// Output defines where logs go (stdout, file, syslog)
	Output string `json:"output" yaml:"output"`

	// File path for file output
	FilePath string `json:"file_path,omitempty" yaml:"file_path,omitempty"`

	// MaxSize for log rotation (MB)
	MaxSize int `json:"max_size,omitempty" yaml:"max_size,omitempty"`

	// MaxBackups for log retention
	MaxBackups int `json:"max_backups,omitempty" yaml:"max_backups,omitempty"`

	// MaxAge for log retention (days)
	MaxAge int `json:"max_age,omitempty" yaml:"max_age,omitempty"`

	// Compress old log files
	Compress bool `json:"compress,omitempty" yaml:"compress,omitempty"`
}

// MetricsConfig defines metrics collection behavior
type MetricsConfig struct {
	// Enabled indicates if metrics collection is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Port for metrics endpoint
	Port int `json:"port,omitempty" yaml:"port,omitempty"`

	// Path for metrics endpoint
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// Interval for metrics collection
	Interval time.Duration `json:"interval" yaml:"interval"`

	// Custom metrics to collect
	CustomMetrics []string `json:"custom_metrics,omitempty" yaml:"custom_metrics,omitempty"`
}

// CommunicationConfig defines communication behavior
type CommunicationConfig struct {
	// Protocols supported by the agent
	Protocols []string `json:"protocols" yaml:"protocols"`

	// MessageQueueSize for buffering
	MessageQueueSize int `json:"message_queue_size" yaml:"message_queue_size"`

	// ConnectionTimeout for establishing connections
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"`

	// ReadTimeout for message reading
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout"`

	// WriteTimeout for message writing
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// MaxRetries for failed communications
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
}

// MemoryConfig defines memory management behavior
type MemoryConfig struct {
	// MaxMemoryUsage in MB
	MaxMemoryUsage int `json:"max_memory_usage" yaml:"max_memory_usage"`

	// GCInterval for garbage collection
	GCInterval time.Duration `json:"gc_interval" yaml:"gc_interval"`

	// PersistenceEnabled for memory persistence
	PersistenceEnabled bool `json:"persistence_enabled" yaml:"persistence_enabled"`

	// SyncInterval for memory synchronization
	SyncInterval time.Duration `json:"sync_interval" yaml:"sync_interval"`
}

// SecurityConfig defines security settings
type SecurityConfig struct {
	// TLSEnabled for secure communications
	TLSEnabled bool `json:"tls_enabled" yaml:"tls_enabled"`

	// CertificatePath for TLS certificate
	CertificatePath string `json:"certificate_path,omitempty" yaml:"certificate_path,omitempty"`

	// KeyPath for TLS private key
	KeyPath string `json:"key_path,omitempty" yaml:"key_path,omitempty"`

	// AuthenticationRequired for API access
	AuthenticationRequired bool `json:"authentication_required" yaml:"authentication_required"`

	// AllowedOrigins for CORS
	AllowedOrigins []string `json:"allowed_origins,omitempty" yaml:"allowed_origins,omitempty"`
}

// DeploymentStrategy defines deployment approach
type DeploymentStrategy string

const (
	// DeploymentStrategyRecreate kills all existing instances before creating new ones
	DeploymentStrategyRecreate DeploymentStrategy = "Recreate"

	// DeploymentStrategyRollingUpdate gradually replaces instances
	DeploymentStrategyRollingUpdate DeploymentStrategy = "RollingUpdate"

	// DeploymentStrategyBlueGreen creates new instances alongside old ones
	DeploymentStrategyBlueGreen DeploymentStrategy = "BlueGreen"

	// DeploymentStrategyCanary gradually shifts traffic to new instances
	DeploymentStrategyCanary DeploymentStrategy = "Canary"
)

// RollingUpdateConfig defines rolling update behavior
type RollingUpdateConfig struct {
	// MaxUnavailable during update (number or percentage)
	MaxUnavailable string `json:"max_unavailable,omitempty" yaml:"max_unavailable,omitempty"`

	// MaxSurge during update (number or percentage)
	MaxSurge string `json:"max_surge,omitempty" yaml:"max_surge,omitempty"`

	// PauseBeforeScale pauses before scaling operations
	PauseBeforeScale time.Duration `json:"pause_before_scale,omitempty" yaml:"pause_before_scale,omitempty"`
}

// ResourceRequirements defines resource needs
type ResourceRequirements struct {
	// Requests define minimum resource requirements
	Requests ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`

	// Limits define maximum resource usage
	Limits ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// ResourceList defines resource amounts
type ResourceList struct {
	// CPU in millicores (e.g., "100m" or "0.1")
	CPU string `json:"cpu,omitempty" yaml:"cpu,omitempty"`

	// Memory in bytes (e.g., "128Mi", "1Gi")
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`

	// Storage in bytes (e.g., "1Gi", "10Gi")
	Storage string `json:"storage,omitempty" yaml:"storage,omitempty"`

	// Custom resources
	Custom map[string]string `json:"custom,omitempty" yaml:"custom,omitempty"`
}

// Toleration defines node taint tolerations
type Toleration struct {
	// Key is the taint key to tolerate
	Key string `json:"key,omitempty" yaml:"key,omitempty"`

	// Operator defines how the key is compared
	Operator string `json:"operator,omitempty" yaml:"operator,omitempty"`

	// Value is the taint value to match
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	// Effect specifies the taint effect to tolerate
	Effect string `json:"effect,omitempty" yaml:"effect,omitempty"`

	// TolerationSeconds defines how long to tolerate
	TolerationSeconds *int64 `json:"toleration_seconds,omitempty" yaml:"toleration_seconds,omitempty"`
}

// Affinity defines node and pod affinity rules
type Affinity struct {
	// NodeAffinity for node selection preferences
	NodeAffinity *NodeAffinity `json:"node_affinity,omitempty" yaml:"node_affinity,omitempty"`

	// PodAffinity for pod co-location preferences
	PodAffinity *PodAffinity `json:"pod_affinity,omitempty" yaml:"pod_affinity,omitempty"`

	// PodAntiAffinity for pod separation preferences
	PodAntiAffinity *PodAffinity `json:"pod_anti_affinity,omitempty" yaml:"pod_anti_affinity,omitempty"`
}

// NodeAffinity defines node selection rules
type NodeAffinity struct {
	// RequiredDuringSchedulingIgnoredDuringExecution hard requirements
	RequiredDuringSchedulingIgnoredDuringExecution *NodeSelector `json:"required_during_scheduling_ignored_during_execution,omitempty" yaml:"required_during_scheduling_ignored_during_execution,omitempty"`

	// PreferredDuringSchedulingIgnoredDuringExecution soft preferences
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty" yaml:"preferred_during_scheduling_ignored_during_execution,omitempty"`
}

// PodAffinity defines pod affinity rules
type PodAffinity struct {
	// RequiredDuringSchedulingIgnoredDuringExecution hard requirements
	RequiredDuringSchedulingIgnoredDuringExecution []PodAffinityTerm `json:"required_during_scheduling_ignored_during_execution,omitempty" yaml:"required_during_scheduling_ignored_during_execution,omitempty"`

	// PreferredDuringSchedulingIgnoredDuringExecution soft preferences
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferred_during_scheduling_ignored_during_execution,omitempty" yaml:"preferred_during_scheduling_ignored_during_execution,omitempty"`
}

// NodeSelector defines node selection criteria
type NodeSelector struct {
	// NodeSelectorTerms is a list of node selector terms (ORed)
	NodeSelectorTerms []NodeSelectorTerm `json:"node_selector_terms" yaml:"node_selector_terms"`
}

// NodeSelectorTerm defines a single node selector term
type NodeSelectorTerm struct {
	// MatchExpressions is a list of expressions (ANDed)
	MatchExpressions []NodeSelectorRequirement `json:"match_expressions,omitempty" yaml:"match_expressions,omitempty"`

	// MatchFields is a list of field selectors (ANDed)
	MatchFields []NodeSelectorRequirement `json:"match_fields,omitempty" yaml:"match_fields,omitempty"`
}

// NodeSelectorRequirement defines a selector requirement
type NodeSelectorRequirement struct {
	// Key is the label key or field name
	Key string `json:"key" yaml:"key"`

	// Operator defines the comparison operation
	Operator string `json:"operator" yaml:"operator"`

	// Values is the list of values to match
	Values []string `json:"values,omitempty" yaml:"values,omitempty"`
}

// PreferredSchedulingTerm defines a preferred scheduling term
type PreferredSchedulingTerm struct {
	// Weight defines the preference weight (1-100)
	Weight int32 `json:"weight" yaml:"weight"`

	// Preference defines the node selector term
	Preference NodeSelectorTerm `json:"preference" yaml:"preference"`
}

// PodAffinityTerm defines pod affinity constraints
type PodAffinityTerm struct {
	// LabelSelector selects pods by labels
	LabelSelector *LabelSelector `json:"label_selector,omitempty" yaml:"label_selector,omitempty"`

	// Namespaces specifies which namespaces to consider
	Namespaces []string `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`

	// TopologyKey defines the topology domain
	TopologyKey string `json:"topology_key" yaml:"topology_key"`
}

// WeightedPodAffinityTerm defines weighted pod affinity
type WeightedPodAffinityTerm struct {
	// Weight defines the preference weight (1-100)
	Weight int32 `json:"weight" yaml:"weight"`

	// PodAffinityTerm defines the affinity constraint
	PodAffinityTerm PodAffinityTerm `json:"pod_affinity_term" yaml:"pod_affinity_term"`
}

// LabelSelector defines label selection criteria
type LabelSelector struct {
	// MatchLabels is a map of key-value pairs
	MatchLabels map[string]string `json:"match_labels,omitempty" yaml:"match_labels,omitempty"`

	// MatchExpressions is a list of label selector requirements
	MatchExpressions []LabelSelectorRequirement `json:"match_expressions,omitempty" yaml:"match_expressions,omitempty"`
}

// LabelSelectorRequirement defines a label selector requirement
type LabelSelectorRequirement struct {
	// Key is the label key
	Key string `json:"key" yaml:"key"`

	// Operator defines the comparison operation
	Operator string `json:"operator" yaml:"operator"`

	// Values is the list of values to match
	Values []string `json:"values,omitempty" yaml:"values,omitempty"`
}

// Validate validates the agent configuration
func (ac *AgentConfiguration) Validate() error {
	// Implement comprehensive validation logic
	// TODO: Add validation for all configuration fields
	return nil
}

// ToJSON converts the configuration to JSON
func (ac *AgentConfiguration) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ac, "", "  ")
}

// Clone creates a deep copy of the configuration
func (ac *AgentConfiguration) Clone() (*AgentConfiguration, error) {
	data, err := ac.ToJSON()
	if err != nil {
		return nil, err
	}

	var clone AgentConfiguration
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}

	return &clone, nil
}
