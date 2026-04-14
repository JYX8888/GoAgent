package memory

import (
	"fmt"
	"sync"
)

type PerceptualMemory struct {
	mu         sync.RWMutex
	items      map[string]*MemoryItem
	modalities []string
}

func NewPerceptualMemory(modalities []string) *PerceptualMemory {
	if modalities == nil {
		modalities = []string{"text", "image", "audio", "video"}
	}
	return &PerceptualMemory{
		items:      make(map[string]*MemoryItem),
		modalities: modalities,
	}
}

func (p *PerceptualMemory) Add(item *MemoryItem) string {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.items[item.ID] = item
	return item.ID
}

func (p *PerceptualMemory) Retrieve(query string, limit int, opts ...RetrieveOption) []*MemoryItem {
	p.mu.RLock()
	defer p.mu.RUnlock()

	items := make([]*MemoryItem, 0, len(p.items))
	for _, item := range p.items {
		items = append(items, item)
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (p *PerceptualMemory) Update(memoryID string, content string, importance float64, metadata map[string]interface{}) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	item, exists := p.items[memoryID]
	if !exists {
		return false
	}
	if content != "" {
		item.Content = content
	}
	if importance > 0 {
		item.Importance = importance
	}
	if metadata != nil {
		for k, v := range metadata {
			item.Metadata[k] = v
		}
	}
	return true
}

func (p *PerceptualMemory) Remove(memoryID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.items[memoryID]; exists {
		delete(p.items, memoryID)
		return true
	}
	return false
}

func (p *PerceptualMemory) HasMemory(memoryID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.items[memoryID]
	return exists
}

func (p *PerceptualMemory) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.items = make(map[string]*MemoryItem)
}

func (p *PerceptualMemory) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.items)
}

func (p *PerceptualMemory) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"type":       "perceptual",
		"count":      len(p.items),
		"modalities": p.modalities,
	}
}

type Perception struct {
	ID       string
	Modality string
	Content  interface{}
	Metadata map[string]interface{}
}

func NewPerception(modality string, content interface{}) *Perception {
	return &Perception{
		Modality: modality,
		Content:  content,
		Metadata: make(map[string]interface{}),
	}
}

func (p *PerceptualMemory) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("PerceptualMemory(count=%d, modalities=%v)", len(p.items), p.modalities)
}
