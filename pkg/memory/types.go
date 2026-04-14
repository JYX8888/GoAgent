package memory

import (
	"fmt"
	"sync"
	"time"
)

type MemoryType string

const (
	MemoryTypeWorking    MemoryType = "working"
	MemoryTypeEpisodic   MemoryType = "episodic"
	MemoryTypeSemantic   MemoryType = "semantic"
	MemoryTypePerceptual MemoryType = "perceptual"
)

type MemoryItem struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	MemoryType MemoryType             `json:"memory_type"`
	UserID     string                 `json:"user_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Importance float64                `json:"importance"`
	Metadata   map[string]interface{} `json:"metadata"`
}

func NewMemoryItem(content string, memoryType MemoryType, userID string) *MemoryItem {
	return &MemoryItem{
		ID:         generateID(),
		Content:    content,
		MemoryType: memoryType,
		UserID:     userID,
		Timestamp:  time.Now(),
		Importance: 0.5,
		Metadata:   make(map[string]interface{}),
	}
}

func (m *MemoryItem) SetImportance(importance float64) {
	if importance < 0 {
		importance = 0
	}
	if importance > 1 {
		importance = 1
	}
	m.Importance = importance
}

func (m *MemoryItem) WithMetadata(key string, value interface{}) *MemoryItem {
	m.Metadata[key] = value
	return m
}

func (m *MemoryItem) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          m.ID,
		"content":     m.Content,
		"memory_type": m.MemoryType,
		"user_id":     m.UserID,
		"timestamp":   m.Timestamp,
		"importance":  m.Importance,
		"metadata":    m.Metadata,
	}
}

type BaseMemory struct {
	mu         sync.RWMutex
	items      map[string]*MemoryItem
	memoryType MemoryType
}

func NewBaseMemory(memoryType MemoryType) *BaseMemory {
	return &BaseMemory{
		items:      make(map[string]*MemoryItem),
		memoryType: memoryType,
	}
}

func (b *BaseMemory) Add(item *MemoryItem) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items[item.ID] = item
	return item.ID
}

func (b *BaseMemory) Get(memoryID string) *MemoryItem {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.items[memoryID]
}

func (b *BaseMemory) List() []*MemoryItem {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*MemoryItem, 0, len(b.items))
	for _, item := range b.items {
		result = append(result, item)
	}
	return result
}

func (b *BaseMemory) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.items)
}

func (b *BaseMemory) Remove(memoryID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.items[memoryID]; exists {
		delete(b.items, memoryID)
		return true
	}
	return false
}

func (b *BaseMemory) HasMemory(memoryID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.items[memoryID]
	return exists
}

func (b *BaseMemory) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = make(map[string]*MemoryItem)
}

func (b *BaseMemory) GetStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return map[string]interface{}{
		"type":  string(b.memoryType),
		"count": len(b.items),
	}
}

func (b *BaseMemory) String() string {
	return fmt.Sprintf("%s(count=%d)", b.memoryType, b.Count())
}
