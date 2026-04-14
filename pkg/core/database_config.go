package core

import (
	"os"
	"strconv"
	"sync"
)

type QdrantConfig struct {
	URL            string `json:"url" env:"QDRANT_URL"`
	APIKey         string `json:"api_key" env:"QDRANT_API_KEY"`
	CollectionName string `json:"collection_name" env:"QDRANT_COLLECTION"`
	VectorSize     int    `json:"vector_size" env:"QDRANT_VECTOR_SIZE"`
	Distance       string `json:"distance" env:"QDRANT_DISTANCE"`
	Timeout        int    `json:"timeout" env:"QDRANT_TIMEOUT"`
	mu             sync.RWMutex
}

func NewQdrantConfig() *QdrantConfig {
	return &QdrantConfig{
		CollectionName: "hello_agents_vectors",
		VectorSize:     384,
		Distance:       "cosine",
		Timeout:        30,
	}
}

func LoadQdrantConfigFromEnv() *QdrantConfig {
	cfg := NewQdrantConfig()

	if v := os.Getenv("QDRANT_URL"); v != "" {
		cfg.URL = v
	}
	if v := os.Getenv("QDRANT_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("QDRANT_COLLECTION"); v != "" {
		cfg.CollectionName = v
	}
	if v := os.Getenv("QDRANT_VECTOR_SIZE"); v != "" {
		if size, err := strconv.Atoi(v); err == nil {
			cfg.VectorSize = size
		}
	}
	if v := os.Getenv("QDRANT_DISTANCE"); v != "" {
		cfg.Distance = v
	}
	if v := os.Getenv("QDRANT_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil {
			cfg.Timeout = timeout
		}
	}

	return cfg
}

func (c *QdrantConfig) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"url":             c.URL,
		"api_key":         c.APIKey,
		"collection_name": c.CollectionName,
		"vector_size":     c.VectorSize,
		"distance":        c.Distance,
		"timeout":         c.Timeout,
	}
}

type Neo4jConfig struct {
	URI                          string `json:"uri" env:"NEO4J_URI"`
	Username                     string `json:"username" env:"NEO4J_USERNAME"`
	Password                     string `json:"password" env:"NEO4J_PASSWORD"`
	Database                     string `json:"database" env:"NEO4J_DATABASE"`
	MaxConnectionLifetime        int    `json:"max_connection_lifetime" env:"NEO4J_MAX_CONNECTION_LIFETIME"`
	MaxConnectionPoolSize        int    `json:"max_connection_pool_size" env:"NEO4J_MAX_CONNECTION_POOL_SIZE"`
	ConnectionAcquisitionTimeout int    `json:"connection_acquisition_timeout" env:"NEO4J_CONNECTION_TIMEOUT"`
	mu                           sync.RWMutex
}

func NewNeo4jConfig() *Neo4jConfig {
	return &Neo4jConfig{
		URI:                          "bolt://localhost:7687",
		Username:                     "neo4j",
		Password:                     "hello-agents-password",
		Database:                     "neo4j",
		MaxConnectionLifetime:        3600,
		MaxConnectionPoolSize:        50,
		ConnectionAcquisitionTimeout: 60,
	}
}

func LoadNeo4jConfigFromEnv() *Neo4jConfig {
	cfg := NewNeo4jConfig()

	if v := os.Getenv("NEO4J_URI"); v != "" {
		cfg.URI = v
	}
	if v := os.Getenv("NEO4J_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("NEO4J_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("NEO4J_DATABASE"); v != "" {
		cfg.Database = v
	}
	if v := os.Getenv("NEO4J_MAX_CONNECTION_LIFETIME"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			cfg.MaxConnectionLifetime = val
		}
	}
	if v := os.Getenv("NEO4J_MAX_CONNECTION_POOL_SIZE"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			cfg.MaxConnectionPoolSize = val
		}
	}
	if v := os.Getenv("NEO4J_CONNECTION_TIMEOUT"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			cfg.ConnectionAcquisitionTimeout = val
		}
	}

	return cfg
}

func (c *Neo4jConfig) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"uri":                            c.URI,
		"username":                       c.Username,
		"password":                       c.Password,
		"database":                       c.Database,
		"max_connection_lifetime":        c.MaxConnectionLifetime,
		"max_connection_pool_size":       c.MaxConnectionPoolSize,
		"connection_acquisition_timeout": c.ConnectionAcquisitionTimeout,
	}
}

type DatabaseConfig struct {
	Qdrant *QdrantConfig `json:"qdrant"`
	Neo4j  *Neo4jConfig  `json:"neo4j"`
	mu     sync.RWMutex
}

var (
	dbConfig     *DatabaseConfig
	dbConfigOnce sync.Once
	dbConfigLock sync.Mutex
)

func GetDatabaseConfig() *DatabaseConfig {
	dbConfigOnce.Do(func() {
		dbConfig = &DatabaseConfig{
			Qdrant: LoadQdrantConfigFromEnv(),
			Neo4j:  LoadNeo4jConfigFromEnv(),
		}
	})
	return dbConfig
}

func UpdateDatabaseConfig(qdrant *QdrantConfig, neo4j *Neo4jConfig) {
	dbConfigLock.Lock()
	defer dbConfigLock.Unlock()

	if dbConfig == nil {
		dbConfig = GetDatabaseConfig()
	}

	if qdrant != nil {
		dbConfig.Qdrant = qdrant
	}
	if neo4j != nil {
		dbConfig.Neo4j = neo4j
	}
}

func (c *DatabaseConfig) GetQdrantConfig() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Qdrant == nil {
		return nil
	}
	return c.Qdrant.ToMap()
}

func (c *DatabaseConfig) GetNeo4jConfig() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.Neo4j == nil {
		return nil
	}
	return c.Neo4j.ToMap()
}

func (c *DatabaseConfig) ValidateConnections() (map[string]bool, error) {
	results := map[string]bool{
		"qdrant": false,
		"neo4j":  false,
	}

	c.mu.RLock()
	qdrantCfg := c.Qdrant
	neo4jCfg := c.Neo4j
	c.mu.RUnlock()

	if qdrantCfg != nil && qdrantCfg.URL != "" {
		results["qdrant"] = true
	}

	if neo4jCfg != nil && neo4jCfg.URI != "" {
		results["neo4j"] = true
	}

	return results, nil
}

type DatabaseConfigOption func(*DatabaseConfig)

func WithQdrantConfig(cfg *QdrantConfig) DatabaseConfigOption {
	return func(c *DatabaseConfig) { c.Qdrant = cfg }
}

func WithNeo4jConfig(cfg *Neo4jConfig) DatabaseConfigOption {
	return func(c *DatabaseConfig) { c.Neo4j = cfg }
}

func NewDatabaseConfig(opts ...DatabaseConfigOption) *DatabaseConfig {
	cfg := &DatabaseConfig{
		Qdrant: NewQdrantConfig(),
		Neo4j:  NewNeo4jConfig(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func LoadDatabaseConfigFromEnv() *DatabaseConfig {
	return &DatabaseConfig{
		Qdrant: LoadQdrantConfigFromEnv(),
		Neo4j:  LoadNeo4jConfigFromEnv(),
	}
}
