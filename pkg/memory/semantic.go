package memory

import (
	"fmt"
	"sync"
)

type SemanticMemory struct {
	mu        sync.RWMutex
	items     map[string]*MemoryItem
	entities  map[string]*Entity
	relations map[string]*Relation
}

func NewSemanticMemory() *SemanticMemory {
	return &SemanticMemory{
		items:     make(map[string]*MemoryItem),
		entities:  make(map[string]*Entity),
		relations: make(map[string]*Relation),
	}
}

func (s *SemanticMemory) Add(item *MemoryItem) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[item.ID] = item
	return item.ID
}

func (s *SemanticMemory) Retrieve(query string, limit int, opts ...RetrieveOption) []*MemoryItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*MemoryItem, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func (s *SemanticMemory) Update(memoryID string, content string, importance float64, metadata map[string]interface{}) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.items[memoryID]
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

func (s *SemanticMemory) Remove(memoryID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.items[memoryID]; exists {
		delete(s.items, memoryID)
		return true
	}
	return false
}

func (s *SemanticMemory) HasMemory(memoryID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.items[memoryID]
	return exists
}

func (s *SemanticMemory) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]*MemoryItem)
	s.entities = make(map[string]*Entity)
	s.relations = make(map[string]*Relation)
}

func (s *SemanticMemory) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *SemanticMemory) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"type":           "semantic",
		"count":          len(s.items),
		"entity_count":   len(s.entities),
		"relation_count": len(s.relations),
	}
}

func (s *SemanticMemory) AddEntity(entity *Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entities[entity.Name] = entity
}

func (s *SemanticMemory) AddRelation(relation *Relation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.relations[relation.ID] = relation
}

type Entity struct {
	Name       string
	Type       string
	Content    string
	Importance float64
	Metadata   map[string]interface{}
}

func NewEntity(name, entityType, content string) *Entity {
	return &Entity{
		Name:       name,
		Type:       entityType,
		Content:    content,
		Importance: 0.5,
		Metadata:   make(map[string]interface{}),
	}
}

type Relation struct {
	ID       string
	Source   string
	Target   string
	Type     string
	Weight   float64
	Metadata map[string]interface{}
}

func NewRelation(source, relationType, target string) *Relation {
	return &Relation{
		Source:   source,
		Target:   target,
		Type:     relationType,
		Weight:   1.0,
		Metadata: make(map[string]interface{}),
	}
}

func (s *SemanticMemory) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("SemanticMemory(count=%d, entities=%d, relations=%d)", len(s.items), len(s.entities), len(s.relations))
}
