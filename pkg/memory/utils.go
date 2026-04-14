package memory

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func LoadMemoryConfigFromEnv() *MemoryConfig {
	cfg := NewMemoryConfig()

	if v := os.Getenv("MEMORY_STORAGE_PATH"); v != "" {
		cfg.StoragePath = v
	}
	if v := os.Getenv("MEMORY_MAX_CAPACITY"); v != "" {
		if cap, err := strconv.Atoi(v); err == nil {
			cfg.MaxCapacity = cap
		}
	}
	if v := os.Getenv("MEMORY_IMPORTANCE_THRESHOLD"); v != "" {
		if th, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.ImportanceThreshold = th
		}
	}
	if v := os.Getenv("MEMORY_WORKING_CAPACITY"); v != "" {
		if cap, err := strconv.Atoi(v); err == nil {
			cfg.WorkingMemoryCapacity = cap
		}
	}
	if v := os.Getenv("MEMORY_WORKING_TTL"); v != "" {
		if ttl, err := strconv.Atoi(v); err == nil {
			cfg.WorkingMemoryTTLMinutes = ttl
		}
	}

	return cfg
}

func isEpisodicContent(content string) bool {
	keywords := []string{"昨天", "今天", "明天", "上次", "记得", "发生", "经历", "yesterday", "today", "remember", "happened"}
	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func isSemanticContent(content string) bool {
	keywords := []string{"定义", "概念", "规则", "知识", "原理", "方法", "definition", "concept", "knowledge", "principle"}
	contentLower := strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
