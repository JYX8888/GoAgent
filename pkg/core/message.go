package core

import (
	"time"
)

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

type Message struct {
	Content   string                 `json:"content"`
	Role      MessageRole            `json:"role"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Name      string                 `json:"name,omitempty"`
}

func NewMessage(content string, role MessageRole) *Message {
	return &Message{
		Content:   content,
		Role:      role,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

func NewUserMessage(content string) *Message {
	return NewMessage(content, RoleUser)
}

func NewAssistantMessage(content string) *Message {
	return NewMessage(content, RoleAssistant)
}

func NewSystemMessage(content string) *Message {
	return NewMessage(content, RoleSystem)
}

func (m *Message) WithMetadata(key string, value interface{}) *Message {
	m.Metadata[key] = value
	return m
}

func (m *Message) WithName(name string) *Message {
	m.Name = name
	return m
}

func (m *Message) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"role":    m.Role,
		"content": m.Content,
	}
	if m.Name != "" {
		result["name"] = m.Name
	}
	return result
}

func (m *Message) String() string {
	return string(m.Role) + ": " + m.Content
}

func MessagesToMap(messages []*Message) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		result[i] = msg.ToMap()
	}
	return result
}
