package configuration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aosanya/CodeValdCortex/internal/agent"
)

// DefaultValidator implements the Validator interface
type DefaultValidator struct {
	// resourceChecker checks resource availability
	resourceChecker ResourceChecker
}

// ResourceChecker defines the interface for checking resource availability
type ResourceChecker interface {
	// CheckCPUAvailability checks if CPU resources are available
	CheckCPUAvailability(cpu string) error

	// CheckMemoryAvailability checks if memory resources are available
	CheckMemoryAvailability(memory string) error

	// CheckStorageAvailability checks if storage resources are available
	CheckStorageAvailability(storage string) error
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%s': %s", ve.Field, ve.Value, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// NewDefaultValidator creates a new default validator
func NewDefaultValidator(resourceChecker ResourceChecker) *DefaultValidator {
	return &DefaultValidator{
		resourceChecker: resourceChecker,
	}
}

// Validate validates an agent configuration
func (v *DefaultValidator) Validate(config *AgentConfiguration) error {
	var errors ValidationErrors

	// Validate basic fields
	if config.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Value:   config.Name,
			Message: "name cannot be empty",
		})
	}

	if config.AgentType == "" {
		errors = append(errors, ValidationError{
			Field:   "agent_type",
			Value:   config.AgentType,
			Message: "agent type cannot be empty",
		})
	}

	// Validate agent type is supported
	if !isValidAgentType(config.AgentType) {
		errors = append(errors, ValidationError{
			Field:   "agent_type",
			Value:   config.AgentType,
			Message: "unsupported agent type",
		})
	}

	// Validate base configuration
	if err := v.validateBaseConfig(&config.BaseConfig); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "base_config",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	// Validate runtime configuration
	if err := v.validateRuntimeConfig(&config.RuntimeConfig); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "runtime_config",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	// Validate deployment configuration
	if err := v.validateDeploymentConfig(&config.DeploymentConfig); err != nil {
		if ve, ok := err.(ValidationErrors); ok {
			errors = append(errors, ve...)
		} else {
			errors = append(errors, ValidationError{
				Field:   "deployment_config",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	// Validate labels format
	if err := v.validateLabels(config.Labels); err != nil {
		errors = append(errors, ValidationError{
			Field:   "labels",
			Value:   "",
			Message: err.Error(),
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateCompatibility checks if a configuration is compatible with the current system
func (v *DefaultValidator) ValidateCompatibility(config *AgentConfiguration) error {
	var errors ValidationErrors

	// Check if agent type is supported by the system
	if !v.isAgentTypeSupported(config.AgentType) {
		errors = append(errors, ValidationError{
			Field:   "agent_type",
			Value:   config.AgentType,
			Message: "agent type not supported by current system",
		})
	}

	// Check deployment strategy compatibility
	if !v.isDeploymentStrategySupported(config.DeploymentConfig.Strategy) {
		errors = append(errors, ValidationError{
			Field:   "deployment_config.strategy",
			Value:   string(config.DeploymentConfig.Strategy),
			Message: "deployment strategy not supported",
		})
	}

	// Check resource format compatibility
	if err := v.validateResourceFormats(&config.DeploymentConfig.Resources); err != nil {
		errors = append(errors, ValidationError{
			Field:   "deployment_config.resources",
			Value:   "",
			Message: err.Error(),
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateResources checks if resources are available for the configuration
func (v *DefaultValidator) ValidateResources(config *AgentConfiguration) error {
	if v.resourceChecker == nil {
		return nil // Skip resource checking if no checker is provided
	}

	var errors ValidationErrors

	resources := &config.DeploymentConfig.Resources

	// Check CPU resources
	if resources.Requests.CPU != "" {
		if err := v.resourceChecker.CheckCPUAvailability(resources.Requests.CPU); err != nil {
			errors = append(errors, ValidationError{
				Field:   "deployment_config.resources.requests.cpu",
				Value:   resources.Requests.CPU,
				Message: fmt.Sprintf("CPU resources not available: %v", err),
			})
		}
	}

	// Check memory resources
	if resources.Requests.Memory != "" {
		if err := v.resourceChecker.CheckMemoryAvailability(resources.Requests.Memory); err != nil {
			errors = append(errors, ValidationError{
				Field:   "deployment_config.resources.requests.memory",
				Value:   resources.Requests.Memory,
				Message: fmt.Sprintf("memory resources not available: %v", err),
			})
		}
	}

	// Check storage resources
	if resources.Requests.Storage != "" {
		if err := v.resourceChecker.CheckStorageAvailability(resources.Requests.Storage); err != nil {
			errors = append(errors, ValidationError{
				Field:   "deployment_config.resources.requests.storage",
				Value:   resources.Requests.Storage,
				Message: fmt.Sprintf("storage resources not available: %v", err),
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateBaseConfig validates the base agent configuration
func (v *DefaultValidator) validateBaseConfig(config *agent.Config) error {
	var errors ValidationErrors

	if config.MaxConcurrentTasks <= 0 {
		errors = append(errors, ValidationError{
			Field:   "base_config.max_concurrent_tasks",
			Value:   fmt.Sprintf("%d", config.MaxConcurrentTasks),
			Message: "max concurrent tasks must be greater than 0",
		})
	}

	if config.TaskQueueSize <= 0 {
		errors = append(errors, ValidationError{
			Field:   "base_config.task_queue_size",
			Value:   fmt.Sprintf("%d", config.TaskQueueSize),
			Message: "task queue size must be greater than 0",
		})
	}

	if config.HeartbeatInterval <= 0 {
		errors = append(errors, ValidationError{
			Field:   "base_config.heartbeat_interval",
			Value:   config.HeartbeatInterval.String(),
			Message: "heartbeat interval must be greater than 0",
		})
	}

	if config.TaskTimeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "base_config.task_timeout",
			Value:   config.TaskTimeout.String(),
			Message: "task timeout must be greater than 0",
		})
	}

	// Validate resources
	if config.Resources.CPU <= 0 {
		errors = append(errors, ValidationError{
			Field:   "base_config.resources.cpu",
			Value:   fmt.Sprintf("%d", config.Resources.CPU),
			Message: "CPU must be greater than 0",
		})
	}

	if config.Resources.Memory <= 0 {
		errors = append(errors, ValidationError{
			Field:   "base_config.resources.memory",
			Value:   fmt.Sprintf("%d", config.Resources.Memory),
			Message: "memory must be greater than 0",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateRuntimeConfig validates the runtime configuration
func (v *DefaultValidator) validateRuntimeConfig(config *RuntimeConfiguration) error {
	var errors ValidationErrors

	// Validate restart policy
	if err := v.validateRestartPolicy(&config.RestartPolicy); err != nil {
		errors = append(errors, ValidationError{
			Field:   "runtime_config.restart_policy",
			Value:   "",
			Message: err.Error(),
		})
	}

	// Validate health check configuration
	if err := v.validateHealthCheckConfig(&config.HealthCheck); err != nil {
		errors = append(errors, ValidationError{
			Field:   "runtime_config.health_check",
			Value:   "",
			Message: err.Error(),
		})
	}

	// Validate logging configuration
	if err := v.validateLoggingConfig(&config.Logging); err != nil {
		errors = append(errors, ValidationError{
			Field:   "runtime_config.logging",
			Value:   "",
			Message: err.Error(),
		})
	}

	// Validate metrics configuration
	if err := v.validateMetricsConfig(&config.Metrics); err != nil {
		errors = append(errors, ValidationError{
			Field:   "runtime_config.metrics",
			Value:   "",
			Message: err.Error(),
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateDeploymentConfig validates the deployment configuration
func (v *DefaultValidator) validateDeploymentConfig(config *DeploymentConfiguration) error {
	var errors ValidationErrors

	if config.Replicas < 0 {
		errors = append(errors, ValidationError{
			Field:   "deployment_config.replicas",
			Value:   fmt.Sprintf("%d", config.Replicas),
			Message: "replicas cannot be negative",
		})
	}

	// Validate deployment strategy
	if !isValidDeploymentStrategy(config.Strategy) {
		errors = append(errors, ValidationError{
			Field:   "deployment_config.strategy",
			Value:   string(config.Strategy),
			Message: "invalid deployment strategy",
		})
	}

	// Validate rolling update config if strategy is rolling update
	if config.Strategy == DeploymentStrategyRollingUpdate {
		if err := v.validateRollingUpdateConfig(&config.RollingUpdate); err != nil {
			errors = append(errors, ValidationError{
				Field:   "deployment_config.rolling_update",
				Value:   "",
				Message: err.Error(),
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateRestartPolicy validates restart policy configuration
func (v *DefaultValidator) validateRestartPolicy(policy *RestartPolicy) error {
	validPolicies := []string{"Always", "OnFailure", "Never"}
	for _, valid := range validPolicies {
		if policy.Policy == valid {
			break
		}
	}

	if policy.MaxRetries < -1 {
		return fmt.Errorf("max retries cannot be less than -1")
	}

	if policy.BackoffMultiplier <= 0 {
		return fmt.Errorf("backoff multiplier must be greater than 0")
	}

	if policy.InitialDelay < 0 {
		return fmt.Errorf("initial delay cannot be negative")
	}

	if policy.MaxDelay < 0 {
		return fmt.Errorf("max delay cannot be negative")
	}

	if policy.MaxDelay > 0 && policy.InitialDelay > policy.MaxDelay {
		return fmt.Errorf("initial delay cannot be greater than max delay")
	}

	return nil
}

// validateHealthCheckConfig validates health check configuration
func (v *DefaultValidator) validateHealthCheckConfig(config *HealthCheckConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if config.Interval <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	if config.FailureThreshold <= 0 {
		return fmt.Errorf("failure threshold must be greater than 0")
	}

	if config.SuccessThreshold <= 0 {
		return fmt.Errorf("success threshold must be greater than 0")
	}

	if config.InitialDelay < 0 {
		return fmt.Errorf("initial delay cannot be negative")
	}

	return nil
}

// validateLoggingConfig validates logging configuration
func (v *DefaultValidator) validateLoggingConfig(config *LoggingConfig) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if config.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s", config.Level)
	}

	validFormats := []string{"json", "text"}
	validFormat := false
	for _, format := range validFormats {
		if config.Format == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid log format: %s", config.Format)
	}

	validOutputs := []string{"stdout", "file", "syslog"}
	validOutput := false
	for _, output := range validOutputs {
		if config.Output == output {
			validOutput = true
			break
		}
	}
	if !validOutput {
		return fmt.Errorf("invalid log output: %s", config.Output)
	}

	if config.Output == "file" && config.FilePath == "" {
		return fmt.Errorf("file path is required when output is 'file'")
	}

	return nil
}

// validateMetricsConfig validates metrics configuration
func (v *DefaultValidator) validateMetricsConfig(config *MetricsConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if config.Interval <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}

	return nil
}

// validateRollingUpdateConfig validates rolling update configuration
func (v *DefaultValidator) validateRollingUpdateConfig(config *RollingUpdateConfig) error {
	if config.MaxUnavailable != "" {
		if err := validateIntOrPercentage(config.MaxUnavailable); err != nil {
			return fmt.Errorf("invalid max unavailable: %v", err)
		}
	}

	if config.MaxSurge != "" {
		if err := validateIntOrPercentage(config.MaxSurge); err != nil {
			return fmt.Errorf("invalid max surge: %v", err)
		}
	}

	if config.PauseBeforeScale < 0 {
		return fmt.Errorf("pause before scale cannot be negative")
	}

	return nil
}

// validateResourceFormats validates resource format strings
func (v *DefaultValidator) validateResourceFormats(resources *ResourceRequirements) error {
	if resources.Requests.CPU != "" {
		if err := validateResourceQuantity(resources.Requests.CPU); err != nil {
			return fmt.Errorf("invalid CPU request format: %v", err)
		}
	}

	if resources.Limits.CPU != "" {
		if err := validateResourceQuantity(resources.Limits.CPU); err != nil {
			return fmt.Errorf("invalid CPU limit format: %v", err)
		}
	}

	if resources.Requests.Memory != "" {
		if err := validateResourceQuantity(resources.Requests.Memory); err != nil {
			return fmt.Errorf("invalid memory request format: %v", err)
		}
	}

	if resources.Limits.Memory != "" {
		if err := validateResourceQuantity(resources.Limits.Memory); err != nil {
			return fmt.Errorf("invalid memory limit format: %v", err)
		}
	}

	if resources.Requests.Storage != "" {
		if err := validateResourceQuantity(resources.Requests.Storage); err != nil {
			return fmt.Errorf("invalid storage request format: %v", err)
		}
	}

	if resources.Limits.Storage != "" {
		if err := validateResourceQuantity(resources.Limits.Storage); err != nil {
			return fmt.Errorf("invalid storage limit format: %v", err)
		}
	}

	return nil
}

// validateLabels validates label format and content
func (v *DefaultValidator) validateLabels(labels map[string]string) error {
	labelRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-_\.]*[a-zA-Z0-9])?$`)

	for key, value := range labels {
		if len(key) == 0 || len(key) > 63 {
			return fmt.Errorf("label key '%s' must be 1-63 characters", key)
		}

		if !labelRegex.MatchString(key) {
			return fmt.Errorf("label key '%s' contains invalid characters", key)
		}

		if len(value) > 63 {
			return fmt.Errorf("label value '%s' must be at most 63 characters", value)
		}

		if value != "" && !labelRegex.MatchString(value) {
			return fmt.Errorf("label value '%s' contains invalid characters", value)
		}
	}

	return nil
}

// Helper functions

func isValidAgentType(agentType string) bool {
	validTypes := []string{"worker", "coordinator", "monitor", "proxy", "gateway"}
	for _, valid := range validTypes {
		if agentType == valid {
			return true
		}
	}
	return false
}

func isValidDeploymentStrategy(strategy DeploymentStrategy) bool {
	validStrategies := []DeploymentStrategy{
		DeploymentStrategyRecreate,
		DeploymentStrategyRollingUpdate,
		DeploymentStrategyBlueGreen,
		DeploymentStrategyCanary,
	}
	for _, valid := range validStrategies {
		if strategy == valid {
			return true
		}
	}
	return false
}

func (v *DefaultValidator) isAgentTypeSupported(agentType string) bool {
	// TODO: Check with the actual system capabilities
	return isValidAgentType(agentType)
}

func (v *DefaultValidator) isDeploymentStrategySupported(strategy DeploymentStrategy) bool {
	// TODO: Check with the actual system capabilities
	return isValidDeploymentStrategy(strategy)
}

func validateIntOrPercentage(value string) error {
	if strings.HasSuffix(value, "%") {
		percentStr := strings.TrimSuffix(value, "%")
		percent, err := strconv.Atoi(percentStr)
		if err != nil {
			return fmt.Errorf("invalid percentage: %s", value)
		}
		if percent < 0 || percent > 100 {
			return fmt.Errorf("percentage must be 0-100: %s", value)
		}
		return nil
	}

	// Parse as integer
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("must be integer or percentage: %s", value)
	}
	return nil
}

func validateResourceQuantity(quantity string) error {
	// Validate Kubernetes-style resource quantities
	// Examples: "100m", "1", "1.5", "100Mi", "1Gi", "1Ti"

	if quantity == "" {
		return fmt.Errorf("quantity cannot be empty")
	}

	// CPU quantities (can have 'm' suffix for millicores)
	cpuRegex := regexp.MustCompile(`^(\d+(\.\d+)?)(m)?$`)
	if cpuRegex.MatchString(quantity) {
		return nil
	}

	// Memory/Storage quantities (can have binary suffixes)
	memoryRegex := regexp.MustCompile(`^(\d+(\.\d+)?)(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)?$`)
	if memoryRegex.MatchString(quantity) {
		return nil
	}

	return fmt.Errorf("invalid resource quantity format: %s", quantity)
}
