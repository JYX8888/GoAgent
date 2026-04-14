package core

import (
	"fmt"
	"sync"
)

type Agent interface {
	Run(input string, opts ...AgentOption) (string, error)
	Name() string
	GetHistory() []*Message
	ClearHistory()
	AddMessage(msg *Message)
}

type BaseAgent struct {
	Name_        string
	LLM          *LLM
	SystemPrompt string
	Config       *Config
	history      []*Message
	mu           sync.RWMutex
}

func NewBaseAgent(name string, llm *LLM, systemPrompt string, cfg *Config) *BaseAgent {
	if cfg == nil {
		cfg = NewConfig()
	}
	return &BaseAgent{
		Name_:        name,
		LLM:          llm,
		SystemPrompt: systemPrompt,
		Config:       cfg,
		history:      make([]*Message, 0),
	}
}

func (a *BaseAgent) Name() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.Name_
}

func (a *BaseAgent) GetHistory() []*Message {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]*Message, len(a.history))
	copy(result, a.history)
	return result
}

func (a *BaseAgent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.history = a.history[:0]
}

func (a *BaseAgent) AddMessage(msg *Message) {
	a.mu.Lock()
	defer a.mu.Unlock()

	maxLen := a.Config.MaxHistoryLength
	if len(a.history) >= maxLen {
		a.history = a.history[1:]
	}
	a.history = append(a.history, msg)
}

func (a *BaseAgent) buildMessages(input string) []ChatMessage {
	a.mu.RLock()
	defer a.mu.RUnlock()

	messages := make([]ChatMessage, 0, len(a.history)+2)

	if a.SystemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: a.SystemPrompt,
		})
	}

	for _, msg := range a.history {
		messages = append(messages, ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: input,
	})

	return messages
}

func (a *BaseAgent) String() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return fmt.Sprintf("Agent(name=%s, provider=%s)", a.Name_, a.LLM.Provider)
}

type AgentOption func(*BaseAgent)

func WithSystemPrompt(prompt string) AgentOption {
	return func(a *BaseAgent) { a.SystemPrompt = prompt }
}

func WithConfig(cfg *Config) AgentOption {
	return func(a *BaseAgent) { a.Config = cfg }
}
