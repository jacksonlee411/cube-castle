package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port" default:"9090"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"30s"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" default:"120s"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host" default:"localhost"`
	Port            int           `mapstructure:"port" default:"5432"`
	Database        string        `mapstructure:"database" default:"cubecastle"`
	Username        string        `mapstructure:"username" default:"user"`
	Password        string        `mapstructure:"password" default:"password"`
	MaxConnections  int           `mapstructure:"max_connections" default:"25"`
	MinConnections  int           `mapstructure:"min_connections" default:"5"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime" default:"30m"`
	SSLMode         string        `mapstructure:"ssl_mode" default:"disable"`
}

type KafkaConfig struct {
	Brokers    []string `mapstructure:"brokers"`
	EventTopic string   `mapstructure:"event_topic" default:"organization.events"`
	ClientID   string   `mapstructure:"client_id" default:"organization-command-service"`
	Acks       string   `mapstructure:"acks" default:"all"`
	Retries    int      `mapstructure:"retries" default:"3"`
	BatchSize  int      `mapstructure:"batch_size" default:"16384"`
	LingerMS   int      `mapstructure:"linger_ms" default:"10"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host" default:"localhost"`
	Port     int    `mapstructure:"port" default:"6379"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database" default:"0"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level" default:"info"`
	Format     string `mapstructure:"format" default:"json"`
	Output     string `mapstructure:"output" default:"stdout"`
	TimeFormat string `mapstructure:"time_format" default:"2006-01-02T15:04:05.000Z07:00"`
}

type AppConfig struct {
	Name             string `mapstructure:"name" default:"Organization Command Service"`
	Version          string `mapstructure:"version" default:"1.0.0"`
	Environment      string `mapstructure:"environment" default:"development"`
	DefaultTenantID  string `mapstructure:"default_tenant_id" default:"3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"`
	DefaultTenantName string `mapstructure:"default_tenant_name" default:"高谷集团"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Set config file search paths and name
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("../configs")
		viper.AddConfigPath("/etc/organization-command-server/")
	}
	
	// Environment variable handling
	viper.SetEnvPrefix("ORG_CMD")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()
	
	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is not fatal - we can work with env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 9090)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "cubecastle")
	viper.SetDefault("database.username", "user")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.max_connections", 25)
	viper.SetDefault("database.min_connections", 5)
	viper.SetDefault("database.max_conn_lifetime", "30m")
	viper.SetDefault("database.ssl_mode", "disable")
	
	// Kafka defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.event_topic", "organization.events")
	viper.SetDefault("kafka.client_id", "organization-command-service")
	viper.SetDefault("kafka.acks", "all")
	viper.SetDefault("kafka.retries", 3)
	viper.SetDefault("kafka.batch_size", 16384)
	viper.SetDefault("kafka.linger_ms", 10)
	
	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.time_format", "2006-01-02T15:04:05.000Z07:00")
	
	// App defaults
	viper.SetDefault("app.name", "Organization Command Service")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.default_tenant_id", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
	viper.SetDefault("app.default_tenant_name", "高谷集团")
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if len(config.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}
	
	if config.App.DefaultTenantID == "" {
		return fmt.Errorf("default tenant ID is required")
	}
	
	return nil
}

// GetDatabaseConnectionString returns formatted database connection string
func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}