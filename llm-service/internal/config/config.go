package config

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"llm-service/internal/logger"

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

type LLM struct {
	BaseURL         string `mapstructure:"base_url"`
	APIKey          string `mapstructure:"api_key"`
	Model           string `mapstructure:"model"`
	ReasoningEffort string `mapstructure:"reasoning_effort"`
	TokenLimit      int    `mapstructure:"token_limit"`
	// Специальные модели для конкретных задач
	TitleGenerationModel string `mapstructure:"title_generation_model"`
}

type JWT struct {
	Secret string `mapstructure:"secret"`
}

type Proxy struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Protocol string `mapstructure:"protocol"`
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

type CoreService struct {
	Address string `mapstructure:"address"`
}

type DocsProcessor struct {
	Address string `mapstructure:"address"`
}

type Tavily struct {
	APIKey string `mapstructure:"api_key"`
}

// Config holds all runtime-configurable settings
type Config struct {
	mu sync.RWMutex

	ServiceName   string        `mapstructure:"service_name"`
	HTTP          HTTP          `mapstructure:"http"`
	GRPC          GRPC          `mapstructure:"grpc"`
	DB            DB            `mapstructure:"db"`
	LLM           LLM           `mapstructure:"llm"`
	Proxy         *Proxy        `mapstructure:"proxy"`
	JWT           JWT           `mapstructure:"jwt"`
	Jaeger        Jaeger        `mapstructure:"jaeger"`
	CoreService   CoreService   `mapstructure:"core_service"`
	DocsProcessor DocsProcessor `mapstructure:"docs_processor"`
	Tavily        Tavily        `mapstructure:"tavily"`
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

	viper.SetDefault("service_name", "llm-agent")
	viper.SetDefault("http.address", ":8084")
	viper.SetDefault("http.path_prefix", "/api")
	viper.SetDefault("http.enable_gateway", true)
	viper.SetDefault("grpc.port", 50063)
	viper.SetDefault("jaeger.endpoint", "http://localhost:14268/api/traces")
	viper.SetDefault("db.ssl_mode", "disable")
	viper.SetDefault("llm.token_limit", 500000)
	viper.SetDefault("jwt.secret", "")
	viper.SetDefault("core_service.address", "localhost:50051")
	viper.SetDefault("docs_processor.address", "localhost:50052")
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
	// Parse port from address like ":8084"
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

// GetProxyUrl returns the current Proxy URL
func (c *Config) GetProxyUrl() *url.URL {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Proxy == nil {
		return nil
	}

	return &url.URL{
		Scheme: c.Proxy.Protocol,
		Host:   fmt.Sprintf("%s:%d", c.Proxy.Host, c.Proxy.Port),
		User:   url.UserPassword(c.Proxy.User, c.Proxy.Password),
	}
}

// GetLLMBaseURL returns the current LLM Base URL
func (c *Config) GetLLMBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.BaseURL
}

// GetLLMAPIKey returns the current LLM API Key
func (c *Config) GetLLMAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.APIKey
}

// GetLLMModel returns the current LLM Model
func (c *Config) GetLLMModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.Model
}

// GetLLMReasoningEffort returns the current LLM Reasoning Effort
func (c *Config) GetLLMReasoningEffort() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.ReasoningEffort
}

// GetLLMTokenLimit returns the current LLM Token Limit
func (c *Config) GetLLMTokenLimit() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LLM.TokenLimit
}

// GetLLMTitleGenerationModel returns the model for title generation
func (c *Config) GetLLMTitleGenerationModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Fallback to main model if not specified
	if c.LLM.TitleGenerationModel == "" {
		return c.LLM.Model
	}
	return c.LLM.TitleGenerationModel
}

// GetJWTSecret returns the JWT secret from config
func (c *Config) GetJWTSecret() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.JWT.Secret
}

// GetJaegerEndpoint returns the Jaeger endpoint from config
func (c *Config) GetJaegerEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Jaeger.Endpoint
}

// GetCoreServiceAddress returns the Core Service address from config
func (c *Config) GetCoreServiceAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CoreService.Address
}

// GetDocsProcessorAddress returns the Docs Processor address from config
func (c *Config) GetDocsProcessorAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DocsProcessor.Address
}

// GetTavilyAPIKey returns the Tavily API key from config
func (c *Config) GetTavilyAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Tavily.APIKey
}
