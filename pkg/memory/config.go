package memory

import (
		"sync"
)

type MemoryConfig struct {
	StoragePath         string  `json:"storage_path" env:"MEMORY_STORAGE_PATH"`
	MaxCapacity         int     `json:"max_capacity"`
	ImportanceThreshold float64 `json:"importance_threshold"`
	DecayFactor         float64 `json:"decay_factor"`

	WorkingMemoryCapacity   int `json:"working_memory_capacity"`
	WorkingMemoryTokens     int `json:"working_memory_tokens"`
	WorkingMemoryTTLMinutes int `json:"working_memory_ttl_minutes"`

	PerceptualMemoryModalities []string `json:"perceptual_memory_modalities"`

	mu sync.RWMutex
}

func NewMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		StoragePath:                "./memory_data",
		MaxCapacity:                100,
		ImportanceThreshold:        0.1,
		DecayFactor:                0.95,
		WorkingMemoryCapacity:      10,
		WorkingMemoryTokens:        2000,
		WorkingMemoryTTLMinutes:    120,
		PerceptualMemoryModalities: []string{"text", "image", "audio", "video"},
	}
}

func (c *MemoryConfig) SetStoragePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StoragePath = path
}

func (c *MemoryConfig) GetStoragePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.StoragePath
}

func (c *MemoryConfig) SetMaxCapacity(capacity int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MaxCapacity = capacity
}

func (c *MemoryConfig) GetMaxCapacity() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.MaxCapacity
}

func (c *MemoryConfig) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"storage_path":                 c.StoragePath,
		"max_capacity":                 c.MaxCapacity,
		"importance_threshold":         c.ImportanceThreshold,
		"decay_factor":                 c.DecayFactor,
		"working_memory_capacity":      c.WorkingMemoryCapacity,
		"working_memory_tokens":        c.WorkingMemoryTokens,
		"working_memory_ttl_minutes":   c.WorkingMemoryTTLMinutes,
		"perceptual_memory_modalities": c.PerceptualMemoryModalities,
	}
}
