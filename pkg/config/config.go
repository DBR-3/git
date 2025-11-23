package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	Server     ServerConfig
	ClickHouse ClickHouseConfig
	Redis      RedisConfig
	Kafka      KafkaConfig
	Logger     LoggerConfig
	Tracing    TracingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int
	GRPCPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// ClickHouseConfig holds ClickHouse configuration
type ClickHouseConfig struct {
	Addresses        []string
	Database         string
	Username         string
	Password         string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	BatchSize        int
	FlushInterval    time.Duration
	CompressionLevel int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address      string
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers           []string
	GroupID           string
	TopicDataCollect  string
	TopicIndicators   string
	TopicSignals      string
	TopicErrors       string
	BatchSize         int
	BatchTimeout      time.Duration
	CompressionCodec  string
	MaxAttempts       int
	RequiredAcks      int
	EnableAutoCommit  bool
	CommitInterval    time.Duration
	SessionTimeout    time.Duration
	RebalanceTimeout  time.Duration
	StartOffset       string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string
	Format     string // json or text
	Output     string // stdout or file path
	MaxSize    int    // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled     bool
	ServiceName string
	JaegerURL   string
	SamplerType string
	SamplerRate float64
}

// LoadConfig loads configuration from file and environment
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from config file
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read from environment variables
	v.SetEnvPrefix("STOCK")
	v.AutomaticEnv()

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.grpcport", 50051)
	v.SetDefault("server.readtimeout", 30*time.Second)
	v.SetDefault("server.writetimeout", 30*time.Second)

	// ClickHouse defaults
	v.SetDefault("clickhouse.addresses", []string{"localhost:9000"})
	v.SetDefault("clickhouse.database", "stock_market")
	v.SetDefault("clickhouse.username", "stock_app")
	v.SetDefault("clickhouse.password", "stock_password")
	v.SetDefault("clickhouse.maxopenconns", 50)
	v.SetDefault("clickhouse.maxidleconns", 10)
	v.SetDefault("clickhouse.connmaxlifetime", 1*time.Hour)
	v.SetDefault("clickhouse.batchsize", 10000)
	v.SetDefault("clickhouse.flushinterval", 10*time.Second)
	v.SetDefault("clickhouse.compressionlevel", 1)

	// Redis defaults
	v.SetDefault("redis.address", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.poolsize", 100)
	v.SetDefault("redis.minidleconns", 10)
	v.SetDefault("redis.maxretries", 3)
	v.SetDefault("redis.dialtimeout", 5*time.Second)
	v.SetDefault("redis.readtimeout", 3*time.Second)
	v.SetDefault("redis.writetimeout", 3*time.Second)

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.groupid", "stock-market-consumer")
	v.SetDefault("kafka.topicdatacollect", "stock.data.collected")
	v.SetDefault("kafka.topicindicators", "stock.indicators.calculated")
	v.SetDefault("kafka.topicsignals", "stock.signals.generated")
	v.SetDefault("kafka.topicerrors", "stock.errors")
	v.SetDefault("kafka.batchsize", 100)
	v.SetDefault("kafka.batchtimeout", 1*time.Second)
	v.SetDefault("kafka.compressioncodec", "snappy")
	v.SetDefault("kafka.maxattempts", 3)
	v.SetDefault("kafka.requiredacks", 1)
	v.SetDefault("kafka.enableautocommit", true)
	v.SetDefault("kafka.commitinterval", 5*time.Second)
	v.SetDefault("kafka.sessiontimeout", 10*time.Second)
	v.SetDefault("kafka.rebalancetimeout", 60*time.Second)
	v.SetDefault("kafka.startoffset", "latest")

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.output", "stdout")
	v.SetDefault("logger.maxsize", 100)
	v.SetDefault("logger.maxbackups", 3)
	v.SetDefault("logger.maxage", 7)
	v.SetDefault("logger.compress", true)

	// Tracing defaults
	v.SetDefault("tracing.enabled", false)
	v.SetDefault("tracing.servicename", "stock-market-service")
	v.SetDefault("tracing.jaegerurl", "http://localhost:14268/api/traces")
	v.SetDefault("tracing.samplertype", "probabilistic")
	v.SetDefault("tracing.samplerrate", 0.1)
}
