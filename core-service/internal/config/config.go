package config

import (
	"context"
	"core-service/internal/logger"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type DB struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type HTTP struct {
	Address       string `mapstructure:"address"`
	PathPrefix    string `mapstructure:"path_prefix"`
	EnableGateway bool   `mapstructure:"enable_gateway"`
}

type GRPC struct {
	Port int `mapstructure:"port"`
}

type JWT struct {
	Secret    string `mapstructure:"secret"`
	AccessTTL string `mapstructure:"access_ttl"`
}

func (db *DB) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
		db.SSLMode,
	)
}

type Jaeger struct {
	Endpoint string `mapstructure:"endpoint"`
}

type RabbitMQ struct {
	URL       string `mapstructure:"url"`
	QueueName string `mapstructure:"queue_name"`
}

type Telegram struct {
	BotTokens          []string `mapstructure:"bot_tokens"`
	InitDataTTLSeconds int      `mapstructure:"init_data_ttl_seconds"`
}

// Config holds all runtime-configurable settings
type Config struct {
	mu sync.RWMutex

	ServiceName string   `mapstructure:"service_name"`
	HTTP        HTTP     `mapstructure:"http"`
	GRPC        GRPC     `mapstructure:"grpc"`
	DB          DB       `mapstructure:"db"`
	JWT         JWT      `mapstructure:"jwt"`
	Jaeger      Jaeger   `mapstructure:"jaeger"`
	RabbitMQ    RabbitMQ `mapstructure:"rabbitmq"`
	Telegram    Telegram `mapstructure:"telegram"`
}

var (
	instance *Config
	once     sync.Once
)

// Get returns the singleton config instance
func Get() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.loadDefaults()
	})
	return instance
}

// Initialize sets up viper to watch config file
func Initialize(configPath string) error {
	cfg := Get()

	viper.SetConfigFile(configPath)

	// Enable environment variable substitution
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	// Load initial config
	if err := cfg.reload(); err != nil {
		return fmt.Errorf("failed to load initial config: %w", err)
	}

	// Watch for config changes
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		ctx := context.Background()
		logger.Info(ctx, fmt.Sprintf("Config file changed: %s", e.Name))
		if err := cfg.reload(); err != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to reload config: %v", err))
		} else {
			logger.Info(ctx, "Config reloaded successfully")
		}
	})

	return nil
}

// loadDefaults sets default values for all config fields
func (c *Config) loadDefaults() {
	c.mu.Lock()
	defer c.mu.Unlock()

	viper.SetDefault("service_name", "core-service")
	viper.SetDefault("http.address", ":8080")
	viper.SetDefault("http.path_prefix", "/api")
	viper.SetDefault("http.enable_gateway", true)
	viper.SetDefault("grpc.port", 50051)
	viper.SetDefault("jaeger.endpoint", "http://localhost:14268/api/traces")
	viper.SetDefault("db.ssl_mode", "disable")
	viper.SetDefault("jwt.secret", "")
	viper.SetDefault("jwt.access_ttl", "1h")
	viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("telegram.bot_tokens", []string{})
	viper.SetDefault("telegram.init_data_ttl_seconds", 86400)
}

// reload reads the config file and updates all values
func (c *Config) reload() error {
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Unmarshal the config into the struct
	if err := viper.Unmarshal(c); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// GetDB returns the current DB config
func (c *Config) GetDBDSN() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DB.GetDSN()
}

func (c *Config) GetServiceName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ServiceName
}

func (c *Config) GetHTTPAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HTTP.Address
}

func (c *Config) GetHTTPPathPrefix() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HTTP.PathPrefix
}

func (c *Config) GetHTTPEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HTTP.EnableGateway
}

func (c *Config) GetHTTPPort() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Parse port from address like ":8080"
	var port int
	fmt.Sscanf(c.HTTP.Address, ":%d", &port)
	return port
}

func (c *Config) GetGRPCPort() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GRPC.Port
}

func (c *Config) GetGRPCEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf(":%d", c.GRPC.Port)
}

// GetJWTSecret returns the JWT secret from config
func (c *Config) GetJWTSecret() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.JWT.Secret
}

func (c *Config) GetJWTAccessTTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.JWT.AccessTTL == "" {
		return time.Hour
	}
	if ttl, err := time.ParseDuration(c.JWT.AccessTTL); err == nil {
		return ttl
	}
	return time.Hour
}

// GetJaegerEndpoint returns the Jaeger endpoint from config
func (c *Config) GetJaegerEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Jaeger.Endpoint
}

// GetRabbitMQURL returns the RabbitMQ URL from config
func (c *Config) GetRabbitMQURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RabbitMQ.URL
}

func (c *Config) GetRabbitMQQueueName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RabbitMQ.QueueName
}

func (c *Config) GetTelegramBotTokens() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Telegram.BotTokens
}

func (c *Config) GetTelegramInitDataTTL() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Telegram.InitDataTTLSeconds <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(c.Telegram.InitDataTTLSeconds) * time.Second
}
