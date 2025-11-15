package config

import (
	"context"
	"fmt"
	"sync"

	"docs-processor/internal/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type HTTP struct {
	Port          int    `mapstructure:"port"`
	PathPrefix    string `mapstructure:"path_prefix"`
	EnableGateway bool   `mapstructure:"enable_gateway"`
}

type GRPC struct {
	Port int `mapstructure:"port"`
}

type RabbitMQ struct {
	URL           string `mapstructure:"url"`
	QueueName     string `mapstructure:"queue_name"`
	ConsumerTag   string `mapstructure:"consumer_tag"`
	PrefetchCount int    `mapstructure:"prefetch_count"`
}

type S3 struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type OpenSearch struct {
	Addresses      []string `mapstructure:"addresses"`
	Username       string   `mapstructure:"username"`
	Password       string   `mapstructure:"password"`
	IndexName      string   `mapstructure:"index_name"`
	TemplatesIndex string   `mapstructure:"templates_index"`
}

type Embeddings struct {
	BaseURL   string `mapstructure:"base_url"`
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	BatchSize int    `mapstructure:"batch_size"`
}

type Chunking struct {
	MaxChunkSize int `mapstructure:"max_chunk_size"`
	OverlapSize  int `mapstructure:"overlap_size"`
}

type Jaeger struct {
	Endpoint string `mapstructure:"endpoint"`
}

type CoreService struct {
	Address string `mapstructure:"address"`
}

type Config struct {
	mu sync.RWMutex

	ServiceName string      `mapstructure:"service_name"`
	HTTP        HTTP        `mapstructure:"http"`
	GRPC        GRPC        `mapstructure:"grpc"`
	RabbitMQ    RabbitMQ    `mapstructure:"rabbitmq"`
	S3          S3          `mapstructure:"s3"`
	OpenSearch  OpenSearch  `mapstructure:"opensearch"`
	Embeddings  Embeddings  `mapstructure:"embeddings"`
	Chunking    Chunking    `mapstructure:"chunking"`
	Jaeger      Jaeger      `mapstructure:"jaeger"`
	CoreService CoreService `mapstructure:"core_service"`
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		instance = &Config{}
		instance.loadDefaults()
	})
	return instance
}

func (c *Config) loadDefaults() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ServiceName = "docs-processor"
	c.HTTP.Port = 8081
	c.HTTP.EnableGateway = true
	c.GRPC.Port = 50052
	c.RabbitMQ.QueueName = "document_processing"
	c.RabbitMQ.ConsumerTag = "doc-processor-worker"
	c.RabbitMQ.PrefetchCount = 1
	c.S3.Region = "us-east-1"
	c.S3.Bucket = "documents"
	c.OpenSearch.IndexName = "documents"
	c.Embeddings.Model = "text-embedding-3-small"
	c.Embeddings.BatchSize = 100
	c.Chunking.MaxChunkSize = 1000
	c.Chunking.OverlapSize = 200
	c.CoreService.Address = "localhost:50051"
}

func Initialize(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	cfg := Get()
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		cfg.mu.Lock()
		defer cfg.mu.Unlock()

		if err := v.Unmarshal(cfg); err != nil {
			logger.Error(context.Background(), "failed to reload config", "error", err)
		} else {
			logger.Info(context.Background(), "config reloaded", "file", e.Name)
		}
	})

	return nil
}

func (c *Config) GetServiceName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ServiceName
}

func (c *Config) GetHTTPPort() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HTTP.Port
}

func (c *Config) GetHTTPPathPrefix() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HTTP.PathPrefix
}

func (c *Config) GetEnableGateway() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.HTTP.EnableGateway
}

func (c *Config) GetGRPCPort() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.GRPC.Port
}

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

func (c *Config) GetRabbitMQConsumerTag() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RabbitMQ.ConsumerTag
}

func (c *Config) GetRabbitMQPrefetchCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RabbitMQ.PrefetchCount
}

func (c *Config) GetS3Endpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.S3.Endpoint
}

func (c *Config) GetS3AccessKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.S3.AccessKey
}

func (c *Config) GetS3SecretKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.S3.SecretKey
}

func (c *Config) GetS3Bucket() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.S3.Bucket
}

func (c *Config) GetS3Region() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.S3.Region
}

func (c *Config) GetS3UseSSL() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.S3.UseSSL
}

func (c *Config) GetOpenSearchAddresses() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OpenSearch.Addresses
}

func (c *Config) GetOpenSearchUsername() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OpenSearch.Username
}

func (c *Config) GetOpenSearchPassword() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OpenSearch.Password
}

func (c *Config) GetOpenSearchIndexName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OpenSearch.IndexName
}

func (c *Config) GetOpenSearchTemplatesIndex() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.OpenSearch.TemplatesIndex
}

func (c *Config) GetEmbeddingsBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Embeddings.BaseURL
}

func (c *Config) GetEmbeddingsAPIKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Embeddings.APIKey
}

func (c *Config) GetEmbeddingsModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Embeddings.Model
}

func (c *Config) GetEmbeddingsBatchSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Embeddings.BatchSize
}

func (c *Config) GetChunkingMaxChunkSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Chunking.MaxChunkSize
}

func (c *Config) GetChunkingOverlapSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Chunking.OverlapSize
}

func (c *Config) GetJaegerEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Jaeger.Endpoint
}

func (c *Config) GetCoreServiceAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.CoreService.Address
}
