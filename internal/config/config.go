package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Application settings
	AppName   string `mapstructure:"app_name"`
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`

	// Server configuration
	Server ServerConfig `mapstructure:"server"`

	// Database configuration
	Database DatabaseConfig `mapstructure:"database"`

	// Kubernetes configuration
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`

	// Agent configuration
	Agent AgentConfig `mapstructure:"agent"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	TLSEnabled   bool   `mapstructure:"tls_enabled"`
	TLSCertFile  string `mapstructure:"tls_cert_file"`
	TLSKeyFile   string `mapstructure:"tls_key_file"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	Namespace  string `mapstructure:"namespace"`
	InCluster  bool   `mapstructure:"in_cluster"`
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	DefaultImage     string            `mapstructure:"default_image"`
	DefaultResources map[string]string `mapstructure:"default_resources"`
	MaxInstances     int               `mapstructure:"max_instances"`
	HealthCheckPath  string            `mapstructure:"health_check_path"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		// Set defaults
		AppName:   "CodeValdCortex",
		LogLevel:  "info",
		LogFormat: "text",
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
			TLSEnabled:   false,
		},
		Database: DatabaseConfig{
			Type:     "arangodb",
			Host:     "localhost",
			Port:     8529,
			Database: "codevaldcortex",
			Username: "root",
			SSLMode:  "disable",
		},
		Kubernetes: KubernetesConfig{
			Namespace: "default",
			InCluster: false,
		},
		Agent: AgentConfig{
			DefaultImage:    "codevaldcortex/agent:latest",
			MaxInstances:    100,
			HealthCheckPath: "/health",
			DefaultResources: map[string]string{
				"cpu":    "100m",
				"memory": "128Mi",
			},
		},
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config paths
	if configPath != "" {
		if filepath.IsAbs(configPath) {
			viper.SetConfigFile(configPath)
		} else {
			viper.AddConfigPath(filepath.Dir(configPath))
			viper.SetConfigName(filepath.Base(configPath[:len(configPath)-len(filepath.Ext(configPath))]))
		}
	}

	// Add common config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/codevaldcortex")

	// Environment variable support
	viper.SetEnvPrefix("CVXC")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found is acceptable, we'll use defaults and env vars
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	// Override with environment variables
	if password := os.Getenv("CVXC_DATABASE_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if port := os.Getenv("CVXC_SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if dbPort := os.Getenv("CVXC_DATABASE_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}

	return config, nil
}
