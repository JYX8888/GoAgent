package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	LLM    LLMConfig
	Embed  EmbedConfig
	DB     DBConfig
	Memory MemoryConfig
	Log    LogConfig
}

type LLMConfig struct {
	Model    string
	BaseURL  string
	APIKey   string
	Provider string
	Timeout  int
}

type EmbedConfig struct {
	ModelType string
	ModelName string
	APIKey    string
	BaseURL   string
}

type DBConfig struct {
	QdrantURL     string
	QdrantAPIKey  string
	Neo4jURI      string
	Neo4jUsername string
	Neo4jPassword string
}

type MemoryConfig struct {
	StoragePath     string
	MaxCapacity     int
	WorkingCapacity int
	WorkingTTL      int
}

type LogConfig struct {
	Level string
	Debug bool
}

func Load() *Config {
	return &Config{
		LLM: LLMConfig{
			Model:    getEnv("LLM_MODEL_ID", "gpt-3.5-turbo"),
			BaseURL:  getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			APIKey:   getEnv("LLM_API_KEY", ""),
			Provider: getEnv("LLM_PROVIDER", "openai"),
			Timeout:  getEnvInt("LLM_TIMEOUT", 60),
		},
		Embed: EmbedConfig{
			ModelType: getEnv("EMBED_MODEL_TYPE", "dashscope"),
			ModelName: getEnv("EMBED_MODEL_NAME", "text-embedding-v3"),
			APIKey:    getEnv("EMBED_API_KEY", ""),
			BaseURL:   getEnv("EMBED_BASE_URL", ""),
		},
		DB: DBConfig{
			QdrantURL:     getEnv("QDRANT_URL", ""),
			QdrantAPIKey:  getEnv("QDRANT_API_KEY", ""),
			Neo4jURI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
			Neo4jUsername: getEnv("NEO4J_USERNAME", "neo4j"),
			Neo4jPassword: getEnv("NEO4J_PASSWORD", ""),
		},
		Memory: MemoryConfig{
			StoragePath:     getEnv("MEMORY_STORAGE_PATH", "./memory_data"),
			MaxCapacity:     getEnvInt("MEMORY_MAX_CAPACITY", 100),
			WorkingCapacity: getEnvInt("MEMORY_WORKING_CAPACITY", 10),
			WorkingTTL:      getEnvInt("MEMORY_WORKING_TTL", 120),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "INFO"),
			Debug: getEnvBool("DEBUG", false),
		},
	}
}

func LoadFromFile(envFile string) (*Config, error) {
	data, err := os.ReadFile(envFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}

	return Load(), nil
}

func LoadFromEnvFile() (*Config, error) {
	envPath := FindEnvFile()
	if envPath == "" {
		return Load(), nil
	}
	return LoadFromFile(envPath)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	switch key {
	case "OPENAI_API_KEY":
		if value := os.Getenv("OPENAI_API_KEY"); value != "" {
			return value
		}
	case "DEEPSEEK_API_KEY":
		if value := os.Getenv("DEEPSEEK_API_KEY"); value != "" {
			return value
		}
	case "DASHSCOPE_API_KEY":
		if value := os.Getenv("DASHSCOPE_API_KEY"); value != "" {
			return value
		}
	case "KIMI_API_KEY":
		if value := os.Getenv("KIMI_API_KEY"); value != "" {
			return value
		}
	case "ZHIPU_API_KEY":
		if value := os.Getenv("ZHIPU_API_KEY"); value != "" {
			return value
		}
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func GetProjectRoot() string {
	execPath, _ := os.Executable()
	dir := filepath.Dir(execPath)
	return dir
}

func FindEnvFile() string {
	paths := []string{
		".env",
		".env.local",
		filepath.Join(GetProjectRoot(), ".env"),
		filepath.Join(GetProjectRoot(), "config", ".env"),
		filepath.Join(GetProjectRoot(), "config", ".env.local"),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (c *Config) GetAPIKey() string {
	switch c.LLM.Provider {
	case "openai":
		return getEnv("OPENAI_API_KEY", c.LLM.APIKey)
	case "deepseek":
		return getEnv("DEEPSEEK_API_KEY", c.LLM.APIKey)
	case "qwen", "dashscope":
		return getEnv("DASHSCOPE_API_KEY", c.LLM.APIKey)
	case "modelscope":
		return getEnv("MODELSCOPE_API_KEY", c.LLM.APIKey)
	case "kimi":
		return getEnv("KIMI_API_KEY", c.LLM.APIKey)
	case "zhipu":
		return getEnv("ZHIPU_API_KEY", c.LLM.APIKey)
	default:
		return c.LLM.APIKey
	}
}

func (c *Config) GetBaseURL() string {
	switch c.LLM.Provider {
	case "openai":
		return getEnv("LLM_BASE_URL", "https://api.openai.com/v1")
	case "deepseek":
		return getEnv("LLM_BASE_URL", "https://api.deepseek.com")
	case "qwen", "dashscope":
		return getEnv("LLM_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	case "modelscope":
		return getEnv("LLM_BASE_URL", "https://api-inference.modelscope.cn/v1/")
	case "kimi":
		return getEnv("LLM_BASE_URL", "https://api.moonshot.cn/v1")
	case "zhipu":
		return getEnv("LLM_BASE_URL", "https://open.bigmodel.cn/api/paas/v4")
	case "ollama":
		return getEnv("OLLAMA_HOST", "http://localhost:11434/v1")
	case "vllm":
		return getEnv("VLLM_HOST", "http://localhost:8000/v1")
	default:
		return c.LLM.BaseURL
	}
}

func (c *Config) String() string {
	return fmt.Sprintf(`Config:
  LLM:
    Model: %s
    Provider: %s
    BaseURL: %s
  Embed:
    ModelType: %s
    ModelName: %s
  Memory:
    StoragePath: %s
    MaxCapacity: %d
  Log:
    Level: %s
    Debug: %v`,
		c.LLM.Model, c.LLM.Provider, c.LLM.BaseURL,
		c.Embed.ModelType, c.Embed.ModelName,
		c.Memory.StoragePath, c.Memory.MaxCapacity,
		c.Log.Level, c.Log.Debug)
}
