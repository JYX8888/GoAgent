package memory

import (
	"sync"
	"time"
)

type RetrieveOption func(*retrieveOptions)

type retrieveOptions struct {
	userID        string
	minImportance float64
	timeRange     *TimeRange
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func WithUserID(userID string) RetrieveOption {
	return func(o *retrieveOptions) { o.userID = userID }
}

func WithMinImportance(min float64) RetrieveOption {
	return func(o *retrieveOptions) { o.minImportance = min }
}

func WithTimeRange(start, end time.Time) RetrieveOption {
	return func(o *retrieveOptions) { o.timeRange = &TimeRange{Start: start, End: end} }
}

type EpisodicMemory struct {
	mu    sync.RWMutex
	items map[string]*MemoryItem
}

func NewEpisodicMemory() *EpisodicMemory {
	return &EpisodicMemory{
		items: make(map[string]*MemoryItem),
	}
}

func (e *EpisodicMemory) Add(item *MemoryItem) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.items[item.ID] = item
	return item.ID
}

func (e *EpisodicMemory) Retrieve(query string, limit int, opts ...RetrieveOption) []*MemoryItem {
	e.mu.RLock()
	defer e.mu.RUnlock()

	items := make([]*MemoryItem, 0, len(e.items))
	for _, item := range e.items {
		items = append(items, item)
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (e *EpisodicMemory) Update(memoryID string, content string, importance float64, metadata map[string]interface{}) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	item, exists := e.items[memoryID]
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

func (e *EpisodicMemory) Remove(memoryID string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.items[memoryID]; exists {
		delete(e.items, memoryID)
		return true
	}
	return false
}

func (e *EpisodicMemory) HasMemory(memoryID string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, exists := e.items[memoryID]
	return exists
}

func (e *EpisodicMemory) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.items = make(map[string]*MemoryItem)
}

func (e *EpisodicMemory) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.items)
}

func (e *EpisodicMemory) GetStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"type":  "episodic",
		"count": len(e.items),
	}
}

type Episode struct {
	ID         string
	Content    string
	UserID     string
	Importance float64
	Metadata   map[string]interface{}
}

func NewEpisode(content, userID string) *Episode {
	return &Episode{
		Content:    content,
		UserID:     userID,
		Importance: 0.5,
		Metadata:   make(map[string]interface{}),
	}
}
