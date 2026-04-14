package memory

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type workingMemoryItem struct {
	priority  float64
	timestamp time.Time
	item      *MemoryItem
	index     int
}

type WorkingMemoryQueue []*workingMemoryItem

func (q WorkingMemoryQueue) Len() int { return len(q) }

func (q WorkingMemoryQueue) Less(i, j int) bool {
	return q[i].priority > q[j].priority
}

func (q WorkingMemoryQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *WorkingMemoryQueue) Push(x interface{}) {
	item := x.(*workingMemoryItem)
	item.index = len(*q)
	*q = append(*q, item)
}

func (q *WorkingMemoryQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	*q = old[0 : n-1]
	return item
}

type WorkingMemory struct {
	mu            sync.RWMutex
	items         map[string]*MemoryItem
	maxCapacity   int
	maxTokens     int
	maxAgeMinutes int
	currentTokens int
	sessionStart  time.Time
	memoryHeap    WorkingMemoryQueue
}

func NewWorkingMemory(config *MemoryConfig) *WorkingMemory {
	return &WorkingMemory{
		items:         make(map[string]*MemoryItem),
		maxCapacity:   config.WorkingMemoryCapacity,
		maxTokens:     config.WorkingMemoryTokens,
		maxAgeMinutes: config.WorkingMemoryTTLMinutes,
		currentTokens: 0,
		sessionStart:  time.Now(),
		memoryHeap:    make(WorkingMemoryQueue, 0),
	}
}

func (w *WorkingMemory) Add(item *MemoryItem) string {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.expireOldMemories()

	priority := w.calculatePriority(item)
	heapItem := &workingMemoryItem{
		priority:  priority,
		timestamp: item.Timestamp,
		item:      item,
	}
	heap.Push(&w.memoryHeap, heapItem)
	w.items[item.ID] = item

	w.currentTokens += len(splitWords(item.Content))
	w.enforceCapacityLimits()

	return item.ID
}

func (w *WorkingMemory) Retrieve(query string, limit int, opts ...RetrieveOption) []*MemoryItem {
	w.mu.RLock()
	defer w.mu.RUnlock()

	w.expireOldMemoriesRLocked()

	if len(w.items) == 0 {
		return nil
	}

	results := make([]*MemoryItem, 0, limit)
	for _, item := range w.items {
		if item.Metadata["forgotten"] == nil || !item.Metadata["forgotten"].(bool) {
			results = append(results, item)
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func (w *WorkingMemory) Update(memoryID string, content string, importance float64, metadata map[string]interface{}) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	item, exists := w.items[memoryID]
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

func (w *WorkingMemory) Remove(memoryID string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	item, exists := w.items[memoryID]
	if !exists {
		return false
	}

	w.currentTokens -= len(splitWords(item.Content))
	delete(w.items, memoryID)
	return true
}

func (w *WorkingMemory) HasMemory(memoryID string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	_, exists := w.items[memoryID]
	return exists
}

func (w *WorkingMemory) Clear() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.items = make(map[string]*MemoryItem)
	w.memoryHeap = make(WorkingMemoryQueue, 0)
	w.currentTokens = 0
	w.sessionStart = time.Now()
}

func (w *WorkingMemory) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.items)
}

func (w *WorkingMemory) GetStats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return map[string]interface{}{
		"type":           "working",
		"count":          len(w.items),
		"max_capacity":   w.maxCapacity,
		"current_tokens": w.currentTokens,
		"max_tokens":     w.maxTokens,
		"session_start":  w.sessionStart.Format(time.RFC3339),
	}
}

func (w *WorkingMemory) calculatePriority(item *MemoryItem) float64 {
	priority := item.Importance

	ageMinutes := time.Since(item.Timestamp).Minutes()
	if ageMinutes > 0 {
		decay := 1.0 - (ageMinutes / float64(w.maxAgeMinutes*10))
		if decay < 0.1 {
			decay = 0.1
		}
		priority *= decay
	}

	return priority
}

func (w *WorkingMemory) expireOldMemories() {
	cutoff := time.Now().Add(-time.Duration(w.maxAgeMinutes) * time.Minute)

	toRemove := make([]string, 0)
	for _, item := range w.items {
		if item.Timestamp.Before(cutoff) {
			toRemove = append(toRemove, item.ID)
		}
	}

	for _, id := range toRemove {
		item := w.items[id]
		w.currentTokens -= len(splitWords(item.Content))
		delete(w.items, id)
	}
}

func (w *WorkingMemory) expireOldMemoriesRLocked() {
	cutoff := time.Now().Add(-time.Duration(w.maxAgeMinutes) * time.Minute)

	toRemove := make([]string, 0)
	for _, item := range w.items {
		if item.Timestamp.Before(cutoff) {
			toRemove = append(toRemove, item.ID)
		}
	}

	for _, id := range toRemove {
		item := w.items[id]
		w.currentTokens -= len(splitWords(item.Content))
		delete(w.items, id)
	}
}

func (w *WorkingMemory) enforceCapacityLimits() {
	for len(w.items) > w.maxCapacity {
		if len(w.memoryHeap) > 0 {
			lowest := heap.Pop(&w.memoryHeap).(*workingMemoryItem)
			delete(w.items, lowest.item.ID)
			w.currentTokens -= len(splitWords(lowest.item.Content))
		}
	}

	for w.currentTokens > w.maxTokens {
		if len(w.memoryHeap) > 0 {
			lowest := heap.Pop(&w.memoryHeap).(*workingMemoryItem)
			delete(w.items, lowest.item.ID)
			w.currentTokens -= len(splitWords(lowest.item.Content))
		}
	}
}

func (w *WorkingMemory) String() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return fmt.Sprintf("WorkingMemory(count=%d, tokens=%d/%d)", len(w.items), w.currentTokens, w.maxTokens)
}

func splitWords(s string) []string {
	var words []string
	word := ""
	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(ch)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
