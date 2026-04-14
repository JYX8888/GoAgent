package memory

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type MemoryManager struct {
	Config     *MemoryConfig
	UserID     string
	working    *WorkingMemory
	episodic   *EpisodicMemory
	semantic   *SemanticMemory
	perceptual *PerceptualMemory
	mu         sync.RWMutex
}

type MemoryManagerOption func(*MemoryManager)

func NewMemoryManager(opts ...MemoryManagerOption) *MemoryManager {
	m := &MemoryManager{
		Config:     NewMemoryConfig(),
		UserID:     "default_user",
		working:    NewWorkingMemory(NewMemoryConfig()),
		episodic:   NewEpisodicMemory(),
		semantic:   NewSemanticMemory(),
		perceptual: NewPerceptualMemory(nil),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func WithUserIDOpt(userID string) MemoryManagerOption {
	return func(m *MemoryManager) { m.UserID = userID }
}

func (m *MemoryManager) AddMemory(content string, memoryType MemoryType, importance float64, metadata map[string]interface{}) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if memoryType == "" {
		memoryType = m.classifyMemoryType(content, metadata)
	}

	if importance == 0 {
		importance = m.calculateImportance(content, metadata)
	}

	item := &MemoryItem{
		ID:         generateID(),
		Content:    content,
		MemoryType: memoryType,
		UserID:     m.UserID,
		Timestamp:  time.Now(),
		Importance: importance,
		Metadata:   metadata,
	}

	switch memoryType {
	case MemoryTypeWorking:
		return m.working.Add(item)
	case MemoryTypeEpisodic:
		return m.episodic.Add(item)
	case MemoryTypeSemantic:
		return m.semantic.Add(item)
	case MemoryTypePerceptual:
		return m.perceptual.Add(item)
	}

	return ""
}

func (m *MemoryManager) RetrieveMemories(query string, memoryTypes []MemoryType, limit int, minImportance float64) []*MemoryItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if memoryTypes == nil {
		memoryTypes = []MemoryType{MemoryTypeWorking, MemoryTypeEpisodic, MemoryTypeSemantic}
	}

	if limit == 0 {
		limit = 10
	}

	perTypeLimit := limit / len(memoryTypes)
	if perTypeLimit < 1 {
		perTypeLimit = 1
	}

	var allResults []*MemoryItem

	for _, memType := range memoryTypes {
		var results []*MemoryItem
		switch memType {
		case MemoryTypeWorking:
			results = m.working.Retrieve(query, perTypeLimit)
		case MemoryTypeEpisodic:
			results = m.episodic.Retrieve(query, perTypeLimit)
		case MemoryTypeSemantic:
			results = m.semantic.Retrieve(query, perTypeLimit)
		case MemoryTypePerceptual:
			results = m.perceptual.Retrieve(query, perTypeLimit)
		}
		allResults = append(allResults, results...)
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Importance > allResults[j].Importance
	})

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults
}

func (m *MemoryManager) UpdateMemory(memoryID string, content string, importance float64, metadata map[string]interface{}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.working.HasMemory(memoryID) {
		return m.working.Update(memoryID, content, importance, metadata)
	}
	if m.episodic.HasMemory(memoryID) {
		return m.episodic.Update(memoryID, content, importance, metadata)
	}
	if m.semantic.HasMemory(memoryID) {
		return m.semantic.Update(memoryID, content, importance, metadata)
	}
	if m.perceptual.HasMemory(memoryID) {
		return m.perceptual.Update(memoryID, content, importance, metadata)
	}
	return false
}

func (m *MemoryManager) RemoveMemory(memoryID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.working.HasMemory(memoryID) {
		return m.working.Remove(memoryID)
	}
	if m.episodic.HasMemory(memoryID) {
		return m.episodic.Remove(memoryID)
	}
	if m.semantic.HasMemory(memoryID) {
		return m.semantic.Remove(memoryID)
	}
	if m.perceptual.HasMemory(memoryID) {
		return m.perceptual.Remove(memoryID)
	}
	return false
}

func (m *MemoryManager) ForgetMemories(strategy string, threshold float64, maxAgeDays int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := m.working.Count() + m.episodic.Count() + m.semantic.Count() + m.perceptual.Count()

	m.working.Clear()
	m.episodic.Clear()
	m.semantic.Clear()
	m.perceptual.Clear()

	return total
}

func (m *MemoryManager) ConsolidateMemories(fromType, toType MemoryType, importanceThreshold float64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return 0
}

func (m *MemoryManager) GetMemoryStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"user_id":        m.UserID,
		"enabled_types":  []string{"working", "episodic", "semantic", "perceptual"},
		"total_memories": m.working.Count() + m.episodic.Count() + m.semantic.Count() + m.perceptual.Count(),
		"memories_by_type": map[string]interface{}{
			"working":    m.working.GetStats(),
			"episodic":   m.episodic.GetStats(),
			"semantic":   m.semantic.GetStats(),
			"perceptual": m.perceptual.GetStats(),
		},
	}
}

func (m *MemoryManager) ClearAllMemories() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.working.Clear()
	m.episodic.Clear()
	m.semantic.Clear()
	m.perceptual.Clear()
}

func (m *MemoryManager) classifyMemoryType(content string, metadata map[string]interface{}) MemoryType {
	if metadata != nil {
		if t, ok := metadata["type"].(string); ok {
			return MemoryType(t)
		}
	}

	if isEpisodicContent(content) {
		return MemoryTypeEpisodic
	}
	if isSemanticContent(content) {
		return MemoryTypeSemantic
	}

	return MemoryTypeWorking
}

func (m *MemoryManager) calculateImportance(content string, metadata map[string]interface{}) float64 {
	importance := 0.5

	if len(content) > 100 {
		importance += 0.1
	}

	importantKeywords := []string{"重要", "关键", "必须", "注意", "警告", "错误", "important", "critical"}
	for _, keyword := range importantKeywords {
		if contains(content, keyword) {
			importance += 0.2
			break
		}
	}

	if metadata != nil {
		if priority, ok := metadata["priority"].(string); ok {
			if priority == "high" {
				importance += 0.3
			} else if priority == "low" {
				importance -= 0.2
			}
		}
	}

	if importance > 1 {
		importance = 1
	}
	if importance < 0 {
		importance = 0
	}

	return importance
}

func (m *MemoryManager) String() string {
	stats := m.GetMemoryStats()
	return fmt.Sprintf("MemoryManager(user=%s, total=%d)", m.UserID, stats["total_memories"])
}
